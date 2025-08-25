package cashier

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

// ProductHandler 收银产品处理程序
type ProductHandler struct {
	productSrv service.IProductSrv // 产品服务
}

// GetProductList 获取收银产品列表
// @Summary 获取收银产品列表
// @Description 获取收银产品列表
// @Tags 收银端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页条数"
// @Success 200 {object} product_resp.ProductListWithPaginationResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/product/list [get]
func (h *ProductHandler) GetProductList(c *gin.Context) {
	// 绑定请求参数
	req := req.ProductListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}

	// 获取收银产品列表
	res, err := h.productSrv.GetProductList(helper.GetContext(c), req)

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetProductCategoryList 获取收银产品类别列表
// @Summary 获取收银产品类别列表
// @Description 获取收银产品类别列表
// @Tags 收银端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/product/category/list [get]
func (h *ProductHandler) GetProductCategoryList(c *gin.Context) {
	// 获取收银产品类别列表
	res, err := h.productSrv.GetProductCategoryList(helper.GetCompanyUuid(c))

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// RegisterProductHandlers 注册收银产品路由
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

	// 创建收银产品处理程序
	wrapper := ProductHandler{
		productSrv: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
			settingSrv,
			cache,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/list", wrapper.GetProductList)                  // 获取收银产品列表
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取收银产品类别列表
	}
}
