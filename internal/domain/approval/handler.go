package approval

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApprovalHandler interface {
	base.BaseHandler
}

type approvalHandler struct {
	approvalService ApprovalService
}

func NewApprovalHandler(approvalService ApprovalService) ApprovalHandler {
	return &approvalHandler{approvalService: approvalService}
}

func (h *approvalHandler) GetAll(c *gin.Context) {
	entities := h.approvalService.GetAll(c.Request.Context())
	c.JSON(200, models.NewListResponse(entities, nil, ""))
}

func (h *approvalHandler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	
	entity, err := h.approvalService.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, models.NewErrorResponse("Not found", err.Error()))
		return
	}
	
	c.JSON(200, models.NewSingleResponse[Approval](entity, ""))
}

func (h *approvalHandler) Create(c *gin.Context) {
	var entity Approval
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, models.NewErrorResponse(err.Error(), err.Error()))
		return
	}
	
	created, err := h.approvalService.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}
	
	c.JSON(201, models.NewSingleResponse[Approval](created, "Created successfully"))
}

func (h *approvalHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	
	var updatePayload ApprovalUpdateRequest
	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(400, models.NewErrorResponse(err.Error(), err.Error()))
		return
	}
	
	updated, err := h.approvalService.Update(c.Request.Context(), id, &updatePayload)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}
	
	c.JSON(200, models.NewSingleResponse[Approval](updated, "Updated successfully"))
}

func (h *approvalHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	
	err = h.approvalService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}
	
	c.JSON(200, models.NewSingleResponse[Approval](nil, "Deleted successfully"))
}
