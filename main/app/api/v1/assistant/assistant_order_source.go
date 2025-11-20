package assistant

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// OrderSourceHandler 订单来源处理器
//
// 任务: story-order-source-nationality Phase 3.6
// 需求: R3, R4
//
// @version v2.10.0
type OrderSourceHandler struct {
	orderSourceSrv service.IOrderSourceSrv
}

// GetList 获取订单来源列表
// @Summary 获取订单来源列表
// @Description 获取所有订单来源配置（供终端选择）
// @Tags 点餐助手端.订单来源
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OrderSourceListResp}
// @Router /assistant/order_source/list [get]
func (h *OrderSourceHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	list, err := h.orderSourceSrv.GetList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, list, "获取成功")
}

// RegisterOrderSourceRoutes 注册订单来源路由
func RegisterOrderSourceRoutes(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	orderSourceSrv := service.NewOrderSourceSrv(dbm)
	handler := &OrderSourceHandler{
		orderSourceSrv: orderSourceSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order_source/list", handler.GetList)
	}
}
