package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CashierInstantHandler 收银点餐处理程序
type CashierInstantHandler struct {
	orderService service.IOrderSrv // 订单服务
}

// CreateOrder 创建订单
func (h *CashierInstantHandler) CreateOrder(c *gin.Context) {
	// 创建订单
	res, err := h.orderService.CreateOrder(1)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, res)
}

// RegisterCashierOrderHandlers 注册收银订单路由
func RegisterCashierOrderHandlers(router gin.IRouter, dbm *database.DBManager) {
	// 创建收银产品处理程序
	wrapper := CashierInstantHandler{
		orderService: service.NewOrderSrv(dbm), // 订单服务
	}

	router.POST("/instant/order/create", wrapper.CreateOrder)
}
