package admin

import (
	"context"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	dbm *database.DBManager
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
	companyResp, err := erp.NewIErpSrv(h.dbm).GetCompanyList(context.Background(), erpnextSiteCompanyReq)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, companyResp)
}

// 初始化店铺
// @Summary 初始化店铺
// @Description 初始化店铺
// @Tags admin
// @Accept json
// @Produce json
// @Param init_shop_req body req.InitShopReq true "初始化店铺请求"
// @Success 200 {object} dto.Response
// @Router /admin/erpnext/shop/init [post]
func (h *Handler) InitShop(c *gin.Context) {
	ctx := helper.GetContext(c)
	var initShopReq req.InitShopReq
	if err := c.ShouldBindJSON(&initShopReq); err != nil {
		helper.HandleValidationError(c, err, initShopReq, nil)
		return
	}
	initShopResp, err := erp.NewIErpSrv(h.dbm).InitShop(ctx, initShopReq)

	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	go func() {
		// NOTE 后续如果有其他的site，需要做区分
		erp.NewIErpSrv(h.dbm).SyncUomAndAttribute(helper.GetContext(c.Copy()), req.SyncUomAndAttributeReq{
			SiteCode: initShopReq.SiteCode,
		})
	}()
	helper.Success(c, initShopResp)
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	wrapper := &Handler{
		dbm: dbm,
	}
	router.GET("/erpnext/site/company", middleware.Internal(), wrapper.GetErpnextSiteCompany)
	router.POST("/erpnext/shop/init", middleware.Internal(), wrapper.InitShop)
}
