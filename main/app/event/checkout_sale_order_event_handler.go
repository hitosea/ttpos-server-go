package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_checkout_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	checkoutSaleOrderEventHandler()
}

// checkoutSaleOrderEventHandler "结账"事件处理器
func checkoutSaleOrderEventHandler() {
	once_checkout_sale_order_event_handler.Do(func() {
		// 打印
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			_, err := printer.NewPrinterRepo(payload.Ctx).PrintingStatementOrder(
				constant.PrinterTemplateBilling,
				payload.SaleBill,
				payload.SaleOrderUuid,
				0,
			)
			if err != nil {
				logger.Logger.Error("SubscribeCheckoutZeroSaleOrderEvent process, PrintStatementOrder failed", zap.Any("payload", payload), zap.Error(err))
				return
			}
		})
		// 创建操作记录
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleBillOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderSettle,
				Remark:        "结账",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCheckoutZeroSaleOrderEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:结账 %+v", payload), zap.Uint64("record", uuid))
		})
		// 扣减库存
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			ReduceStock(db, payload.SaleBillUuid)
		})
		// 扣减会员余额
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			// db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			// 判断该订单的付款单中是否存在会员余额的支付方式。如果存在，则创建余额明细记录-扣减
			// ReduceMemberBalance(db, payload.SaleBillUuid)
		})

		// 发布会员余额变动事件。 获取所有未处理的余额明细记录，计算出会员余额变动金额，更新会员余额。

		// 发放积分
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			if payload.SaleOrderUuid == 0 {
				return
			}
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			saleOrder := payload.SaleBill.GetSaleOrder(payload.SaleOrderUuid)
			// 如果订单有会员且开启积分赠送且赠送比例大于0，则发放积分
			if saleOrder.ConsumerUuid != 0 && saleOrder.GiftPointsRate > 0 {
				// 加锁, 避免并发问题
				lock.NewSystemLock().LockUuid(saleOrder.ConsumerUuid)
				defer lock.NewSystemLock().UnlockUuid(saleOrder.ConsumerUuid)

				// 获取最新的会员信息
				member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
				if err != nil {
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetMemberRecord failed", zap.Any("payload", payload), zap.Error(err))
					return
				}
				// 创建积分发放记录
				saleOrder.HandleMemberPoints(&member)

				if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
					// 更新会员积分
					if err := repository.NewMemberRepo(tx).Update(member.Uuid, map[string]any{
						"frozen_point": member.FrozenPoint,
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewMemberPointLog()
					if _, err := repository.NewMemberPointLogRepo(tx).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
					return nil
				}); err != nil {
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, Transaction failed", zap.Any("payload", payload), zap.Error(err))
					return
				}

				// 发布“积分变动”事件
				go HandleMemberPoints(db, saleOrder.ConsumerUuid)
			}
		})
	})
}

// 处理积分变动
func HandleMemberPoints(db *gorm.DB, memberUuid uint64) {
	// 加锁, 避免并发问题
	lock.NewSystemLock().LockUuid(constant.LockNameMemberPoints)
	defer lock.NewSystemLock().UnlockUuid(constant.LockNameMemberPoints)

	// 查询积分变动
	memberPointLogRepo := repository.NewMemberPointLogRepo(db)
	memberPointLogs, err := memberPointLogRepo.GetMemberPointLogNotProcessed()
	if err != nil {
		logger.Logger.Info("HandleMemberPoints process, GetMemberPointLogNotProcessed failed", zap.Any("memberUuid", memberUuid), zap.Error(err))
		return
	}

	type MemberUuid uint64
	memberChangePoint := make(map[MemberUuid]decimal.Decimal) // 各个会员的积分变动
	for _, memberPointLog := range memberPointLogs {
		// 累计同一个会员的积分变动
		pre := memberChangePoint[MemberUuid(memberPointLog.MemberUuid)]
		memberChangePoint[MemberUuid(memberPointLog.MemberUuid)] = pre.Add(decimal.NewFromFloat(memberPointLog.Value))
	}

	memberUuids := make([]uint64, 0) // 积分有变动的会员
	for memberUuid := range memberChangePoint {
		memberUuids = append(memberUuids, uint64(memberUuid))
	}

	// 获取会员信息
	members, err := repository.NewMemberRepo(db).GetMembersByUuids(memberUuids)
	if err != nil {
		logger.Logger.Info("HandleMemberPoints process, GetMembersByUuids failed", zap.Any("memberUuids", memberUuids), zap.Error(err))
		return
	}
	// 更新会员积分
	logMemberInfoMap := make(map[string][]float64)
	for _, member := range members {
		beforePoint := member.Point
		memberChangePoint := memberChangePoint[MemberUuid(member.Uuid)].InexactFloat64()
		member.UpdatePoint(memberChangePoint)
		logMemberInfoMap[member.Nickname] = []float64{beforePoint, memberChangePoint, member.Point}
	}

	uuids := make([]uint64, 0)
	for _, memberPointLog := range memberPointLogs {
		uuids = append(uuids, memberPointLog.Uuid)
	}

	// 更新会员积分,更新到数据库
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		for _, member := range members {
			if err := repository.NewMemberRepo(tx).Update(member.Uuid, map[string]any{
				"point":        member.Point,
				"frozen_point": member.FrozenPoint,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 标记所有记录为已经处理
		if err := repository.NewMemberPointLogRepo(tx).UpdateProcessed(uuids); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		logger.Logger.Info("HandleMemberPoints process, Transaction failed", zap.Any("members", members), zap.Error(err))
		return
	}
	// 记录日志
	for nickname, info := range logMemberInfoMap {
		logger.Logger.Info("HandleMemberPoints process, UpdateMemberPoint", zap.Any("member", nickname), zap.Any("before point", info[0]), zap.Any("change point", info[1]), zap.Any("after point", info[2]))
	}
}
