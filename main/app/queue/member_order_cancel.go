package queue

import (
	"github.com/hdt3213/delayqueue"
	"sync"
	"ttpos-server-go/app/queue/consumer"
	"ttpos-server-go/pkg/cache"
)

var (
	MemberOrderCancelQueue *delayqueue.DelayQueue
	memberOrderInit        sync.Once
)

// InitMemberOrderCancel 初始化会员订单自动取消队列
func InitMemberOrderCancel() {
	memberOrderInit.Do(func() {
		if cache.Global.GetClusterClient() != nil {
			MemberOrderCancelQueue = delayqueue.NewQueueOnCluster(MEMBER_ORDER_CANCEL, cache.Global.GetClusterClient(), consumer.ProcessMemberOrderCancel).WithConcurrent(5)
		} else {
			MemberOrderCancelQueue = delayqueue.NewQueue(MEMBER_ORDER_CANCEL, cache.Global.GetClient(), consumer.ProcessMemberOrderCancel).WithConcurrent(5)
		}
		MemberOrderCancelQueue.StartConsume()
	})
}
