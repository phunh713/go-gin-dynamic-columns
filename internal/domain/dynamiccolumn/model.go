package dynamiccolumn

type Dependency struct {
	RecordIdsSelector string   // SQL query to find which records are affected by this dependency
	Columns           []string // Columns from the dependency that are used
}

type DynamicColumn struct {
	ID           int64                 `json:"id" gorm:"primaryKey;column:id"`
	Name         string                `json:"name" gorm:"column:name"`
	TableName    string                `json:"table_name" gorm:"column:table_name"`
	Formula      string                `json:"formula" gorm:"column:formula"`
	DefaultValue string                `json:"default_value" gorm:"column:default_value"`
	Type         string                `json:"type" gorm:"column:type"`
	Dependencies map[string]Dependency `json:"dependencies" gorm:"column:dependencies;type:jsonb;serializer:json"`
}

type DynamicColumnWithIds struct {
	DynamicColumn
	Ids []int64 `json:"ids"`
}

type CtxObjIds = map[string]struct {
	Ids []int64 `json:"ids"`
}
