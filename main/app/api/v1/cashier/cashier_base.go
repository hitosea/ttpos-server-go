package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// BaseHandler 结构体
type BaseHandler struct {
	authSrv          service.IAuthSrv
	settingSrv       setting.ISrv
	paymentMethodSrv service.IPaymentMethodSrv
}

// GetCashierBase 收银端基础信息
// @Summary 收银端基础信息
// @Description 收银端基础信息
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CashierBase}
// @Router /cashier/base [get]
func (h *BaseHandler) GetCashierBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.CashierBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, info)
}

// GetLanguage 收银端语言
// @Summary 收银端语言
// @Description 收银端语言
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.LanguageResp}
// @Router /cashier/language [get]
func (h *BaseHandler) GetLanguage(c *gin.Context) {
	ctx := helper.GetContext(c)
	language, err := h.settingSrv.GetCashierLanguage(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, language)
}

// GetAd 收银端副屏广告
// @Summary 收银端副屏广告
// @Description 收银端副屏广告
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.Ads}
// @Router /cashier/ad [get]
func (h *BaseHandler) GetAd(c *gin.Context) {
	ctx := helper.GetContext(c)
	ads, err := h.settingSrv.GetCashierAd(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, ads)
}

// VerifyCashBoxPassword 验证钱箱密码
// @Summary 验证钱箱密码
// @Description 验证钱箱密码
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证钱箱密码参数"
// @Success 200 {object} dto.Response
// @Router /cashier/verify_cash_box_password [post]
func (h *BaseHandler) VerifyCashBoxPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.CashierVerifyPassword(ctx, constant.PasswordTypeCashBox, passwordReq.Password, helper.GetCompanyUuid(c))
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// VerifyAdvancedPassword 验证高级密码
// @Summary 验证高级密码
// @Description 验证高级密码
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /cashier/verify_advanced_password [post]
func (h *BaseHandler) VerifyAdvancedPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.CashierVerifyPassword(ctx, constant.PasswordTypeAdvanced, passwordReq.Password, helper.GetCompanyUuid(c))
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// VerifyLockPassword 验证锁屏密码
// @Summary 验证锁屏密码
// @Description 验证锁屏密码
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /cashier/verify_lock_password [post]
func (h *BaseHandler) VerifyLockPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.CashierVerifyPassword(ctx, constant.PasswordTypeLock, passwordReq.Password, helper.GetCompanyUuid(c))
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// checkUpdate 检查更新
// @Summary 检查更新
// @Description 检查更新
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param brand query string true "品牌参数"
// @Success 200 {object} dto.Response
// @Router /cashier/check_update [get]
func (h *BaseHandler) checkUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateInfo, err := h.settingSrv.CheckUpdate(ctx, constant.AppTypeCashier, c.Query("brand"), i18n.GetAcceptLanguage(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, updateInfo)
}

// GetPaymentMethodList 获取支付方式列表
// @Summary 获取支付方式列表
// @Description 获取支付方式列表
// @Tags 收银端.支付方式
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.PaymentMethodList}
// @Router /cashier/payment_method/list [get]
func (h *BaseHandler) GetPaymentMethodList(c *gin.Context) {
	helper.Success(c, h.paymentMethodSrv.GetList(helper.GetContext(c)))
}

func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)

	wrapper := &BaseHandler{
		authSrv:          authSrv,
		settingSrv:       settingSrv,
		paymentMethodSrv: paymentMethodSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/base", wrapper.GetCashierBase)                              // 获取基础信息
		privateApi.GET("/language", wrapper.GetLanguage)                             // 获取语言
		privateApi.GET("/ad", wrapper.GetAd)                                         // 收银机副屏广告
		privateApi.POST("/verify_cash_box_password", wrapper.VerifyCashBoxPassword)  // 验证钱箱密码
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword) // 验证高级密码
		privateApi.POST("/verify_lock_password", wrapper.VerifyLockPassword)         // 验证锁屏密码
		privateApi.GET("/check_update", wrapper.checkUpdate)                         // 检查更新
		privateApi.GET("/print_data", nil)                                           // todo 获取打印数据

		privateApi.GET("/payment_method/list", wrapper.GetPaymentMethodList) // 获取支付方式列表
	}
}
