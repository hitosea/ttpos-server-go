package kitchen

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
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

// GetUnprocessedCallList 未处理呼叫列表
// @Summary 未处理呼叫列表
// @Description 未处理呼叫列表
// @Tags 厨显端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.UnprocessedCallList}
// @Router /kitchen/call/unprocessed/list [get]
func (h *CallHandler) GetUnprocessedCallList(c *gin.Context) {
	var listReq req.UnprocessedCallListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.callSrv.GetUnprocessedCallList(helper.GetCompanyUuid(c), listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetUnprocessed 获取未处理消息数量
// @Summary 获取未处理消息数量
// @Description 获取未处理消息数量
// @Tags 厨显端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.UnprocessedResp}
// @Router /kitchen/call/unprocessed [get]
func (h *CallHandler) GetUnprocessed(c *gin.Context) {
	res, err := h.callSrv.GetUnprocessed(helper.GetCompanyUuid(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// Processed 处理呼叫
// @Summary 处理呼叫
// @Description 处理呼叫
// @Tags 厨显端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ProcessedCallReq true "处理呼叫参数"
// @Router /kitchen/call/processed [post]
func (h *CallHandler) Processed(c *gin.Context) {
	var processedReq req.ProcessedCallReq
	if err := c.ShouldBindJSON(&processedReq); err != nil {
		helper.HandleValidationError(c, err, processedReq, nil)
		return
	}
	err := h.callSrv.Processed(helper.GetCompanyUuid(c), processedReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "成功")
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
		privateApi.GET("/call/unprocessed/list", wrapper.GetUnprocessedCallList) // 分页获取呼叫列表
		privateApi.POST("/call/processed", wrapper.Processed)                    // 处理呼叫
		privateApi.GET("/call/unprocessed", wrapper.GetUnprocessed)              // 获取未处理消息数量
	}
}
