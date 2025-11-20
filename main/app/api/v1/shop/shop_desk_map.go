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

// DeskMapHandler 桌台地图处理器
//
// 任务: story-admin-desktop-table-map Phase 2.6
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapHandler struct {
	deskMapSrv service.IDeskMapSrv
}

// GetAreaList 获取区域列表及地图配置状态
// @Summary 获取区域列表及地图配置状态
// @Description 获取所有区域及其桌台地图配置状态
// @Tags 商家端.桌台地图
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.DeskMapAreaListResp}
// @Router /shop/desk/map/areas [get]
func (h *DeskMapHandler) GetAreaList(c *gin.Context) {
	ctx := helper.GetContext(c)

	list, err := h.deskMapSrv.GetAreaListWithStatus(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, list, "获取成功")
}

// GetLayoutDetail 获取某区域布局详情
// @Summary 获取某区域布局详情
// @Description 获取指定区域的桌台列表和布局信息
// @Tags 商家端.桌台地图
// @Accept json
// @Produce json
// @Security JwtToken
// @Param area_uuid query uint64 true "区域UUID"
// @Success 200 {object} dto.Response{data=resp.DeskMapLayoutResp}
// @Router /shop/desk/map/layout_detail [get]
func (h *DeskMapHandler) GetLayoutDetail(c *gin.Context) {
	ctx := helper.GetContext(c)

	var detailReq req.DeskMapLayoutDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	detail, err := h.deskMapSrv.GetLayoutDetail(ctx, detailReq.AreaUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, detail, "获取成功")
}

// SaveLayout 保存区域布局
// @Summary 保存区域布局
// @Description 保存指定区域的桌台地图布局配置
// @Tags 商家端.桌台地图
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeskMapSaveLayoutReq true "布局数据"
// @Success 200 {object} dto.Response
// @Router /shop/desk/map/save_layout [post]
func (h *DeskMapHandler) SaveLayout(c *gin.Context) {
	ctx := helper.GetContext(c)

	var saveReq req.DeskMapSaveLayoutReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	if err := h.deskMapSrv.SaveLayout(ctx, saveReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "保存成功")
}

// RegisterDeskMapHandlers 注册桌台地图相关路由
func RegisterDeskMapHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	deskMapSrv := service.NewDeskMapSrv(dbm)

	handler := &DeskMapHandler{
		deskMapSrv: deskMapSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/desk/map/areas", handler.GetAreaList)
		privateApi.GET("/desk/map/layout_detail", handler.GetLayoutDetail)
		privateApi.POST("/desk/map/save_layout", handler.SaveLayout)
	}
}
