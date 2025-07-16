package member

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

// OrderHandler 认证鉴权控制器
type OrderHandler struct {
	orderSrv service.IOrderSrv
}

// CreateOrder 创建会员端订单
// @Summary 创建会员端订单
// @Description 创建会员端订单
// @Tags 会员端
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/create [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.CreateMemberOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("创建会员端订单", zap.Any("params", params))

	// 创建会员端订单
	res, checkRes, err := h.orderSrv.CreateMemberOrder(ctx, params)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("提交订单检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	// 初始化处理器
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, service.NewMustPlanSrv(dbm), service.NewPaymentMethodSrv(dbm, settingSrv), service.NewMemberSrv(dbm), service.NewCashBoxSrv(dbm))
	wrapper := &OrderHandler{
		orderSrv: orderSrv,
	}
	// 需要认证
	// privateApi := router.Group("")
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.POST("/order/create", wrapper.CreateOrder) // 创建订单
	}
}
