package event

import (
	"sync"
	"ttpos-server-go/app/modules/printer"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
)

var once_finish_menu_event_handler sync.Once

// finishMenuEventHandler "完成制作"事件处理器
func FinishMenuEventHandler() {
	once_finish_menu_event_handler.Do(func() {

		// 创建送厨单打印记录
		event.NewSystemBus().SubscribeFinishMenuEvent(func(payload event.FinishMenuPayload) {
			if len(payload.Products) == 0 {
				return
			}
			utils.Go(func() {
				products := printer_model.Products{}
				copier.Copy(&products, payload.Products)
				repo := printer.NewPrinterRepo(payload.Ctx, "")
				repo.SetFinishedTime(payload.FinishedTime)
				repo.PrintingDishes(
					printerConstant.PrinterProductTypeOutMenu,
					payload.SaleBillUuid,
					payload.SaleOrderUuid,
					products,
				)
			})
		})

	})
}
