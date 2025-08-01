package service

import "github.com/hdt3213/delayqueue"

type queue struct {
	MemberOrderCancelQueue *delayqueue.DelayQueue
}

var Queue = new(queue)

func (s *queue) RegisterMemberOrderCancelQueue(q *delayqueue.DelayQueue) {
	s.MemberOrderCancelQueue = q
}
