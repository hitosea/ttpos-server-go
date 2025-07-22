package event

import (
	"sync"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_reject_member_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	rejectMemberSaleOrderEventHandler()
}

// rejectMemberSaleOrderEventHandler "外送拒单"事件处理器
func rejectMemberSaleOrderEventHandler() {
	once_reject_member_sale_order_event_handler.Do(func() {
		// 外送拒单后，自动退款
		event.NewSystemBus().SubscribeRejectMemberSaleOrderEvent(func(payload event.RejectMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// +++++++++++++++++++++++ 设置SaleBill的当班编号
			setDutyNoForSaleBill(db, payload.BasePayload)
		})
	})
}
