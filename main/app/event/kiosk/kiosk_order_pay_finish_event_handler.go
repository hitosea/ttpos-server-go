package event

import (
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_pay_finish_kiosk_order_event_handler sync.Once

// PayFinishKioskOrderEventHandler 自助点餐机订单支付完成事件处理器
func PayFinishKioskOrderEventHandler() {
	once_pay_finish_kiosk_order_event_handler.Do(func() {
		// 创建订单操作日志
		// event.NewSystemBus().SubscribePayFinishKioskOrderEvent(func(payload event.PayFinishKioskOrderPayload) {
		// 	db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
		// 	orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
		// 	record := model.SaleOrderOperationRecord{
		// 		Source:        payload.Source,
		// 		Action:        constant.OrderPayFinishKioskOrder,
		// 		Remark:        "自助点餐机订单支付成功",
		// 		SaleBillUuid:  payload.SaleBillUuid,
		// 		SaleOrderUuid: payload.SaleOrderUuid,
		// 		OperatorUuid:  payload.GetOperatorUuid(),
		// 	}
		// 	record.Data = payload.ToJsonString()
		// 	uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
		// 	if err != nil {
		// 		logger.Logger.Error("SubscribePayFinishKioskOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
		// 		return
		// 	}
		// 	logger.Logger.Info("操作记录:自助点餐机订单支付成功", zap.Uint64("record", uuid), zap.Any("payload", payload))
		// })

		// 执行送厨逻辑
		event.NewSystemBus().SubscribePayFinishKioskOrderEvent(func(payload event.PayFinishKioskOrderPayload) {
			dbm := database.GetDBManager(config.DatabaseConf{})
			db := dbm.GetDB(payload.CompanyUuid)

			// 初始化订单服务
			settingSrv := setting.NewSrv(dbm, cache.Global)
			cashBoxSrv := service.NewCashBoxSrv(dbm)
			localeSrv := service.NewLocaleSrv()
			mustPlanSrv := service.NewMustPlanSrv(dbm)
			paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
			memberSrv := service.NewMemberSrv(dbm, cache.Global)
			orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

			// 获取销售账单信息（包含商品列表）
			saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(payload.SaleBillUuid)
			if err != nil {
				logger.Logger.Error("SubscribePayFinishKioskOrderEvent process, GetSaleBillAllInfo failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
				return
			}

			// 检查是否有未送厨的商品
			unCookingProducts := saleBill.GetSaleOrderProductUnCooking()
			if len(unCookingProducts) == 0 {
				logger.Logger.Info("SubscribePayFinishKioskOrderEvent process, no uncooking product", zap.Uint64("saleBillUuid", payload.SaleBillUuid))
				return
			}

			// 设置上下文来源为 kiosk
			payload.Ctx.SetSource(constant.SourceKiosk)

			// 执行送厨
			_, checkRes, err := orderSrv.InstantOrderCartProductCooking(payload.Ctx, req.OrderCartProductCookingReq{
				SaleBillUuid: payload.SaleBillUuid,
				IgnoreMust:   true, // Kiosk 订单忽略必点方案检查
			})
			if err != nil {
				logger.Logger.Error("SubscribePayFinishKioskOrderEvent process, InstantOrderCartProductCooking failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
				return
			}
			if checkRes != nil {
				logger.Logger.Warn("SubscribePayFinishKioskOrderEvent process, InstantOrderCartProductCooking check failed", zap.Any("payload", utils.ToJson(payload)), zap.Any("checkRes", checkRes))
				return
			}

			logger.Logger.Info("SubscribePayFinishKioskOrderEvent process, InstantOrderCartProductCooking success", zap.Uint64("saleBillUuid", payload.SaleBillUuid))

			// 完成订单（结账）
			_, err = orderSrv.InstantOrderPaymentFinish(payload.Ctx, req.InstantOrderPaymentFinishReq{
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
			})
			if err != nil {
				logger.Logger.Error("SubscribePayFinishKioskOrderEvent process, InstantOrderPaymentFinish failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
				return
			}

			logger.Logger.Info("SubscribePayFinishKioskOrderEvent process, order completed", zap.Uint64("saleBillUuid", payload.SaleBillUuid))
		})
	})
}
