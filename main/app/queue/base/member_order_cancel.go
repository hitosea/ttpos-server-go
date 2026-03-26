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

const MEMBER_ORDER_CANCEL = "member_order_cancel"

var (
	memberOrderCancelQueue *delayqueue.DelayQueue
	memberOrderInit        sync.Once

	// 服务实例缓存 - 避免重复创建
	serviceInstances struct {
		settingSrv  setting.ISrv
		orderSrv    service.IOrderSrv
		localeSrv   service.ILocaleSrv
		mustPlanSrv service.IMustPlanSrv
		paymentSrv  service.IPaymentMethodSrv
		memberSrv   service.IMemberSrv
		cashBoxSrv  service.ICashBoxSrv
		initialized bool
		mu          sync.RWMutex
	}
)

// InitMemberOrderCancel 初始化会员订单自动取消队列
func InitMemberOrderCancel() {
	memberOrderInit.Do(func() {
		// 初始化服务实例缓存
		initServiceInstances()

		// 增加并发数以提高性能
		concurrent := 10 // 从5增加到10
		if cache.Global.GetClusterClient() != nil {
			memberOrderCancelQueue = delayqueue.NewQueueOnCluster(MEMBER_ORDER_CANCEL, cache.Global.GetClusterClient(), ProcessMemberOrderCancel).WithConcurrent(uint(concurrent))
		} else {
			memberOrderCancelQueue = delayqueue.NewQueue(MEMBER_ORDER_CANCEL, cache.Global.GetClient(), ProcessMemberOrderCancel).WithConcurrent(uint(concurrent))
		}
		service.Queue.RegisterMemberOrderCancelQueue(memberOrderCancelQueue)
		memberOrderCancelQueue.StartConsume()
	})
}

// initServiceInstances 初始化服务实例缓存
func initServiceInstances() {
	serviceInstances.mu.Lock()
	defer serviceInstances.mu.Unlock()

	if serviceInstances.initialized {
		return
	}

	// 获取DB管理器
	dbm := database.GetDBManager(config.DatabaseConf{})

	// 创建服务实例
	serviceInstances.settingSrv = setting.NewSrv(dbm, cache.Global)
	serviceInstances.localeSrv = service.NewLocaleSrv()
	serviceInstances.mustPlanSrv = service.NewMustPlanSrv(dbm)
	serviceInstances.paymentSrv = service.NewPaymentMethodSrv(dbm, serviceInstances.settingSrv)
	serviceInstances.memberSrv = service.NewMemberSrv(dbm, cache.Global)
	serviceInstances.cashBoxSrv = service.NewCashBoxSrv(dbm)
	serviceInstances.orderSrv = service.NewOrderSrv(
		dbm,
		serviceInstances.localeSrv,
		serviceInstances.settingSrv,
		serviceInstances.mustPlanSrv,
		serviceInstances.paymentSrv,
		serviceInstances.memberSrv,
		serviceInstances.cashBoxSrv,
	)

	serviceInstances.initialized = true
	logger.Logger.Info("服务实例缓存初始化完成")
}

func GetMemberOrderCancelQueue() *delayqueue.DelayQueue {
	return memberOrderCancelQueue
}

// MemberOrderCancelParams 会员订单取消队列参数
type MemberOrderCancelParams struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员订单UUID（外送订单使用）
	SaleBillUuid        uint64 `json:"sale_bill_uuid"`         // 销售账单UUID（堂食订单使用）
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

	// 使用缓存的服务实例 - 避免重复创建
	serviceInstances.mu.RLock()
	orderSrv := serviceInstances.orderSrv
	serviceInstances.mu.RUnlock()

	if orderSrv == nil {
		logger.Logger.Error("订单服务实例未初始化")
		return false
	}

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

	// 堂食订单支付超时自动取消订单
	if params.CancelScene == constant.MemberSaleOrderSceneDineInPaymentTimeout {
		err := orderSrv.DineInOrderPayTimeoutAutoCancel(ctx, params.SaleBillUuid)
		if err != nil {
			logger.Logger.Error("处理堂食订单自动取消失败",
				zap.Uint64("company_uuid", params.CompanyUuid),
				zap.Uint64("sale_bill_uuid", params.SaleBillUuid),
				zap.Error(err))
			return false
		}
		return true
	}

	return true
}
