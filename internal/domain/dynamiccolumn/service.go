package dynamiccolumn

import (
	"context"
	"fmt"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/utils"
	"strings"
)

type DynamicColumnService interface {
	RefreshDynamicColumnsOfRecordIds(ctx context.Context, table string, ids []int64, action constants.Action, changes map[string]Dependency, originalRecordId *int64, actionPayload interface{}) error
	CheckShouldRefreshDynamicColumn(ctx context.Context, table string, action constants.Action, changes map[string]Dependency, payload interface{}) (bool, map[string]Dependency)
}

type dynamicColumnService struct {
	dynamicColumnRepo DynamicColumnRepository
	modelsMap         map[string]interface{}
}

func NewDynamicColumnService(dynamicColumnRepo DynamicColumnRepository, modelsMap map[string]interface{}) DynamicColumnService {
	return &dynamicColumnService{dynamicColumnRepo: dynamicColumnRepo, modelsMap: modelsMap}
}

func (r *dynamicColumnService) RefreshDynamicColumnsOfRecordIds(
	ctx context.Context, table string, ids []int64, action constants.Action, changes map[string]Dependency, originalRecordId *int64, actionPayload interface{}) error {
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
	for _, col := range orderedDynamicCols {
		err := r.dynamicColumnRepo.RefreshDynamicColumn(ctx, col)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckShouldRefreshDynamicColumn checks if the action requires refreshing dynamic columns
func (r *dynamicColumnService) CheckShouldRefreshDynamicColumn(
	ctx context.Context, table string, action constants.Action,
	changes map[string]Dependency, payload interface{}) (bool, map[string]Dependency) {
	if changes == nil {
		changes = make(map[string]Dependency)
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
func (r *dynamicColumnService) addColumnsToDependency(changes map[string]Dependency, table string, columns []string) {
	dep := changes[table]
	dep.Columns = utils.AppendUnique(dep.Columns, columns...)
	changes[table] = dep
}

func (r *dynamicColumnService) getAllDynamicColumnsFromChanges(ctx context.Context, table string, changes map[string]Dependency) []DynamicColumn {
	var result []DynamicColumn
	currentChanges := changes
	for {
		dynamicColumns := r.dynamicColumnRepo.GetAllDependantsByChanges(ctx, table, currentChanges)
		if len(dynamicColumns) == 0 {
			break
		}
		result = append(result, dynamicColumns...)
		for _, colChange := range dynamicColumns {
			currentChanges = make(map[string]Dependency)
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
func (r *dynamicColumnService) determineRefreshOrder(ctx context.Context, table string, ids []int64, dynamicCols []DynamicColumn, originalRecordId *int64) []DynamicColumnWithMetadata {
	result := make([]DynamicColumnWithMetadata, 0)
	processed := make(map[string]bool) // O(1) lookup instead of O(n)
	refreshColNames := r.buildRefreshColumnNames(dynamicCols)

	// Use index-based loop so appending to dynamicCols extends the loop
	for i := 0; i < len(dynamicCols); i++ {
		col := dynamicCols[i]
		colName := col.TableName + "." + col.Name

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
		refreshColNames = append(refreshColNames, col.TableName+"."+col.Name)
	}
	return refreshColNames
}

// extractDependencyColumnNames extracts all dependency column names in "table.column" format
func (r *dynamicColumnService) extractDependencyColumnNames(dependencies map[string]Dependency) []string {
	deps := make([]string, 0)
	for depTable, dep := range dependencies {
		for _, depCol := range dep.Columns {
			deps = append(deps, depTable+"."+depCol)
		}
	}
	return deps
}

// resolveIdsFromOriginalTable resolves IDs based on the original table and record selector
// This is used for the first level of dynamic columns that directly depend on the original changed record
func (r *dynamicColumnService) resolveIdsFromOriginalTable(ctx context.Context, table string, ids []int64, selector string, originalRecordId *int64) []int64 {
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
		table: map[string]string{
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
			if matchName != depCol.TableName+"."+depCol.Name {
				continue
			}

			idsStr := make([]string, len(depCol.Ids))
			for i, v := range depCol.Ids {
				idsStr[i] = fmt.Sprintf("%d", v)
			}

			ctxObj := map[string]interface{}{
				depCol.TableName: map[string]string{
					"ids": strings.Join(idsStr, ","),
				},
			}

			query := col.Dependencies[depCol.TableName].RecordIdsSelector
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
