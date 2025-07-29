package queue

import (
	"encoding/json"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"

	"github.com/hdt3213/delayqueue"
)

var (
	memberOrderCancelQueue *delayqueue.DelayQueue
	memberOrderInit        sync.Once
)

// InitMemberOrderCancel 初始化会员订单自动取消队列
func initMemberOrderCancel() {
	memberOrderInit.Do(func() {
		if cache.Global.GetClusterClient() != nil {
			memberOrderCancelQueue = delayqueue.NewQueueOnCluster(MEMBER_ORDER_CANCEL, cache.Global.GetClusterClient(), ProcessMemberOrderCancel).WithConcurrent(5)
		} else {
			memberOrderCancelQueue = delayqueue.NewQueue(MEMBER_ORDER_CANCEL, cache.Global.GetClient(), ProcessMemberOrderCancel).WithConcurrent(5)
		}
		service.Queue.RegisterMemberOrderCancelQueue(memberOrderCancelQueue)
		memberOrderCancelQueue.StartConsume()
	})
}

func GetMemberOrderCancelQueue() *delayqueue.DelayQueue {
	return memberOrderCancelQueue
}

// MemberOrderCancelParams 会员订单取消队列参数
type MemberOrderCancelParams struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员订单UUID
	CompanyUuid         uint64 `json:"company_uuid"`           // 公司UUID
	Reason              string `json:"reason"`                 // 取消原因
}

// ProcessMemberOrderCancel 处理会员订单自动取消
func ProcessMemberOrderCancel(paramsJson string) bool {
	var params MemberOrderCancelParams
	if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
		logger.Logger.Error("ProcessMemberOrderCancel 处理会员订单自动取消失败", zap.String("paramsJson", paramsJson), zap.Error(err))
		return true
	}

	// 获取DB
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(params.CompanyUuid)

	// 1. 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(params.MemberSaleOrderUuid)
	if err != nil {
		logger.Logger.Error("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		return false
	}

	// 2. 检查订单状态是否可以取消 - 只有“选购中”状态的订单才能自动取消
	if memberSaleOrder.Status != constant.MemberSaleOrderStatusSelecting {
		return true
	}

	// 3. 执行取消操作
	memberSaleOrder.SetCancel(params.Reason)
	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		logger.Logger.Error("更新会员订单状态失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		return false
	}

	// 创建上下文
	ctx := context.NewContext(
		context.WithCompanyUuid(params.CompanyUuid),
		context.WithLogger(logger.Logger),
	)
	ctx.SetDB(db)

	// 4. 取消订单（如果有关联的销售账单）
	if memberSaleOrder.SaleBillUuid > 0 {
		if err := repository.NewOrderRepo(db).CancelOrder(ctx, memberSaleOrder.SaleBillUuid, 0, params.Reason); err != nil {
			logger.Logger.Error("取消销售账单失败",
				zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid),
				zap.Uint64("saleBillUuid", memberSaleOrder.SaleBillUuid),
				zap.Error(err))
			// 这里不返回false，因为会员订单状态已经更新成功了
		}
	}

	// 发布“订单取消”操作事件
	go func() {
		event.NewSystemBus().PublishCancelMemberOrderEvent(event.CancelMemberOrderPayload{
			BasePayload: event.BasePayload{ // 基础信息
				Ctx:                 ctx,
				CompanyUuid:         ctx.GetCompanyUuid(),
				Source:              constant.SourceMember,
				SaleBillUuid:        memberSaleOrder.SaleBillUuid,
				SaleOrderUuid:       memberSaleOrder.Uuid,
				MemberUuid:          memberSaleOrder.MemberUuid,
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
			},
			Data: event.CancelMemberOrderPayloadData{
				Type: "timeout_cancel",
			},
		})
	}()

	return true
}
