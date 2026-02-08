package contract

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContractHandler interface {
	base.BaseHandler
}

type contractHandler struct {
	contractService ContractService
}

func NewContractHandler(contractService ContractService) ContractHandler {
	return &contractHandler{contractService: contractService}
}

func (h *contractHandler) GetAll(c *gin.Context) {
	entities := h.contractService.GetAll(c.Request.Context())
	c.JSON(200, types.NewListResponse(entities, nil, ""))
}

func (h *contractHandler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	entity, err := h.contractService.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, types.NewErrorResponse("Not found", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse(entity, ""))
}

func (h *contractHandler) Create(c *gin.Context) {
	var entity Contract
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, types.NewErrorResponse(err.Error(), err.Error()))
		return
	}

	created, err := h.contractService.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(201, types.NewSingleResponse(created, "Created successfully"))
}

func (h *contractHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	var updatePayload ContractUpdateRequest
	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(400, types.NewErrorResponse(err.Error(), err.Error()))
		return
	}

	updated, err := h.contractService.Update(c.Request.Context(), id, &updatePayload)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse(updated, "Updated successfully"))
}

func (h *contractHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	err = h.contractService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse[Contract](nil, "Deleted successfully"))
}
