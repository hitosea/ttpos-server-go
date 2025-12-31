package event

import (
	"context"
	"fmt"
	"sync"

	"ttpos-server-go/app/modules/takeout/domain/event"
	takeoutDomainService "ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_erp_sync_event_handler sync.Once

// RegisterErpSyncEventHandler 注册 ERP 同步事件处理器
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func RegisterErpSyncEventHandler() {
	once_erp_sync_event_handler.Do(func() {
		// 订阅外卖订单接单事件
		event.GetDispatcher().Register(&erpSyncEventHandler{})
	})
}

// erpSyncEventHandler ERP 同步事件处理器（基础设施层）
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
// @responsibility 订阅外卖订单接单事件，触发 ERP 同步
type erpSyncEventHandler struct{}

// SubscribedEvents 返回订阅的事件类型
func (h *erpSyncEventHandler) SubscribedEvents() []string {
	return []string{"takeout.order.accepted", "takeout.order.ready"}
}

// Handle 处理事件
func (h *erpSyncEventHandler) Handle(domainEvent event.DomainEvent) error {
	var orderUuid uint64
	var platform string
	var platformOrderId string
	var shortOrderNumber string
	var takeoutOrderUuid string
	var acceptedBy uint64
	var companyUuid uint64

	// 类型断言 - 支持两种事件类型
	switch e := domainEvent.(type) {
	case event.OrderAcceptedEvent:
		// 手动接单事件
		orderUuid = e.OrderUuid
		platform = e.Platform
		platformOrderId = e.PlatformOrderId
		shortOrderNumber = e.ShortOrderNumber
		takeoutOrderUuid = e.TakeoutOrderUuid
		acceptedBy = e.AcceptedBy
		companyUuid = e.CompanyUuid
	case event.OrderReadyEvent:
		// 呼叫骑手事件（自动接单订单在此时触发 ERP 同步）
		orderUuid = e.OrderUuid
		platform = e.Platform
		platformOrderId = e.PlatformOrderId
		shortOrderNumber = e.ShortOrderNumber
		takeoutOrderUuid = e.TakeoutOrderUuid
		companyUuid = e.CompanyUuid
		acceptedBy = e.AcceptedBy
	default:
		return nil
	}

	db := database.GetDBManager(config.Database).GetDB(companyUuid)
	if db == nil {
		return fmt.Errorf("数据库连接失败")
	}

	// 创建上下文
	ctx := appContext.NewContext(
		appContext.WithContext(context.Background()),
		appContext.WithStaffUuid(acceptedBy),
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

	// 构造统一的事件对象用于 ERP 同步
	orderAcceptedEvent := event.OrderAcceptedEvent{
		OrderUuid:        orderUuid,
		Platform:         platform,
		PlatformOrderId:  platformOrderId,
		ShortOrderNumber: shortOrderNumber,
		TakeoutOrderUuid: takeoutOrderUuid,
		AcceptedBy:       acceptedBy,
		CompanyUuid:      companyUuid,
	}

	// 异步同步到 ERP
	utils.Go(func() {
		erpSyncService := takeoutDomainService.NewTakeoutErpSyncService()
		if err := erpSyncService.SyncOrderToERP(ctx, orderUuid); err != nil {
			logger.Logger.Error("同步 Grab 订单到 ERP 失败",
				zap.Uint64("orderUuid", orderAcceptedEvent.OrderUuid),
				zap.String("platformOrderId", orderAcceptedEvent.PlatformOrderId),
				zap.Error(err))
		}
	})

	return nil
}
