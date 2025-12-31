package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	printerService "ttpos-server-go/app/modules/printer/service"
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
	authSrv              service.IAuthSrv
	settingSrv           setting.ISrv
	paymentMethodSrv     service.IPaymentMethodSrv
	otherSrv             service.IOtherSrv
	printerLogSrv        printerService.IPrinterLogSrv
	staffShiftSrv        service.IStaffShiftSrv
	printerSrv           printerService.IPrinterSrv
	marketingActivitySrv service.IMarketingActivitySrv
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 基本信息
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CashierBase}
// @Router /cashier/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
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

// EditAcceptMemberOrderSetting 修改接单会员订单设置
// @Summary 修改接单会员订单设置
// @Description 修改接单会员订单设置
// @Tags 收银端.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdateAcceptMemberOrderSetting true "修改接单会员订单设置参数"
// @Success 200 {object} dto.Response
// @Router /cashier/setting/edit_accept_member_order [post]
func (h *BaseHandler) EditAcceptMemberOrderSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var acceptMemberOrderSetting req.UpdateAcceptMemberOrderSetting
	if err := c.ShouldBindJSON(&acceptMemberOrderSetting); err != nil {
		helper.HandleValidationError(c, err, acceptMemberOrderSetting, req.UpdateAcceptMemberOrderSettingMessage)
		return
	}
	err := h.settingSrv.EditAcceptMemberOrderSetting(ctx, acceptMemberOrderSetting)
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
// @Success 200 {object} dto.Response{data=resp.ReturnFoodReasonResp}
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
// @Success 200 {object} dto.Response{data=resp.GiftOrFreeOrderReasonResp}
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
// @Success 200 {object} dto.Response{data=resp.ShiftSubmit}
// @Router /cashier/shift [post]
func (h *BaseHandler) SubmitShift(c *gin.Context) {
	ctx := helper.GetContext(c)
	var submitReq req.SubmitShiftReq
	if err := c.ShouldBindJSON(&submitReq); err != nil {
		helper.HandleValidationError(c, err, submitReq, nil)
		return
	}

	info, err := h.staffShiftSrv.SubmitShift(ctx, submitReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info, "交班成功")
}

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

// SubmitReport 提交报备信息
// @Summary 提交报备信息
// @Description 提交报备信息
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CashierReportReq true "提交报备信息"
// @Success 200 {object} dto.Response{data=resp.CashierReportResp}
// @Router /cashier/report [post]
func (h *BaseHandler) SubmitReport(c *gin.Context) {
	ctx := helper.GetContext(c)
	var reportReq req.CashierReportReq
	if err := c.ShouldBindJSON(&reportReq); err != nil {
		helper.HandleValidationError(c, err, reportReq, nil)
		return
	}
	err := h.staffShiftSrv.SubmitCashierReport(ctx, reportReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetOpenCashBoxPrinterConfig 获取打开钱箱的打印机配置
// @Summary 获取打开钱箱的打印机配置
// @Description 获取打开钱箱的打印机配置
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.PrinterData}
// @Router /cashier/open_cash_box_printer_config [get]
func (h *BaseHandler) GetOpenCashBoxPrinterConfig(c *gin.Context) {
	ctx := helper.GetContext(c)
	resp, err := h.printerLogSrv.GetStaticOpenCashBoxPrinterConfig(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// ShiftWithdraw 交班取钱
// @Summary 交班取钱
// @Description 交班取钱
// @Tags 收银端.交班
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ShiftWithdrawReq true "交班取钱参数"
// @Success 200 {object} dto.Response
// @Router /cashier/shift/withdraw [post]
func (h *BaseHandler) ShiftWithdraw(c *gin.Context) {
	ctx := helper.GetContext(c)
	var withdrawReq req.ShiftWithdrawReq
	if err := c.ShouldBindJSON(&withdrawReq); err != nil {
		helper.HandleValidationError(c, err, withdrawReq, nil)
		return
	}

	err := h.staffShiftSrv.ShiftWithdraw(ctx, withdrawReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// ShiftDeposit 交班存钱
// @Summary 交班存钱
// @Description 交班存钱
// @Tags 收银端.交班
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ShiftDepositReq true "交班存钱参数"
// @Success 200 {object} dto.Response
// @Router /cashier/shift/deposit [post]
func (h *BaseHandler) ShiftDeposit(c *gin.Context) {
	ctx := helper.GetContext(c)
	var depositReq req.ShiftDepositReq
	if err := c.ShouldBindJSON(&depositReq); err != nil {
		helper.HandleValidationError(c, err, depositReq, nil)
		return
	}

	err := h.staffShiftSrv.ShiftDeposit(ctx, depositReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// ShiftPrinter 交班打印
// @Summary 交班打印
// @Description 交班打印
// @Tags 收银端.交班
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ShiftPrinterReq true "交班打印参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/shift/printer [post]
func (h *BaseHandler) ShiftPrinter(c *gin.Context) {
	ctx := helper.GetContext(c)
	var printerReq req.ShiftPrinterReq
	if err := c.ShouldBindJSON(&printerReq); err != nil {
		helper.HandleValidationError(c, err, printerReq, nil)
		return
	}
	printerData, err := h.staffShiftSrv.ShiftPrinter(ctx, printerReq, 1, false)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, printerData, "发送成功")
}

// UsbPrinterReport usb 打印机上报
// @Summary usb 打印机上报
// @Description usb 打印机上报
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UsbPrinterReportReq true "usb 打印机上报参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/usb/printer/report [post]
func (h *BaseHandler) UsbPrinterReport(c *gin.Context) {
	ctx := helper.GetContext(c)
	var reportReq req.UsbPrinterReportReq
	if err := c.ShouldBindJSON(&reportReq); err != nil {
		helper.HandleValidationError(c, err, reportReq, nil)
		return
	}
	data, err := h.printerSrv.UsbPrinterReport(ctx, reportReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, data)
}

// 解密活动二维码
// @Summary 解密活动二维码
// @Description 解密活动二维码
// @Tags 收银端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DecryptQrCodeReq true "解密活动二维码参数"
// @Success 200 {object} dto.Response{data=resp.DecryptQrCodeResp}
// @Router /cashier/decrypt/activity/qr_code [post]
func (h *BaseHandler) DecryptQrCode(c *gin.Context) {
	ctx := helper.GetContext(c)
	var decryptQrCodeReq req.DecryptQrCodeReq
	if err := c.ShouldBindJSON(&decryptQrCodeReq); err != nil {
		helper.HandleValidationError(c, err, decryptQrCodeReq, nil)
		return
	}
	decryptQrCodeResp, err := h.marketingActivitySrv.DecryptQrCode(ctx, decryptQrCodeReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, decryptQrCodeResp)
}

func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	otherSrv := service.NewOtherSrv(dbm, cache, settingSrv)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	printerLogSrv := printerService.NewPrinterLogSrv(dbm, settingSrv)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	printerSrv := printerService.NewPrinterSrv(dbm, cache)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	marketingActivitySrv := service.NewMarketingActivitySrv(dbm, cache)

	wrapper := &BaseHandler{
		authSrv:              authSrv,
		settingSrv:           settingSrv,
		paymentMethodSrv:     paymentMethodSrv,
		otherSrv:             otherSrv,
		printerLogSrv:        printerLogSrv,
		staffShiftSrv:        staffShiftSrv,
		printerSrv:           printerSrv,
		marketingActivitySrv: marketingActivitySrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.GetBase)                                             // 获取基础信息
		privateApi.GET("/language", wrapper.GetLanguage)                                     // 获取语言
		privateApi.GET("/ad", wrapper.GetAd)                                                 // 收银机副屏广告
		privateApi.POST("/verify_cash_box_password", wrapper.VerifyCashBoxPassword)          // 验证钱箱密码
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword)         // 验证高级密码
		privateApi.POST("/verify_lock_password", wrapper.VerifyLockPassword)                 // 验证锁屏密码
		privateApi.GET("/check_update", wrapper.CheckUpdate)                                 // 检查更新
		privateApi.GET("/payment_method/list", wrapper.GetPaymentMethodList)                 // 获取支付方式列表
		privateApi.GET("/return_reason", wrapper.GetReturnReason)                            // 获取退菜原因
		privateApi.GET("/free_or_gift_reason", wrapper.GetFreeOrGiftReason)                  // 获取退菜原因
		privateApi.GET("/print_data", wrapper.GetPrintData)                                  // 获取打印数据
		privateApi.GET("/report", wrapper.GetReport)                                         // 获取报备信息
		privateApi.POST("/report", wrapper.SubmitReport)                                     // 提交报备信息
		privateApi.GET("/open_cash_box_printer_config", wrapper.GetOpenCashBoxPrinterConfig) // 获取打开钱箱的打印机配置

		// 保存接单设置
		privateApi.POST("/setting/edit_accept_order", wrapper.EditAcceptOrderSetting)              // 修改接单设置
		privateApi.POST("/setting/edit_accept_member_order", wrapper.EditAcceptMemberOrderSetting) // 修改接单外送订单设置
		privateApi.POST("/setting/edit_system", wrapper.EditSystemSetting)                         // 修改系统设置
		privateApi.GET("/setting", wrapper.GetSetting)                                             // 获取设置

		// 交班
		privateApi.GET("/shift", wrapper.GetShiftInfo)            // 获取交班信息
		privateApi.POST("/shift", wrapper.SubmitShift)            // 提交交班
		privateApi.POST("/shift/withdraw", wrapper.ShiftWithdraw) // 取钱
		privateApi.POST("/shift/deposit", wrapper.ShiftDeposit)   // 存钱
		privateApi.POST("/shift/printer", wrapper.ShiftPrinter)   // 打印

		// usb 打印
		privateApi.POST("/usb/printer/report", wrapper.UsbPrinterReport) // 交班打印

		// 活动二维码
		privateApi.POST("/decrypt/activity/qr_code", wrapper.DecryptQrCode) // 解密活动二维码
	}
}
