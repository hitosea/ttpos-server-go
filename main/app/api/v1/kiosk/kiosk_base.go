package kiosk

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础相关控制器
type BaseHandler struct {
	authSrv    service.IAuthSrv
	settingSrv setting.ISrv
	deviceSrv  service.IDeviceSrv
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 获取自助点餐机首页基本信息
// @Tags 自助点餐机.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.KioskBase}
// @Router /kiosk/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.KioskBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &BaseHandler{
		authSrv:    authSrv,
		settingSrv: settingSrv,
		deviceSrv:  deviceSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.GetBase) // 获取基本信息
	}
}
