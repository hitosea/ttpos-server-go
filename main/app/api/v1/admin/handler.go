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

	cc "ttpos-server-go/pkg/context"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	dbm *database.DBManager
}

// 获取ERPNext站点公司名称
// @Summary 获取ERPNext站点公司名称
// @Description 获取ERPNext站点公司名称
// @Tags 商家管理端.商家授权
// @Accept json
// @Produce json
// @Param site_code query string true "ERPNext站点编码"
// @Param company_name query string false "公司名称"
// @Param company_abbr query string false "公司缩写"
// @Param parent_company query string false "父公司名称"
// @Success 200 {object} dto.Response
// @Router /admin/erpnext/site/company [get]
func (h *Handler) GetErpnextSiteCompany(c *gin.Context) {
	var siteCompanyReq req.ErpnextSiteCompanyReq
	if err := c.ShouldBindQuery(&siteCompanyReq); err != nil {
		helper.HandleValidationError(c, err, siteCompanyReq, nil)
		return
	}
	erpnextSiteCompanyReq := req.ErpnextSiteCompanyReq{
		SiteCode:      siteCompanyReq.SiteCode,
		CompanyName:   siteCompanyReq.CompanyName,
		CompanyAbbr:   siteCompanyReq.CompanyAbbr,
		ParentCompany: siteCompanyReq.ParentCompany,
	}
	// 调用erpnext服务，获取公司名称
	companyResp, err := erp.NewIErpSrv(h.dbm).GetCompanyList(context.Background(), erpnextSiteCompanyReq)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	posProfileResp, err := erp.NewIErpSrv(h.dbm).GetPosProfileList(context.Background(), req.GetPosProfileListReq{
		SiteCode: siteCompanyReq.SiteCode,
	})
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{
		"company_list":     companyResp.List,
		"pos_profile_list": posProfileResp.ProfileList,
	})
}

// 初始化店铺
// @Summary 初始化店铺
// @Description 初始化店铺
// @Tags 商家管理端.商家授权
// @Accept json
// @Produce json
// @Param init_shop_req body req.InitShopReq true "初始化店铺请求"
// @Success 200 {object} dto.Response
// @Router /admin/erpnext/shop/init [post]
func (h *Handler) InitShop(c *gin.Context) {
	var initShopReq req.InitShopReq
	if err := c.ShouldBindJSON(&initShopReq); err != nil {
		helper.HandleValidationError(c, err, initShopReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	ctx.SetCompanyUuid(initShopReq.CompanyUuid)
	initShopResp, err := erp.NewIErpSrv(h.dbm).InitShop(ctx, initShopReq)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}
	go func(ctx cc.Context) {
		erp.NewIErpSrv(h.dbm).SyncUomAndAttribute(ctx, req.SyncUomAndAttributeReq{
			SiteCode:    initShopReq.SiteCode,
			CompanyAbbr: initShopReq.DefaultCompanyAbbr,
		})
	}(ctx)
	helper.Success(c, initShopResp)
}

// 支付方式列表
// @Summary 支付方式列表
// @Description 支付方式列表
// @Tags 商家管理端.支付管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /admin/erpnext/lianlian/payment/add [post]
func (h *Handler) AddLianLianPayment(c *gin.Context) {
	var addLianPaymentReq req.ErpnextSiteAddLianLianPaymentReq
	if err := c.ShouldBindJSON(&addLianPaymentReq); err != nil {
		helper.HandleValidationError(c, err, addLianPaymentReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	ctx.SetCompanyUuid(addLianPaymentReq.CompanyUuid)
	err := erp.NewIErpSrv(h.dbm).AddLianPayment(ctx, addLianPaymentReq)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	wrapper := &Handler{
		dbm: dbm,
	}
	router.GET("/erpnext/site/company", middleware.Internal(), wrapper.GetErpnextSiteCompany)
	router.POST("/erpnext/shop/init", middleware.Internal(), wrapper.InitShop)
	router.POST("/erpnext/lianlian/payment/add", middleware.Internal(), wrapper.AddLianLianPayment)
}
