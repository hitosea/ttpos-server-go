package member_req

import "ttpos-server-go/app/dto"

type CouponListReq struct {
	MemberUuid uint64 `form:"member_uuid"` // 会员ID
	IsHistory  int    `form:"is_history"`  // 是否历史 1: 是 0: 否
}

type PointsRecordListReq struct {
	MemberUuid uint64 `form:"member_uuid"` // 会员ID
	Type       int    `form:"type"`        // 积分类型 0:全部 1:奖励 2:消费
	dto.PageReq
}
