package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_order_reverse_settle_event_handler sync.Once

// orderReverseSettleEventHandler 订单"反结账"事件处理器
func orderReverseSettleEventHandler() {
	once_order_reverse_settle_event_handler.Do(func() {
		event.NewSystemBus().SubscribeOrderReverseSettleEvent(func(payload event.OrderReverseSettlePayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderReverseSettle,
				Remark:        "反结账",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeOrderReverseSettleEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:反结账 %+v", payload), zap.Uint64("record", uuid))
		})

		event.NewSystemBus().SubscribeOrderReverseSettleEvent(func(payload event.OrderReverseSettlePayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// 反结账处理所有会员的等级变化。账单有多个销售订单时，要处理多个会员的等级变化
			saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(payload.SaleBillUuid)
			if errSaleBill != nil {
				logger.Logger.Error("GetSaleBillAllInfo", zap.Error(fmt.Errorf("%v %s", payload.SaleBillUuid, errSaleBill)))
				return
			}
			memberSrv := service.NewMemberSrv(database.GetDBManager(config.DatabaseConf{}), cache.Global)
			for _, saleOrder := range saleBill.SaleOrders {
				if saleOrder.ConsumerUuid != 0 {
					// 处理会员升级
					utils.Go(func() {
						memberSrv.HandleMemberUpgrade(payload.CompanyUuid, saleOrder.ConsumerUuid)
					})
				}
			}
		})
	})
}
