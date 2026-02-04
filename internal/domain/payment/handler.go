package payment

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaymentHandler interface {
	base.BaseHandler
}

type paymentHandler struct {
	paymentService PaymentService
}

func NewPaymentHandler(paymentService PaymentService) PaymentHandler {
	return &paymentHandler{paymentService: paymentService}
}

func (h *paymentHandler) GetAll(c *gin.Context) {
	entities := h.paymentService.GetAllPayments(c.Request.Context())
	c.JSON(200, models.NewListResponse(entities, nil, ""))
}

func (h *paymentHandler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid ID", err.Error()))
		return
	}

	entity, err := h.paymentService.GetById(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, models.NewErrorResponse("Not found", err.Error()))
		return
	}

	c.JSON(200, models.NewSingleResponse(entity, "Payment fetched successfully"))
}

func (h *paymentHandler) Create(c *gin.Context) {
	var entity Payment
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(400, models.NewErrorResponse("Invalid request", err.Error()))
		return
	}

	created, err := h.paymentService.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(500, models.NewErrorResponse("Failed to create payment", err.Error()))
		return
	}

	c.JSON(201, models.NewSingleResponse(created, "Created successfully"))
}

func (h *paymentHandler) Update(c *gin.Context) {
	c.JSON(501, models.NewErrorResponse("Not implemented", ""))
}

func (h *paymentHandler) Delete(c *gin.Context) {
	c.JSON(501, models.NewErrorResponse("Not implemented", ""))
}
