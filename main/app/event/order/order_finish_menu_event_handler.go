package event

import (
	"sync"
	"ttpos-server-go/app/modules/printer"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/repository"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
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
			// 获取订单信息
			dbm := database.GetDBManager(config.DatabaseConf{})
			db := dbm.GetDB(payload.CompanyUuid)

			// 异步打印送厨单 如果TakeoutOrderUuid不为0，则打印
			if payload.TakeoutOrderUuid != 0 {
				utils.Go(func() {
					takeoutSrv := takeoutService.NewTakeoutSrv(dbm, nil, nil, nil, nil, nil)
					_, err := takeoutSrv.PrintProductionOrder(payload.Ctx, payload.TakeoutOrderUuid, printerConstant.PrinterProductTypeOutMenu)
					if err != nil {
						logger.Logger.Error("打印送厨单失败",
							zap.Uint64("takeoutOrderUuid", payload.TakeoutOrderUuid),
							zap.Error(err))
					}
				})
			} else {
				// 异步打印出菜单
				utils.Go(func() {
					products := printer_model.Products{}
					copier.Copy(&products, payload.Products)
					repo := printer.NewPrinterRepo(payload.Ctx, "")
					repo.SetFinishedTime(payload.FinishedTime)

					// 获取订单信息
					billInfo, err := repository.NewOrderRepo(db).GetSaleBill(
						repository.CommonRepo.WhereByUuid(payload.SaleBillUuid),
						repository.CommonRepo.Preload(repository.WithPreload{Query: "Desk"}),
					)
					if err != nil {
						logger.Logger.Error("FinishMenuEventHandler process, GetSaleBill failed", zap.Error(err))
						return
					}
					repo.PrintingDishes(
						printerConstant.PrinterProductTypeOutMenu,
						printer_model.Order{
							Uuid:                   payload.SaleBillUuid,
							SaleOrderUuid:          payload.SaleOrderUuid,
							OrderNo:                billInfo.OrderNo,
							MealNum:                billInfo.MealNum,
							IsTakeoutBill:          billInfo.IsTakeoutBill(),
							OrderSourceTakeoutText: billInfo.GetOrderSourceTakeoutText(),
							SerialNo:               billInfo.SerialNo,
							OrderRemark:            billInfo.GetLatestOrderRemarkRes(),
							DeskUuid:               billInfo.DeskUuid,
							Desk: &printer_model.OrderDesk{
								DeskNo: func() string {
									if billInfo.Desk == nil {
										return ""
									}
									return billInfo.Desk.DeskNo
								}(),
								RegionUuid: func() uint64 {
									if billInfo.Desk == nil {
										return 0
									}
									return billInfo.Desk.RegionUuid
								}(),
							},
							UpdateTime: billInfo.UpdateTime,
							FinishTime: billInfo.FinishTime,
							Products:   products,
						},
					)
				})
			}
		})

	})
}
