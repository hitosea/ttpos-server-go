package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/cashier_req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// ProductHandler 收银产品处理程序
type ProductHandler struct {
	productService service.IProductSrv // 产品服务
}

// GetProductList 获取收银产品列表
func (h *ProductHandler) GetProductList(c *gin.Context) {
	dbId := helper.GetCompanyId(c)
	req := cashier_req.ProductListReq{}
	res, err := h.productService.GetProductList(dbId, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, res)
}

// RegisterProductHandlers 注册收银产品路由
func RegisterProductHandlers(router gin.IRouter, dbm *database.DBManager) {
	wrapper := ProductHandler{
		productService: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
		),
	}

	router.GET("/product/list", wrapper.GetProductList) // 获取收银产品列表
}
