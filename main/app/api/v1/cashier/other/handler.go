package other

import (
	"net/http"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service/cashier"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 结构体
type Handler struct {
	cashierService cashier.CashierServiceInterface
}

// GetCashierDeskList 处理获取收银台列表
// @Summary 获取收银台列表
// @Description 获取收银台列表
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "收银台列表"
// @Failure 404 {object} nil "未找到"
// @Router /desk/list [get]
func (siw *Handler) GetCashierDeskList(c *gin.Context) {
	// 处理获取收银台列表的逻辑
}

// GetCashierDeskRegionAndType 处理获取收银台的区域和类型
// @Summary 获取收银台的区域和类型
// @Description 获取收银台的区域和类型
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "收银台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /desk/regionAndType [get]
func (siw *Handler) GetCashierDeskRegionAndType(c *gin.Context) {
	// 处理获取收银台的区域和类型的逻辑
}

// PostCashierLogin 处理收银员登录
// @Summary 收银员登录
// @Description 处理收银员登录
// @Tags cashier
// @Accept json
// @Produce json
// @Param loginDetails body nil true "登录详情"
// @Success 200 {object} nil "登录成功"
// @Failure 400 {object} nil "错误请求"
// @Router /login [post]
func (siw *Handler) PostCashierLogin(c *gin.Context) {
	// 处理收银员登录的逻辑
}

// GetCashierLoginCaptcha 处理获取收银员登录验证码
// @Summary 获取收银员登录验证码
// @Description 获取收银员登录验证码
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {object} nil "验证码详情"
// @Failure 404 {object} nil "未找到"
// @Router /login/captcha [get]
func (siw *Handler) GetCashierLoginCaptcha(c *gin.Context) {
	// 处理获取收银员登录验证码的逻辑
}

// PostCashierLoginPublicKey 处理获取收银员登录的公钥
// @Summary 获取收银员登录的公钥
// @Description 获取收银员登录的公钥
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {object} nil "公钥详情"
// @Failure 404 {object} nil "未找到"
// @Router /login/publicKey [post]
func (siw *Handler) PostCashierLoginPublicKey(c *gin.Context) {
	// 处理获取收银员登录的公钥的逻辑
}

// GetCashierMemberInfo 处理获取收银员会员信息
// @Summary 获取收银员会员信息
// @Description 获取收银员会员信息
// @Tags cashier
// @Accept json
// @Produce json
// @Param memberId path string true "会员ID"
// @Success 200 {object} nil "会员详情"
// @Failure 404 {object} nil "未找到"
// @Router /member/info/{memberId} [get]
func (siw *Handler) GetCashierMemberInfo(c *gin.Context) {
	// 处理获取收银员会员信息的逻辑
}

// GetCashierMemberSearch 处理搜索收银员会员
// @Summary 搜索收银员会员
// @Description 搜索收银员会员
// @Tags cashier
// @Accept json
// @Produce json
// @Param query query string true "搜索查询"
// @Success 200 {array} nil "会员列表"
// @Failure 404 {object} nil "未找到"
// @Router /member/search [get]
func (siw *Handler) GetCashierMemberSearch(c *gin.Context) {
	// 处理搜索收银员会员的逻辑
}

// PostCashierOrderHideOrder 处理隐藏收银订单
// @Summary 隐藏收银订单
// @Description 隐藏收银订单
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /order/hideOrder/{orderId} [post]
func (siw *Handler) PostCashierOrderHideOrder(c *gin.Context) {
	// 处理隐藏收银订单的逻辑
}

// PostCashierOrderPack 处理打包收银订单
// @Summary 打包收银订单
// @Description 打包收银订单
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /order/pack/{orderId} [post]
func (siw *Handler) PostCashierOrderPack(c *gin.Context) {
	// 处理打包收银订单的逻辑
}

// DeleteCashierOrderProductCancelGift 处理取消收银订单产品的礼物
// @Summary 取消收银订单产品的礼物
// @Description 取消收银订单产品的礼物
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /order/product/cancelGift/{productId} [delete]
func (siw *Handler) DeleteCashierOrderProductCancelGift(c *gin.Context) {
	// 处理取消收银订单产品的礼物的逻辑
}

// PostCashierOrderProductGift 处理发布收银订单产品的礼物
// @Summary 发布收银订单产品的礼物
// @Description 发布收银订单产品的礼物
// @Tags cashier
// @Accept json
// @Produce json
// @Param gift body nil true "礼物详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /order/product/gift [post]
func (siw *Handler) PostCashierOrderProductGift(c *gin.Context) {
	// 处理发布收银订单产品的礼物的逻辑
}

// PostCashierOrderProductPrice 处理发布收银订单产品的价格
// @Summary 发布收银订单产品的价格
// @Description 发布收银订单产品的价格
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Param price body float64 true "新价格"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /order/product/price/{productId} [post]
func (siw *Handler) PostCashierOrderProductPrice(c *gin.Context) {
	// 处理发布收银订单产品的价格的逻辑
}

// GetCashierOrderProductRemark 处理获取收银订单产品的备注
// @Summary 获取收银订单产品的备注
// @Description 获取收银订单产品的备注
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 200 {object} nil "备注详情"
// @Failure 404 {object} nil "未找到"
// @Router /order/product/remark/{productId} [get]
func (siw *Handler) GetCashierOrderProductRemark(c *gin.Context) {
	// 处理获取收银订单产品的备注的逻辑
}

// PostCashierOrderProductRemark 处理发布收银订单产品的备注
// @Summary 发布收银订单产品的备注
// @Description 发布收银订单产品的备注
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Param remark body nil true "备注详情"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /order/product/remark/{productId} [post]
func (siw *Handler) PostCashierOrderProductRemark(c *gin.Context) {
	// 处理发布收银订单产品的备注的逻辑
}

// GetCashierOrderShowOrderList 处理获取收银订单列表
// @Summary 获取收银订单列表
// @Description 获取收银订单列表
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /order/showOrder/list [get]
func (siw *Handler) GetCashierOrderShowOrderList(c *gin.Context) {
	// 处理获取收银订单列表的逻辑
}

// PostCashierOrderUnpack 处理拆包收银订单
// @Summary 拆包收银订单
// @Description 拆包收银订单
// @Tags cashier
// @Accept json
// @Produce json
// @Param orderId path string true "订单ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /order/unpack/{orderId} [post]
func (siw *Handler) PostCashierOrderUnpack(c *gin.Context) {
	// 处理拆包收银订单的逻辑
}

// GetCashierPaymentTypeList 处理获取收银支付类型列表
// @Summary 获取收银支付类型列表
// @Description 获取收银支付类型列表
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "支付类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /payment/typeList [get]
func (siw *Handler) GetCashierPaymentTypeList(c *gin.Context) {
	// 处理获取收银支付类型列表的逻辑
}

// GetCashierProductCategory 处理获取收银产品类别
// @Summary 获取收银产品类别
// @Description 获取收银产品类别
// @Tags cashier
// @Accept json
// @Produce json
// @Param token header string true "认证令牌"
// @Param language header string true "语言"
// @Success 200 {object} dto.Response{data=resp.ProductCategory} "产品类别列表"
// @Router /cashier/product/category [get]
func (siw *Handler) GetCashierProductCategory(c *gin.Context) {
	companyId := helper.GetCompanyId(c)
	if companyId == 0 {
		helper.Fail(c, constant.CodeBadRequest, "参数错误")
		logger.Logger.Info("从token中获取公司ID失败")
		return
	}
	language := helper.GetLanguage(c)
	logger.Logger.Info("从token中获取公司ID成功", zap.Any("companyId", companyId), zap.String("从header中获取语言", language))
	// 处理获取收银产品类别的逻辑
	productCategory, err := siw.cashierService.GetProductCategory(companyId, language)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusInternalServerError, err)
		return
	}
	helper.Success(c, productCategory)
}

// GetCashierProductInfo 处理获取收银产品信息
// @Summary 获取收银产品信息
// @Description 获取收银产品信息
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 200 {object} nil "产品详情"
// @Failure 404 {object} nil "未找到"
// @Router /product/info/{productId} [get]
func (siw *Handler) GetCashierProductInfo(c *gin.Context) {
	// 处理获取收银产品信息的逻辑
}

// GetCashierProductList 处理获取收银产品列表
// @Summary 获取收银产品列表
// @Description 获取收银产品列表
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "产品列表"
// @Failure 404 {object} nil "未找到"
// @Router /product/list [get]
func (siw *Handler) GetCashierProductList(c *gin.Context) {
	// 处理获取收银产品列表的逻辑
}

// PostCashierProductionCreate 处理创建收银生产
// @Summary 创建收银生产
// @Description 创建收银生产
// @Tags cashier
// @Accept json
// @Produce json
// @Param production body nil true "生产详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /production/create [post]
func (siw *Handler) PostCashierProductionCreate(c *gin.Context) {
	// 处理创建收银生产的逻辑
}

// PostCashierProductionReturn 处理返回收银生产
// @Summary 返回收银生产
// @Description 返回收银生产
// @Tags cashier
// @Accept json
// @Produce json
// @Param productionId path string true "生产ID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /production/return/{productionId} [post]
func (siw *Handler) PostCashierProductionReturn(c *gin.Context) {
	// 处理返回收银生产的逻辑
}

// GetCashierProductionReturnReturnFoodReason 处理获取退货原因
// @Summary 获取退货原因
// @Description 获取退货原因
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "退货原因列表"
// @Failure 404 {object} nil "未找到"
// @Router /production/return/returnFoodReason [get]
func (siw *Handler) GetCashierProductionReturnReturnFoodReason(c *gin.Context) {
	// 处理获取退货原因的逻辑
}

// GetCashierReasonGiftFood 处理获取赠送食物的原因
// @Summary 获取赠送食物的原因
// @Description 获取赠送食物的原因
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {array} nil "赠送原因列表"
// @Failure 404 {object} nil "未找到"
// @Router /reason/giftFood [get]
func (siw *Handler) GetCashierReasonGiftFood(c *gin.Context) {
	// 处理获取赠送食物的原因的逻辑
}

// GetCashierShoppingCartInfo 处理获取收银购物车信息
// @Summary 获取收银购物车信息
// @Description 获取收银购物车信息
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {object} nil "购物车详情"
// @Failure 404 {object} nil "未找到"
// @Router /shoppingCart/info [get]
func (siw *Handler) GetCashierShoppingCartInfo(c *gin.Context) {
	// 处理获取收银购物车信息的逻辑
}

// DeleteCashierShoppingCartProduct 处理从购物车中删除产品
// @Summary 从购物车中删除产品
// @Description 从购物车中删除产品
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Success 204 {object} nil "无内容"
// @Failure 400 {object} nil "错误请求"
// @Router /shoppingCart/product/{productId} [delete]
func (siw *Handler) DeleteCashierShoppingCartProduct(c *gin.Context) {
	// 处理从购物车中删除产品的逻辑
}

// PostCashierShoppingCartProductCreate 处理在购物车中创建产品
// @Summary 在购物车中创建产品
// @Description 在购物车中创建产品
// @Tags cashier
// @Accept json
// @Produce json
// @Param product body nil true "产品详情"
// @Success 201 {object} nil "已创建"
// @Failure 400 {object} nil "错误请求"
// @Router /shoppingCart/product/create [post]
func (siw *Handler) PostCashierShoppingCartProductCreate(c *gin.Context) {
	// 处理在购物车中创建产品的逻辑
}

// PostCashierShoppingCartProductNumber 处理更新购物车中产品的数量
// @Summary 更新购物车中产品的数量
// @Description 更新购物车中产品的数量
// @Tags cashier
// @Accept json
// @Produce json
// @Param productId path string true "产品ID"
// @Param number body int true "新数量"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shoppingCart/product/number/{productId} [post]
func (siw *Handler) PostCashierShoppingCartProductNumber(c *gin.Context) {
	// 处理更新购物车中产品的数量的逻辑
}

// GetCashierShoppingCartSubBill 处理获取购物车的子账单
// @Summary 获取购物车的子账单
// @Description 获取购物车的子账单
// @Tags cashier
// @Accept json
// @Produce json
// @Success 200 {object} nil "子账单详情"
// @Failure 404 {object} nil "未找到"
// @Router /shoppingCart/subBill [get]
func (siw *Handler) GetCashierShoppingCartSubBill(c *gin.Context) {
}

// PostCashierVerifyAdvancedPassword 处理验证收银员的高级密码
// @Summary 验证收银员的高级密码
// @Description 验证收银员的高级密码
// @Tags cashier
// @Accept json
// @Produce json
// @Param password body string true "高级密码"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /verifyAdvancedPassword [post]
func (siw *Handler) PostCashierVerifyAdvancedPassword(c *gin.Context) {
	// Handler logic for PostCashierVerifyAdvancedPassword
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager) {

	wrapper := Handler{
		cashierService: cashier.NewCashierServiceImpl(dbm),
	}

	router.GET("/desk/list", wrapper.GetCashierDeskList)
	router.GET("/desk/regionAndType", wrapper.GetCashierDeskRegionAndType)
	router.POST("/login", wrapper.PostCashierLogin)
	router.GET("/login/captcha", wrapper.GetCashierLoginCaptcha)
	router.POST("/login/publicKey", wrapper.PostCashierLoginPublicKey)
	router.GET("/member/info", wrapper.GetCashierMemberInfo)
	router.GET("/member/search", wrapper.GetCashierMemberSearch)
	router.GET("/payment/typeList", wrapper.GetCashierPaymentTypeList)
	router.GET("/product/category", wrapper.GetCashierProductCategory)
	router.GET("/product/info", wrapper.GetCashierProductInfo)
	router.GET("/product/list", wrapper.GetCashierProductList)
	router.POST("/production/create", wrapper.PostCashierProductionCreate)
	router.POST("/production/return", wrapper.PostCashierProductionReturn)
	router.GET("/production/return/returnFoodReason", wrapper.GetCashierProductionReturnReturnFoodReason)
	router.GET("/reason/giftFood", wrapper.GetCashierReasonGiftFood)
	router.GET("/shoppingCart/info", wrapper.GetCashierShoppingCartInfo)
	router.DELETE("/shoppingCart/product", wrapper.DeleteCashierShoppingCartProduct)
	router.POST("/shoppingCart/product/create", wrapper.PostCashierShoppingCartProductCreate)
	router.POST("/shoppingCart/product/number", wrapper.PostCashierShoppingCartProductNumber)
	router.GET("/shoppingCart/subBill", wrapper.GetCashierShoppingCartSubBill)
	router.POST("/verifyAdvancedPassword", wrapper.PostCashierVerifyAdvancedPassword)
}
