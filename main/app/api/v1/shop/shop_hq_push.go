package shop

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

// HqPushHandler 总部推送控制器
type HqPushHandler struct {
	hqPushSrv service.IHqPushSrv
}

// GetControlSetting 获取总部控制设置
// @Summary 获取总部控制设置
// @Description 获取总部对各字段的控制模式（统一控制/分开控制）
// @Tags 商家端.总部推送
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=resp.HqControlSettingResp}
// @Router /shop/product/hq_control_setting [get]
func (h *HqPushHandler) GetControlSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.hqPushSrv.GetControlSetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "")
}

// UpdateControlSetting 更新总部控制设置
// @Summary 更新总部控制设置
// @Description 更新总部对各字段的控制模式，从分开切换为统一时会触发强制推送
// @Tags 商家端.总部推送
// @Accept json
// @Produce json
// @Param data body req.HqControlSettingUpdateReq true "控制设置参数"
// @Success 200 {object} dto.Response
// @Router /shop/product/hq_control_setting [post]
func (h *HqPushHandler) UpdateControlSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateReq := req.HqControlSettingUpdateReq{}
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}
	if err := h.hqPushSrv.UpdateControlSetting(ctx, updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "")
}

// GetBatchPushStoreList 获取可推送门店列表
// @Summary 获取可推送门店列表
// @Description 获取总部下所有子店列表，用于选择批量推送目标
// @Tags 商家端.总部推送
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=resp.HqBatchPushStoreListResp}
// @Router /shop/product/hq_batch_push_store_list [get]
func (h *HqPushHandler) GetBatchPushStoreList(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.hqPushSrv.GetBatchPushStoreList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "")
}

// BatchPush 批量推送到子店
// @Summary 批量推送到子店
// @Description 批量推送商品/物品数据到指定子店，支持字段级推送和全量同步
// @Tags 商家端.总部推送
// @Accept json
// @Produce json
// @Param data body req.HqBatchPushReq true "推送参数"
// @Success 200 {object} dto.Response{data=resp.HqBatchPushResp}
// @Router /shop/product/hq_batch_push [post]
func (h *HqPushHandler) BatchPush(c *gin.Context) {
	ctx := helper.GetContext(c)
	pushReq := req.HqBatchPushReq{}
	if err := c.ShouldBindJSON(&pushReq); err != nil {
		helper.HandleValidationError(c, err, pushReq, nil)
		return
	}
	result, err := h.hqPushSrv.BatchPush(ctx, pushReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "")
}

// RegisterHqPushHandlers 注册总部推送相关路由
func RegisterHqPushHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := HqPushHandler{
		hqPushSrv: service.NewHqPushSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/hq_control_setting", wrapper.GetControlSetting)           // 获取总部控制设置
		privateApi.POST("/product/hq_control_setting", wrapper.UpdateControlSetting)       // 更新总部控制设置
		privateApi.GET("/product/hq_batch_push_store_list", wrapper.GetBatchPushStoreList) // 获取可推送门店列表
		privateApi.POST("/product/hq_batch_push", wrapper.BatchPush)                       // 批量推送到子店
	}
}
