package kitchen

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础相关控制器
type BaseHandler struct {
	authSrv    service.IAuthSrv
	deviceSrv  service.IDeviceSrv
	settingSrv setting.ISrv
	pinterSrv  service.IPrinterSrv
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 基本信息
// @Tags 厨显端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.KitchenBase}
// @Router /kitchen/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.KitchenBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// VerifyAdvancedPassword 验证高级密码
// @Summary 验证高级密码
// @Description 验证高级密码
// @Tags 厨显端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /kitchen/verify_advanced_password [post]
func (h *BaseHandler) VerifyAdvancedPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceKitchen, constant.PasswordTypeAdvanced, passwordReq.Password)
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// CheckUpdate 检查更新
// @Summary 检查更新
// @Description 检查更新
// @Tags 厨显端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param brand query string true "品牌参数"
// @Success 200 {object} dto.Response
// @Router /kitchen/check_update [get]
func (h *BaseHandler) CheckUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateInfo, err := h.settingSrv.CheckUpdate(ctx, constant.AppTypeKitchen, c.Query("brand"), i18n.GetAcceptLanguage(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, updateInfo)
}

// GetProductPrinterList 获取打印档口列表
// @Summary 获取打印档口列表
// @Description 获取打印档口列表
// @Tags 厨显端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ProductPrinterList}
// @Router /kitchen/product_printer_list [get]
func (h *BaseHandler) GetProductPrinterList(c *gin.Context) {
	ctx := helper.GetContext(c)
	data, err := h.pinterSrv.GetProductPrinterList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, data)
}

// Bind 绑定商品打印、设置工作模式
// @Summary 绑定商品打印、设置工作模式
// @Description 绑定商品打印、设置工作模式
// @Tags 厨显端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenBindReq true "绑定商品打印、设置工作模式参数"
// @Success 200 {object} dto.Response
// @Router /kitchen/bind [post]
func (h *BaseHandler) Bind(c *gin.Context) {
	var kitchenBindReq req.KitchenBindReq
	if err := c.ShouldBindJSON(&kitchenBindReq); err != nil {
		helper.HandleValidationError(c, err, kitchenBindReq, nil)
		return
	}
	_, err := h.deviceSrv.AddDevice(helper.GetContext(c), req.AddDeviceReq{
		DeviceId:           helper.GetDeviceSn(c),
		Brand:              kitchenBindReq.Brand,
		Source:             constant.SourceKitchen,
		FinallyLoginUuid:   helper.GetStaffUuid(c),
		CompanyUuid:        helper.GetCompanyUuid(c),
		ProductPrinterUuid: kitchenBindReq.ProductPrinterUuid,
		Remark:             kitchenBindReq.Remark,
		KdsMode:            kitchenBindReq.Mode,
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &BaseHandler{
		authSrv:    authSrv,
		deviceSrv:  deviceSrv,
		settingSrv: settingSrv,
		pinterSrv:  service.NewPrinterSrv(dbm, cache),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.GetBase)                                     // 获取基本信息
		privateApi.POST("/bind", wrapper.Bind)                                       // 绑定商品打印（修改设置）
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword) // 验证高级密码
		privateApi.GET("/check_update", wrapper.CheckUpdate)                         // 检查更新
		privateApi.GET("/product_printer_list", wrapper.GetProductPrinterList)       // 获取打印档口列表（商品打印列表）
	}
}
