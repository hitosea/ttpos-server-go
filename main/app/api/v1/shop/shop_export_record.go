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

type exportRecordHandler struct {
	exportRecordSrv service.IExportRecordSrv
}

// GetExportRecordList 获取导出记录列表
// @Summary 获取导出记录列表
// @Description 获取导出记录列表，支持按类型筛选
// @Tags 商家端.导出记录
// @Accept json
// @Produce json
// @Security JwtToken
// @Param export_type query string false "导出类型（可选）"
// @Param page_no query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} dto.Response{data=resp.ExportRecordListPaginationResp} "导出记录列表"
// @Router /shop/export_record/list [get]
func (h *exportRecordHandler) GetExportRecordList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.ExportRecordListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	result, err := h.exportRecordSrv.GetExportRecordList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// DeleteExportRecords 批量删除导出记录
// @Summary 批量删除导出记录
// @Description 批量删除导出记录（软删除）
// @Tags 商家端.导出记录
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ExportRecordDeleteReq true "删除参数"
// @Success 200 {object} dto.Response "删除成功"
// @Router /shop/export_record/delete [delete]
func (h *exportRecordHandler) DeleteExportRecords(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.ExportRecordDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.exportRecordSrv.DeleteExportRecords(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// RegisterExportRecordHandlers 注册导出记录相关路由
func RegisterExportRecordHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	exportRecordSrv := service.NewExportRecordSrv(dbm)

	handler := &exportRecordHandler{
		exportRecordSrv: exportRecordSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/export_record/list", handler.GetExportRecordList)
		privateApi.DELETE("/export_record/delete", handler.DeleteExportRecords)
	}
}
