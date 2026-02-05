package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/pkg/rocketmq"

	"github.com/hdt3213/delayqueue"
)

// IOperationDurationQueue 操作耗时记录队列接口
type IOperationDurationQueue interface {
	Push(record *req.OperationDurationRecord)
	Start()
	Stop()
	GetQueueLen() int
	GetInstanceID() string
}

type queue struct {
	MemberOrderCancelQueue *delayqueue.DelayQueue
	Manager                *rocketmq.Manager
	OperationDurationQueue IOperationDurationQueue // 操作耗时记录队列
}

var Queue = new(queue)

func (s *queue) RegisterMemberOrderCancelQueue(q *delayqueue.DelayQueue) {
	s.MemberOrderCancelQueue = q
}

func (s *queue) RegisterManager(q *rocketmq.Manager) {
	s.Manager = q
}

// RegisterOperationDurationQueue 注册操作耗时记录队列
func (s *queue) RegisterOperationDurationQueue(q IOperationDurationQueue) {
	s.OperationDurationQueue = q
}
