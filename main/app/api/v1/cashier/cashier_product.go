package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req/cashier_req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CashierProductHandler 收银产品处理程序
type CashierProductHandler struct {
	productService service.IProductSrv // 产品服务
}

// GetProductList 获取收银产品列表
// @Summary 获取收银产品列表
// @Description 获取收银产品列表
// @Tags 收银端
// @Accept json
// @Produce json
// @Param pageNo query int true "页码"
// @Param pageSize query int true "每页条数"
// @Success 200 {object} cashier_resp.ProductListWithPaginationResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/product/list [get]
func (h *CashierProductHandler) GetProductList(c *gin.Context) {
	// 绑定请求参数
	req := cashier_req.ProductListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}

	// 获取收银产品列表
	res, err := h.productService.GetProductList(
		1,
		req,
	)

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetProductCategoryList 获取收银产品类别列表
// @Summary 获取收银产品类别列表
// @Description 获取收银产品类别列表
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {object} cashier_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /cashier/product/category/list [get]
func (h *CashierProductHandler) GetProductCategoryList(c *gin.Context) {
	// 获取收银产品类别列表
	res, err := h.productService.GetProductCategoryList(1)

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// RegisterProductHandlers 注册收银产品路由
func RegisterProductHandlers(router gin.IRouter, dbm *database.DBManager) {
	// 创建收银产品处理程序
	wrapper := CashierProductHandler{
		productService: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
		),
	}

	// 注册收银产品路由
	router.GET("/product/list", wrapper.GetProductList)                  // 获取收银产品列表
	router.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取收银产品类别列表
}
