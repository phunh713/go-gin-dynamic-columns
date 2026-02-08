package company

import (
	"gin-demo/internal/shared/base"
	"gin-demo/internal/shared/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CompanyHandler interface {
	base.BaseHandler
}

type companyHandler struct {
	companyService CompanyService
}

func NewCompanyHandler(companyService CompanyService) CompanyHandler {
	return &companyHandler{companyService: companyService}
}

func (h *companyHandler) GetAll(c *gin.Context) {
	companies := h.companyService.GetAllCompanies(c.Request.Context())
	c.JSON(200, types.NewListResponse(companies, nil, ""))
}

func (h *companyHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid company ID", err.Error()))
		return
	}

	company, err := h.companyService.GetById(c.Request.Context(), idInt64)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to get company", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse(company, ""))
}

func (h *companyHandler) Create(c *gin.Context) {
	var company Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid request", err.Error()))
		return
	}

	createdCompany, err := h.companyService.Create(c.Request.Context(), &company)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to create company", err.Error()))
		return
	}

	c.JSON(201, types.NewSingleResponse(createdCompany, "Company created successfully"))
}

func (h *companyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid company ID", err.Error()))
		return
	}

	var payload CompanyUpdateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, types.NewErrorResponse("Invalid request", err.Error()))
		return
	}

	err = h.companyService.Update(c.Request.Context(), idInt64, &payload)
	if err != nil {
		c.JSON(500, types.NewErrorResponse("Failed to update company", err.Error()))
		return
	}

	c.JSON(200, types.NewSingleResponse[Company](nil, "Company updated successfully"))
}

func (h *companyHandler) Delete(c *gin.Context) {

}
