package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type ProductHandler struct {
	productSrv service.IProductSrv
}

// GetProductCategoryList 获取会员端产品类别列表
// @Summary 获取会员端产品类别列表
// @Description 获取会员端产品类别列表
// @Tags 会员端.产品列表
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/product/category/list [get]
func (h *ProductHandler) GetProductCategoryList(c *gin.Context) {
	// 获取会员端产品类别列表
	res, err := h.productSrv.GetProductCategoryList(helper.GetCompanyUuid(c))

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetProductList 获取会员端产品列表
// @Summary 获取会员端产品列表
// @Description 获取会员端产品列表，支持区分外送和堂食订单类型
// @Tags 会员端.产品列表
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页条数"
// @Param order_type query int false "订单类型：0-外送（默认，价格应用外送折扣率），1-堂食/到店自取（价格与收银机相同）"
// @Success 200 {object} product_resp.ProductListWithPaginationResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/product/list [get]
func (h *ProductHandler) GetProductList(c *gin.Context) {
	// 绑定请求参数
	productListReq := req.ProductListReq{}
	if err := c.ShouldBindQuery(&productListReq); err != nil {
		helper.HandleValidationError(c, err, productListReq, dto.PageReqMessage)
		return
	}

	// 获取会员端产品列表
	productListReq.IsMember = true
	res, err := h.productSrv.GetProductList(helper.GetContext(c), productListReq)

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetProductCategoryList 获取会员端商家推荐商品列表
// @Summary 获取会员端商家推荐商品列表
// @Description 获取会员端商家推荐商品列表
// @Tags 会员端.产品列表
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/product/recommend/list [get]
func (h *ProductHandler) GetProductRecommendList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 获取会员端产品类别列表
	res, err := h.productSrv.GetProductRecommendList(ctx, req.ProductRecommendListReq{})

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// SearchProducts 搜索商品
// @Summary 搜索商品
// @Description 根据关键词搜索商品，不分页返回
// @Tags 会员端.产品列表
// @Accept json
// @Produce json
// @Security JwtToken
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} dto.Response{data=[]product_resp.Product} "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/product/search [get]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	// 绑定请求参数
	searchReq := req.ProductSearchReq{}
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		helper.HandleValidationError(c, err, searchReq, nil)
		return
	}

	// 设置为会员端查询
	searchReq.IsMember = true

	// 搜索商品
	res, err := h.productSrv.SearchProducts(helper.GetContext(c), searchReq)

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

func RegisterProductHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	translateSrv := service.NewTranslateSrv(dbm, cache)
	// 初始化处理器
	productSrv := service.NewProductSrv(dbm, service.NewLocaleSrv(), settingSrv, cache, translateSrv)
	wrapper := &ProductHandler{
		productSrv: productSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.GET("/product/list", wrapper.GetProductList)                    // 获取商品列表
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList)   // 获取商品类别列表
		privateApi.GET("/product/recommend/list", wrapper.GetProductRecommendList) // 获取商品推荐列表
		privateApi.GET("/product/search", wrapper.SearchProducts)                  // 搜索商品
	}
}
