package event

import (
	"sync"
	"time"
	"ttpos-server-go/app/modules/printer"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_sent_cooking_pre_event_handler sync.Once

// sentCookingPreEventHandler "预送厨"事件处理器
func SentCookingPreEventHandler() {
	once_sent_cooking_pre_event_handler.Do(func() {

		event.NewSystemBus().SubscribeSentCookingPreEvent(func(payload event.SentCookingPrePayload) {
			if len(payload.Products) == 0 {
				return
			}

			if payload.IsPreBatchPrint { // 是预先分批打印送厨单，处理预先分批打印
				// 参考 handleBatchProductsMergePrint 的逻辑，处理预先分批打印
				utils.Go(func() {
					handlePreBatchProductsMergePrint(payload)
				})
			} else {
				utils.Go(func() {
					products := printer_model.Products{}
					copier.Copy(&products, payload.Products)
					// 获取订单信息
					db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
					billInfo, err := repository.NewOrderRepo(db).GetSaleBill(
						repository.CommonRepo.WhereByUuid(payload.SaleBillUuid),
						repository.CommonRepo.Preload(repository.WithPreload{Query: "Desk"}),
					)
					if err != nil {
						logger.Logger.Error("SentCookingPreEventHandler process, GetSaleBill failed", zap.Error(err))
						return
					}
					printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
						printerConstant.PrinterProductTypeKitchen,
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

// handlePreBatchProductsMergePrint 处理预送厨商品的预先分批打印逻辑
func handlePreBatchProductsMergePrint(payload event.SentCookingPrePayload) {
	// payload.Products 中传入的商品列表已经是 pre_batch_print_time 为0的商品了
	// 按 BatchTagUuid 分组
	typeGroups := make(map[uint64][]event.OrderProductPre)
	productUuids := make([]uint64, 0)
	for _, product := range payload.Products {
		batchTagUuid := product.BatchTagUuid
		if batchTagUuid == 0 {
			// 如果没有分批类型，跳过不打印
			continue
		}
		if typeGroups[batchTagUuid] == nil {
			typeGroups[batchTagUuid] = []event.OrderProductPre{product}
		} else {
			typeGroups[batchTagUuid] = append(typeGroups[batchTagUuid], product)
		}
		productUuids = append(productUuids, product.OrderProductId)
	}

	if len(typeGroups) == 0 {
		return
	}

	db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

	// 更新 pre_batch_print_time 为当前时间戳
	{
		// 打印完成后，更新这些商品的 pre_batch_print_time 为当前时间戳
		if db == nil {
			logger.Logger.Error("获取数据库连接失败", zap.Uint64("companyUuid", payload.CompanyUuid))
			return
		}
		printTime := time.Now().Unix()
		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			return repository.NewSaleOrderProductRepo(tx).Update(map[string]interface{}{
				"pre_batch_print_time": printTime,
			}, repository.CommonRepo.WhereInUuids(productUuids))
		}); err != nil {
			logger.Logger.Error("更新 pre_batch_print_time 失败", zap.Error(err), zap.Uint64("saleBillUuid", payload.SaleBillUuid))
			return
		}
	}

	// 每个组进行一次打印送厨单
	for _, products := range typeGroups {
		if len(products) == 0 {
			continue
		}

		// 转换为打印模型
		printProducts := printer_model.Products{}
		// 使用 Copy 函数将 OrderProductPre 转换为 printer_model.OrderProduct
		for _, productPre := range products {
			printProduct := printer_model.OrderProduct{}
			copier.Copy(&printProduct, productPre)
			printProduct.ShowDelayTag = true // 标记显示为延迟送厨
			printProducts = append(printProducts, printProduct)
		}

		if len(printProducts) == 0 {
			continue
		}

		// 使用第一个商品的销售订单UUID（从 payload 中获取）
		saleOrderUuid := payload.SaleOrderUuid
		if saleOrderUuid > 0 {
			// 获取订单信息
			billInfo, err := repository.NewOrderRepo(db).GetSaleBill(
				repository.CommonRepo.WhereByUuid(payload.SaleBillUuid),
				repository.CommonRepo.Preload(repository.WithPreload{Query: "Desk"}),
			)
			if err != nil {
				logger.Logger.Error("handlePreBatchProductsMergePrint process, GetSaleBill failed", zap.Error(err))
				return
			}
			printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
				printerConstant.PrinterProductTypeKitchen,
				printer_model.Order{
					Uuid:                   payload.SaleBillUuid,
					SaleOrderUuid:          saleOrderUuid,
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
					Products:   printProducts,
				},
			)
		} else {
			logger.Logger.Warn("打印商品列表不为空但未找到销售订单UUID", zap.Uint64("saleBillUuid", payload.SaleBillUuid), zap.Int("productCount", len(printProducts)))
		}
	}
}
