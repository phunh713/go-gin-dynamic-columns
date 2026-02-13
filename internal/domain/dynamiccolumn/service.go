package dynamiccolumn

import (
	"context"
	"errors"
	"fmt"
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/types"
	"gin-demo/internal/shared/utils"
	"log/slog"
	"regexp"
	"strings"
)

type DynamicColumnService interface {
	RefreshDynamicColumnsOfRecordIds(ctx context.Context, table constants.TableName, ids []int64, action constants.Action, originalRecordId *int64, actionPayload interface{}) error
	CheckShouldRefreshDynamicColumn(ctx context.Context, table constants.TableName, action constants.Action, payload interface{}) (bool, map[constants.TableName]Dependency)
	BuildFormula(table constants.TableName, col string, userFormula string, userVars string) (string, error)
	ResolveTablesRelationLink(comparor constants.TableName, target constants.TableName, prev []RelationLink, visited map[constants.TableName]bool) ([]RelationLink, error)
	Create(ctx context.Context, payload *DynamicColumnCreateRequest) (*DynamicColumn, error)
}

type dynamicColumnService struct {
	dynamicColumnRepo DynamicColumnRepository
	modelsMap         types.ModelsMap
	modelRelationsMap types.ModelRelationsMap
	logger            *slog.Logger
	base.BaseHelper
}

func NewDynamicColumnService(dynamicColumnRepo DynamicColumnRepository,
	modelsMap types.ModelsMap,
	modelRelationsMap types.ModelRelationsMap,
	logger *slog.Logger,
) DynamicColumnService {
	return &dynamicColumnService{dynamicColumnRepo: dynamicColumnRepo,
		modelsMap:         modelsMap,
		modelRelationsMap: modelRelationsMap,
		logger:            logger,
	}
}

func (r *dynamicColumnService) RefreshDynamicColumnsOfRecordIds(
	ctx context.Context, table constants.TableName, ids []int64, action constants.Action, originalRecordId *int64, actionPayload interface{}) error {
	logPayload := r.GetLogPayload(ctx)
	(*logPayload)["refresh_table"] = table
	(*logPayload)["action_lead_to_refresh"] = action

	// Check if action requires refreshing dynamic columns.
	// Get changes slice to refresh dependant tables later.
	shouldCheck, changes := r.CheckShouldRefreshDynamicColumn(ctx, table, action, actionPayload)
	(*logPayload)["should_refresh"] = shouldCheck
	(*logPayload)["changes"] = changes

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
		(*logPayload)["error"] = fmt.Sprintf("Error creating temp ids table: %v", err)
		return err
	}
	for _, col := range orderedDynamicCols {
		err := r.dynamicColumnRepo.CopyIdsToTempTable(ctx, col.Ids)
		if err != nil {
			(*logPayload)["error"] = fmt.Sprintf("Error copying ids to temp ids table: %v", err)
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
	payload interface{}) (bool, map[constants.TableName]Dependency) {
	changes := make(map[constants.TableName]Dependency)

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
* determineRefreshOrder determines the sequence of refreshing dynamic columns based on their dependencies.
* table is the original table where the changes happened
 */
func (r *dynamicColumnService) determineRefreshOrder(
	ctx context.Context,
	table constants.TableName,
	ids []int64,
	dynamicCols []DynamicColumn,
	originalRecordId *int64,
) []DynamicColumnWithMetadata {
	result := make([]DynamicColumnWithMetadata, 0)
	processed := make(map[string]bool)
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
	// Step 1: Validate and parse variables
	vars, err := r.resolveVariables(userVars)
	if err != nil {
		return "", err
	}

	// Step 2: Resolve formula and related tables
	resolvedFormula, relatedTables := r.resolveFormula(userFormula, table, vars)

	// Step 3: Resolve CTEs for related tables
	resolvedCtes, err := r.resolveCte(relatedTables, table, vars)
	if err != nil {
		return "", err
	}

	// Step 4: Build final formula string from template and built components
	res := constants.FORMULA_TEMPLATE
	res = strings.ReplaceAll(res, "{{t_name}}", string(table))
	res = strings.ReplaceAll(res, "{{c_name}}", col)
	res = strings.ReplaceAll(res, "{{formula}}", resolvedFormula)
	res = strings.ReplaceAll(res, "{{cte}}", resolvedCtes.CteValues)
	res = strings.ReplaceAll(res, "{{cte_joins}}", resolvedCtes.CteJoinStrs)
	return res, nil
}

// resolveVariables parses the variable definitions from a string and returns a slice of Variable structs
func (r *dynamicColumnService) resolveVariables(varStr string) ([]Variable, error) {
	res := make([]Variable, 0)
	if varStr == "" {
		return res, nil
	}

	splitVarStr := strings.Split(varStr, "\n")

	for _, i := range splitVarStr {
		i = strings.TrimSpace(i)
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

		v.Name = matches[1]
		vValue := matches[2]
		valueRegex := regexp.MustCompile(`{{(\w+)}}\.\w+`)

		valueMatches := valueRegex.FindAllStringSubmatch(vValue, -1)
		prev := valueMatches[0][1]
		for i := 1; i < len(valueMatches); i++ {
			if valueMatches[i][1] != prev {
				return nil, errors.New("Variable definition error: Multiple table references in one variable is not supported")
			}
		}
		vValue = strings.ReplaceAll(vValue, "{{", "")
		vValue = strings.ReplaceAll(vValue, "}}", "")
		v.Value = vValue
		v.Table = constants.TableName(valueMatches[0][1])

		res = append(res, v)
	}

	return res, nil
}

// resolveFormula replaces table and column placeholders in the formula string
// placeholders are in the format {{table}}.column
func (r *dynamicColumnService) resolveFormula(formulaStr string, table constants.TableName, vars []Variable) (string, RelatedTables) {
	re := regexp.MustCompile(`{{(\w+)}}\.(\w+)`)
	relatedTables := make(RelatedTables)

	replacedFormula := re.ReplaceAllStringFunc(formulaStr, func(m string) string {
		sub := re.FindStringSubmatch(m)
		t := constants.TableName(sub[1])
		col := sub[2]
		res := string(t) + "." + col
		if t != table {
			relatedTables[t] = utils.AppendUnique(relatedTables[t], col)
			res = "cte_" + res
		}
		return res
	})

	for _, v := range vars {
		relatedTables[v.Table] = utils.AppendUnique(relatedTables[v.Table], v.Name)
	}

	return replacedFormula, relatedTables
}

// resolveCte builds CTE strings for related tables
func (r *dynamicColumnService) resolveCte(relatedTables RelatedTables, rootTable constants.TableName, vars []Variable) (*CteStrings, error) {
	ctes := make([]FormulaCte, 0)
	for relatedTable, relatedCol := range relatedTables {
		cteLinks, err := r.ResolveTablesRelationLink(rootTable, relatedTable, nil, nil)
		if err != nil {
			return nil, err
		}
		cte, err := r.createCte(rootTable, cteLinks, relatedCol, vars)
		if err != nil {
			return nil, err
		}
		ctes = append(ctes, *cte)
	}
	result := &CteStrings{}
	for _, cte := range ctes {
		result.CteValues += cte.Value + ",\n"
		result.CteJoinStrs += cte.Join + "\n"
	}
	return result, nil
}

func (r *dynamicColumnService) createCte(
	rootTable constants.TableName,
	cteLinks []RelationLink,
	joinCols []string,
	vars []Variable,
) (*FormulaCte, error) {
	var cte FormulaCte
	joinTable := cteLinks[len(cteLinks)-1].Table
	cte.Name = string(rootTable) + "_" + string(joinTable)
	cte.Join = fmt.Sprintf("LEFT JOIN %s cte_%s ON cte_%s.id = %s.id", cte.Name, joinTable, joinTable, rootTable)
	for i, col := range joinCols {
		joinCols[i] = string(joinTable) + "." + col
	}
	selectCols := strings.Join(joinCols, ", ")
	if selectCols != "" {
		selectCols = ", " + selectCols
	}
	for _, v := range vars {
		selectCols = strings.ReplaceAll(selectCols, string(joinTable)+"."+v.Name, v.Value+" AS "+v.Name)
	}

	groupByCols := make([]string, 0)
	for _, joinCol := range joinCols {
		foundVar := false
		for _, variable := range vars {
			if string(variable.Table)+"."+variable.Name == joinCol {
				foundVar = true
				break
			}
		}
		if !foundVar {
			groupByCols = utils.AppendUnique(groupByCols, joinCol)
		}
	}
	groupByColsStr := strings.Join(groupByCols, ", ")
	if groupByColsStr != "" {
		groupByColsStr = ", " + groupByColsStr
	}

	selectId := rootTable + "." + "id"
	joinStms := r.createJoinStmFromCteLinks(cteLinks, rootTable)
	cte.Value = fmt.Sprintf(`
			%s_%s AS (
				SELECT %s %s
				FROM %s
				JOIN %s tdi ON %s.id = tdi.id
				%s
				GROUP BY %s %s
			)
		`, rootTable, joinTable, selectId, selectCols, rootTable, constants.TEMP_TABLE_NAME, rootTable, strings.Join(joinStms, " \n"), selectId, groupByColsStr)
	return &cte, nil
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

func (r *dynamicColumnService) createJoinStmFromCteLinks(cteLinks []RelationLink, rootTable constants.TableName) []string {
	joinStms := make([]string, 0)
	for i, cteLink := range cteLinks {
		joinerCol := ""
		joineeCol := ""
		prevTableName := ""
		if cteLink.Relation == constants.TableRelationManyToOne {
			if i == 0 {
				prevTableName = string(rootTable)
			} else {
				prevTableName = string(cteLinks[i-1].Table)
			}
			joinerCol = prevTableName + "." + string(cteLink.Table) + "_id"
			joineeCol = string(cteLink.Table) + ".id"
		}
		if cteLink.Relation == constants.TableRelationOneToMany {
			if i == 0 {
				prevTableName = string(rootTable)
			} else {
				prevTableName = string(cteLinks[i-1].Table)
			}
			joinerCol = string(cteLink.Table) + "." + string(prevTableName) + "_id"
			joineeCol = prevTableName + ".id"
		}
		joinStms = append(joinStms, fmt.Sprintf("LEFT JOIN %s ON %s = %s AND %s.is_deleted = false", cteLink.Table, joinerCol, joineeCol, cteLink.Table))
	}
	return joinStms
}

func (r *dynamicColumnService) buildDependencies(formula string, variables string, rootTable constants.TableName) (map[constants.TableName]Dependency, error) {
	resolvedVars, err := r.resolveVariables(variables)
	if err != nil {
		return nil, err
	}
	for _, v := range resolvedVars {
		formula = strings.ReplaceAll(formula, v.Name, v.Value)
	}

	re := regexp.MustCompile(`{{(\w+)}}\.(\w+)`)

	// Find All tables in the formula
	matches := re.FindAllStringSubmatch(formula+variables, -1)

	dependencies := make(map[constants.TableName]Dependency)

	// Loop through all tables to start building dependencies
	for _, match := range matches {
		tableName := constants.TableName(match[1])

		// find out how table is related to root table (maybe through some middle tables)
		cteLinks := make([]RelationLink, 0)
		if tableName != rootTable {
			cteLinks, err = r.ResolveTablesRelationLink(tableName, rootTable, nil, nil)
			if err != nil {
				return nil, err
			}
		}
		names := []constants.TableName{tableName}
		for _, link := range cteLinks {
			names = append(names, link.Table)
		}

		for _, name := range names {
			dep := dependencies[name]

			// if table (name) is in the formula, add the column of the table to columns
			// if not, it must be middle table, so id and is_deleted are added as dependencies
			if name == constants.TableName(match[1]) {
				dep.Columns = utils.AppendUnique(dep.Columns, match[2])
			} else {
				dep.Columns = utils.AppendUnique(dep.Columns, "is_deleted", "id")
			}
			joinStm := ""

			// root table does not need record selector because it's directly refreshed based on the changed record ids
			// while other tables need to be joined back to root table to select the affected records
			if name != rootTable {
				joinStm, err = r.buildDependencySelector(name, rootTable)
				if err != nil {
					return nil, err
				}
			}
			dep.RecordIdsSelector = joinStm
			dependencies[name] = dep
		}
	}

	return dependencies, nil
}

func (r *dynamicColumnService) buildDependencySelector(depTable constants.TableName, rootTable constants.TableName) (string, error) {
	cteLinks, err := r.ResolveTablesRelationLink(depTable, rootTable, nil, nil)
	if err != nil {
		return "", err
	}
	joinStms := r.createJoinStmFromCteLinks(cteLinks, depTable)
	joinStr := fmt.Sprintf("SELECT %s.id FROM %s %s WHERE %s.id IN ({%s.ids}) GROUP BY %s.id", rootTable, depTable, strings.Join(joinStms, " "), depTable, depTable, rootTable)
	joinStr = strings.ReplaceAll(joinStr, "LEFT JOIN", "JOIN") // Use INNER JOIN for selector to ensure only matching records are returned
	return joinStr, nil
}

func (r *dynamicColumnService) Create(ctx context.Context, payload *DynamicColumnCreateRequest) (*DynamicColumn, error) {
	formula, err := r.BuildFormula(payload.TableName, payload.Name, payload.Formula, payload.Variables)
	if err != nil {
		return nil, err
	}
	dependencies, err := r.buildDependencies(payload.Formula, payload.Variables, payload.TableName)
	if err != nil {
		return nil, err
	}
	dynamicColumn := &DynamicColumn{
		TableName:    payload.TableName,
		Name:         payload.Name,
		Type:         payload.Type,
		Formula:      formula,
		Dependencies: dependencies,
		Variables:    payload.Variables,
	}

	created, err := r.dynamicColumnRepo.Create(ctx, dynamicColumn)
	if err != nil {
		return nil, err
	}
	return created, nil
}
