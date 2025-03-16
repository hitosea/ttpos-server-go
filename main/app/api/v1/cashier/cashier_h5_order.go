package cashier

import (
	"strconv"
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
	res, err := h.h5OrderSrv.GetH5OrderList(helper.GetCompanyUuid(c), h5OrderListReq)
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
// @Param order_uuid query int true "h5订单uuid"
// @Success 200 {object} dto.Response{data=resp.H5OrderDetailResp}
// @Router /cashier/h5_order/detail [get]
func (h *H5OrderHandler) GetH5OrderDetail(c *gin.Context) {
	orderUuid, err := strconv.ParseUint(c.Query("order_uuid"), 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err, "参数错误"))
	}
	res, err := h.h5OrderSrv.GetH5OrderDetail(helper.GetCompanyUuid(c), orderUuid)
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
// @Param order_uuid query int true "h5订单uuid"
// @Success 200 {object} dto.Response
// @Router /cashier/h5_order/reject [post]
func (h *H5OrderHandler) RejectH5Order(c *gin.Context) {
	orderUuid, err := strconv.ParseUint(c.Query("order_uuid"), 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err, "参数错误"))
	}

	err = h.h5OrderSrv.RejectH5Order(helper.GetContext(c), orderUuid)
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
// @Param order_uuid query int true "h5订单uuid"
// @Success 200 {object} dto.Response
// @Router /cashier/h5_order/accept [post]
func (h *H5OrderHandler) AcceptH5Order(c *gin.Context) {
	orderUuid, err := strconv.ParseUint(c.Query("order_uuid"), 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err, "参数错误"))
	}
	err = h.h5OrderSrv.AcceptH5Order(helper.GetContext(c), orderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	// 初始化处理器
	wrapper := H5OrderHandler{
		h5OrderSrv: service.NewH5OrderSrv(dbm),
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
