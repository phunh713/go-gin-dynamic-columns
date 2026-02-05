package dynamiccolumn

import (
	"context"
	"database/sql"
	"fmt"
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/utils"
	"strings"
)

type DynamicColumnRepository interface {
	GetAll(ctx context.Context) []DynamicColumn
	Create(ctx context.Context, column *DynamicColumn) (*DynamicColumn, error)
	GetRefreshRecordById(ctx context.Context, table string, id int64) (interface{}, error)
	GetRecordByDependency(ctx context.Context, dependency string) []DynamicColumn
	RefreshDynamicColumn(ctx context.Context, col DynamicColumnWithMetadata) error
	FindDependantTableAndIds(ctx context.Context, table string, ctxObj map[string]interface{}, changes map[string]Dependency) *map[string][]int64
	GetAllDependantsByChanges(ctx context.Context, table string, changes map[string]Dependency) []DynamicColumn
	GetAllSelectorIds(ctx context.Context, querySelector string, ctxObj map[string]interface{}) []int64
}

type dynamicColumnRepository struct {
	base.BaseHelper
	ModelsMap map[string]interface{}
}

func NewDynamicColumnRepository(modelMaps map[string]interface{}) DynamicColumnRepository {
	return &dynamicColumnRepository{
		ModelsMap: modelMaps,
	}
}

func (r *dynamicColumnRepository) GetAll(ctx context.Context) []DynamicColumn {
	// Dummy data for illustration
	return []DynamicColumn{}
}

func (r *dynamicColumnRepository) GetAllDependantsByChanges(ctx context.Context, table string, changes map[string]Dependency) []DynamicColumn {
	if len(changes) == 0 {
		return nil
	}

	tx := r.GetDbTx(ctx).Debug()
	var columns []DynamicColumn

	depTables := ""
	for tableName := range changes {
		depTables += fmt.Sprintf("'%s',", tableName)
	}

	// Remove the last comma
	depTables = depTables[:len(depTables)-1]

	query := fmt.Sprintf("SELECT * FROM dynamic_columns WHERE dependencies ?| ARRAY[%s]", depTables)
	err := tx.Raw(query).Scan(&columns).Error
	if err != nil {
		return nil
	}

	return r.compareDepColumns(columns, changes)
}

/*
* compareDepColumns filters dynamic columns whose dependencies intersect with the provided changes.
 */
func (r *dynamicColumnRepository) compareDepColumns(columns []DynamicColumn, changes map[string]Dependency) []DynamicColumn {
	res := make([]DynamicColumn, 0)
	for i := range columns {
		colDeps := columns[i].Dependencies
		for colDepKey, colDepVal := range changes {
			existingDep, exists := colDeps[colDepKey]
			if !exists {
				continue
			}
			if len(utils.StringSlicesIntersect(existingDep.Columns, colDepVal.Columns)) > 0 {
				res = append(res, columns[i])
			}
		}
	}
	return res
}

func (r *dynamicColumnRepository) Create(ctx context.Context, column *DynamicColumn) (*DynamicColumn, error) {
	tx := r.GetDbTx(ctx)

	err := tx.Create(column).Error
	if err != nil {
		return nil, err
	}

	return column, nil
}

func (r *dynamicColumnRepository) GetRefreshRecordById(ctx context.Context, table string, id int64) (interface{}, error) {
	tx := r.GetDbTx(ctx)
	modelType, exists := r.ModelsMap[table]
	if !exists {
		return nil, fmt.Errorf("model not found for table: %s", table)
	}

	// Create a new addressable instance of the model type
	result := utils.NewInstance(modelType)

	err := tx.Table(table).Where("id = ?", id).Scan(result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

/*
* GetRecordByDependency retrieves dynamic columns that depend on a specific dependency.
* dependency string sample: "invoices.total_amount"
* dependencies column is a JSONB. Sample: {"invoices.total_amount": <SQL expression>}
 */
func (r *dynamicColumnRepository) GetRecordByDependency(ctx context.Context, dependency string) []DynamicColumn {
	tx := r.GetDbTx(ctx)
	var results []DynamicColumn
	// Use parameterized query with JSONB ? operator
	err := tx.Raw("SELECT * FROM dynamic_columns WHERE dependencies \\? ?", dependency).Scan(&results).Error
	if err != nil {
		return nil
	}
	return results
}

func (r *dynamicColumnRepository) RefreshDynamicColumn(ctx context.Context, col DynamicColumnWithMetadata) error {
	query := strings.Join(strings.Fields(col.Formula), " ")
	tx := r.GetDbTx(ctx).Debug()
	setStm := fmt.Sprintf("%s = %s", col.Name, query)

	idsStr := make([]string, len(col.Ids))
	for i, id := range col.Ids {
		idsStr[i] = fmt.Sprintf("%d", id)
	}

	if len(idsStr) == 0 {
		idsStr = append(idsStr, "NULL")
	}

	distinctFromFilter := fmt.Sprintf("%s IS DISTINCT FROM %s", col.Name, query)
	updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE id IN (%s) AND %s", col.TableName, setStm, strings.Join(idsStr, ","), distinctFromFilter)
	err := tx.Exec(updateQuery).Error

	if err != nil {
		return err
	}
	return nil
}

func (r *dynamicColumnRepository) FindDependantTableAndIds(ctx context.Context, table string, ctxObj map[string]interface{}, changes map[string]Dependency) *map[string][]int64 {
	tx := r.GetDbTx(ctx)
	result := make(map[string][]int64)
	dynamicColumns := r.GetAllDependantsByChanges(ctx, table, changes)
	queries := r.getSelectorQueries(dynamicColumns, changes)
	for tableName, recordSelectors := range queries {
		for _, recordSelector := range recordSelectors {
			if recordSelector == "" {
				continue
			}
			query := utils.BuildFormulaSQL(recordSelector, ctxObj)
			// Use nullable int64 to handle NULL values from SQL
			var nullableIds []sql.NullInt64
			err := tx.Raw(query).Scan(&nullableIds).Error
			if err != nil {
				continue
			}
			// Filter out NULL values
			for _, id := range nullableIds {
				if id.Valid {
					result[tableName] = append(result[tableName], id.Int64)
				}
			}
		}
	}
	return &result
}

func (r *dynamicColumnRepository) getSelectorQueries(columns []DynamicColumn, changes map[string]Dependency) map[string][]string {
	res := map[string][]string{}
	for _, col := range columns {
		// If no changes provided, return all record selectors by table name
		if changes == nil {
			res[col.TableName] = append(res[col.TableName], col.Dependencies[col.TableName].RecordIdsSelector)
			continue
		}
		for changedTable, change := range changes {
			colDep, exists := col.Dependencies[changedTable]

			// Skip if no dependency found
			if !exists {
				continue
			}

			// Check if dynamic column's dependency columns intersect with provided change columns
			if len(utils.StringSlicesIntersect(colDep.Columns, change.Columns)) > 0 {
				res[col.TableName] = append(res[col.TableName], colDep.RecordIdsSelector)
			}
		}
	}

	return res
}

func (r *dynamicColumnRepository) GetAllSelectorIds(ctx context.Context, querySelector string, ctxObj map[string]interface{}) []int64 {
	var ids []sql.NullInt64
	tx := r.GetDbTx(ctx)
	query := utils.BuildFormulaSQL(querySelector, ctxObj)
	err := tx.Raw(query).Scan(&ids).Error
	if err != nil {
		return nil
	}
	// Convert []sql.NullInt64 to []int64
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id.Valid {
			result = append(result, id.Int64)
		}
	}
	return result
}
