package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_takeout_accept_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	takeoutAcceptOrderEventHandler()
}

// checkoutSaleOrderEventHandler "结账"事件处理器
func takeoutAcceptOrderEventHandler() {
	once_takeout_accept_order_event_handler.Do(func() {

		// 创建结账单打印
		event.NewSystemBus().SubscribeAcceptMemberSaleOrderEvent(func(payload event.AcceptMemberSaleOrderPayload) {
			// 设置订单备注
			payload.MemberSaleOrder.SaleBill.Remark = payload.MemberSaleOrder.Remark
			// 打印外送单
			_, err := printer.NewPrinterRepo(payload.Ctx).PrintingTakeoutOrder(
				payload.MemberSaleOrder,
				payload.MemberSaleOrder.SaleBill,
				payload.SaleOrderUuid,
			)
			if err != nil {
				fmt.Println("CheckoutSaleOrderEvent process, PrintingStatementOrder failed ", err)
			}
		})

	})
}
