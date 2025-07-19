package event

import (
	"go.uber.org/zap"
	"sync"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_rider_accept_member_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	riderAcceptMemberSaleOrderEventHandler()
}

// riderAcceptMemberSaleOrderEventHandler "外送拒单"事件处理器
func riderAcceptMemberSaleOrderEventHandler() {
	once_rider_accept_member_sale_order_event_handler.Do(func() {
		// 骑手接单后，更新订单状态
		event.NewSystemBus().SubscribeRiderAcceptMemberSaleOrderEvent(func(payload event.RiderAcceptMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			memberSaleOrder := model.MemberSaleOrder{
				BaseModel: model.BaseModel{
					Uuid: payload.MemberSaleOrderUuid,
				},
			}
			memberSaleOrder.RiderAccept()
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRiderAccept(memberSaleOrder); err != nil {
				payload.Ctx.Log().Error("更新会员端销售订单-骑手接单失败", zap.Error(err))
			}
		})
	})
}
