package bill

import "github.com/gin-gonic/gin"

// Handler 结构体
type Handler struct {
}

// PostCashierOrderHideOrder 处理隐藏收银订单
// @Summary 隐藏收银订单
// @Description 隐藏收银订单
// @Tags 收银端
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/hide_order/{orderId} [post]
func (h *Handler) PostCashierOrderHideOrder(c *gin.Context) {
	// 处理隐藏收银订单的逻辑
}

// PostCashierOrderPack 处理打包收银订单
// @Summary 打包收银订单
// @Description 打包收银订单
// @Tags 收银端
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/pack/{orderId} [post]
func (h *Handler) PostCashierOrderPack(c *gin.Context) {
	// 处理打包收银订单的逻辑
}

// DeleteCashierOrderProductCancelGift 处理取消收银订单产品的礼物
// @Summary 取消收银订单产品的礼物
// @Description 取消收银订单产品的礼物
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/product/cancel_gift/{productId} [delete]
func (h *Handler) DeleteCashierOrderProductCancelGift(c *gin.Context) {
	// 处理取消收银订单产品的礼物的逻辑
}

// PostCashierOrderProductGift 处理发布收银订单产品的礼物
// @Summary 发布收银订单产品的礼物
// @Description 发布收银订单产品的礼物
// @Tags 收银端
// @Accept json
// @Produce json
// @Param gift body nil true "礼物详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/product/gift [post]
func (h *Handler) PostCashierOrderProductGift(c *gin.Context) {
	// 处理发布收银订单产品的礼物的逻辑
}

// PostCashierOrderProductPrice 处理发布收银订单产品的价格
// @Summary 发布收银订单产品的价格
// @Description 发布收银订单产品的价格
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Param price body float64 true "新价格"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/product/price/{productId} [post]
func (h *Handler) PostCashierOrderProductPrice(c *gin.Context) {
	// 处理发布收银订单产品的价格的逻辑
}

// GetCashierOrderProductRemark 处理获取收银订单产品的备注
// @Summary 获取收银订单产品的备注
// @Description 获取收银订单产品的备注
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 200 {object} nil "备注详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/product/remark/{productId} [get]
func (h *Handler) GetCashierOrderProductRemark(c *gin.Context) {
	// 处理获取收银订单产品的备注的逻辑
}

// PostCashierOrderProductRemark 处理发布收银订单产品的备注
// @Summary 发布收银订单产品的备注
// @Description 发布收银订单产品的备注
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Param remark body nil true "备注详情"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/product/remark/{productId} [post]
func (h *Handler) PostCashierOrderProductRemark(c *gin.Context) {
	// 处理发布收银订单产品的备注的逻辑
}

// GetCashierOrderShowOrderList 处理获取收银订单列表
// @Summary 获取收银订单列表
// @Description 获取收银订单列表
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {array} nil "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/show_order/list [get]
func (h *Handler) GetCashierOrderShowOrderList(c *gin.Context) {
	// 处理获取收银订单列表的逻辑
}

// PostCashierOrderUnpack 处理拆包收银订单
// @Summary 拆包收银订单
// @Description 拆包收银订单
// @Tags 收银端
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/order/unpack/{orderId} [post]
func (h *Handler) PostCashierOrderUnpack(c *gin.Context) {
	// 处理拆包收银订单的逻辑
}

func RegisterHandlers(router gin.IRouter) {

	wrapper := Handler{}

	router.POST("/order/hide_order", wrapper.PostCashierOrderHideOrder)
	router.POST("/order/pack", wrapper.PostCashierOrderPack)
	router.DELETE("/order/product/cancel_gift", wrapper.DeleteCashierOrderProductCancelGift)
	router.POST("/order/product/gift", wrapper.PostCashierOrderProductGift)
	router.POST("/order/product/price", wrapper.PostCashierOrderProductPrice)
	router.GET("/order/product/remark", wrapper.GetCashierOrderProductRemark)
	router.POST("/order/product/remark", wrapper.PostCashierOrderProductRemark)
	router.GET("/order/show_order/list", wrapper.GetCashierOrderShowOrderList)
	router.POST("/order/unpack", wrapper.PostCashierOrderUnpack)
}
