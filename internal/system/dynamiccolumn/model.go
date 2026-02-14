package dynamiccolumn

import "gin-demo/internal/shared/constants"

type Dependency struct {
	RecordIdsSelector string   // SQL query to find which records are affected by this dependency
	Columns           []string // Columns from the dependency that are used
}

type DynamicColumn struct {
	ID           int64                              `json:"id" gorm:"primaryKey;column:id"`
	Name         string                             `json:"name" gorm:"column:name"`
	TableName    constants.TableName                `json:"table_name" gorm:"column:table_name"`
	Formula      string                             `json:"formula" gorm:"column:formula"`
	DefaultValue string                             `json:"default_value" gorm:"column:default_value"`
	Type         string                             `json:"type" gorm:"column:type"`
	Dependencies map[constants.TableName]Dependency `json:"dependencies" gorm:"column:dependencies;type:jsonb;serializer:json"`
	Variables    string                             `json:"variables" gorm:"column:variables"`
}

type DynamicColumnWithMetadata struct {
	DynamicColumn
	Ids    []int64 `json:"ids"`
	ctxObj map[constants.TableName]interface{}
}

type CtxObjIds = map[string]struct {
	Ids []int64 `json:"ids"`
}

type DynamicColumnCreateRequest struct {
	TableName constants.TableName `json:"table_name"`
	Name      string              `json:"name"`
	Formula   string              `json:"formula"`
	Variables string              `json:"variables"`
	Type      string              `json:"type"`
}

type Variable struct {
	Name  string
	Value string
	Table constants.TableName
}

type FormulaCte struct {
	Name  string
	Value string
	Join  string
}

type RelationLink struct {
	Table    constants.TableName
	Relation constants.TableRelation
}

type RelatedTables map[constants.TableName][]string

type CteStrings struct {
	CteJoinStrs string
	CteValues   string
}
