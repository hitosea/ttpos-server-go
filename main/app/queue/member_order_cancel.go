package queue

import (
	"sync"
	"ttpos-server-go/app/queue/consumer"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/cache"

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
			memberOrderCancelQueue = delayqueue.NewQueueOnCluster(MEMBER_ORDER_CANCEL, cache.Global.GetClusterClient(), consumer.ProcessMemberOrderCancel).WithConcurrent(5)
		} else {
			memberOrderCancelQueue = delayqueue.NewQueue(MEMBER_ORDER_CANCEL, cache.Global.GetClient(), consumer.ProcessMemberOrderCancel).WithConcurrent(5)
		}
		service.Queue.RegisterMemberOrderCancelQueue(memberOrderCancelQueue)
		memberOrderCancelQueue.StartConsume()
	})
}

func GetMemberOrderCancelQueue() *delayqueue.DelayQueue {
	return memberOrderCancelQueue
}
