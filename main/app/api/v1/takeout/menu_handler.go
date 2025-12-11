package takeout

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// Handler 外卖菜单 Handler
type Handler struct {
	menuAppSrv application.ITakeoutMenuAppService
}

// NewHandler 创建 Handler
func NewHandler(dbm *database.DBManager, cache cache.Cache) *Handler {
	menuRepo := persistence.NewMenuDataRepository(dbm)
	menuAppSrv := application.NewTakeoutMenuAppService(dbm, menuRepo, cache)

	return &Handler{
		menuAppSrv: menuAppSrv,
	}
}

// ExportMenu 导出菜单
// @Summary 导出菜单到外卖平台格式
// @Description 将 TTPOS 菜单数据转换为指定外卖平台（如 Grab）的格式
// @Tags 外卖菜单
// @Accept json
// @Produce json
// @Security JwtAuth
// @Param body body req.TakeoutMenuExportReq true "导出请求"
// @Success 200 {object} dto.Response{data=resp.TakeoutMenuExportResp} "成功"
// @Router /takeout/menu/export [post]
func (h *Handler) ExportMenu(c *gin.Context) {
	var exportReq req.TakeoutMenuExportReq
	if err := c.ShouldBindJSON(&exportReq); err != nil {
		helper.HandleValidationError(c, err, exportReq, nil)
		return
	}

	ctx := helper.GetContext(c)

	// 如果未指定公司，使用当前公司
	if exportReq.CompanyUuid == 0 {
		exportReq.CompanyUuid = ctx.GetCompanyUuid()
	}

	// 调用应用服务
	menuData, err := h.menuAppSrv.ExportMenu(ctx, application.ExportMenuRequest{
		Platform:       exportReq.Platform,
		CompanyUuid:    exportReq.CompanyUuid,
		CategoryIDs:    exportReq.CategoryIDs,
		SellingTimeIDs: exportReq.SellingTimeIDs,
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "导出菜单失败"))
		return
	}

	helper.Success(c, resp.TakeoutMenuExportResp{
		Platform: exportReq.Platform,
		MenuData: menuData,
	})
}

// RegisterTakeoutHandlers 注册外卖菜单路由（共用 shop 的认证）
func RegisterTakeoutHandlers(router *gin.RouterGroup, dbm *database.DBManager, cache cache.Cache) {
	takeoutHandler := NewHandler(dbm, cache)

	// 需要认证
	privateApi := router.Group("")
	{
		privateApi.POST("/menu/export", takeoutHandler.ExportMenu)
	}
}
