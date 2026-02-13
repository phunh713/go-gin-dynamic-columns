package dynamiccolumn

import (
	"context"
	"database/sql"
	"fmt"
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/types"
	"gin-demo/internal/shared/utils"
	"strings"
)

type DynamicColumnRepository interface {
	GetAll(ctx context.Context) []DynamicColumn
	Create(ctx context.Context, column *DynamicColumn) (*DynamicColumn, error)
	GetRefreshRecordById(ctx context.Context, table constants.TableName, id int64) (interface{}, error)
	GetRecordByDependency(ctx context.Context, dependency string) []DynamicColumn
	RefreshDynamicColumn(ctx context.Context, col DynamicColumnWithMetadata) error
	GetAllDependantsByChanges(ctx context.Context, table constants.TableName, changes map[constants.TableName]Dependency) []DynamicColumn
	GetAllSelectorIds(ctx context.Context, querySelector string, ctxObj map[string]interface{}) []int64
	CreateTempIdsTable(ctx context.Context) error
	CopyIdsToTempTable(ctx context.Context, ids []int64) error
	TruncateTempTable(ctx context.Context) error
}

type dynamicColumnRepository struct {
	base.BaseHelper
	ModelsMap         types.ModelsMap
	modelRelationsMap types.ModelRelationsMap
}

func NewDynamicColumnRepository(modelsMap types.ModelsMap, modelRelationsMap types.ModelRelationsMap) DynamicColumnRepository {
	return &dynamicColumnRepository{
		ModelsMap:         modelsMap,
		modelRelationsMap: modelRelationsMap,
	}
}

func (r *dynamicColumnRepository) GetAll(ctx context.Context) []DynamicColumn {
	// Dummy data for illustration
	return []DynamicColumn{}
}

func (r *dynamicColumnRepository) GetAllDependantsByChanges(ctx context.Context, table constants.TableName, changes map[constants.TableName]Dependency) []DynamicColumn {
	if len(changes) == 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)
	var columns []DynamicColumn

	depTables := ""
	for tableName := range changes {
		depTables += fmt.Sprintf("'%s',", tableName)
	}

	// Remove the last comma
	depTables = depTables[:len(depTables)-1]

	query := fmt.Sprintf("SELECT * FROM dynamic_column WHERE dependencies ?| ARRAY[%s]", depTables)
	err := tx.Raw(query).Scan(&columns).Error
	if err != nil {
		return nil
	}

	return r.compareDepColumns(columns, changes)
}

/*
* compareDepColumns filters dynamic columns whose dependencies intersect with the provided changes.
 */
func (r *dynamicColumnRepository) compareDepColumns(columns []DynamicColumn, changes map[constants.TableName]Dependency) []DynamicColumn {
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

func (r *dynamicColumnRepository) GetRefreshRecordById(ctx context.Context, table constants.TableName, id int64) (interface{}, error) {
	tx := r.GetDbTx(ctx)
	modelType, exists := r.ModelsMap[table]
	if !exists {
		return nil, fmt.Errorf("model not found for table: %s", table)
	}

	// Create a new addressable instance of the model type
	result := utils.NewInstance(modelType)

	err := tx.Table(string(table)).Where("id = ?", id).Scan(result).Error
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
	err := tx.Raw("SELECT * FROM dynamic_column WHERE dependencies \\? ?", dependency).Scan(&results).Error
	if err != nil {
		return nil
	}
	return results
}

func (r *dynamicColumnRepository) CreateTempIdsTable(ctx context.Context) error {
	tx := r.GetDbTx(ctx)

	err := tx.Exec(fmt.Sprintf(`
		CREATE TEMP TABLE %s (
			id BIGINT PRIMARY KEY
		) ON COMMIT DROP;
	`, constants.TEMP_TABLE_NAME)).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *dynamicColumnRepository) TruncateTempTable(ctx context.Context) error {
	tx := r.GetDbTx(ctx)
	err := tx.Exec(fmt.Sprintf("TRUNCATE %s", constants.TEMP_TABLE_NAME)).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *dynamicColumnRepository) CopyIdsToTempTable(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	tx := r.GetDbTx(ctx)

	// For millions of IDs, use optimized batched multi-row VALUES insert
	// This is much faster than individual inserts or unnest for large datasets
	batchSize := 5000 // Optimal batch size for PostgreSQL

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		// Build multi-row VALUES clause: (1),(2),(3)...
		var valueStrings strings.Builder
		args := make([]interface{}, len(batch))

		for j, id := range batch {
			if j > 0 {
				valueStrings.WriteString(",")
			}
			valueStrings.WriteString("(?)")
			args[j] = id
		}

		// Single INSERT with multiple VALUES for maximum performance
		query := fmt.Sprintf("INSERT INTO %s (id) VALUES %s",
			constants.TEMP_TABLE_NAME,
			valueStrings.String())

		err := tx.Exec(query, args...).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *dynamicColumnRepository) RefreshDynamicColumn(ctx context.Context, col DynamicColumnWithMetadata) error {
	tx := r.GetDbTx(ctx)
	query := strings.Join(strings.Fields(col.Formula), " ")
	err := tx.Exec(query).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *dynamicColumnRepository) getSelectorQueries(columns []DynamicColumn, changes map[constants.TableName]Dependency) map[constants.TableName][]string {
	res := map[constants.TableName][]string{}
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
