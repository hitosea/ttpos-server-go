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

// PrinterHandler 打印相关控制器
type PrinterHandler struct {
	printerLogSrv printerService.IPrinterLogSrv
}

// GetPrinterList 获取打印列表-查询条件
// @Summary 获取打印列表-查询条件
// @Description 获取打印列表-查询条件
// @Tags 收银端.打印记录
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.PrinterBaseResp "打印数据列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/printer/base [get]
func (h *PrinterHandler) GetPrinterBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	resp, err := h.printerLogSrv.GetPrinterBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// GetPrinterList 获取打印列表
// @Summary 获取打印列表
// @Description 获取打印列表
// @Tags 收银端.打印记录
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.PrinterListReq true "列表参数"
// @Success 200 {object} resp.PrinterListPaginationResp "打印数据列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/printer/list [get]
func (h *PrinterHandler) GetPrinterList(c *gin.Context) {
	ctx := helper.GetContext(c)
	req := req.PrinterListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	resp, err := h.printerLogSrv.GetPrinterLogList(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// PrinterPrint 打印
// @Summary 打印
// @Description 打印
// @Tags 收银端.打印记录
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.PrinterPrintReq true "打印参数"
// @Success 200 {object} resp.PrinterData "打印数据"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/printer/print [post]
func (h *PrinterHandler) PrinterPrint(c *gin.Context) {
	ctx := helper.GetContext(c)
	req := req.PrinterPrintReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	resp, err := h.printerLogSrv.PrinterPrint(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp, "发送成功")
}

// PrinterReport 打印报告
// @Summary 打印报告
// @Description 打印报告
// @Tags 收银端.打印记录
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.PrinterReportReqs true "打印参数"
// @Success 200 {object} dto.Response
// @Failure 404 {object} nil "未找到"
// @Router /cashier/printer/report [post]
func (h *PrinterHandler) PrinterReport(c *gin.Context) {
	ctx := helper.GetContext(c)
	req := req.PrinterReportReqs{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	err := h.printerLogSrv.PrinterReport(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

func RegisterPrinterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	printerLogSrv := printerService.NewPrinterLogSrv(dbm, settingSrv)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &PrinterHandler{
		printerLogSrv: printerLogSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/printer/base", wrapper.GetPrinterBase)   // 获取打印数据
		privateApi.GET("/printer/list", wrapper.GetPrinterList)   // 获取打印数据
		privateApi.POST("/printer/print", wrapper.PrinterPrint)   // 打印
		privateApi.POST("/printer/report", wrapper.PrinterReport) // 打印报告
	}
}
