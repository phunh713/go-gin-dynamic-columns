package config

import (
	"github.com/gin-gonic/gin"
)

type App struct {
	*gin.Engine
	Port string
}

func NewServer(config *ConfigEnv, middlewares ...gin.HandlerFunc) *App {
	app := gin.Default()
	for _, m := range middlewares {
		app.Use(m)
	}

	return &App{
		Engine: app,
		Port:   config.AppPort,
	}
}
