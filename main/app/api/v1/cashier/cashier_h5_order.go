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
	"go.uber.org/zap"
)

// H5OrderHandler 接单相关控制器
type H5OrderHandler struct {
	h5OrderSrv service.IH5OrderSrv
}

// GetH5OrderList h5订单列表
// @Summary h5订单列表
// @Description h5订单列表
// @Tags 收银端.接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query int true "状态:0-待接单；1-已处理"
// @Param desk_region_uuid query int false "桌台区域uuid"
// @Success 200 {object} dto.Response{data=resp.H5OrderList}
// @Router /cashier/h5_order/list [get]
func (h *H5OrderHandler) GetH5OrderList(c *gin.Context) {
	var h5OrderListReq req.H5OrderListReq
	if err := c.ShouldBindQuery(&h5OrderListReq); err != nil {
		helper.HandleValidationError(c, err, h5OrderListReq, nil)
		return
	}
	res, err := h.h5OrderSrv.GetH5OrderList(helper.GetContext(c), h5OrderListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetH5OrderDetail h5订单详情
// @Summary h5订单详情
// @Description h5订单详情
// @Tags 收银端.接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param h5_order_uuid query int true "h5订单uuid"
// @Success 200 {object} dto.Response{data=resp.H5OrderDetailResp}
// @Router /cashier/h5_order/detail [get]
func (h *H5OrderHandler) GetH5OrderDetail(c *gin.Context) {
	var h5OrderDetailReq req.H5OrderDetailReq
	if err := c.ShouldBindQuery(&h5OrderDetailReq); err != nil {
		helper.HandleValidationError(c, err, h5OrderDetailReq, nil)
		return
	}
	res, err := h.h5OrderSrv.GetH5OrderDetail(helper.GetContext(c), h5OrderDetailReq.H5OrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// RejectH5Order 拒单
// @Summary 拒单
// @Description 拒单
// @Tags 收银端.接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RejectH5OrderReq true "拒单参数"
// @Success 200 {object} dto.Response
// @Router /cashier/h5_order/reject [post]
func (h *H5OrderHandler) RejectH5Order(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.RejectH5OrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	tracker := helper.StartTrack(ctx, constant.ActionCashierRejectH5Order).WithBill(0, params.H5OrderUuid).WithPath(c.Request.URL.Path)
	err := h.h5OrderSrv.RejectH5Order(ctx, params.H5OrderUuid)
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// AcceptH5Order 接单
// @Summary 接单
// @Description 接单
// @Tags 收银端.接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AcceptH5OrderReq true "接单参数"
// @Success 200 {object} dto.Response{data=resp.OrderCheckRes}
// @Router /cashier/h5_order/accept [post]
func (h *H5OrderHandler) AcceptH5Order(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.AcceptH5OrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	tracker := helper.StartTrack(ctx, constant.ActionCashierAcceptH5Order).WithBill(0, params.H5OrderUuid).WithPath(c.Request.URL.Path)
	checkRes, err := h.h5OrderSrv.AcceptH5Order(ctx, params.H5OrderUuid)
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code, constant.WithIsH5()))
		return
	}
	helper.Success(c, gin.H{})
}

func RegisterH5OrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))
	// 初始化处理器
	wrapper := H5OrderHandler{
		h5OrderSrv: service.NewH5OrderSrv(dbm, orderSrv),
	}
	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/h5_order/list", wrapper.GetH5OrderList)     // 获取h5订单列表
		privateApi.GET("/h5_order/detail", wrapper.GetH5OrderDetail) // 获取h5订单详情
		privateApi.POST("/h5_order/reject", wrapper.RejectH5Order)   // 拒单
		privateApi.POST("/h5_order/accept", wrapper.AcceptH5Order)   // 接单
	}
}
