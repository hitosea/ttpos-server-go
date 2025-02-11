package cashier

import (
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CashierOrderHandler 收银点餐处理程序
type CashierOrderHandler struct {
	orderService service.IOrderSrv // 订单服务
}

// GetCashierOrderList 处理获取收银订单列表
// @Summary 获取收银订单列表
// @Description 获取收银订单列表
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {array} nil "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/list [get]
func (h *CashierOrderHandler) GetCashierOrderList(c *gin.Context) {
	// 处理获取收银订单列表的逻辑
}

// RegisterOrderHandlers 注册收银订单路由
func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager) {
	// 创建收银产品处理程序
	wrapper := CashierOrderHandler{
		orderService: service.NewOrderSrv(dbm), // 订单服务
	}

	router.GET("/order/list", wrapper.GetCashierOrderList)
}
