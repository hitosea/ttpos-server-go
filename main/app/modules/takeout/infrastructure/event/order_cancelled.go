package event

import (
	"context"
	"fmt"
	"sync"

	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_order_cancelled_event_handler sync.Once

// RegisterErpSyncEventHandler 注册 ERP 同步事件处理器
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func RegisterOrderCancelledEventHandler() {
	once_order_cancelled_event_handler.Do(func() {
		// 订阅外卖订单接单事件
		event.GetDispatcher().Register(&orderCancelledEventHandler{})
	})
}

// erpSyncEventHandler ERP 同步事件处理器（基础设施层）
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
// @responsibility 订阅外卖订单接单事件，触发 ERP 同步
type orderCancelledEventHandler struct{}

// SubscribedEvents 返回订阅的事件类型
func (h *orderCancelledEventHandler) SubscribedEvents() []string {
	return []string{"takeout.order.cancelled"}
}

// Handle 处理事件
func (h *orderCancelledEventHandler) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderCancelEvent, ok := domainEvent.(event.OrderCancelEvent)
	if !ok {
		return nil
	}

	orderUuid := orderCancelEvent.OrderUuid
	companyUuid := orderCancelEvent.CompanyUuid

	db := database.GetDBManager(config.Database).GetDB(companyUuid)
	if db == nil {
		return fmt.Errorf("获取数据库连接失败")
	}

	// 创建上下文
	ctx := appContext.NewContext(
		appContext.WithContext(context.Background()),
		appContext.WithCompanyUuid(companyUuid),
	)
	ctx.SetCompanyUuid(companyUuid)
	ctx.SetDB(db)

	// 查询公司配置信息
	companyRepo := repository.NewCompanyRepo(db)
	company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		return fmt.Errorf("查询公司信息失败: %w", err)
	}
	if company == nil || company.CompanySetting == nil {
		return fmt.Errorf("公司配置信息不存在")
	}

	ctx.SetCompanyUuid(companyUuid)
	ctx.SetCompany(*company)
	ctx.SetCompanySetting(*company.CompanySetting)

	// 异步同步到 ERP
	utils.Go(func() {
		if err := service.NewTakeoutErpSyncService().SyncOrderCancelledToERP(ctx, orderUuid); err != nil {
			logger.Logger.Error("同步 Grab 订单取消到 ERP 失败",
				zap.Uint64("orderUuid", orderUuid),
				zap.Error(err))
		}
	})

	return nil
}
