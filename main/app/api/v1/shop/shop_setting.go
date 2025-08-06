package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type SettingHandler struct {
	settingSrv setting.ISrv
}

// SaveSetting 保存常规设置
// @Summary 保存常规设置
// @Description 保存常规设置
// @Tags 商家端.常规设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdateStoreSetting true "更新门店设置"
// @Success 200 {object} dto.Response
// @Router /shop/setting [post]
func (h *SettingHandler) SaveSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateStoreSetting req.UpdateStoreSetting
	if err := c.ShouldBindJSON(&updateStoreSetting); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.settingSrv.EditStoreSetting(ctx, updateStoreSetting)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

func RegisterSettingHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &SettingHandler{
		settingSrv: settingSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/setting", wrapper.SaveSetting) // 保存设置
	}
}
