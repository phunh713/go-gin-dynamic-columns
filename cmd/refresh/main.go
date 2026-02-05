package main

import (
	"context"
	"flag"
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/shared/constants"
	"strings"
)

func main() {
	flag.Parse()
	args := flag.Args()
	table := args[0]
	idsStr := args[1]
	ids := []int64{}
	split := strings.Split(idsStr, ",")
	for _, idStr := range split {
		var id int64
		fmt.Sscanf(idStr, "%d", &id)
		ids = append(ids, id)
	}
	fmt.Println(args)
	ctx := context.Background()
	// Load config
	configEnv := config.LoadEnv()

	// Connect to database
	db := config.NewDB(configEnv)
	ctx = context.WithValue(ctx, config.ContextKeyDB, db)
	c := container.NewContainer()

	c.DynamicColumnService.RefreshDynamicColumnsOfRecordIds(ctx, table, ids, constants.ActionRefresh, nil, nil, nil)
}

func crontab() {

}
