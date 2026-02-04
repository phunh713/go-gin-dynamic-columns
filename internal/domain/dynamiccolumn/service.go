package dynamiccolumn

import (
	"context"
	"fmt"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/utils"
)

type DynamicColumnService interface {
	RefreshDynamicColumnsOfRecordId(ctx context.Context, table string, id int64, action constants.Action, changes map[string]Dependency, originalRecord interface{}) error
	CheckShouldRefreshDynamicColumn(ctx context.Context, table string, action constants.Action, changes map[string]Dependency, payload interface{}) (bool, map[string]Dependency)
}

type dynamicColumnService struct {
	dynamicColumnRepo DynamicColumnRepository
	modelsMap         map[string]interface{}
}

func NewDynamicColumnService(dynamicColumnRepo DynamicColumnRepository, modelsMap map[string]interface{}) DynamicColumnService {
	return &dynamicColumnService{dynamicColumnRepo: dynamicColumnRepo, modelsMap: modelsMap}
}

func (r *dynamicColumnService) RefreshDynamicColumnsOfRecordId(
	ctx context.Context, table string, id int64, action constants.Action, changes map[string]Dependency, originalRecord interface{}) error {
	// Check if action requires refreshing dynamic columns.
	// Get changes slice to refresh dependant tables later.
	shouldCheck, changes := r.CheckShouldRefreshDynamicColumn(ctx, table, action, changes, originalRecord)
	if !shouldCheck {
		return nil
	}

	// Get the record which needs to be refreshed based on table name and id
	refreshRecord, err := r.dynamicColumnRepo.GetRefreshRecordById(ctx, table, id)
	if err != nil {
		return err
	}

	// Get all dynamic columns affected by the changes
	dynamicCols := r.getAllDynamicColumnsFromChanges(ctx, table, changes)
	if len(dynamicCols) == 0 {
		return nil
	}

	// Determine the order of refreshing dynamic columns based on their dependencies
	orderedDynamicCols := r.determineRefreshOrder(ctx, table, id, dynamicCols)
	fmt.Printf("Ordered Dynamic Columns to refresh: %v\n", orderedDynamicCols)

	// build ctxObj for building formula SQL later
	ctxObj := map[string]interface{}{
		table:                             refreshRecord,
		fmt.Sprintf("%s:original", table): originalRecord,
	}
	changes, err = r.dynamicColumnRepo.RefreshDynamicColumns(ctx, table, id, action, changes, ctxObj)
	if err != nil {
		fmt.Println("Error refreshing internal dynamic columns:", err)
		return err
	}

	tableIdMap := r.dynamicColumnRepo.FindDependantTableAndIds(ctx, table, ctxObj, changes)
	if tableIdMap == nil {
		return nil
	}

	// print the detail value of tableIdMap
	fmt.Printf("Dependent Table and IDs to refresh:%v\n", *tableIdMap)
	for table, ids := range *tableIdMap {
		for _, id := range ids {
			// TODO: should I get the original record here?
			r.RefreshDynamicColumnsOfRecordId(ctx, table, id, constants.ActionRefresh, changes, nil)
		}
	}
	return err
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
	case constants.ActionRefresh:
		// Get all dynamic columns for this table
		dynamicCols := r.dynamicColumnRepo.GetAllByTableName(ctx, table, changes, action)
		for _, col := range dynamicCols {
			columns = append(columns, col.Name)
		}

	case constants.ActionCreate, constants.ActionDelete:
		// All fields are affected on create/delete
		model := utils.NewInstance(r.modelsMap[table])
		columns = utils.GetStructFieldJsonTags(model)

	case constants.ActionUpdate:
		// Only updated fields are affected
		columns = utils.GetStructFieldJsonTags(payload)

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
func (r *dynamicColumnService) determineRefreshOrder(ctx context.Context, table string, id int64, dynamicCols []DynamicColumn) []DynamicColumnWithIds {
	result := make([]DynamicColumnWithIds, 0)
	// resultNames is slice of flatten names already in result: ["invoices.status", ...]
	resultNames := make([]string, 0)

	refreshColNames := make([]string, 0)

	for _, col := range dynamicCols {
		refreshColNames = append(refreshColNames, col.TableName+"."+col.Name)
	}

	for _, col := range dynamicCols {
		// deps is slice of flatten names: ["invoices.total_amount", "payments.amount", ...]
		deps := make([]string, 0)
		for depTable, dep := range col.Dependencies {
			for _, depCol := range dep.Columns {
				deps = append(deps, depTable+"."+depCol)
			}
		}

		ids := make([]int64, 0)
		// if the dynamic column "col" does not depend on any of the refreshColNames (the columns to be refreshed)
		// append it to result directly, so that it will be refreshed/recalculated first
		if intersect := utils.StringSlicesIntersect(refreshColNames, deps); len(intersect) == 0 {
			query := col.Dependencies[table].RecordIdsSelector

			// base on the record selector, get all ids to refresh for this dynamic column
			if query == "" {
				ids = append(ids, id)
			} else {
				ctxObj := CtxObjIds{
					table: {
						Ids: []int64{id},
					},
				}

				foundIds := r.dynamicColumnRepo.GetAllSelectorIds(ctx, query, ctxObj)
				ids = append(ids, foundIds...)
			}

			result = append(result, DynamicColumnWithIds{
				DynamicColumn: col,
				Ids:           ids,
			})
			continue
		}

		for _, resCol := range result {
			resultNames = append(resultNames, resCol.TableName+"."+resCol.Name)
		}

		// if the dependecies of "col" are in result already, append it to result
		// So that it will be refreshed/recalculated after its dependencies ready
		if intersect := utils.StringSlicesIntersect(resultNames, deps); len(intersect) > 0 {
			for _, matchName := range intersect {
				// find the ids of match in result
				for _, depCol := range result {
					if matchName != depCol.TableName+"."+depCol.Name {
						continue
					}

					ctxObj := CtxObjIds{
						depCol.TableName: {
							Ids: depCol.Ids,
						},
					}
					foundIds := r.dynamicColumnRepo.GetAllSelectorIds(ctx, col.Dependencies[depCol.TableName].RecordIdsSelector, ctxObj)
					ids = append(ids, foundIds...)
				}
				result = append(result, DynamicColumnWithIds{
					DynamicColumn: col,
					Ids:           ids,
				})
				continue
			}

			// else, put it back to the end of dynamicCols to check again later
			dynamicCols = append(dynamicCols, col)

		}
	}
	return result
}
