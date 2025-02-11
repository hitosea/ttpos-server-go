package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
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
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data query req.GetOrderListReq true "列表参数"
// @Success 200 {object} resp.CashierOrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/list [get]
func (h *CashierOrderHandler) GetCashierOrderList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	req := req.GetOrderListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.orderService.GetCashierOrderList(companyUuid, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterOrderHandlers 注册收银订单路由
func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	// 初始化处理器
	wrapper := CashierOrderHandler{
		orderService: service.NewOrderSrv(dbm, cache), // 订单服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/order/list", wrapper.GetCashierOrderList)
	}
}
