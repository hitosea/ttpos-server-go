package cashier

import (
	"fmt"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	printerService "ttpos-server-go/app/modules/printer/service"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrderOldHandler 收银点餐处理程序
type OrderOldHandler struct {
	orderSrv      service.IOrderSrv             // 订单服务
	deskSrv       service.IDeskSrv              // 桌台服务
	printerLogSrv printerService.IPrinterLogSrv // 打印日志服务
	settingSrv    setting.ISrv                  // 设置服务
	memberSrv     service.IMemberSrv            // 会员服务
	cashBoxSrv    service.ICashBoxSrv           // 钱箱服务
}

// GetCashierOrderList 处理获取订单列表
// @Summary 获取订单列表
// @Description 获取订单列表
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.OrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/old/list [get]
func (h *OrderOldHandler) GetCashierOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	//
	dataType := "all"
	switch req.Status {
	case 0:
		dataType = "payment"
	case 1:
		dataType = "complete"
	case 2:
		dataType = "cancel"
	}
	// 订单类型
	orderSource := 0
	switch req.BillType {
	case 0:
		orderSource = 10
	case 1:
		orderSource = 20
	}
	// 日期类型
	dateType := 0
	switch req.DateType {
	case 0:
		dateType = 1
	case 1:
		dateType = 2
	case 2:
		dateType = 3
	}
	//
	timeMode := []int{}
	if req.EnableCreateTime {
		timeMode = append(timeMode, 0)
	}
	if req.EnablePayTime {
		timeMode = append(timeMode, 1)
	}
	//
	orderType := -1
	switch req.DiningMethod {
	case 0:
		orderType = 1
	case 1:
		orderType = 0
	}
	// 时间
	times := []string{}
	tz := utils.IfString(ctx.GetGin().GetHeader("tz") != "", ctx.GetGin().GetHeader("tz"), "UTC+8")
	if req.QueryStartTime > 0 {
		times = append(times, utils.SetTimezone(tz).FormatUnixTime(int64(req.QueryStartTime), ""))
	}
	if req.QueryEndTime > 0 {
		times = append(times, utils.SetTimezone(tz).FormatUnixTime(int64(req.QueryEndTime), ""))
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/index", map[string]interface{}{
		"dataType":     dataType,
		"order_source": orderSource,
		"time_type":    dateType,
		"time_mode":    timeMode,
		"time":         times,
		"order_no":     req.OrderNo,
		"order_type":   orderType,
		"page":         req.PageNo,
		"list_rows":    req.PageSize,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 判断返回数据是否为空数组
	if data, ok := result["data"].([]interface{}); ok && len(data) == 0 {
		helper.Success(c, gin.H{})
		return
	}
	//
	helper.Success(c, result["data"])
}

// GetOrderInfo 处理获取订单详情
// @Summary 获取订单详情
// @Description 获取订单详情
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderInfoReq true "详情参数"
// @Success 200 {object} resp.OrderInfosResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/info [get]
func (h *OrderOldHandler) GetOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/detail", map[string]interface{}{
		"sale_bill_uuid":  req.SaleBillUuid,
		"sale_order_uuid": req.SaleOrderUuid,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	helper.Success(c, result["data"])
}

// CancelOrder 处理取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/cancel [post]
func (h *OrderOldHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/orderCancel", map[string]interface{}{
		"order_id": req.SaleBillUuid,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// DeleteOrder 处理删除订单
// @Summary 删除订单
// @Description 删除订单
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDeleteReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/delete [delete]
func (h *OrderOldHandler) DeleteOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderDeleteReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/delete", map[string]interface{}{
		"order_id": req.SaleBillUuid,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderPrint 打印小票
// @Summary 打印小票
// @Description 打印小票
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/old/order/print [post]
func (h *OrderOldHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	//
	if printReq.PrintLang == "" {
		// 获取打印机设置
		printerSetting, err := h.settingSrv.GetPrinterSetting(ctx, nil)
		if err != nil {
			logger.Logger.Error("获取打印机设置失败", zap.Error(err))
			fmt.Println("获取打印机设置失败", zap.Error(err))
		}
		printReq.PrintLang = printerSetting.DefaultLanguage
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/print", map[string]interface{}{
		"order_id":   printReq.SaleBillUuid,
		"print_lang": printReq.PrintLang,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}

	res, err := h.printerLogSrv.GetOldOrderPrinterConfig(ctx, result["data"].(map[string]interface{})["printer_data"].(string))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, res, "发送成功")
}

// OrderPrint 打印发票
// @Summary 打印发票
// @Description 打印发票
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintInvoiceReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/old/order/print/invoice [post]
func (h *OrderOldHandler) OrderPrintInvoice(c *gin.Context) {
	ctx := helper.GetContext(c)
	//
	var printReq req.OrderPrintInvoiceReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	// 获取打印机设置
	if printReq.PrintLang == "" {
		printerSetting, err := h.settingSrv.GetPrinterSetting(ctx, nil)
		if err != nil {
			logger.Logger.Error("获取打印机设置失败", zap.Error(err))
			fmt.Println("获取打印机设置失败", zap.Error(err))
		}
		printReq.PrintLang = printerSetting.DefaultLanguage
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/printInvoice", map[string]interface{}{
		"order_id":   printReq.SaleBillUuid,
		"print_lang": printReq.PrintLang,
		"invoice_info": map[string]interface{}{
			"company_name":       printReq.CompanyName,
			"company_addr":       printReq.CompanyAddr,
			"company_tax_number": printReq.CompanyTaxNumber,
			"company_phone":      printReq.CompanyPhone,
		},
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}
	res, err := h.printerLogSrv.GetOldOrderPrinterConfig(ctx, result["data"].(map[string]interface{})["printer_data"].(string))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// OrderInvoiceInfo 获取发票信息
// @Summary 获取发票信息
// @Description 获取发票信息
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @Success 200 {object} dto.Response{data=resp.SaleOrderInvoiceInfo} "发票信息"
// @Router /cashier/old/order/invoice [get]
func (h *OrderOldHandler) OrderInvoiceInfo(c *gin.Context) {
	var invoiceReq req.OrderInvoiceInfoReq
	if err := c.ShouldBindQuery(&invoiceReq); err != nil {
		helper.HandleValidationError(c, err, invoiceReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/invoiceInfo", map[string]interface{}{
		"order_id": invoiceReq.SaleBillUuid,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}
	if result["data"] == nil {
		helper.Success(c, resp.SaleOrderInvoiceInfo{})
		return
	}
	//
	helper.Success(c, resp.SaleOrderInvoiceInfo{
		CompanyName:      result["data"].(map[string]interface{})["company_name"].(string),
		CompanyAddr:      result["data"].(map[string]interface{})["company_addr"].(string),
		CompanyTaxNumber: result["data"].(map[string]interface{})["company_tax_number"].(string),
		CompanyPhone:     result["data"].(map[string]interface{})["company_phone"].(string),
	})
}

// ReturnOrderInfo 获取退款信息
// @Summary 获取退款信息
// @Description 获取退款信息
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderReturnInfoReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.OrderReturnInfoResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/return [get]
func (h *OrderOldHandler) ReturnOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReturnInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	result, err := utils.HttpGet("http://nginx/api/cashier/order.order/orderRefund", map[string]interface{}{
		"order_id": req.SaleBillUuid,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}
	// 返回结果
	helper.Success(c, result["data"])
}

// ReturnOrder 处理退款订单
// @Summary 退款订单
// @Description 退款订单
// @Tags 收银端.历史订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/return [post]
func (h *OrderOldHandler) ReturnOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReturnReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	result, err := utils.HttpPost("http://nginx/api/cashier/order.order/orderRefund", map[string]interface{}{
		"order_id": req.SaleBillUuid,
	}, map[string]string{
		"token":    ctx.GetToken(),
		"language": ctx.GetLanguage(),
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	if result["code"].(float64) != 1 {
		helper.ErrorWithDetail(c, constant.CodeFail, fmt.Errorf("%v", result["msg"].(string)))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// HandleMemberBalance 处理会员余额
func (h *OrderOldHandler) HandleMemberBalance(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 检查是否为内网IP
	clientIP := c.ClientIP()
	if !utils.IsInternalIP(clientIP) {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("只能在内网环境下访问"))
		return
	}
	// 绑定请求参数
	req := req.MemberBalanceChangeReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	if err := h.memberSrv.HandleMemberBalance(ctx, service.MemberBalanceChangeReq{
		MemberUuid:  req.MemberUuid,
		GiftMoney:   req.Money,
		Scene:       constant.MemberBalanceLogRefund,
		Describe:    "旧订单退款: " + req.OrderNo,
		RelatedUuid: req.RelatedUuid,
	}); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// HandleCashBoxBalance 处理钱箱余额
func (h *OrderOldHandler) HandleCashBoxBalance(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 检查是否为内网IP
	clientIP := c.ClientIP()
	if !utils.IsInternalIP(clientIP) {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("只能在内网环境下访问"))
		return
	}
	// 绑定请求参数
	req := req.CashBoxBalanceChangeReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 更新钱箱余额
	if err := h.cashBoxSrv.UpdateBalance(ctx, service.UpdateCashBalanceParam{
		Amount:    req.Amount,
		Scene:     constant.CashBoxLogSceneRefund,
		OrderUuid: req.RelatedUuid,
		Remark:    "旧订单退款: " + req.OrderNo,
	}); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// RegisterOrderOldHandler 注册收银订单路由
func RegisterOrderOldHandler(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	// 初始化处理器
	wrapper := OrderOldHandler{
		orderSrv:      orderSrv,
		deskSrv:       service.NewDeskSrv(dbm, service.NewLocaleSrv(), orderSrv, settingSrv, deviceSrv, mustPlanSrv),
		printerLogSrv: printerService.NewPrinterLogSrv(dbm, settingSrv),
		settingSrv:    settingSrv,
		memberSrv:     memberSrv,
		cashBoxSrv:    cashBoxSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/old/order/list", wrapper.GetCashierOrderList)            // 获取订单列表
		privateApi.GET("/old/order/info", wrapper.GetOrderInfo)                   // 获取订单详情
		privateApi.POST("/old/order/cancel", wrapper.CancelOrder)                 // 取消订单
		privateApi.GET("/old/order/return", wrapper.ReturnOrderInfo)              // 获取退款弹窗信息
		privateApi.POST("/old/order/return", wrapper.ReturnOrder)                 // 整单退款或部分退款
		privateApi.DELETE("/old/order/delete", wrapper.DeleteOrder)               // 删除订单
		privateApi.POST("/old/order/print", wrapper.OrderPrint)                   // 打印
		privateApi.POST("/old/order/print/invoice", wrapper.OrderPrintInvoice)    // 打印发票
		privateApi.GET("/old/order/invoice", wrapper.OrderInvoiceInfo)            // 获取发票信息
		privateApi.POST("/old/order/member/balance", wrapper.HandleMemberBalance) // 处理会员余额
		privateApi.POST("/old/order/cash/balance", wrapper.HandleCashBoxBalance)  // 处理钱箱余额
	}
}
