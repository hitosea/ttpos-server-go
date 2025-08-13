package assistant

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
	authSrv          service.IAuthSrv
	settingSrv       setting.ISrv
	paymentMethodSrv service.IPaymentMethodSrv
	otherSrv         service.IOtherSrv
	deviceSrv        service.IDeviceSrv
}

// GetLanguage 语言
// @Summary 语言
// @Description 语言
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.LanguageResp}
// @Router /assistant/language [get]
func (h *BaseHandler) GetLanguage(c *gin.Context) {
	ctx := helper.GetContext(c)
	language, err := h.settingSrv.GetCashierLanguage(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, language)
}

// GetOnlineCashiers 获取在线收银机
// @Summary 获取在线收银机
// @Description 获取在线收银机
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OnlineCashierList}
// @Router /assistant/online_cashiers [get]
func (h *BaseHandler) GetOnlineCashiers(c *gin.Context) {
	helper.Success(c, h.authSrv.GetOnlineCashiers(helper.GetCompanyUuid(c)))
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 基本信息
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.AssistantBase}
// @Router /assistant/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.AssistantBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// VerifyAdvancedPassword 验证高级密码
// @Summary 验证高级密码
// @Description 验证高级密码
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /assistant/verify_advanced_password [post]
func (h *BaseHandler) VerifyAdvancedPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceAssistant, constant.PasswordTypeAdvanced, passwordReq.Password)
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// VerifyLockPassword 验证锁屏密码
// @Summary 验证锁屏密码
// @Description 验证锁屏密码
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /assistant/verify_lock_password [post]
func (h *BaseHandler) VerifyLockPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceAssistant, constant.PasswordTypeLock, passwordReq.Password)
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// CheckUpdate 检查更新
// @Summary 检查更新
// @Description 检查更新
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param brand query string true "品牌参数"
// @Success 200 {object} dto.Response
// @Router /assistant/check_update [get]
func (h *BaseHandler) CheckUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateInfo, err := h.settingSrv.CheckUpdate(ctx, constant.AppTypeAssistant, c.Query("brand"), i18n.GetAcceptLanguage(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, updateInfo)
}

// GetPaymentMethodList 获取支付方式列表
// @Summary 获取支付方式列表
// @Description 获取支付方式列表
// @Tags 点餐助手端.支付方式
// @Accept json
// @Produce json
// @Security JwtToken
// @param type query string false "显示类型 all-默认全部 checkout-结账 recharge-充值"
// @Success 200 {object} dto.Response{data=resp.PaymentMethodList}
// @Router /assistant/payment_method/list [get]
func (h *BaseHandler) GetPaymentMethodList(c *gin.Context) {
	var paymentListReq req.PaymentMethodListReq
	if err := c.ShouldBindQuery(&paymentListReq); err != nil {
		helper.HandleValidationError(c, err, paymentListReq, nil)
		return
	}
	helper.Success(c, h.paymentMethodSrv.GetList(helper.GetContext(c), paymentListReq.Type))
}

// GetReturnReason 获取退菜原因
// @Summary 获取退菜原因
// @Description 获取退菜原因
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ReturnFoodReasonResp}
// @Router /assistant/return_reason [get]
func (h *BaseHandler) GetReturnReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	respSetting, err := h.otherSrv.GetReturnFoodReasonList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, respSetting)
}

// GetFreeOrGiftReason 获取免单/赠菜原因
// @Summary 获取免单/赠菜原因
// @Description 获取免单/赠菜原因
// @Tags 点餐助手端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.GiftOrFreeOrderReasonResp}
// @Router /assistant/free_or_gift_reason [get]
func (h *BaseHandler) GetFreeOrGiftReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	respSetting, err := h.otherSrv.GetGiftOrFreeReasonList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, respSetting)
}

// EditSetting 修改设置
// @Summary 修改设置
// @Description 修改设置
// @Tags 点餐助手端.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.EditDeviceRemarkReq true "修改设置参数"
// @Success 200 {object} dto.Response
// @Router /assistant/setting [post]
func (h *BaseHandler) EditSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var settingReq req.EditDeviceRemarkReq
	if err := c.ShouldBindJSON(&settingReq); err != nil {
		helper.HandleValidationError(c, err, settingReq, nil)
		return
	}
	err := h.deviceSrv.UpdateRemark(ctx, settingReq)
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
	otherSrv := service.NewOtherSrv(dbm, cache, settingSrv)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)

	wrapper := &BaseHandler{
		authSrv:          authSrv,
		settingSrv:       settingSrv,
		paymentMethodSrv: paymentMethodSrv,
		otherSrv:         otherSrv,
		deviceSrv:        deviceSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/language", wrapper.GetLanguage)                             // 获取语言
		privateApi.GET("/online_cashiers", wrapper.GetOnlineCashiers)                // 获取在线的收银机
		privateApi.GET("/base", wrapper.GetBase)                                     // 获取基本信息
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword) // 验证高级密码
		privateApi.POST("/verify_lock_password", wrapper.VerifyLockPassword)         // 验证锁屏密码
		privateApi.GET("/check_update", wrapper.CheckUpdate)                         // 检查更新
		privateApi.GET("/payment_method/list", wrapper.GetPaymentMethodList)         // 获取支付方式列表
		privateApi.GET("/return_reason", wrapper.GetReturnReason)                    // 获取退菜原因
		privateApi.GET("/free_or_gift_reason", wrapper.GetFreeOrGiftReason)          // 获取退菜原因
		privateApi.POST("/setting", wrapper.EditSetting)                             // 修改机器备注
	}
}
