package cashier

import (
	"errors"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CashierInstantHandler 收银点餐处理程序
type CashierInstantHandler struct {
	orderService service.IOrderSrv // 订单服务
}

// CreateInstantOrder 创建点餐订单
// @Summary 创建点餐订单
// @Description 创建点餐订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Success 200 {object} resp.CreateOrderResp
// @Router /cashier/instant/order/create [post]
func (h *CashierInstantHandler) CreateInstantOrder(c *gin.Context) {
	// 创建订单
	res, err := h.orderService.CreateInstantOrder(helper.GetCompanyUuid(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("创建订单失败"))
		return
	}
	helper.Success(c, res)
}

// RegisterInstantHandlers 注册收银订单路由
func RegisterInstantHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	// 创建收银产品处理程序
	wrapper := CashierInstantHandler{
		orderService: service.NewOrderSrv(dbm, cache), // 订单服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.POST("/instant/order/create", wrapper.CreateInstantOrder)
	}
}
