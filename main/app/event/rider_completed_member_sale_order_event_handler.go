package event

import (
	"sync"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_rider_completed_member_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	riderCompletedMemberSaleOrderEventHandler()
}

// riderCompletedMemberSaleOrderEventHandler "骑手配送完成"事件处理器
func riderCompletedMemberSaleOrderEventHandler() {
	once_rider_completed_member_sale_order_event_handler.Do(func() {
		// 骑手配送完成后，更新订单状态
		event.NewSystemBus().SubscribeRiderCompletedMemberSaleOrderEvent(func(payload event.RiderCompletedMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			memberSaleOrder := model.MemberSaleOrder{
				BaseModel: model.BaseModel{
					Uuid: payload.MemberSaleOrderUuid,
				},
			}
			memberSaleOrder.RiderCompleted()
			repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRiderCompleted(memberSaleOrder)
		})
	})
}
