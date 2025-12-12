package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// TakeoutHandler 外卖处理程序
type TakeoutHandler struct {
	takeoutSrv service.ITakeoutSrv
	menuAppSrv application.ITakeoutMenuAppService
}

// NewTakeoutHandler 创建外卖 Handler
func NewTakeoutHandler(dbm *database.DBManager, cache cache.Cache, productSrv service.IProductSrv, translateSrv service.ITranslateSrv) *TakeoutHandler {
	menuAppSrv := application.NewTakeoutMenuAppService(dbm, cache)
	takeoutSrv := service.NewTakeoutSrv(dbm, cache, productSrv, translateSrv)
	return &TakeoutHandler{
		takeoutSrv: takeoutSrv,
		menuAppSrv: menuAppSrv,
	}
}

// GetGrabBindingLink 获取绑定链接
// @Summary 获取 Grab 绑定链接
// @Description 获取 Grab 平台的绑定链接，用户可跳转到 Grab 页面完成配置
// @Tags 商家端.Grab外卖
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.GrabBindingLinkResp} "成功"
// @Router /shop/grab/binding-link [post]
func (h *TakeoutHandler) GetGrabBindingLink(c *gin.Context) {

	ctx := helper.GetContext(c)

	// 调用应用服务
	result, err := h.menuAppSrv.GetBindingLink(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取绑定链接失败"))
		return
	}

	helper.Success(c, resp.GrabBindingLinkResp{
		BindingLink: result.BindingLink,
		ExpiresAt:   result.ExpiresAt,
	})
}

// CheckGrabBindingStatus 检查绑定状态
// @Summary 检查 Grab 绑定状态
// @Description 验证是否已经绑定 Grab 平台，前端可以定时查询
// @Tags 商家端.Grab外卖
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.GrabBindingStatusResp} "成功"
// @Router /shop/grab/binding-status [post]
func (h *TakeoutHandler) CheckGrabBindingStatus(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 调用应用服务
	result, err := h.menuAppSrv.CheckBindingStatus(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "检查绑定状态失败"))
		return
	}

	helper.Success(c, resp.GrabBindingStatusResp{
		IsBound:      result.IsBound,
		BoundAt:      result.BoundAt,
		MerchantID:   result.MerchantID,
		MerchantName: result.MerchantName,
	})
}

// GetGrabMenu 获取 Grab 菜单
// @Summary 获取 Grab 商品菜单
// @Description 从 Grab API 获取商品菜单数据
// @Tags 商家端.Grab外卖
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.GrabMenuResp} "成功"
// @Router /shop/grab/menu [post]
func (h *TakeoutHandler) GetGrabMenu(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 调用应用服务
	result, err := h.menuAppSrv.GetGrabMenu(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取 Grab 菜单失败"))
		return
	}

	helper.Success(c, resp.GrabMenuResp{
		Menu: result.Menu,
	})
}

// ImportMenu 导入 Grab 菜单
// @Summary 导入 Grab 菜单
// @Description 按规则导入 Grab 菜单（分类/商品/规格/属性/单位映射或创建）
// @Tags 商家端.Grab外卖
// @Accept json
// @Produce json
// @Security JwtToken
// @Param body body req.TakeoutMenuImportReq true "Grab 菜单 JSON（grab_models.go 结构）"
// @Success 200 {object} dto.Response{data=resp.GrabMenuImportResp} "成功"
// @Router /shop/grab/menu/import [post]
func (h *TakeoutHandler) ImportMenu(c *gin.Context) {
	var importReq req.TakeoutMenuImportReq
	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, nil)
		return
	}

	ctx := helper.GetContext(c)

	result, err := h.takeoutSrv.ImportMenu(ctx, "Grab", importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "导入 Grab 菜单失败"))
		return
	}

	helper.Success(c, resp.GrabMenuImportResp{
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		CreatedItems: result.CreatedItems,
		UpdatedItems: result.UpdatedItems,
		Failures:     result.Failures,
	})
}

// RegisterTakeoutHandlers 注册外卖路由
func RegisterTakeoutHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	productSrv := service.NewProductSrv(dbm, service.NewLocaleSrv(), settingSrv, cache, service.NewTranslateSrv(dbm, cache))
	translateSrv := service.NewTranslateSrv(dbm, cache)
	takeoutHandler := NewTakeoutHandler(dbm, cache, productSrv, translateSrv)

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/grab/binding-link", takeoutHandler.GetGrabBindingLink)
		privateApi.POST("/grab/binding-status", takeoutHandler.CheckGrabBindingStatus)
		privateApi.POST("/grab/menu", takeoutHandler.GetGrabMenu)
		privateApi.POST("/grab/menu/import", takeoutHandler.ImportMenu)
	}
}
