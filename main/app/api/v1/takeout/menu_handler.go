package takeout

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/types/request"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// Handler 外卖菜单 Handler
type Handler struct {
	menuAppSrv application.ITakeoutMenuAppService
	dbm        *database.DBManager
	cache      cache.Cache
}

// NewHandler 创建 Handler
func NewHandler(dbm *database.DBManager, cache cache.Cache) *Handler {
	menuAppSrv := application.NewTakeoutMenuAppService(dbm)

	return &Handler{
		menuAppSrv: menuAppSrv,
		dbm:        dbm,
		cache:      cache,
	}
}

// ExportMenu 导出菜单
// @Summary 导出菜单到外卖平台格式
// @Description 将 TTPOS 菜单数据转换为指定外卖平台（如 Grab）的格式
// @Tags 外卖菜单
// @Accept json
// @Produce json
// @Security JwtAuth
// @Param body body request.ExportMenuRequest true "导出请求"
// @Param X-TTPOS-SECRET header string true "TTPOS 导出密钥"
// @Success 200 {object} dto.Response{data=resp.TakeoutMenuExportResp} "成功"
// @Router /takeout/menu/export [post]
func (h *Handler) ExportMenu(c *gin.Context) {
	var exportReq request.ExportMenuRequest
	if err := c.ShouldBindJSON(&exportReq); err != nil {
		helper.HandleValidationError(c, err, exportReq, nil)
		return
	}

	// 验证头部参数
	auth := c.GetHeader("X-TTPOS-SECRET")
	if auth != config.Takeout.TakeoutTtposSecret {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("无效的认证类型"))
		return
	}

	// 平台不能为空
	if exportReq.Platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台不能为空"))
		return
	}

	// 公司ID不能为空
	db := h.dbm.GetDB(exportReq.CompanyUuid)
	if db == nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("公司ID不存在"))
		return
	}

	ctx := helper.GetContext(c)
	ctx.SetCompanyUuid(exportReq.CompanyUuid)
	ctx.SetDB(h.dbm.GetDB(exportReq.CompanyUuid))
	currencySetting, err := setting.NewSrv(h.dbm, h.cache).GetCurrencySetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取货币设置失败"))
		return
	}

	// 调用应用服务
	menuData, err := h.menuAppSrv.ExportMenu(ctx, request.ExportMenuRequest{
		Platform:     exportReq.Platform,
		CurrencyUnit: currencySetting.Unit,
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
