package invoice

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler interface {
	base.BaseHandler
}

type invoiceHandler struct {
	invoiceService InvoiceService
}

func NewInvoiceHandler(invoiceService InvoiceService) InvoiceHandler {
	return &invoiceHandler{invoiceService: invoiceService}
}

func (h *invoiceHandler) GetAll(c *gin.Context) {
	invoices := h.invoiceService.GetAll(c.Request.Context())
	c.JSON(200, types.NewListResponse(invoices, nil, ""))
}

func (h *invoiceHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	invoice, err := h.invoiceService.GetById(c.Request.Context(), idInt64)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to get invoice", err.Error()))
		return
	}
	c.JSON(200, types.NewSingleResponse(invoice, ""))
}

func (h *invoiceHandler) Create(c *gin.Context) {
	var payload Invoice
	err := c.ShouldBind(&payload)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid request", err.Error()))
		return
	}
	invoice, err := h.invoiceService.Create(c.Request.Context(), &payload)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to create invoice", err.Error()))
		return
	}
	c.JSON(200, types.NewSingleResponse(invoice, "Invoice created successfully"))
}

func (h *invoiceHandler) Update(c *gin.Context) {
	id, _ := c.Params.Get("id")
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	var updatePayload InvoiceUpdateRequest
	err = c.ShouldBindJSON(&updatePayload)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid request", err.Error()))
		return
	}
	invoice, err := h.invoiceService.Update(c.Request.Context(), idInt64, &updatePayload)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to update invoice", err.Error()))
		return
	}
	c.JSON(200, types.NewSingleResponse(invoice, "Invoice updated successfully"))
}

func (h *invoiceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid ID", err.Error()))
		return
	}
	err = h.invoiceService.Delete(c.Request.Context(), idInt64)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to delete invoice", err.Error()))
		return
	}
	c.JSON(200, types.NewSingleResponse[Invoice](nil, "Invoice deleted successfully"))
}
