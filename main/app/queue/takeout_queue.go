package queue

import (
	"context"
	"github.com/hdt3213/delayqueue"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"
)

var (
	takeoutCancelQueue *delayqueue.DelayQueue
	takeoutInit        sync.Once
)

// InitTakeoutCancel 初始化订单取消队列
func initTakeoutCancel() {
	takeoutInit.Do(func() {
		if cache.Global.GetClusterClient() != nil {
			takeoutCancelQueue = delayqueue.NewQueueOnCluster(TAKEOUT, cache.Global.GetClusterClient(), TakeoutCancelFunc).WithConcurrent(5)
		} else {
			takeoutCancelQueue = delayqueue.NewQueue(TAKEOUT, cache.Global.GetClient(), TakeoutCancelFunc).WithConcurrent(5)
		}
		service.Queue.RegisterTakeoutCancelQueue(takeoutCancelQueue)
		takeoutCancelQueue.StartConsume()
	})
}

func TakeoutCancelFunc(shopRefNo string) bool {
	// 解析参数，判断是外送订单UUID还是会员订单UUID
	_, err := strconv.ParseUint(shopRefNo, 10, 64)
	if err != nil {
		logger.Logger.Error("解析订单UUID失败", zap.String("shopRefNo", shopRefNo), zap.Error(err))
		return true // 返回true表示消费成功，避免重复处理
	}

	// 这里需要根据业务逻辑判断是否是会员端订单自动取消
	// 暂时先调用现有的外送取消逻辑，后续可以优化为更通用的订单取消处理
	if err := takeout.NewTakeoutSrv().CancelOrder(context.Background(), &req.CancelTakeoutOrderReq{
		ShopOrderUuid: shopRefNo,
	}); err != nil {
		logger.Logger.Error("自动取消订单失败", zap.String("memberSaleOrderUuid", shopRefNo), zap.Error(err))
		// 这里可以根据错误类型决定是否重试
		return false // 返回false会触发重试
	}

	logger.Logger.Info("自动取消订单成功", zap.String("memberSaleOrderUuid", shopRefNo))
	return true
}

func GetTakeoutCancelQueue() *delayqueue.DelayQueue {
	return takeoutCancelQueue
}
