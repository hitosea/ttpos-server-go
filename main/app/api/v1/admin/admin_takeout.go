package admin

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/types/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// TakeoutHandler Admin 端外卖处理程序
type TakeoutHandler struct {
	dbm           *database.DBManager
	takeoutAppSrv application.ITakeoutAppService
}

// NewTakeoutHandler 创建外卖 Handler
func NewTakeoutHandler(dbm *database.DBManager) *TakeoutHandler {
	return &TakeoutHandler{
		dbm:           dbm,
		takeoutAppSrv: application.NewTakeoutAppService(dbm),
	}
}

// GetTakeoutImportLogsReq 获取外卖导入日志列表请求
type GetTakeoutImportLogsReq struct {
	CompanyUuid uint64 `form:"company_uuid" json:"company_uuid" binding:"required"`          // 商户 UUID
	Platform    string `form:"platform" json:"platform"`                                     // 外卖平台(grab/lineman等)，为空查询所有
	ImportType  int8   `form:"import_type" json:"import_type"`                               // 导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)，0 查询所有
	Status      int8   `form:"status" json:"status" binding:"omitempty,oneof=-1 0 1 2"`      // 导入状态(0-进行中 1-成功 2-失败)，-1 查询所有
	Page        int    `form:"page" json:"page" binding:"omitempty,min=1"`                   // 页码，默认 1
	PageSize    int    `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量，默认 20，最大 100
}

// GetTakeoutImportLogs 获取外卖导入日志列表
// @Summary 获取外卖导入日志列表
// @Description 分页查询外卖导入日志，支持按平台、类型、状态筛选。平台管理员可查看所有商户，商户管理员只能查看自己的。
// @Tags 平台端.外卖管理
// @Accept json
// @Produce json
// @Security InternalToken
// @Param platform query string false "外卖平台(grab/lineman等)，为空查询所有"
// @Param import_type query int false "导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)，0 查询所有"
// @Param status query int false "导入状态(0-进行中 1-成功 2-失败)，-1 查询所有"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页数量，默认 20，最大 100"
// @Param company_uuid query string false "商户 UUID(平台管理员可指定，商户管理员只能查自己的)"
// @Success 200 {object} dto.Response
// @Router /admin/takeout/logs [get]
func (h *TakeoutHandler) GetTakeoutImportLogs(c *gin.Context) {
	// 解析请求参数
	var req GetTakeoutImportLogsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	// 转换为 Application Service 的请求参数
	var importType *int8
	if req.ImportType != 0 {
		importType = &req.ImportType
	}
	var status *int8
	if req.Status != -1 {
		status = &req.Status
	}

	reqData := request.GetImportLogsRequest{
		Platform:   req.Platform,
		ImportType: importType,
		Status:     status,
		PageNo:     req.Page,
		PageSize:   req.PageSize,
	}

	ctx := helper.GetContext(c)
	db := h.dbm.GetDB(req.CompanyUuid)
	if db == nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("门店不存在"))
		return
	}
	companyInfo, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(req.CompanyUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "门店不存在"))
		return
	}
	ctx.SetBasicInfo(db, companyInfo)

	// 调用 Application Service
	logsResp, err := h.takeoutAppSrv.GetImportLogs(ctx, reqData)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取外卖导入日志失败"))
		return
	}

	helper.Success(c, logsResp)
}

// RegisterTakeoutHandlers 注册 Admin 端外卖路由
func RegisterTakeoutHandlers(router gin.IRouter, dbm *database.DBManager) {
	handler := NewTakeoutHandler(dbm)

	// 使用 Internal 中间件保护
	router.GET("/takeout/logs", middleware.Internal(), handler.GetTakeoutImportLogs)
}
