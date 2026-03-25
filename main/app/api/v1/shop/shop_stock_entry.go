package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StockEntryHandler Stock Entry 控制器
type StockEntryHandler struct {
	dbm *database.DBManager
}

// TriggerStockEntryDeduction 触发 Stock Entry 合并扣减
// @Summary 触发 Stock Entry 合并扣减
// @Description 查询未扣减库存的已结账订单，合并提交 Stock Entry 到 ERPNext
// @Tags 商家端.库存管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response "成功"
// @Router /shop/stock_entry/trigger_deduction [post]
func (h *StockEntryHandler) TriggerStockEntryDeduction(c *gin.Context) {
	ctx := helper.GetContext(c)

	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsErpSalesInvoiceMode() {
		helper.Success(c, nil, "当前未启用ERP Sales Invoice模式")
		return
	}

	companyUuid := ctx.GetCompanyUuid()
	erpStockEntrySrv := service.NewErpStockEntrySrv(h.dbm)

	utils.Go(func() {
		if err := erpStockEntrySrv.TriggerStockEntryDeduction(ctx, companyUuid); err != nil {
			logger.Logger.Error("触发Stock Entry合并扣减失败",
				zap.Uint64("company_uuid", companyUuid),
				zap.Error(err))
		}
	})

	helper.Success(c, nil, "Stock Entry合并扣减已触发")
}

// RegisterStockEntryHandlers 注册 Stock Entry 相关路由
func RegisterStockEntryHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &StockEntryHandler{
		dbm: dbm,
	}

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/stock_entry/trigger_deduction", wrapper.TriggerStockEntryDeduction) // 触发 Stock Entry 合并扣减
	}
}
