package event

import (
	"sync"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_pay_finish_member_dine_in_order_event_handler sync.Once

// PayFinishMemberDineInOrderEventHandler 会员端堂食订单支付完成事件处理器
func PayFinishMemberDineInOrderEventHandler() {
	once_pay_finish_member_dine_in_order_event_handler.Do(func() {
		// 执行送厨逻辑并完成订单
		event.NewSystemBus().SubscribePayFinishMemberDineInOrderEvent(func(payload event.PayFinishMemberDineInOrderPayload) {
			// dbm := database.GetDBManager(config.DatabaseConf{})
			// db := dbm.GetDB(payload.CompanyUuid)

			// // 初始化订单服务
			// settingSrv := setting.NewSrv(dbm, cache.Global)
			// cashBoxSrv := service.NewCashBoxSrv(dbm)
			// localeSrv := service.NewLocaleSrv()
			// mustPlanSrv := service.NewMustPlanSrv(dbm)
			// paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
			// memberSrv := service.NewMemberSrv(dbm, cache.Global)
			// orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

			// // 获取销售账单信息（包含商品列表）
			// saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(payload.SaleBillUuid)
			// if err != nil {
			// 	logger.Logger.Error("SubscribePayFinishMemberDineInOrderEvent process, GetSaleBillAllInfo failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
			// 	return
			// }

			// // 检查是否有未送厨的商品
			// unCookingProducts := saleBill.GetSaleOrderProductUnCooking()
			// if len(unCookingProducts) == 0 {
			// 	logger.Logger.Info("SubscribePayFinishMemberDineInOrderEvent process, no uncooking product", zap.Uint64("saleBillUuid", payload.SaleBillUuid))
			// 	return
			// }

			// // 设置上下文来源为会员端
			// payload.Ctx.SetSource(constant.SourceMember)

			// // 执行送厨
			// _, checkRes, err := orderSrv.InstantOrderCartProductCooking(payload.Ctx, req.OrderCartProductCookingReq{
			// 	SaleBillUuid: payload.SaleBillUuid,
			// 	IgnoreMust:   true, // 会员端堂食订单忽略必点方案检查
			// })
			// if err != nil {
			// 	logger.Logger.Error("SubscribePayFinishMemberDineInOrderEvent process, InstantOrderCartProductCooking failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
			// 	return
			// }
			// if checkRes != nil {
			// 	logger.Logger.Warn("SubscribePayFinishMemberDineInOrderEvent process, InstantOrderCartProductCooking check failed", zap.Any("payload", utils.ToJson(payload)), zap.Any("checkRes", checkRes))
			// 	return
			// }

			// logger.Logger.Info("SubscribePayFinishMemberDineInOrderEvent process, InstantOrderCartProductCooking success", zap.Uint64("saleBillUuid", payload.SaleBillUuid))

			// // 完成订单（结账）
			// _, err = orderSrv.InstantOrderPaymentFinish(payload.Ctx, req.InstantOrderPaymentFinishReq{
			// 	SaleBillUuid:  payload.SaleBillUuid,
			// 	SaleOrderUuid: payload.SaleOrderUuid,
			// })
			// if err != nil {
			// 	logger.Logger.Error("SubscribePayFinishMemberDineInOrderEvent process, InstantOrderPaymentFinish failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
			// 	return
			// }

			// logger.Logger.Info("SubscribePayFinishMemberDineInOrderEvent process, order completed", zap.Uint64("saleBillUuid", payload.SaleBillUuid))
		})
	})
}
