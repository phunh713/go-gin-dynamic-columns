package base

import (
	"github.com/gin-gonic/gin"
)

type BaseHandler interface {
	GetAll(c *gin.Context)
	GetById(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type HandlerConfig struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}
