package types

import "gin-demo/internal/shared/constants"

type ModelsMap map[constants.TableName]interface{}

type ModelRelationsMap map[constants.TableRelation]map[constants.TableName][]constants.TableName
