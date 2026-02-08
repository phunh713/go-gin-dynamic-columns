package employee

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler interface {
	base.BaseHandler
}

type employeeHandler struct {
	employeeService EmployeeService
}

func NewEmployeeHandler(employeeService EmployeeService) EmployeeHandler {
	return &employeeHandler{employeeService: employeeService}
}

func (h *employeeHandler) GetAll(c *gin.Context) {
	entities := h.employeeService.GetAll(c.Request.Context())
	c.JSON(200, types.NewListResponse(entities, nil, ""))
}

func (h *employeeHandler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	entity, err := h.employeeService.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, types.NewErrorResponse("Not found", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse(entity, ""))
}

func (h *employeeHandler) Create(c *gin.Context) {
	var entity Employee
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, types.NewErrorResponse(err.Error(), err.Error()))
		return
	}

	created, err := h.employeeService.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(201, types.NewSingleResponse(created, "Created successfully"))
}

func (h *employeeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	var updatePayload EmployeeUpdateRequest
	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(400, types.NewErrorResponse(err.Error(), err.Error()))
		return
	}

	updated, err := h.employeeService.Update(c.Request.Context(), id, &updatePayload)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse(updated, "Updated successfully"))
}

func (h *employeeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	err = h.employeeService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse[Employee](nil, "Deleted successfully"))
}
