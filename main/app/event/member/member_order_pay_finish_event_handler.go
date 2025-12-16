package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
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

var once_pay_finish_member_sale_order_event_handler sync.Once

// payFinishMemberSaleOrderEventHandler "支付完成会员端销售订单"事件处理器
func PayFinishMemberSaleOrderEventHandler() {
	once_pay_finish_member_sale_order_event_handler.Do(func() {
		// 创建订单操作日志
		event.NewSystemBus().SubscribePayFinishMemberSaleOrderEvent(func(payload event.PayFinishMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:              payload.Source,
				Action:              constant.OrderPayFinishMemberSaleOrder,
				Remark:              "订单支付成功",
				SaleBillUuid:        payload.SaleBillUuid,
				SaleOrderUuid:       payload.SaleOrderUuid,
				OperatorUuid:        payload.GetOperatorUuid(),
				MemberUuid:          payload.MemberUuid,
				MemberSaleOrderUuid: payload.MemberSaleOrderUuid,
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribePayFinishMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:订单支付成功 %+v", payload), zap.Uint64("record", uuid))
		})

		// 修改订单状态
		event.NewSystemBus().SubscribePayFinishMemberSaleOrderEvent(func(payload event.PayFinishMemberSaleOrderPayload) {
			dbm := database.GetDBManager(config.DatabaseConf{})
			db := dbm.GetDB(payload.CompanyUuid)
			orderRepo := repository.NewMemberSaleOrderRepo(db)

			if AutoAcceptMemberSaleOrder(payload) {
				// 自动接单成功，结束事件处理
				return
			} else {
				// 自动接单失败，更新订单状态为“待商家接单”
				// 或未开启自动接单
				// 更新订单状态为“待商家接单”
				if err := orderRepo.UpdateMemberSaleOrderPendingMerchantAccept(payload.MemberSaleOrderUuid); err != nil {
					logger.Logger.Error("SubscribePayFinishMemberSaleOrderEvent process, UpdateMemberSaleOrderPendingMerchantAccept failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
					return
				}
			}
		})
	})
}

func AutoAcceptMemberSaleOrder(payload event.PayFinishMemberSaleOrderPayload) bool {
	dbm := database.GetDBManager(config.DatabaseConf{})
	cache := cache.Global
	db := dbm.GetDB(payload.CompanyUuid)
	orderRepo := repository.NewMemberSaleOrderRepo(db)

	// 如果自动接单，更新订单状态为“商家备餐中”
	settingSrv := setting.NewSrv(dbm, cache)
	cashierSetting, err := settingSrv.GetCashierSetting(payload.Ctx, nil)
	if err != nil {
		logger.Logger.Error("SubscribePayFinishMemberSaleOrderEvent process, GetCashierSetting failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
		return false
	}

	memberSaleOrder, err := orderRepo.GetMemberSaleOrderRecordOnly(payload.MemberSaleOrderUuid)
	if err != nil {
		logger.Logger.Error("SubscribePayFinishMemberSaleOrderEvent process, GetMemberSaleOrder failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
		return false
	}

	// 初始化订单服务
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	if cashierSetting.IsAutoMemberOrderBool() {
		limitAmount := cashierSetting.AutoMemberOrderLimitValue()
		amount := memberSaleOrder.Amount
		if limitAmount >= amount {
			// 设置上下文来源为收银机，用于显示在操作日志中的来源
			// payload.Ctx.SetSource(constant.SourceCashier)
			// 设置上下文场景为会员端订单
			payload.Ctx.SetScene(constant.SceneMemberOrder)
			// 设置上下文日志
			if err := orderSrv.AcceptMemberSaleOrder(payload.Ctx, req.AcceptOrderReq{
				MemberSaleOrderUuid: payload.MemberSaleOrderUuid,
			}, service.WithIsAutoAccept()); err != nil {
				logger.Logger.Info("SubscribePayFinishMemberSaleOrderEvent process, AcceptMemberSaleOrder failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
				return false
			}
			return true
		}
	}
	return false
}
