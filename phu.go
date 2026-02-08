package main

import (
	"fmt"
	"gin-demo/internal/application/container"
)

func main() {
	c := container.NewContainer()

	// formula, _ := c.DynamicColumnService.BuildFormula("contract", "status", constants.SAMPLE_FORMULA, constants.SAMPLE_VARIABLES)

	// fmt.Println(formula)

	rel, err := c.DynamicColumnService.ResolveTablesRelationLink("payment", "approval", nil, nil)
	if err != nil {
		fmt.Println("Error resolving tables relation link:", err)
	} else {
		fmt.Println("Relation links:", rel)
	}
}
