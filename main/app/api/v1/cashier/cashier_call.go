package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	printerService "ttpos-server-go/app/modules/printer/service"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CallHandler 呼叫相关控制器
type CallHandler struct {
	callSrv       service.ICallSrv
	printerLogSrv printerService.IPrinterLogSrv
}

// GetAbnormalPrintList 异常打印列表
// @Summary 异常打印列表
// @Description 异常打印列表
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.AbnormalPrintList}
// @Router /cashier/call/abnormal_print/list [get]
func (h *CallHandler) GetAbnormalPrintList(c *gin.Context) {
	var listReq req.AbnormalPrintListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.callSrv.GetAbnormalPrintList(helper.GetCompanyUuid(c), listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetUnprocessedCallList 未处理呼叫列表
// @Summary 未处理呼叫列表
// @Description 未处理呼叫列表
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.UnprocessedCallList}
// @Router /cashier/call/unprocessed/list [get]
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
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.UnprocessedResp}
// @Router /cashier/call/unprocessed [get]
func (h *CallHandler) GetUnprocessed(c *gin.Context) {
	res, err := h.callSrv.GetUnprocessed(helper.GetCompanyUuid(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetUnprocessedNotice 获取未处理消息
// @Summary 获取未处理消息
// @Description 获取未处理消息
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.UnprocessedListResp}
// @Router /cashier/call/unprocessed_notice [get]
func (h *CallHandler) GetUnprocessedNotice(c *gin.Context) {
	res, err := h.callSrv.GetUnprocessedNotice(helper.GetContext(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// Processed 处理呼叫
// @Summary 处理呼叫
// @Description 处理呼叫
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ProcessedCallReq true "处理呼叫参数"
// @Router /cashier/call/processed [post]
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

// DeletePrint 删除打印
// @Summary 删除打印
// @Description 删除打印
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.PrinterLogReq true "删除打印参数"
// @Router /cashier/call/print [delete]
func (h *CallHandler) DeletePrint(c *gin.Context) {
	var printerLogReq req.PrinterLogReq
	if err := c.ShouldBindJSON(&printerLogReq); err != nil {
		helper.HandleValidationError(c, err, printerLogReq, nil)
		return
	}
	err := h.callSrv.DeletePrint(helper.GetCompanyUuid(c), printerLogReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "成功")
}

// Reprint 重新打印
// @Summary 重新打印
// @Description 重新打印
// @Tags 收银端.呼叫
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.PrinterPrintReq true "打印参数"
// @Success 200 {object} resp.PrinterData "打印数据"
// @Router /cashier/call/reprint [post]
func (h *CallHandler) Reprint(c *gin.Context) {
	req := req.PrinterPrintReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	ctx := helper.GetContext(c)
	resp, err := h.printerLogSrv.PrinterPrint(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp, "发送成功")
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
	printerLogSrv := printerService.NewPrinterLogSrv(dbm, settingSrv)

	// 初始化处理器
	wrapper := CallHandler{
		callSrv:       service.NewCallSrv(dbm),
		printerLogSrv: printerLogSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/call/unprocessed/list", wrapper.GetUnprocessedCallList)  // 分页获取呼叫列表
		privateApi.POST("/call/processed", wrapper.Processed)                     // 处理呼叫
		privateApi.GET("/call/abnormal_print/list", wrapper.GetAbnormalPrintList) // 异常打印列表
		privateApi.DELETE("/call/print", wrapper.DeletePrint)                     // 删除异常打印
		privateApi.POST("/call/reprint", wrapper.Reprint)                         // 重新打印
		privateApi.GET("/call/unprocessed", wrapper.GetUnprocessed)               // 获取未处理消息数量
		privateApi.GET("/call/unprocessed_notice", wrapper.GetUnprocessedNotice)  // 获取未处理消息
	}
}
