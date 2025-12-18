package kiosk

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

// CallHandler 呼叫相关控制器
type CallHandler struct {
	callSrv service.ICallSrv
}

// Call 发起呼叫
// @Summary 发起呼叫
// @Description 发起呼叫服务员
// @Tags 自助点餐机.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CallReq true "呼叫请求"
// @Success 200 {object} dto.Response
// @Router /kiosk/call [post]
func (h *CallHandler) Call(c *gin.Context) {
	var callReq req.CallReq
	if err := c.ShouldBindJSON(&callReq); err != nil {
		helper.HandleValidationError(c, err, callReq, nil)
		return
	}
	err := h.callSrv.Call(helper.GetContext(c), callReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "已呼叫服务员，请稍等")
}

func RegisterCallHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	wrapper := CallHandler{
		callSrv: service.NewCallSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/call", wrapper.Call) // 呼叫
	}
}

