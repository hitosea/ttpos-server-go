package menu

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	productSrv service.IProductSrv // 产品服务
	settingSrv setting.ISrv        // 设置
}

// GetProductCategoryList 获取产品类别列表
// @Summary 获取产品类别列表
// @Description 获取产品类别列表
// @Tags 电子菜单
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=product_resp.ProductCategoryListResp} "成功"
// @Router /menu/product/category/list [get]
func (h *Handler) GetProductCategoryList(c *gin.Context) {
	res, err := h.productSrv.GetProductCategoryList(helper.GetCompanyUuid(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductList 获取产品列表
// @Summary 获取产品列表
// @Description 获取产品列表
// @Tags 电子菜单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=product_resp.ProductListWithPaginationResp} "成功"
// @Router /menu/product/list [get]
func (h *Handler) GetProductList(c *gin.Context) {
	listReq := req.ProductListReq{}
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductList(helper.GetContext(c), listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// BaseInfo 获取基础信息
// @Summary 获取基础信息
// @Description 获取基础信息
// @Tags 电子菜单
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.MenuBaseInfo}
// @Router /menu/base [get]
func (h *Handler) BaseInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	currencySetting, err1 := h.settingSrv.GetCurrencySetting(ctx)
	cloudBasic, err2 := h.settingSrv.GetCloudBasicSetting(ctx)
	cashierSetting, err3 := h.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	h5Setting, err4 := h.settingSrv.GetH5Setting(ctx, nil)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("获取配置失败")))
		return
	}
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	companyResp := resp.Company{ // menu
		Uuid:                  company.Uuid,
		Name:                  company.Name,
		Logo:                  utils.AddImageDomain(company.Logo, utils.GetBaseURL(c.Request), true),
		TimeZone:              companySetting.Timezone,
		ExpireTime:            company.ExpireTime,
		IsOpenMember:          companySetting.IsOpenMember,
		IsOpenBuffet:          companySetting.IsOpenBuffet,
		IsOpenH5Order:         companySetting.IsOpenH5Order,
		IsOpenRider:           companySetting.IsOpenRider(),
		IsEnableErp:           company.IsOpenErp(),
		IsOpenGrabDelivery:    companySetting.IsOpenGrabDelivery(),
		IsOpenLINEMANDelivery: companySetting.IsOpenLINEMANDelivery(),
	}
	helper.Success(c, resp.MenuBaseInfo{
		Currency:      currencySetting,
		CloudBasic:    cloudBasic,
		Company:       companyResp,
		IsShowSoldOut: cashierSetting.MenuShowSoldOut == "1",
		Menu: resp.Menu{
			LanguageList:    h5Setting.LanguageList,
			Language:        h5Setting.Language,
			DefaultLanguage: h5Setting.DefaultLanguage,
		},
	})
}

// RegisterMenuHandlers 注册菜单路由
func RegisterMenuHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	translateSrv := service.NewTranslateSrv(dbm, cache)
	productService := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
	// 初始化处理器
	wrapper := Handler{
		productSrv: productService,
		settingSrv: settingSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.BusinessMenu(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.BaseInfo)                                // 获取货币信息
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取产品类别列表
		privateApi.GET("/product/list", wrapper.GetProductList)                  // 获取产品列表
	}
}
