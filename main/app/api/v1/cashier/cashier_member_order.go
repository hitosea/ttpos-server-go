package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MemberOrderHandler 接单相关控制器
type MemberOrderHandler struct {
	memberOrderSrv service.IMemberOrderSrv
}

// GetMemberOrderList 会员端订单列表
// @Summary 会员端订单列表
// @Description 会员端订单列表
// @Tags 收银端.外送单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query string true "状态:"unaccept" 待接单, "accept" 备餐中, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderListResp}
// @Router /cashier/member_order/list [get]
func (h *MemberOrderHandler) GetMemberOrderList(c *gin.Context) {
	var memberOrderListReq req.MemberOrderListReq
	if err := c.ShouldBindQuery(&memberOrderListReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderListReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberOrderList(helper.GetContext(c), memberOrderListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetMemberOrderDetail 会员端订单详情
// @Summary 会员端订单详情
// @Description 会员端订单详情
// @Tags 收银端.外送单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query int true "会员端销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderDetailResp}
// @Router /cashier/member_order/detail [get]
func (h *MemberOrderHandler) GetMemberOrderDetail(c *gin.Context) {
	var memberOrderDetailReq req.GetMemberOrderDetailReq
	if err := c.ShouldBindQuery(&memberOrderDetailReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderDetailReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberOrderDetail(helper.GetContext(c), memberOrderDetailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

func RegisterMemberOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))
	// 初始化处理器
	wrapper := MemberOrderHandler{
		memberOrderSrv: orderSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/member_order/list", wrapper.GetMemberOrderList)     // 获取会员端订单列表
		privateApi.GET("/member_order/detail", wrapper.GetMemberOrderDetail) // 获取会员端订单详情
	}
}
