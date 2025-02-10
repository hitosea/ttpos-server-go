package cashier

import (
	"errors"
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

// CreateOrder 创建点餐订单
// @Summary 创建点餐订单
// @Description 创建点餐订单
// @Tags 收银端.点餐订单
// @Accept json
// @Produce json
// @Success 200 {object} cashier_resp.CreateOrderResp
// @Failure 400 {object} helper.Error
// @Failure 500 {object} helper.Error
// @Router /cashier/instant/order/create [post]
func (h *CashierInstantHandler) CreateOrder(c *gin.Context) {
	// 创建订单
	res, err := h.orderService.CreateOrder(1, constant.OrderSourceInstant)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("创建订单失败"))
		return
	}
	helper.Success(c, res)
}

// RegisterInstantHandlers 注册收银订单路由
func RegisterInstantHandlers(router gin.IRouter, dbm *database.DBManager) {
	// 创建收银产品处理程序
	wrapper := CashierInstantHandler{
		orderService: service.NewOrderSrv(dbm), // 订单服务
	}

	router.POST("/instant/order/create", wrapper.CreateOrder)
}
