package event

import (
	"sync"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_rider_delivery_member_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	riderDeliveryMemberSaleOrderEventHandler()
}

// riderDeliveryMemberSaleOrderEventHandler "骑手配送中"事件处理器
func riderDeliveryMemberSaleOrderEventHandler() {
	once_rider_delivery_member_sale_order_event_handler.Do(func() {
		// 骑手配送中后，更新订单状态
		event.NewSystemBus().SubscribeRiderDeliveryMemberSaleOrderEvent(func(payload event.RiderDeliveryMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			memberSaleOrder := model.MemberSaleOrder{
				BaseModel: model.BaseModel{
					Uuid: payload.MemberSaleOrderUuid,
				},
			}
			memberSaleOrder.RiderDelivery()
			repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRiderDelivery(memberSaleOrder)
		})
	})
}
