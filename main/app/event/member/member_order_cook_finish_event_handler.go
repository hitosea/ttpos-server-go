package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_cook_finish_member_sale_order_event_handler sync.Once

// cookFinishMemberSaleOrderEventHandler "外送订单“备餐完成”事件处理器"
func CookFinishMemberSaleOrderEventHandler() {
	once_cook_finish_member_sale_order_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCookFinishMemberSaleOrderEvent(func(payload event.CookFinishMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderPickMemberSaleOrder,
				Remark:        "备餐完成",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCookFinishMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:备餐完成 %+v", payload), zap.Uint64("record", uuid))

			// +++++++++++++++++++++++ 设置SaleBill的当班编号
			SetDutyNoForSaleBill(db, payload.BasePayload)

			// 结束SaleBill和SaleOrder
			EndSaleBillAndSaleOrder(db, payload)
		})
	})
}

// 备餐完成时，结束SaleBill和SaleOrder
func EndSaleBillAndSaleOrder(db *gorm.DB, payload event.CookFinishMemberSaleOrderPayload) {
	saleBill := payload.MemberSaleOrder.SaleBill
	staff := payload.Ctx.GetStaff()
	saleBill.EndSaleBillAndSaleOrder(staff.DutyNo, staff.Uuid, staff.GetUserName())
	service.CalcAndSaveSaleBill(payload.Ctx, db, saleBill)
}
