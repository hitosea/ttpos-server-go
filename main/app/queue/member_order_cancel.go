package queue

import (
	"encoding/json"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
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
	CancelScene         string `json:"cancel_scene"`           // 取消场景
}

// ProcessMemberOrderCancel 处理会员订单自动取消
func ProcessMemberOrderCancel(paramsJson string) bool {
	var params MemberOrderCancelParams
	if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
		logger.Logger.Error("ProcessMemberOrderCancel 处理会员订单自动取消失败", zap.String("paramsJson", paramsJson), zap.Error(err))
		return true
	}

	// 获取DB实例
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(params.CompanyUuid)

	// 创建上下文
	ctx := context.NewContext(
		context.WithCompanyUuid(params.CompanyUuid),
		context.WithLogger(logger.Logger),
	)
	ctx.SetDB(db)

	// 创建订单服务
	settingSrv := setting.NewSrv(dbm, cache.Global)
	orderSrv := service.NewOrderSrv(
		dbm,
		service.NewLocaleSrv(),
		settingSrv,
		service.NewMustPlanSrv(dbm),
		service.NewPaymentMethodSrv(dbm, settingSrv),
		service.NewMemberSrv(dbm, cache.Global),
		service.NewCashBoxSrv(dbm),
	)

	// 选购超时自动取消订单
	if params.CancelScene == constant.MemberSaleOrderSceneSelectingTimeout {
		err := orderSrv.MemberOrderSelectingTimeoutAutoCancel(ctx, params.MemberSaleOrderUuid)
		if err != nil {
			logger.Logger.Error("处理会员订单自动取消失败", zap.Error(err))
			return false
		}
		return true
	}

	// 支付超时自动取消订单
	if params.CancelScene == constant.MemberSaleOrderScenePaymentTimeout {
		err := orderSrv.MemberOrderPayTimeoutAutoCancel(ctx, params.MemberSaleOrderUuid)
		if err != nil {
			logger.Logger.Error("处理会员订单自动取消失败", zap.Error(err))
			return false
		}
		return true
	}

	// 骑手接单超时自动取消订单
	if params.CancelScene == constant.MemberSaleOrderSceneRiderPickupTimeout {
		err := orderSrv.MemberOrderRiderPickupTimeoutAutoCancel(ctx, params.MemberSaleOrderUuid)
		if err != nil {
			logger.Logger.Error("处理会员订单自动取消失败", zap.Error(err))
			return false
		}
		return true
	}

	return true
}
