package dynamiccolumn

import "gin-demo/internal/shared/constants"

type Variable struct {
	Name  string
	Value string
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
