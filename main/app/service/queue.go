package service

import (
	"ttpos-server-go/pkg/rocketmq"

	"github.com/hdt3213/delayqueue"
)

type queue struct {
	MemberOrderCancelQueue *delayqueue.DelayQueue
	Manager                *rocketmq.Manager
}

var Queue = new(queue)

func (s *queue) RegisterMemberOrderCancelQueue(q *delayqueue.DelayQueue) {
	s.MemberOrderCancelQueue = q
}

func (s *queue) RegisterManager(q *rocketmq.Manager) {
	s.Manager = q
}
