package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	printerService "ttpos-server-go/app/printer/service"
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
	printerLogSrv    printerService.IPrinterLogSrv
	staffShiftSrv    service.IStaffShiftSrv
}

// GetCashierBase 基本信息
// @Summary 基本信息
// @Description 基本信息
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
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// GetLanguage 语言
// @Summary 语言
// @Description 语言
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
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, language)
}

// GetAd 副屏广告
// @Summary 副屏广告
// @Description 副屏广告
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
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceCashier, constant.PasswordTypeCashBox, passwordReq.Password)
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
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceCashier, constant.PasswordTypeAdvanced, passwordReq.Password)
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
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceCashier, constant.PasswordTypeLock, passwordReq.Password)
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// CheckUpdate 检查更新
// @Summary 检查更新
// @Description 检查更新
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param brand query string true "品牌参数"
// @Success 200 {object} dto.Response
// @Router /cashier/check_update [get]
func (h *BaseHandler) CheckUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateInfo, err := h.settingSrv.CheckUpdate(ctx, constant.AppTypeCashier, c.Query("brand"), i18n.GetAcceptLanguage(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
// @param type query string false "显示类型 all-默认全部 checkout-结账 recharge-充值"
// @Success 200 {object} dto.Response{data=resp.PaymentMethodList}
// @Router /cashier/payment_method/list [get]
func (h *BaseHandler) GetPaymentMethodList(c *gin.Context) {
	var paymentListReq req.PaymentMethodListReq
	if err := c.ShouldBindQuery(&paymentListReq); err != nil {
		helper.HandleValidationError(c, err, paymentListReq, nil)
		return
	}

	helper.Success(c, h.paymentMethodSrv.GetList(helper.GetContext(c), paymentListReq.Type))
}

// EditAcceptOrderSetting 修改接单设置
// @Summary 修改接单设置
// @Description 修改接单设置
// @Tags 收银端.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdateAcceptOrderSetting true "修改接单设置参数"
// @Success 200 {object} dto.Response
// @Router /cashier/setting/edit_accept_order [post]
func (h *BaseHandler) EditAcceptOrderSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var acceptOrderSetting req.UpdateAcceptOrderSetting
	if err := c.ShouldBindJSON(&acceptOrderSetting); err != nil {
		helper.HandleValidationError(c, err, acceptOrderSetting, req.UpdateAcceptOrderSettingMessage)
		return
	}
	err := h.settingSrv.EditAcceptOrderSetting(ctx, acceptOrderSetting)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// EditSystemSetting 修改系统设置
// @Summary 修改系统设置
// @Description 修改系统设置
// @Tags 收银端.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdateSystemSetting true "修改系统设置参数"
// @Success 200 {object} dto.Response
// @Router /cashier/setting/edit_system [post]
func (h *BaseHandler) EditSystemSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var systemSetting req.UpdateSystemSetting
	if err := c.ShouldBindJSON(&systemSetting); err != nil {
		helper.HandleValidationError(c, err, systemSetting, req.UpdateAcceptOrderSettingMessage)
		return
	}
	err := h.settingSrv.EditSystemSetting(ctx, systemSetting)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetSetting 获取设置
// @Summary 获取设置
// @Description 获取设置
// @Tags 收银端.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CashierBaseSetting}
// @Router /cashier/setting [get]
func (h *BaseHandler) GetSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	respSetting, err := h.settingSrv.GetCashierBaseSetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, respSetting)
}

// GetReturnReason 获取退菜原因
// @Summary 获取退菜原因
// @Description 获取退菜原因
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ReturnFoodReasonResps}
// @Router /cashier/return_reason [get]
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
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.GiftOrFreeOrderReasonResps}
// @Router /cashier/free_or_gift_reason [get]
func (h *BaseHandler) GetFreeOrGiftReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	respSetting, err := h.otherSrv.GetGiftOrFreeReasonList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, respSetting)
}

// GetPrintData 获取打印数据
// @Summary 获取打印数据
// @Description 获取打印数据
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.PrinterDataList}
// @Router /cashier/print_data [get]
func (h *BaseHandler) GetPrintData(c *gin.Context) {
	ctx := helper.GetContext(c)
	respSetting, err := h.printerLogSrv.GetPrinterData(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, respSetting)
}

// GetShiftInfo 获取交班信息
// @Summary 获取交班信息
// @Description 获取交班信息
// @Tags 收银端.交班
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ShiftInfo}
// @Router /cashier/shift [get]
func (h *BaseHandler) GetShiftInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.staffShiftSrv.GetShiftInfo(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// SubmitShift 提交交班
// @Summary 提交交班
// @Description 提交交班
// @Tags 收银端.交班
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SubmitShiftReq true "提交交班参数"
// @Success 200 {object} dto.Response
// @Router /cashier/shift [post]
// func (h *BaseHandler) SubmitShift(c *gin.Context) {
// 	ctx := helper.GetContext(c)
// 	var submitReq req.SubmitShiftReq
// 	if err := c.ShouldBindJSON(&submitReq); err != nil {
// 		helper.HandleValidationError(c, err, submitReq, nil)
// 		return
// 	}

// 	err := h.staffShiftSrv.SubmitShift(ctx, submitReq)
// 	if err != nil {
// 		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
// 		return
// 	}
// 	helper.Success(c, gin.H{}, "交班成功")
// }

// GetReport 获取报备信息
// @Summary 获取报备信息
// @Description 获取报备信息
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CashierReportResp}
// @Router /cashier/report [get]
func (h *BaseHandler) GetReport(c *gin.Context) {
	ctx := helper.GetContext(c)
	report, err := h.staffShiftSrv.GetCashierReport(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, report)
}

func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	otherSrv := service.NewOtherSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	printerLogSrv := printerService.NewPrinterLogSrv(dbm, settingSrv)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)

	wrapper := &BaseHandler{
		authSrv:          authSrv,
		settingSrv:       settingSrv,
		paymentMethodSrv: paymentMethodSrv,
		otherSrv:         otherSrv,
		printerLogSrv:    printerLogSrv,
		staffShiftSrv:    staffShiftSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.GetCashierBase)                              // 获取基础信息
		privateApi.GET("/language", wrapper.GetLanguage)                             // 获取语言
		privateApi.GET("/ad", wrapper.GetAd)                                         // 收银机副屏广告
		privateApi.POST("/verify_cash_box_password", wrapper.VerifyCashBoxPassword)  // 验证钱箱密码
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword) // 验证高级密码
		privateApi.POST("/verify_lock_password", wrapper.VerifyLockPassword)         // 验证锁屏密码
		privateApi.GET("/check_update", wrapper.CheckUpdate)                         // 检查更新
		privateApi.GET("/payment_method/list", wrapper.GetPaymentMethodList)         // 获取支付方式列表
		privateApi.GET("/return_reason", wrapper.GetReturnReason)                    // 获取退菜原因
		privateApi.GET("/free_or_gift_reason", wrapper.GetFreeOrGiftReason)          // 获取退菜原因
		privateApi.GET("/print_data", wrapper.GetPrintData)                          // 获取打印数据
		privateApi.GET("/report", wrapper.GetReport)                                 // 获取报备信息

		// 保存接单设置
		privateApi.POST("/setting/edit_accept_order", wrapper.EditAcceptOrderSetting) // 修改接单设置
		privateApi.POST("/setting/edit_system", wrapper.EditSystemSetting)            // 修改系统设置
		privateApi.GET("/setting", wrapper.GetSetting)                                // 获取设置

		// 交班
		privateApi.GET("/shift", wrapper.GetShiftInfo) // 获取交班信息
		// privateApi.POST("/shift", wrapper.SubmitShift) // 提交交班
	}
}
