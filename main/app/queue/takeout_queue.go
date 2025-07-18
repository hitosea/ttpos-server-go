package queue

import (
	"context"
	"github.com/hdt3213/delayqueue"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/pkg/logger"

	"ttpos-server-go/app/service/rpc/takeout"
)

var (
	TakeoutCancelQueue *delayqueue.DelayQueue
	takeoutInit        sync.Once
)

// InitTakeoutCancel 初始化订单取消队列
func InitTakeoutCancel(client *redis.Client) {
	takeoutInit.Do(func() {
		TakeoutCancelQueue = delayqueue.NewQueue(TAKEOUT, client, func(shopRefNo string) bool {
			//TODO 这里填逻辑
			if err := takeout.NewTakeoutSrv().CancelOrder(context.Background(), &req.CancelTakeoutOrderReq{
				ShopOrderUuid: shopRefNo,
			}); err != nil {
				logger.Logger.Error("取消订单失败: %v", zap.Error(err))
			}
			return true
		}).WithConcurrent(5)
		TakeoutCancelQueue.StartConsume()
		//<-done
	})
}
