package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MemberOrderManageHandler 外送订单管理相关控制器
type MemberOrderManageHandler struct {
	memberOrderSrv service.IMemberOrderSrv
}

// GetMemberOrderList 获取外送订单管理页面，订单列表
// @Summary 获取外送订单管理页面，订单列表
// @Description 获取外送订单管理页面，订单列表
// @Tags 收银端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query string true "状态: unaccept-待接单, accept-备餐中, undelivery-待配送, delivery-配送中, completed-已完成, cancel-已取消"
// @Param order_no query string true "订单编号"
// @Param serial_no query string true "订单序号"
// @Param date_range query int true "日期类型 -1=全都、 0=今天、 1=昨天、 2=本周"
// @Param time_type query int true "时间类型  1=下单时间、 2=支付时间"
// @Param query_start_time query int true "查询开始时间戳"
// @Param query_end_time query int true "查询结束时间戳"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderManageListResp}
// @Router /cashier/member_order_manage/list [get]
func (h *MemberOrderManageHandler) GetMemberOrderList(c *gin.Context) {
	var memberOrderListReq req.MemberOrderManageListReq
	if err := c.ShouldBindQuery(&memberOrderListReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderListReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberOrderManageList(helper.GetContext(c), memberOrderListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetMemberOrderDetail 获取外送订单管理页面，订单详情
// @Summary 获取外送订单管理页面，订单详情
// @Description 获取外送订单管理页面，订单详情
// @Tags 收银端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "订单UUID"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderManageDetailResp}
// @Router /cashier/member_order_manage/detail [get]
func (h *MemberOrderManageHandler) GetMemberOrderDetail(c *gin.Context) {
	var memberOrderDetailReq req.GetMemberOrderManageDetailReq
	if err := c.ShouldBindQuery(&memberOrderDetailReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderDetailReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberOrderManageDetail(helper.GetContext(c), memberOrderDetailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetMemberOrderReturnInfo 获取外送订单退款弹窗信息
// @Summary 获取外送订单退款弹窗信息
// @Description 获取外送订单退款弹窗信息
// @Tags 收银端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query int true "会员端销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.OrderReturnInfoResp}
// @Router /cashier/member_order_manage/return_info [get]
func (h *MemberOrderManageHandler) GetMemberOrderReturnInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	reqParam := member_req.MemberOrderReturnInfoReq{}
	if err := c.ShouldBindQuery(&reqParam); err != nil {
		helper.HandleValidationError(c, err, reqParam, nil)
		return
	}
	//
	res, err := h.memberOrderSrv.GetMemberOrderReturnInfo(ctx, reqParam)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// MemberOrderReturn 外送订单退款/部分退款
// @Summary 外送订单退款/部分退款
// @Description 外送订单退款/部分退款
// @Tags 收银端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.OrderReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/member_order_manage/return [post]
func (h *MemberOrderManageHandler) MemberOrderReturn(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	reqParam := req.OrderReturnReq{}
	if err := c.ShouldBindJSON(&reqParam); err != nil {
		helper.HandleValidationError(c, err, reqParam, nil)
		return
	}
	//
	err, codeFail := h.memberOrderSrv.MemberOrderReturn(ctx, reqParam)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// MemberOrderReReturn 外送订单重新退款
// @Summary 外送订单重新退款
// @Description 外送订单重新退款
// @Tags 收银端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.MemberOrderReReturnReq true "详情参数"
// @Success 200 {object} nil "重新退款成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/member_order_manage/re_return [post]
func (h *MemberOrderManageHandler) MemberOrderReReturn(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	reqParam := req.OrderReReturnReq{}
	if err := c.ShouldBindJSON(&reqParam); err != nil {
		helper.HandleValidationError(c, err, reqParam, nil)
		return
	}
	//
	err, codeFail := h.memberOrderSrv.MemberOrderReReturn(ctx, reqParam)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

func RegisterMemberOrderManageHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	wrapper := MemberOrderManageHandler{
		memberOrderSrv: orderSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/member_order_manage/list", wrapper.GetMemberOrderList)              // 获取外送订单管理页面，订单列表
		privateApi.GET("/member_order_manage/detail", wrapper.GetMemberOrderDetail)          // 获取外送订单管理详情
		privateApi.GET("/member_order_manage/return_info", wrapper.GetMemberOrderReturnInfo) // 获取外送订单退款弹窗信息
		privateApi.POST("/member_order_manage/return", wrapper.MemberOrderReturn)            // 外送订单退款/部分退款
		privateApi.POST("/member_order_manage/re_return", wrapper.MemberOrderReReturn)       // 外送订单重新退款
	}
}
