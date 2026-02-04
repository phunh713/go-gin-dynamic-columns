package config

import (
	"github.com/gin-gonic/gin"
)

type App struct {
	*gin.Engine
	port string
}

func NewApp(config *ConfigEnv, middlewares ...gin.HandlerFunc) *App {
	app := gin.Default()
	for _, m := range middlewares {
		app.Use(m)
	}

	return &App{
		Engine: app,
		port:   config.AppPort,
	}
}
