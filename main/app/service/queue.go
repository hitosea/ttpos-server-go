package service

import "github.com/hdt3213/delayqueue"

type queue struct {
	MemberOrderCancelQueue *delayqueue.DelayQueue
	TakeoutCancelQueue     *delayqueue.DelayQueue
}

var Queue = new(queue)

func (s *queue) RegistryMemberOrderCancelQueue(q *delayqueue.DelayQueue) {
	s.MemberOrderCancelQueue = q
}

func (s *queue) RegistryTakeoutCancelQueue(q *delayqueue.DelayQueue) {
	s.TakeoutCancelQueue = q
}
