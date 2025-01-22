package bill

import "github.com/gin-gonic/gin"

// Handler 结构体
type Handler struct {
}

// DeleteCashierBillCancel 处理删除收银账单取消
// @Summary 删除收银账单取消
// @Description 删除收银账单取消
// @Tags cashier
// @Accept json
// @Produce json
// @Param id path string true "账单ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/cancel/{id} [delete]
func (siw *Handler) DeleteCashierBillCancel(c *gin.Context) {
}

// DeleteCashierBillDiscount 处理删除收银账单折扣
// @Summary 删除收银账单折扣
// @Description 删除收银账单折扣
// @Tags cashier
// @Accept json
// @Produce json
// @Param id path string true "折扣ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/discount/{id} [delete]
func (siw *Handler) DeleteCashierBillDiscount(c *gin.Context) {
	// 处理删除收银账单折扣的逻辑
}

// PostCashierBillDiscount 处理发布收银账单折扣
// @Summary 发布收银账单折扣
// @Description 发布收银账单折扣
// @Tags cashier
// @Accept json
// @Produce json
// @Param discount body nil true "折扣详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/discount [post]
func (siw *Handler) PostCashierBillDiscount(c *gin.Context) {
	// 处理发布收银账单折扣的逻辑
}

// PostCashierBillDiscountChangePrice 处理更改收银账单折扣价格
// @Summary 更改收银账单折扣价格
// @Description 更改收银账单折扣价格
// @Tags cashier
// @Accept json
// @Produce json
// @Param id path string true "折扣ID"
// @Param price body float64 true "新价格"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/discount/changePrice/{id} [post]
func (siw *Handler) PostCashierBillDiscountChangePrice(c *gin.Context) {
	// 处理更改收银账单折扣价格的逻辑
}

// PostCashierBillDiscountGift 处理发布收银账单折扣礼物
// @Summary 发布收银账单折扣礼物
// @Description 发布收银账单折扣礼物
// @Tags cashier
// @Accept json
// @Produce json
// @Param gift body nil true "礼物详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/discount/gift [post]
func (siw *Handler) PostCashierBillDiscountGift(c *gin.Context) {
	// 处理发布收银账单折扣礼物的逻辑
}

// PostCashierBillDiscountSmallChange 处理发布收银账单小额变更
// @Summary 发布收银账单小额变更
// @Description 发布收银账单小额变更
// @Tags cashier
// @Accept json
// @Produce json
// @Param change body nil true "小额变更详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/discount/smallChange [post]
func (siw *Handler) PostCashierBillDiscountSmallChange(c *gin.Context) {
	// 处理发布收银账单小额变更的逻辑
}

// GetCashierBillOrderCheck 处理检查收银账单订单
// @Summary 检查收银账单订单
// @Description 检查收银账单订单
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /bill/order/check/{orderId} [get]
func (siw *Handler) GetCashierBillOrderCheck(c *gin.Context) {
	// 处理检查收银账单订单的逻辑
}

// PostCashierBillOrderFinishPayment 处理完成收银账单订单支付
// @Summary 完成收银账单订单支付
// @Description 完成收银账单订单支付
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/finishPayment/{orderId} [post]
func (siw *Handler) PostCashierBillOrderFinishPayment(c *gin.Context) {
	// 处理完成收银账单订单支付的逻辑
}

// DeleteCashierBillOrderMember 处理删除收银账单订单中的成员
// @Summary 删除收银账单订单中的成员
// @Description 删除收银账单订单中的成员
// @Tags cashier
// @Accept json
// @Produce json
// @Param memberId path string true "成员ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/member/{memberId} [delete]
func (siw *Handler) DeleteCashierBillOrderMember(c *gin.Context) {
	// 处理删除收银账单订单中的成员的逻辑
}

// PostCashierBillOrderMember 处理添加成员到收银账单订单
// @Summary 添加成员到收银账单订单
// @Description 添加成员到收银账单订单
// @Tags cashier
// @Accept json
// @Produce json
// @Param member body nil true "成员详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/member [post]
func (siw *Handler) PostCashierBillOrderMember(c *gin.Context) {
	// 处理添加成员到收银账单订单的逻辑
}

// PostCashierBillOrderPay 处理收银账单订单支付
// @Summary 收银账单订单支付
// @Description 收银账单订单支付
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/pay/{orderId} [post]
func (siw *Handler) PostCashierBillOrderPay(c *gin.Context) {
	// 处理收银账单订单支付的逻辑
}

// PostCashierBillOrderPaymentSmallChangeStrategy 处理收银账单订单的小额支付策略
// @Summary 收银账单订单的小额支付策略
// @Description 处理收银账单订单的小额支付策略
// @Tags cashier
// @Accept json
// @Produce json
// @Param strategy body nil true "小额支付策略详情"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/paymentSmallChangeStrategy [post]
func (siw *Handler) PostCashierBillOrderPaymentSmallChangeStrategy(c *gin.Context) {
	// 处理收银账单订单的小额支付策略的逻辑
}

// PostCashierBillOrderPrinterInvoice 处理打印收银账单订单发票
// @Summary 打印收银账单订单发票
// @Description 打印收银账单订单发票
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/printer/invoice/{orderId} [post]
func (siw *Handler) PostCashierBillOrderPrinterInvoice(c *gin.Context) {
	// 处理打印收银账单订单发票的逻辑
}

// PostCashierBillOrderPrinterPaymentBill 处理打印收银账单订单支付账单
// @Summary 打印收银账单订单支付账单
// @Description 打印收银账单订单支付账单
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/printer/paymentBill/{orderId} [post]
func (siw *Handler) PostCashierBillOrderPrinterPaymentBill(c *gin.Context) {
	// 处理打印收银账单订单支付账单的逻辑
}

// PostCashierBillOrderPrinterPrePaymentBill 处理打印收银账单订单预付款账单
// @Summary 打印收银账单订单预付款账单
// @Description 打印收银账单订单预付款账单
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/printer/prePaymentBill/{orderId} [post]
func (siw *Handler) PostCashierBillOrderPrinterPrePaymentBill(c *gin.Context) {
	// 处理打印收银账单订单预付款账单的逻辑
}

// PostCashierBillOrderProductionCancelReturn 处理取消收银账单订单的退货
// @Summary 取消收银账单订单的退货
// @Description 取消收银账单订单的退货
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/production/cancelReturn/{orderId} [post]
func (siw *Handler) PostCashierBillOrderProductionCancelReturn(c *gin.Context) {
	// 处理取消收银账单订单的退货的逻辑
}

// GetCashierBillOrderProductionOrderFinishedProductStatus 处理获取收银账单订单的成品状态
// @Summary 获取收银账单订单的成品状态
// @Description 获取收银账单订单的成品状态
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "产品状态"
// @Failure 404 {object} nil "未找到"
// @Router /bill/order/productionOrder/finishedProduct/status/{orderId} [get]
func (siw *Handler) GetCashierBillOrderProductionOrderFinishedProductStatus(c *gin.Context) {
	// 处理获取收银账单订单的成品状态的逻辑
}

// PostCashierBillOrderRevokePay 处理撤销收银账单订单支付
// @Summary 撤销收银账单订单支付
// @Description 撤销收银账单订单支付
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/order/revokePay/{orderId} [post]
func (siw *Handler) PostCashierBillOrderRevokePay(c *gin.Context) {
	// 处理撤销收银账单订单支付的逻辑
}

// DeleteCashierBillSubBill 处理删除收银账单的子账单
// @Summary 删除收银账单的子账单
// @Description 删除收银账单的子账单
// @Tags cashier
// @Accept json
// @Produce json
// @Param subBillId path string true "子账单ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/subBill/{subBillId} [delete]
func (siw *Handler) DeleteCashierBillSubBill(c *gin.Context) {
	// 处理删除收银账单的子账单的逻辑
}

// PostCashierBillSubBill 处理发布收银账单的子账单
// @Summary 发布收银账单的子账单
// @Description 发布收银账单的子账单
// @Tags cashier
// @Accept json
// @Produce json
// @Param subBill body nil true "子账单详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/subBill [post]
func (siw *Handler) PostCashierBillSubBill(c *gin.Context) {
	// 处理发布收银账单的子账单的逻辑
}

// PostCashierBillSubBillProductMove 处理在子账单中移动产品
// @Summary 在子账单中移动产品
// @Description 在子账单中移动产品
// @Tags cashier
// @Accept json
// @Produce json
// @Param moveDetails body nil true "移动详情"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /bill/subBill/product/move [post]
func (siw *Handler) PostCashierBillSubBillProductMove(c *gin.Context) {
	// 处理在子账单中移动产品的逻辑
}

func RegisterHandlers(router gin.IRouter) {

	wrapper := Handler{}

	router.DELETE("/bill/cancel", wrapper.DeleteCashierBillCancel)
	router.DELETE("/bill/discount", wrapper.DeleteCashierBillDiscount)
	router.POST("/bill/discount", wrapper.PostCashierBillDiscount)
	router.POST("/bill/discount/changePrice", wrapper.PostCashierBillDiscountChangePrice)
	router.POST("/bill/discount/gift", wrapper.PostCashierBillDiscountGift)
	router.POST("/bill/discount/smallChange", wrapper.PostCashierBillDiscountSmallChange)
	router.GET("/bill/order/check", wrapper.GetCashierBillOrderCheck)
	router.POST("/bill/order/finishPayment", wrapper.PostCashierBillOrderFinishPayment)
	router.DELETE("/bill/order/member", wrapper.DeleteCashierBillOrderMember)
	router.POST("/bill/order/member", wrapper.PostCashierBillOrderMember)
	router.POST("/bill/order/pay", wrapper.PostCashierBillOrderPay)
	router.POST("/bill/order/paymentSmallChangeStrategy", wrapper.PostCashierBillOrderPaymentSmallChangeStrategy)
	router.POST("/bill/order/printer/invoice", wrapper.PostCashierBillOrderPrinterInvoice)
	router.POST("/bill/order/printer/paymentBill", wrapper.PostCashierBillOrderPrinterPaymentBill)
	router.POST("/bill/order/printer/prePaymentBill", wrapper.PostCashierBillOrderPrinterPrePaymentBill)
	router.POST("/bill/order/production/cancelReturn", wrapper.PostCashierBillOrderProductionCancelReturn)
	router.GET("/bill/order/productionOrder/finishedProduct/status", wrapper.GetCashierBillOrderProductionOrderFinishedProductStatus)
	router.POST("/bill/order/revokePay", wrapper.PostCashierBillOrderRevokePay)
	router.DELETE("/bill/subBill", wrapper.DeleteCashierBillSubBill)
	router.POST("/bill/subBill", wrapper.PostCashierBillSubBill)
	router.POST("/bill/subBill/product/move", wrapper.PostCashierBillSubBillProductMove)
}
