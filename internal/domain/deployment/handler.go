package deployment

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeploymentHandler interface {
	base.BaseHandler
}

type deploymentHandler struct {
	deploymentService DeploymentService
}

func NewDeploymentHandler(deploymentService DeploymentService) DeploymentHandler {
	return &deploymentHandler{deploymentService: deploymentService}
}

func (h *deploymentHandler) GetAll(c *gin.Context) {
	entities := h.deploymentService.GetAll(c.Request.Context())
	c.JSON(200, models.NewListResponse(entities, nil, ""))
}

func (h *deploymentHandler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	entity, err := h.deploymentService.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, models.NewErrorResponse("Not found", err.Error()))
		return
	}

	c.JSON(200, models.NewSingleResponse[Deployment](entity, ""))
}

func (h *deploymentHandler) Create(c *gin.Context) {
	var entity Deployment
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, models.NewErrorResponse(err.Error(), err.Error()))
		return
	}

	created, err := h.deploymentService.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(201, models.NewSingleResponse[Deployment](created, "Created successfully"))
}

func (h *deploymentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	var updatePayload DeploymentUpdateRequest
	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(400, models.NewErrorResponse(err.Error(), err.Error()))
		return
	}

	updated, err := h.deploymentService.Update(c.Request.Context(), id, &updatePayload)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(200, models.NewSingleResponse[Deployment](updated, "Updated successfully"))
}

func (h *deploymentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	err = h.deploymentService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(200, models.NewSingleResponse[Deployment](nil, "Deleted successfully"))
}
