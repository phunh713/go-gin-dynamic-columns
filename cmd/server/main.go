package main

import (
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/application/container"
	"gin-demo/internal/application/middlewares"
	"gin-demo/internal/shared/utils"
)

func main() {
	configEnv := config.LoadEnv()
	logger := config.NewLogger()
	dbPool := config.NewDB(configEnv)
	app := config.NewServer(
		configEnv, middlewares.LogMiddleware(logger),
		middlewares.DbMiddleware(dbPool),
	)

	// Start Dependency Injection and Route Setup
	container := container.NewContainer()
	SetupRoutes(app, container)

	utils.PrettyPrintRoutes(app.Routes())

	app.Run(fmt.Sprintf(":%s", app.Port))
}
