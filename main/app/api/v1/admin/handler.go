package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	dtoReq "ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	companyService service.ICompanySrv
}

func NewAdminHandler(companyService service.ICompanySrv) *Handler {
	return &Handler{companyService: companyService}
}

// CreateCompany 创建公司
// @Summary 创建公司
// @Tags 公司
// @Accept json
// @Produce json
// @param data body req.ParamCreateCompany true "创建公司参数"
// @Success 200 {object} dto.Response
// @Router /admin/company [post]
func (h *Handler) CreateCompany(c *gin.Context) {
	// 创建公司的逻辑. 新建一个company表, 然后新建一个company_setting表
	// 1. 获取参数
	param := dtoReq.ParamCreateCompany{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    constant.CodeBadRequest,
			Message: err.Error(),
		})
		return
	}

	// 创建公司
	if errCreateCompany := h.companyService.CreateCompany(param); errCreateCompany != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    constant.CodeFail,
			Message: errCreateCompany.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Message: "success", Code: constant.CodeSuccess})
}

// UpdateCompany 更新公司
// @Summary 更新公司
// @Tags 公司
// @Accept json
// @Produce json
// @param data body req.ParamUpdateCompany true "更新公司参数"
// @Success 200 {object} dto.Response
// @Router /admin/company [put]
func (h *Handler) UpdateCompany(c *gin.Context) {
	// 更新公司的逻辑. 更新company表, 然后更新company_setting表
	// 1. 获取参数
	param := dtoReq.ParamUpdateCompany{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    constant.CodeBadRequest,
			Message: err.Error(),
		})
		return
	}
	// 更新公司
	if errUpdateCompany := h.companyService.UpdateCompany(param); errUpdateCompany != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    constant.CodeFail,
			Message: errUpdateCompany.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Message: "success", Code: constant.CodeSuccess})
}

// DeleteCompany 删除公司
// @Summary 删除公司
// @Tags 公司
// @Accept json
// @Produce json
// @param id path int true "公司ID"
// @Success 200 {object} dto.Response
// @Router /admin/company [delete]
func (h *Handler) DeleteCompany(c *gin.Context) {
	// 1. 获取参数
	companyId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    constant.CodeBadRequest,
			Message: err.Error(),
		})
		return
	}
	fmt.Println("companyId", companyId)

	// 2. 删除公司
	if errDeleteCompany := h.companyService.DeleteCompany(uint(companyId)); errDeleteCompany != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    constant.CodeFail,
			Message: errDeleteCompany.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Message: "success", Code: constant.CodeSuccess})
}

// RegisterHandlers 创建与 OpenAPI 规范匹配的 http.Handler。
func RegisterHandlers(router gin.IRouter, companyService service.ICompanySrv) {
	handler := Handler{
		companyService: companyService,
	}

	router.POST("/company", handler.CreateCompany)       // 创建公司
	router.PUT("/company", handler.UpdateCompany)        // 更新公司
	router.DELETE("/company/:id", handler.DeleteCompany) // 删除公司
}
