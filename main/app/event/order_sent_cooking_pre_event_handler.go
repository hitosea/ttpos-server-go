package event

import (
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
)

var once_sent_cooking_pre_event_handler sync.Once

// sentCookingPreEventHandler "预送厨"事件处理器
func sentCookingPreEventHandler() {
	once_sent_cooking_pre_event_handler.Do(func() {

		event.NewSystemBus().SubscribeSentCookingPreEvent(func(payload event.SentCookingPrePayload) {
			if len(payload.Products) == 0 {
				return
			}
			//
			utils.Go(func() {
				products := printer_model.Products{}
				copier.Copy(&products, payload.Products)
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypeKitchen,
					payload.SaleBillUuid,
					payload.SaleOrderUuid,
					products,
				)
			})
		})
	})
}
