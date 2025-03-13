package tablet

import (
	"go.uber.org/zap"
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

// DeskHandler 桌台相关控制器
type DeskHandler struct {
	deskSrv  service.IDeskSrv
	orderSrv service.IOrderSrv
}

// GetDeskList 获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.TabletDeskList}
// @Router /tablet/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	ctx := helper.GetContext(c)
	list, err := h.deskSrv.GetTabletDeskList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, list)
}

// BindDesk 绑定/换绑桌台
// @Summary 绑定/换绑桌台
// @Description 绑定/换绑桌台，调用此接口之后的所有接口，都需要传递x-desk-uuid请求头
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BindDeskReq true "绑定/换绑桌台请求参数"
// @Success 200 {object} dto.Response
// @Router /tablet/desk/bind [post]
func (h *DeskHandler) BindDesk(c *gin.Context) {
	var bindDeskReq req.BindDeskReq
	if err := c.ShouldBindJSON(&bindDeskReq); err != nil {
		helper.HandleValidationError(c, err, bindDeskReq, req.LoginRequestMessage)
		return
	}
	err := h.deskSrv.BindDesk(helper.GetContext(c), bindDeskReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "绑定桌台成功")
}

// CreateDeskOrder 处理创建开台
// @Summary 开台
// @Description 开台
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "开台参数"
// @Success 200 {object} resp.CreateDeskOrderResp "开台成功"
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/open [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {

	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.deskSrv.CreateDeskOrder(ctx, params)
	// 处理错误
	if err != nil {
		ctx.Log().Error("创建桌台订单失败", zap.Error(err))
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetDeskInfo 处理获取收银台列表
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskInfoReq true "详情参数"
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	var infoReq req.DeskInfoReq
	if err := c.ShouldBindQuery(&infoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 获取收银产品列表
	res, err := h.deskSrv.GetDeskInfo(companyUuid, infoReq.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv)

	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv)

	wrapper := &DeskHandler{
		deskSrv:  deskSrv,
		orderSrv: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/desk/list", wrapper.GetDeskList) // 获取可绑定的桌台
		privateApi.POST("/desk/bind", wrapper.BindDesk)   // 绑定桌台

		privateApi.POST("/desk/open", wrapper.CreateDeskOrder) // 创建桌台订单(开桌)
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)      // 获取桌台详情
		privateApi.GET("/desk/place_order", nil)               // todo 桌台下单(送厨)
		privateApi.GET("/desk/get_ordered_product", nil)       // todo 桌台已下单商品
	}
}
