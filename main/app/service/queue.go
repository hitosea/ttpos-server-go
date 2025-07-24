package service

import "github.com/hdt3213/delayqueue"

type queue struct {
	MemberOrderCancelQueue *delayqueue.DelayQueue
	TakeoutCancelQueue     *delayqueue.DelayQueue
}

var Queue = new(queue)

func (s *queue) RegisterMemberOrderCancelQueue(q *delayqueue.DelayQueue) {
	s.MemberOrderCancelQueue = q
}

func (s *queue) RegisterTakeoutCancelQueue(q *delayqueue.DelayQueue) {
	s.TakeoutCancelQueue = q
}
