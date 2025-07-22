package queue

import (
	"github.com/hdt3213/delayqueue"
	"sync"
	"ttpos-server-go/app/queue/consumer"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/cache"
)

var (
	takeoutCancelQueue *delayqueue.DelayQueue
	takeoutInit        sync.Once
)

// InitTakeoutCancel 初始化订单取消队列
func initTakeoutCancel() {
	takeoutInit.Do(func() {
		if cache.Global.GetClusterClient() != nil {
			takeoutCancelQueue = delayqueue.NewQueueOnCluster(TAKEOUT, cache.Global.GetClusterClient(), consumer.TakeoutCancelFunc).WithConcurrent(5)
		} else {
			takeoutCancelQueue = delayqueue.NewQueue(TAKEOUT, cache.Global.GetClient(), consumer.TakeoutCancelFunc).WithConcurrent(5)
		}
		service.Queue.RegistryTakeoutCancelQueue(takeoutCancelQueue)
		takeoutCancelQueue.StartConsume()
	})
}

func GetTakeoutCancelQueue() *delayqueue.DelayQueue {
	return takeoutCancelQueue
}
