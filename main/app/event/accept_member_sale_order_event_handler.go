package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_accept_member_sale_order_event_handler sync.Once

// init 自动注册"外送订单“商家接单”事件处理器
func init() {
	// 只初始化一次
	acceptMemberSaleOrderEventHandler()
}

// acceptMemberSaleOrderEventHandler "外送订单“商家接单”事件处理器"
func acceptMemberSaleOrderEventHandler() {
	once_accept_member_sale_order_event_handler.Do(func() {
		event.NewSystemBus().SubscribeAcceptMemberSaleOrderEvent(func(payload event.AcceptMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderAcceptMemberSaleOrder,
				Remark:        "商家接单",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeAcceptMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:商家接单 %+v", payload), zap.Uint64("record", uuid))

			// +++++++++++++++++++++++ 设置SaleBill的当班编号
			setDutyNoForSaleBill(db, payload.BasePayload)
		})
	})
}

// 设置SaleBill当班编号
func setDutyNoForSaleBill(db *gorm.DB, payload event.BasePayload) {
	if err := repository.NewSaleBillRepo(db).UpdateDutyNo(payload.SaleBillUuid, payload.Staff.DutyNo); err != nil {
		logger.Logger.Error("SubscribeAcceptMemberSaleOrderEvent process, UpdateDutyNo failed", zap.Uint64("sale_bill_uuid", payload.SaleBillUuid), zap.Error(err))
	}
}
