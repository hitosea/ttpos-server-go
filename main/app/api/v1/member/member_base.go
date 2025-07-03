package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/member_service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type BaseHandler struct {
	baseSrv member_service.IBaseSrv
}

// BaseInfo 获取基础信息
// @Summary 获取基础信息
// @Description 获取基础信息
// @Tags 会员端.基础信息
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=member_resp.MemberBaseInfoResp}
// @Router /member/base/info [get]
func (h *BaseHandler) BaseInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	baseInfoResp, err := h.baseSrv.GetBaseInfo(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, baseInfoResp)
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
	// 初始化处理器
	baseSrv := member_service.NewBaseSrv(dbm, cache)
	wrapper := &BaseHandler{
		baseSrv: baseSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.GET("/base/info", wrapper.BaseInfo)
	}
}
