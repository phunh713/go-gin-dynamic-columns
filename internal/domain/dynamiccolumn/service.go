package dynamiccolumn

import (
	"context"
	"errors"
	"fmt"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/types"
	"gin-demo/internal/shared/utils"
	"regexp"
	"strings"
)

type DynamicColumnService interface {
	RefreshDynamicColumnsOfRecordIds(ctx context.Context, table constants.TableName, ids []int64, action constants.Action, changes map[constants.TableName]Dependency, originalRecordId *int64, actionPayload interface{}) error
	CheckShouldRefreshDynamicColumn(ctx context.Context, table constants.TableName, action constants.Action, changes map[constants.TableName]Dependency, payload interface{}) (bool, map[constants.TableName]Dependency)
	BuildFormula(table constants.TableName, col string, userFormula string, userVars string) (string, error)
	ResolveTablesRelationLink(comparor constants.TableName, target constants.TableName, prev []RelationLink, visited map[constants.TableName]bool) ([]RelationLink, error)
}

type dynamicColumnService struct {
	dynamicColumnRepo DynamicColumnRepository
	modelsMap         types.ModelsMap
	modelRelationsMap types.ModelRelationsMap
}

func NewDynamicColumnService(dynamicColumnRepo DynamicColumnRepository, modelsMap types.ModelsMap, modelRelationsMap types.ModelRelationsMap) DynamicColumnService {
	return &dynamicColumnService{dynamicColumnRepo: dynamicColumnRepo, modelsMap: modelsMap, modelRelationsMap: modelRelationsMap}
}

func (r *dynamicColumnService) RefreshDynamicColumnsOfRecordIds(
	ctx context.Context, table constants.TableName, ids []int64, action constants.Action, changes map[constants.TableName]Dependency, originalRecordId *int64, actionPayload interface{}) error {
	// Check if action requires refreshing dynamic columns.
	// Get changes slice to refresh dependant tables later.
	shouldCheck, changes := r.CheckShouldRefreshDynamicColumn(ctx, table, action, changes, actionPayload)
	if !shouldCheck {
		return nil
	}

	// Get all dynamic columns affected by the changes
	dynamicCols := r.getAllDynamicColumnsFromChanges(ctx, table, changes)
	if len(dynamicCols) == 0 {
		return nil
	}
	// Determine the order of refreshing dynamic columns based on their dependencies
	orderedDynamicCols := r.determineRefreshOrder(ctx, table, ids, dynamicCols, originalRecordId)

	// Create a temp table to store ids that need refreshing
	err := r.dynamicColumnRepo.CreateTempIdsTable(ctx)
	if err != nil {
		fmt.Println("Error creating temp ids table:", err)
		return err
	}
	for _, col := range orderedDynamicCols {
		err := r.dynamicColumnRepo.CopyIdsToTempTable(ctx, col.Ids)
		if err != nil {
			fmt.Println("Error copying ids to temp ids table:", err)
			return err
		}
		err = r.dynamicColumnRepo.RefreshDynamicColumn(ctx, col)
		if err != nil {
			return err
		}
		err = r.dynamicColumnRepo.TruncateTempTable(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckShouldRefreshDynamicColumn checks if the action requires refreshing dynamic columns
func (r *dynamicColumnService) CheckShouldRefreshDynamicColumn(
	ctx context.Context, table constants.TableName, action constants.Action,
	changes map[constants.TableName]Dependency, payload interface{}) (bool, map[constants.TableName]Dependency) {
	if changes == nil {
		changes = make(map[constants.TableName]Dependency)
	}

	var columns []string

	switch action {
	case constants.ActionRefresh, constants.ActionCreate, constants.ActionDelete:
		// All fields are affected on create/delete
		model := utils.NewInstance(r.modelsMap[table])
		columns = utils.GetStructFieldJsonTags(model)

	case constants.ActionUpdate:
		// Only updated fields are affected
		columns = utils.GetNonZeroStructFieldJsonTags(payload)

	default:
		return false, nil
	}

	// Add columns to dependency map
	r.addColumnsToDependency(changes, table, columns)

	return true, changes
}

// addColumnsToDependency helper to add columns to the dependency map (unique only)
func (r *dynamicColumnService) addColumnsToDependency(changes map[constants.TableName]Dependency, table constants.TableName, columns []string) {
	dep := changes[table]
	dep.Columns = utils.AppendUnique(dep.Columns, columns...)
	changes[table] = dep
}

func (r *dynamicColumnService) getAllDynamicColumnsFromChanges(ctx context.Context, table constants.TableName, changes map[constants.TableName]Dependency) []DynamicColumn {
	var result []DynamicColumn
	currentChanges := changes
	for {
		dynamicColumns := r.dynamicColumnRepo.GetAllDependantsByChanges(ctx, table, currentChanges)
		if len(dynamicColumns) == 0 {
			break
		}
		result = append(result, dynamicColumns...)
		for _, colChange := range dynamicColumns {
			currentChanges = make(map[constants.TableName]Dependency)
			tableChange := currentChanges[colChange.TableName]
			tableChange.Columns = append(tableChange.Columns, colChange.Name)
			currentChanges[colChange.TableName] = tableChange
		}
	}
	return result
}

/*
* table is the original table where the changes happened
 */
func (r *dynamicColumnService) determineRefreshOrder(ctx context.Context, table constants.TableName, ids []int64, dynamicCols []DynamicColumn, originalRecordId *int64) []DynamicColumnWithMetadata {
	result := make([]DynamicColumnWithMetadata, 0)
	processed := make(map[string]bool) // O(1) lookup instead of O(n)
	refreshColNames := r.buildRefreshColumnNames(dynamicCols)

	// Use index-based loop so appending to dynamicCols extends the loop
	for i := 0; i < len(dynamicCols); i++ {
		col := dynamicCols[i]
		colName := (string(col.TableName) + "." + col.Name)

		// Skip if already processed
		if processed[colName] {
			continue
		}

		deps := r.extractDependencyColumnNames(col.Dependencies)

		// Case 1: Independent column - no dependencies on columns being refreshed
		// This column can be processed immediately because it doesn't wait for other dynamic columns.
		// Example: A column that only depends on static fields (company.name, invoice.created_at)
		if intersect := utils.StringSlicesIntersect(refreshColNames, deps); len(intersect) == 0 {
			ids := r.resolveIdsFromOriginalTable(ctx, table, ids, col.Dependencies[table].RecordIdsSelector, originalRecordId)
			result = append(result, DynamicColumnWithMetadata{
				DynamicColumn: col,
				Ids:           ids,
			})
			processed[colName] = true
			continue
		}

		// Case 2: Dependencies satisfied - at least one dependency has been processed
		// This column can now be calculated because the columns it depends on are ready.
		// Example: companies.status depends on invoices.status (which was just calculated)
		// We resolve IDs by querying based on the already-processed dependency IDs
		if intersect := r.getProcessedDependencies(deps, processed); len(intersect) > 0 {
			ids := r.resolveIdsFromMatchingDependencies(ctx, col, result, intersect)
			result = append(result, DynamicColumnWithMetadata{
				DynamicColumn: col,
				Ids:           ids,
			})
			processed[colName] = true
			continue
		}

		// Case 3: Dependencies pending - defer until dependencies are processed
		// This column depends on other dynamic columns that haven't been calculated yet.
		// Push it to the end of the queue and try again later in the next iteration
		dynamicCols = append(dynamicCols, col)
	}
	return result
}

// getProcessedDependencies returns which dependencies have been processed
func (r *dynamicColumnService) getProcessedDependencies(deps []string, processed map[string]bool) []string {
	result := make([]string, 0)
	for _, dep := range deps {
		if processed[dep] {
			result = append(result, dep)
		}
	}
	return result
}

// buildRefreshColumnNames builds list of "table.column" names from dynamic columns
func (r *dynamicColumnService) buildRefreshColumnNames(dynamicCols []DynamicColumn) []string {
	refreshColNames := make([]string, 0, len(dynamicCols))
	for _, col := range dynamicCols {
		refreshColNames = append(refreshColNames, string(col.TableName)+"."+col.Name)
	}
	return refreshColNames
}

// extractDependencyColumnNames extracts all dependency column names in "table.column" format
func (r *dynamicColumnService) extractDependencyColumnNames(dependencies map[constants.TableName]Dependency) []string {
	deps := make([]string, 0)
	for depTable, dep := range dependencies {
		for _, depCol := range dep.Columns {
			deps = append(deps, string(depTable)+"."+depCol)
		}
	}
	return deps
}

// resolveIdsFromOriginalTable resolves IDs based on the original table and record selector
// This is used for the first level of dynamic columns that directly depend on the original changed record
func (r *dynamicColumnService) resolveIdsFromOriginalTable(ctx context.Context, table constants.TableName, ids []int64, selector string, originalRecordId *int64) []int64 {
	// Build the list of IDs to process
	result := make([]int64, 0)
	if len(ids) > 0 {
		result = append(result, ids...)
	}
	if originalRecordId != nil {
		result = utils.AppendUnique(result, *originalRecordId)
	}

	// If no selector, return the IDs directly
	if selector == "" {
		return result
	}

	// Build context object with comma-separated IDs
	idsStr := make([]string, len(result))
	for i, v := range result {
		idsStr[i] = fmt.Sprintf("%d", v)
	}

	ctxObj := map[string]interface{}{
		string(table): map[string]string{
			"ids": strings.Join(idsStr, ","),
		},
	}

	foundIds := r.dynamicColumnRepo.GetAllSelectorIds(ctx, selector, ctxObj)
	if len(foundIds) > 0 {
		return foundIds
	}
	return []int64{}
}

// resolveIdsFromMatchingDependencies resolves IDs from matching dependencies already in result
func (r *dynamicColumnService) resolveIdsFromMatchingDependencies(ctx context.Context, col DynamicColumn, result []DynamicColumnWithMetadata, matchNames []string) []int64 {
	ids := make([]int64, 0)

	for _, matchName := range matchNames {
		for _, depCol := range result {
			if matchName != string(depCol.TableName)+"."+depCol.Name {
				continue
			}

			idsStr := make([]string, len(depCol.Ids))
			for i, v := range depCol.Ids {
				idsStr[i] = fmt.Sprintf("%d", v)
			}

			ctxObj := map[string]interface{}{
				string(depCol.TableName): map[string]string{
					"ids": strings.Join(idsStr, ","),
				},
			}

			query := col.Dependencies[constants.TableName(depCol.TableName)].RecordIdsSelector
			if query == "" {
				ids = utils.AppendUnique(ids, depCol.Ids...)
			} else {
				foundIds := r.dynamicColumnRepo.GetAllSelectorIds(ctx, query, ctxObj)
				if len(foundIds) > 0 {
					ids = utils.AppendUnique(ids, foundIds...)
				}
			}
		}
	}

	return ids
}

func (r *dynamicColumnService) BuildFormula(table constants.TableName, col string, userFormula string, userVars string) (string, error) {
	vars, err := ResolveVariables(userVars)
	if err != nil {
		fmt.Println("error resolve variables")
		return "", err
	}
	fmt.Println(vars)

	resolvedFormula, resolvedCte := r.resolveFormula(userFormula, table)
	resolvedCteValueStrs := make([]string, 0)
	resolvedCteJoinStrs := make([]string, 0)
	for _, cte := range resolvedCte {
		cteValue := cte.Value
		for _, v := range vars {
			colName := strings.Split(v.Name, ".")[1]
			cteValue = strings.ReplaceAll(cteValue, v.Name, v.Value+" AS "+colName)
		}
		resolvedCteValueStrs = append(resolvedCteValueStrs, cteValue)
		resolvedCteJoinStrs = append(resolvedCteJoinStrs, cte.Join)
	}
	res := constants.FORMULA_TEMPLATE
	res = strings.ReplaceAll(res, "{{t_name}}", string(table))
	res = strings.ReplaceAll(res, "{{c_name}}", col)
	res = strings.ReplaceAll(res, "{{formula}}", resolvedFormula)
	res = strings.ReplaceAll(res, "{{cte}}", strings.Join(resolvedCteValueStrs, ",\n")+",\n")
	res = strings.ReplaceAll(res, "{{cte_joins}}", strings.Join(resolvedCteJoinStrs, "\n"))
	return res, nil
}

func ResolveVariables(varStr string) ([]Variable, error) {
	res := make([]Variable, 0)
	if varStr == "" {
		return res, nil
	}

	splitVarStr := strings.Split(varStr, "\n")

	for _, i := range splitVarStr {
		if i == "" {
			continue
		}

		if !strings.HasPrefix(i, "var") {
			return nil, errors.New("Variable definition error: No 'var' keyword")
		}

		if !strings.Contains(i, "=") {
			return nil, errors.New("Variable definition error: No assign operator '='")
		}

		var v Variable

		re := regexp.MustCompile(`var\s+(.+?)\s*=\s*(.+)`)

		matches := re.FindStringSubmatch(i)

		if len(matches) != 3 {
			return nil, errors.New("Variable definition error: Regex no match")
		}

		vName := matches[1]
		vName = strings.ReplaceAll(vName, "{{", "")
		vName = strings.ReplaceAll(vName, "}}", "")
		v.Name = vName
		vValue := matches[2]
		vValue = strings.ReplaceAll(vValue, "{{", "")
		vValue = strings.ReplaceAll(vValue, "}}", "")
		v.Value = vValue

		res = append(res, v)
	}

	return res, nil
}

func (r *dynamicColumnService) resolveFormula(formulaStr string, table constants.TableName) (string, []FormulaCte) {
	re := regexp.MustCompile(`{{(\w+)}}\.(\w+)`)
	cte := make(map[constants.TableName][]string)

	replacedFormula := re.ReplaceAllStringFunc(formulaStr, func(m string) string {
		sub := re.FindStringSubmatch(m)
		t := constants.TableName(sub[1])
		col := sub[2]
		res := string(t) + "." + col
		if t != table {
			cte[t] = utils.AppendUnique(cte[t], col)
			res = "cte_" + res
		}
		return res
	})

	resolvedCte, _ := r.resolveCte(cte, table)
	return replacedFormula, resolvedCte
}

func (r *dynamicColumnService) resolveCte(cteNames map[constants.TableName][]string, rootTable constants.TableName) ([]FormulaCte, error) {
	result := make([]FormulaCte, 0)
	for joinTable, joinCols := range cteNames {
		cteLinks, err := r.ResolveTablesRelationLink(rootTable, joinTable, nil, nil)
		if err != nil {
			return nil, err
		}
		for i, cteLink := range cteLinks {
			var currentRootTable constants.TableName
			if i == 0 {
				currentRootTable = rootTable
			} else {
				currentRootTable = cteLinks[i-1].Table
			}
			currentJoinCols := joinCols
			if i < len(cteLinks)-1 {
				currentJoinCols = nil
			}
			cte, err := r.createCte(currentRootTable, cteLink, currentJoinCols, result)
			if err != nil {
				return nil, err
			}
			result = append(result, *cte)
		}
	}
	return result, nil
}

func (r *dynamicColumnService) createCte(
	rootTable constants.TableName,
	cteLink RelationLink,
	joinCols []string,
	resolvedCte []FormulaCte,
) (*FormulaCte, error) {
	for _, existingCte := range resolvedCte {
		if existingCte.Name == string(rootTable)+"_"+string(cteLink.Table) {
			return nil, errors.New("CTE already created")
		}
	}
	var cte FormulaCte
	cte.Name = string(rootTable) + "_" + string(cteLink.Table)
	cte.Join = fmt.Sprintf("LEFT JOIN %s cte_%s ON cte_%s.%s_id = %s.id", cte.Name, cteLink.Table, cteLink.Table, rootTable, rootTable)
	for i, col := range joinCols {
		joinCols[i] = string(cteLink.Table) + "." + col
	}
	selectCols := strings.Join(joinCols, ", ")
	if selectCols != "" {
		selectCols = ", " + selectCols
	}
	selectId := ""
	var fromTable constants.TableName = ""
	joinStm := ""
	groupStm := ""
	onId := "id"
	if cteLink.Relation == constants.TableRelationManyToOne {
		selectId = fmt.Sprintf("%s.id AS %s_id", rootTable, rootTable)
		fromTable = rootTable
		joinStm = fmt.Sprintf("JOIN %s ON %s.id = %s.%s_id", cteLink.Table, cteLink.Table, rootTable, cteLink.Table)
	}
	if cteLink.Relation == constants.TableRelationOneToMany {
		selectId = fmt.Sprintf("%s.%s_id", cteLink.Table, rootTable)
		fromTable = cteLink.Table
		onId = fmt.Sprintf("%s_id", rootTable)
		groupStm = fmt.Sprintf("GROUP BY %s.%s_id", cteLink.Table, rootTable)
	}
	cte.Value = fmt.Sprintf(`
			%s_%s AS (
				SELECT %s%s
				FROM %s
				JOIN %s tdi ON %s.%s = tdi.id
				%s
				%s
			)
		`, rootTable, cteLink.Table, selectId, selectCols, fromTable, constants.TEMP_TABLE_NAME, fromTable, onId, joinStm, groupStm)
	return &cte, nil
}

func (r *dynamicColumnService) resolveTableRelation(checkedTable constants.TableName, targetTable constants.TableName) constants.TableRelation {
	checkedTableModel := r.modelsMap[checkedTable]
	_, exists := utils.FindFieldByGormColumn(checkedTableModel, string(targetTable)+"_"+"id")
	if exists {
		return constants.TableRelationManyToOne
	}
	targetTableModel := r.modelsMap[targetTable]
	_, exists = utils.FindFieldByGormColumn(targetTableModel, string(checkedTable)+"_"+"id")
	if exists {
		return constants.TableRelationOneToMany
	}

	return constants.TableRelationNotRelated
}

func (r *dynamicColumnService) ResolveTablesRelationLink(comparor constants.TableName, target constants.TableName, prev []RelationLink, visited map[constants.TableName]bool) ([]RelationLink, error) {
	if prev == nil {
		prev = make([]RelationLink, 0)
	}

	// Build set of all visited tables including the start node
	if visited == nil {
		visited = make(map[constants.TableName]bool)
	}
	visited[comparor] = true

	for _, link := range prev {
		visited[link.Table] = true
	}

	cases := []constants.TableRelation{constants.TableRelationOneToMany, constants.TableRelationManyToOne}
	for _, caseType := range cases {
		relatedTables, exists := r.modelRelationsMap[caseType][comparor]
		if exists {
			// Direct match
			if utils.SliceContains(relatedTables, target) {
				prev = append(prev, RelationLink{
					Table:    target,
					Relation: caseType,
				})
				return prev, nil
			}

			// Recursive search
			for _, relatedTable := range relatedTables {
				// Skip if already visited (prevent circular references)
				if visited[relatedTable] {
					continue
				}

				tmpPrev := make([]RelationLink, len(prev))
				copy(tmpPrev, prev)
				tmpPrev = append(tmpPrev, RelationLink{
					Table:    relatedTable,
					Relation: caseType,
				})
				res, err := r.ResolveTablesRelationLink(relatedTable, target, tmpPrev, visited)
				if err == nil && len(res) != 0 {
					return res, nil
				}
			}
		}
	}
	return nil, errors.New("No relation found")
}
