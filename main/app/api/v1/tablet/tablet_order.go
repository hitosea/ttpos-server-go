package tablet

import (
	"strconv"
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

type OrderHandler struct {
	orderSrv service.IOrderSrv // 订单服务
}

// GetMustPlanList 必点方案列表
// @Summary 必点方案列表
// @Description 必点方案列表
// @Tags 平板端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param X-DESK-UUID header string true "桌台Uuid"
// @param sale_bill_uuid query int true "销售账单Uuid"
// @Success 200 {object} dto.Response{data=resp.ProductMustPlanList}
// @Failure 404 {object} nil "未找到"
// @Router /tablet/order/must_plan_list [get]
func (h *OrderHandler) GetMustPlanList(c *gin.Context) {
	saleBillUuid, err := strconv.ParseUint(c.Query("sale_bill_uuid"), 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("参数错误"))
	}
	ctx := helper.GetContext(c)
	// 创建桌台订单
	res, err := h.orderSrv.GetMustPlanList(ctx, saleBillUuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv)

	// 初始化处理器
	wrapper := OrderHandler{
		orderSrv: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order/must_plan_list", wrapper.GetMustPlanList)
	}
}
