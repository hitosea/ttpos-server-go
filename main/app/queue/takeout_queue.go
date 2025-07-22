package queue

import (
	"github.com/hdt3213/delayqueue"
	"sync"
	"ttpos-server-go/app/queue/consumer"
	"ttpos-server-go/pkg/cache"
)

var (
	TakeoutCancelQueue *delayqueue.DelayQueue
	takeoutInit        sync.Once
)

// InitTakeoutCancel 初始化订单取消队列
func InitTakeoutCancel() {
	takeoutInit.Do(func() {
		if cache.Global.GetClusterClient() != nil {
			TakeoutCancelQueue = delayqueue.NewQueueOnCluster(TAKEOUT, cache.Global.GetClusterClient(), consumer.TakeoutCancelFunc).WithConcurrent(5)
		} else {
			TakeoutCancelQueue = delayqueue.NewQueue(TAKEOUT, cache.Global.GetClient(), consumer.TakeoutCancelFunc).WithConcurrent(5)
		}
		TakeoutCancelQueue.StartConsume()
	})
}
