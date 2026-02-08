package utils

import (
	"gin-demo/internal/shared/constants"
	"gin-demo/internal/shared/types"
)

type TableMapItem struct {
	tableId string
	model   interface{}
}

func BuildRelationMap(modelsMap types.ModelsMap) types.ModelRelationsMap {
	result := types.ModelRelationsMap{
		constants.TableRelationOneToMany: {},
		constants.TableRelationManyToOne: {},
	}

	tablesMap := make(map[constants.TableName]TableMapItem)

	for tableName, tableModel := range modelsMap {
		tablesMap[tableName] = TableMapItem{
			tableId: string(tableName) + "_id",
			model:   tableModel,
		}
	}

	for tableName, tableMap := range tablesMap {
		for relatedTableName, relatedTableMap := range tablesMap {
			if tableName == relatedTableName {
				continue
			}

			_, isOneToMany := FindFieldByGormColumn(relatedTableMap.model, tableMap.tableId)
			if isOneToMany {
				result[constants.TableRelationOneToMany][tableName] = append(result[constants.TableRelationOneToMany][tableName], relatedTableName)
			}

			_, isManyToOne := FindFieldByGormColumn(tableMap.model, relatedTableMap.tableId)
			if isManyToOne {
				result[constants.TableRelationManyToOne][tableName] = append(result[constants.TableRelationManyToOne][tableName], relatedTableName)
			}

		}

	}
	return result
}
