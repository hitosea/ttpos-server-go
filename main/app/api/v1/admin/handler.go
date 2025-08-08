package admin

import (
	"context"
	"net/http"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

// 获取ERPNext站点公司名称
// @Summary 获取ERPNext站点公司名称
// @Description 获取ERPNext站点公司名称
// @Tags admin
// @Accept json
// @Produce json
// @Param erpnext_code query string true "ERPNext站点编码"
// @Param company_name query string false "公司名称"
// @Param company_abbr query string false "公司缩写"
// @Success 200 {object} dto.Response
// @Router /admin/erpnext/site/company [get]
func (h *Handler) GetErpnextSiteCompany(c *gin.Context) {

	var siteCompanyReq req.ErpnextSiteCompanyReq
	if err := c.ShouldBindQuery(&siteCompanyReq); err != nil {
		helper.HandleValidationError(c, err, siteCompanyReq, nil)
		return
	}
	erpnextSiteCompanyReq := req.ErpnextSiteCompanyReq{
		SiteCode:    siteCompanyReq.SiteCode,
		CompanyName: siteCompanyReq.CompanyName,
		CompanyAbbr: siteCompanyReq.CompanyAbbr,
	}
	// 调用erpnext服务，获取公司名称
	companyResp, err := erp.NewIErpSrv().GetCompanyList(context.Background(), erpnextSiteCompanyReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	helper.Success(c, companyResp)
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	wrapper := &Handler{}
	router.GET("/erpnext/site/company", middleware.Internal(), wrapper.GetErpnextSiteCompany)
}
