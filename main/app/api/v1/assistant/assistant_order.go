package assistant

import (
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderSrv service.IOrderSrv // 订单服务
	deskSrv  service.IDeskSrv  // 桌台服务
}

// IsCellClose 判断订单是否可关闭
// @Summary 判断订单是否可关闭
// @Description 判断订单是否可关闭
// @Tags 点餐助手端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderIsCellCloseReq true "详情参数"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/order/is_cell_close [get]
func (h *OrderHandler) IsCellClose(c *gin.Context) {
	ctx := helper.GetContext(c)
	//
	params := req.OrderIsCellCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	var err error
	var productList *resp.CartProductList
	if params.DeskUuid > 0 {
		_, productList, err = h.deskSrv.IsCellCloseDesk(ctx, params.DeskUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err)
			return
		}
	} else if params.SaleBillUuid > 0 {
		productList, err = h.deskSrv.IsCellCloseInstant(ctx, params.SaleBillUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err)
			return
		}
	} else {
		err = errors.New("参数错误")
	}
	if err != nil {
		if strings.Contains(err.Error(), "订单已结账") {
			err = errors.New("当前订单已被部分支付，无法整单取消")
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// GetProductPackageDetail 获取商品选购详情
// @Summary 获取商品选购详情
// @Description 获取商品选购详情
// @Tags 点餐助手端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.GetProductPackageDetailReq true "商品选购详情参数"
// @Success 200 {object} dto.Response{data=resp.ProductPackageDetailRes}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/order/product/package/detail [get]
func (h *OrderHandler) GetProductPackageDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetProductPackageDetailReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取商品选购详情", zap.Any("params", params))
	productPackage, err := h.orderSrv.GetProductPackageDetail(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, productPackage)
}

// CheckAuthorization 检查授权
// @Summary 检查授权
// @Description 检查当前员工是否有权限进行敏感操作（折扣/退款）
// @Tags 点餐助手端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckAuthorizationReq true "检查授权参数"
// @Success 200 {object} dto.Response{data=resp.CheckAuthorizationResp}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/order/check_authorization [post]
func (h *OrderHandler) CheckAuthorization(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.CheckAuthorizationReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 调用 Service 检查授权
	hasPermission, err := h.orderSrv.CheckAuthorization(ctx, req.OperationType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, resp.CheckAuthorizationResp{
		HasPermission: hasPermission,
	})
}

// VerifyPassword 密码验证
// @Summary 密码验证
// @Description 验证授权员工账号和密码（用于敏感操作）
// @Tags 点餐助手端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordForSensitiveOperationReq true "密码验证参数"
// @Success 200 {object} dto.Response{data=resp.VerifyPasswordResp}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/order/verify_password [post]
func (h *OrderHandler) VerifyPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.VerifyPasswordForSensitiveOperationReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 验证密码（折扣操作）
	verified, err := h.orderSrv.VerifyPassword(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, resp.VerifyPasswordResp{
		Verified: verified,
	})
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
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	// 初始化处理器
	wrapper := OrderHandler{
		orderSrv: orderSrv,
		deskSrv:  service.NewDeskSrv(dbm, service.NewLocaleSrv(), orderSrv, settingSrv, deviceSrv, mustPlanSrv),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order/is_cell_close", wrapper.IsCellClose)                      // 判断订单是否可关闭
		privateApi.GET("/order/product/package/detail", wrapper.GetProductPackageDetail) // 获取商品选购详情
		privateApi.POST("/order/check_authorization", wrapper.CheckAuthorization)        // 检查授权
		privateApi.POST("/order/verify_password", wrapper.VerifyPassword)                // 密码验证
	}
}
