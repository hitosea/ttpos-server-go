package member_req

type CouponListReq struct {
	MemberUuid uint64 `form:"member_uuid"` // 会员ID
	IsHistory  int    `form:"is_history"`  // 是否历史 1: 是 0: 否
}
