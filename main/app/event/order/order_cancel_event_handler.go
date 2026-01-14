package event

import (
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/controller"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_cancel_order_event_handler sync.Once

// cancelOrderEventHandler "整单取消"事件处理器
func CancelOrderEventHandler() {
	once_cancel_order_event_handler.Do(func() {

		// 整单取消 - 创建操作记录
		event.NewSystemBus().SubscribeCancelOrderEvent(func(payload event.CancelOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source: payload.Source,
				Action: func() string {
					if payload.MemberSaleOrderUuid != 0 {
						return constant.OrderCancelMemberSaleOrder
					}
					return constant.OrderOrderCancel
				}(),
				Remark:              "整单取消",
				SaleBillUuid:        payload.SaleBillUuid,
				SaleOrderUuid:       payload.SaleOrderUuid,
				H5OrderUuid:         payload.H5OrderUuid,
				OperatorUuid:        payload.GetOperatorUuid(),
				MemberSaleOrderUuid: payload.MemberSaleOrderUuid,
				MemberUuid:          payload.MemberUuid,
				Data: func() string {
					if payload.MemberSaleOrderUuid != 0 {
						return utils.ToJson(payload.Data)
					}
					return ""
				}(),
			}
			if payload.Ctx.GetStaff().Uuid != 0 {
				record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			}
			_, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCancelOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
		})

		// 失效桌台缓存（桌台订单取消后）
		event.NewSystemBus().SubscribeCancelOrderEvent(func(payload event.CancelOrderPayload) {
			if adapter.IsObjectStorageCacheEnabled(payload.CompanyUuid) {
				// 查询销售账单信息，判断是否是桌台订单
				db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
				saleBill, err := repository.NewSaleBillRepo(db).GetSaleBillRecord(payload.SaleBillUuid)
				if err != nil {
					logger.Logger.Error("SubscribeCancelOrderEvent process, GetSaleBillRecord failed", zap.Uint64("saleBillUuid", payload.SaleBillUuid), zap.Error(err))
					return
				}
				// 如果是桌台订单，失效桌台缓存
				if saleBill.IsDeskSaleBill() {
					deskUuid := saleBill.DeskUuid
					deskUuid = persistence.GlobalObjectUuid // 还没有失效单个桌台缓存,全局失效桌台缓存
					if err := controller.GetDeskController().Invalidate(payload.Ctx, deskUuid); err != nil {
						logger.Logger.Error("SubscribeCancelOrderEvent process, Invalidate desk cache failed", zap.Uint64("deskUuid", deskUuid), zap.Error(err))
					}
				}
			}
		})
	})
}
