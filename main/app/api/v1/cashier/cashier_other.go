package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// Handler 结构体
type Handler struct {
	authSrv service.IAuthSrv
}

// PostCashierLogin 收银端登录
// @Summary 收银端登录
// @Description 收银端登录
// @Tags 收银端
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.CashierLoginRequest true "登录参数"
// @Success 200 {object} dto.Response
// @Router /cashier/login [post]
func (h *Handler) PostCashierLogin(c *gin.Context) {
	var loginRequest req.CashierLoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		helper.HandleValidationError(c, err, loginRequest, req.CashierLoginRequestMessage)
		return
	}
	sign := c.GetHeader("X-Sign")
	if sign == "" {
		helper.Fail(c, constant.CodeBadRequest, "验证码签名不能为空")
		return
	}
	token, err := h.authSrv.Login(constant.SourceCashier, loginRequest, sign, loginRequest.Code, c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, gin.H{"token": token})
}

// GetCashierBase 收银端信息
// @Summary 收银端信息
// @Description 收银端信息
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=cashier_resp.Base}
// @Router /cashier/base [get]
func (h *Handler) GetCashierBase(c *gin.Context) {
	info, err := h.authSrv.Base(c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, info)
}

// GetCashierMemberInfo 处理获取收银员会员信息
// @Summary 获取收银员会员信息
// @Description 获取收银员会员信息
// @Tags 收银端
// @Accept json
// @Produce json
// @Param memberId path string true "会员ID"
// @Success 200 {object} nil "会员详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/member/info/{memberId} [get]
func (h *Handler) GetCashierMemberInfo(c *gin.Context) {
	// 处理获取收银员会员信息的逻辑
}

// GetCashierMemberSearch 处理搜索收银员会员
// @Summary 搜索收银员会员
// @Description 搜索收银员会员
// @Tags 收银端
// @Accept json
// @Produce json
// @Param query query string true "搜索查询"
// @Success 200 {array} nil "会员列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/member/search [get]
func (h *Handler) GetCashierMemberSearch(c *gin.Context) {
	// 处理搜索收银员会员的逻辑
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

// GetCashierPaymentTypeList 处理获取收银支付类型列表
// @Summary 获取收银支付类型列表
// @Description 获取收银支付类型列表
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {array} nil "支付类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/payment/type_list [get]
func (h *Handler) GetCashierPaymentTypeList(c *gin.Context) {
	// 处理获取收银支付类型列表的逻辑
}

// GetCashierProductInfo 处理获取收银产品信息
// @Summary 获取收银产品信息
// @Description 获取收银产品信息
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 200 {object} nil "产品详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/product/info/{productId} [get]
func (h *Handler) GetCashierProductInfo(c *gin.Context) {
	// 处理获取收银产品信息的逻辑
}

// PostCashierProductionCreate 处理创建收银生产
// @Summary 创建收银生产
// @Description 创建收银生产
// @Tags 收银端
// @Accept json
// @Produce json
// @Param production body nil true "生产详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/production/create [post]
func (h *Handler) PostCashierProductionCreate(c *gin.Context) {
	// 处理创建收银生产的逻辑
}

// PostCashierProductionReturn 处理返回收银生产
// @Summary 返回收银生产
// @Description 返回收银生产
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productionId path string true "生产ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/production/return/{productionId} [post]
func (h *Handler) PostCashierProductionReturn(c *gin.Context) {
	// 处理返回收银生产的逻辑
}

// GetCashierProductionReturnReturnFoodReason 处理获取退货原因
// @Summary 获取退货原因
// @Description 获取退货原因
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {array} nil "退货原因列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/production/return/return_foodReason [get]
func (h *Handler) GetCashierProductionReturnReturnFoodReason(c *gin.Context) {
	// 处理获取退货原因的逻辑
}

// GetCashierReasonGiftFood 处理获取赠送食物的原因
// @Summary 获取赠送食物的原因
// @Description 获取赠送食物的原因
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {array} nil "赠送原因列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/reason/gift_food [get]
func (h *Handler) GetCashierReasonGiftFood(c *gin.Context) {
	// 处理获取赠送食物的原因的逻辑
}

// GetCashierShoppingCartInfo 处理获取收银购物车信息
// @Summary 获取收银购物车信息
// @Description 获取收银购物车信息
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {object} nil "购物车详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/shopping_cart/info [get]
func (h *Handler) GetCashierShoppingCartInfo(c *gin.Context) {
	// 处理获取收银购物车信息的逻辑
}

// DeleteCashierShoppingCartProduct 处理从购物车中删除产品
// @Summary 从购物车中删除产品
// @Description 从购物车中删除产品
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/shopping_cart/product/{productId} [delete]
func (h *Handler) DeleteCashierShoppingCartProduct(c *gin.Context) {
	// 处理从购物车中删除产品的逻辑
}

// PostCashierShoppingCartProductCreate 处理在购物车中创建产品
// @Summary 在购物车中创建产品
// @Description 在购物车中创建产品
// @Tags 收银端
// @Accept json
// @Produce json
// @Param product body nil true "产品详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/shopping_cart/product/create [post]
func (h *Handler) PostCashierShoppingCartProductCreate(c *gin.Context) {
	// 处理在购物车中创建产品的逻辑
}

// PostCashierShoppingCartProductNumber 处理更新购物车中产品的数量
// @Summary 更新购物车中产品的数量
// @Description 更新购物车中产品的数量
// @Tags 收银端
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Param number body int true "新数量"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/shopping_cart/product/number/{productId} [post]
func (h *Handler) PostCashierShoppingCartProductNumber(c *gin.Context) {
	// 处理更新购物车中产品的数量的逻辑
}

// GetCashierShoppingCartSubBill 处理获取购物车的子账单
// @Summary 获取购物车的子账单
// @Description 获取购物车的子账单
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {object} nil "子账单详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/shopping_cart/sub_bill [get]
func (h *Handler) GetCashierShoppingCartSubBill(c *gin.Context) {
}

// PostCashierVerifyAdvancedPassword 处理验证收银员的高级密码
// @Summary 验证收银员的高级密码
// @Description 验证收银员的高级密码
// @Tags 收银端
// @Accept json
// @Produce json
// @Param password body string true "高级密码"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/verify_advanced_password [post]
func (h *Handler) PostCashierVerifyAdvancedPassword(c *gin.Context) {
	// Handler logic for PostCashierVerifyAdvancedPassword
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	wrapper := &Handler{
		authSrv: authSrv,
	}

	publicApi := router.Group("")
	{
		publicApi.POST("/login", wrapper.PostCashierLogin)
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/base", wrapper.GetCashierBase)
	}

	// 需要认证
	router.GET("/member/info", wrapper.GetCashierMemberInfo)
	router.GET("/member/search", wrapper.GetCashierMemberSearch)
	router.GET("/payment/type_list", wrapper.GetCashierPaymentTypeList)
	router.GET("/product/info", wrapper.GetCashierProductInfo)
	router.POST("/production/create", wrapper.PostCashierProductionCreate)
	router.POST("/production/return", wrapper.PostCashierProductionReturn)
	router.GET("/production/return/return_food_reason", wrapper.GetCashierProductionReturnReturnFoodReason)
	router.GET("/reason/gift_food", wrapper.GetCashierReasonGiftFood)
	router.GET("/shopping_cart/info", wrapper.GetCashierShoppingCartInfo)
	router.DELETE("/shopping_cart/product", wrapper.DeleteCashierShoppingCartProduct)
	router.POST("/shopping_cart/product/create", wrapper.PostCashierShoppingCartProductCreate)
	router.POST("/shopping_cart/product/number", wrapper.PostCashierShoppingCartProductNumber)
	router.GET("/shopping_cart/sub_bill", wrapper.GetCashierShoppingCartSubBill)
	router.POST("/verify_advanced_password", wrapper.PostCashierVerifyAdvancedPassword)
}
