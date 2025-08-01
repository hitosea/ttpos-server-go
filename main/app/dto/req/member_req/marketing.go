package member_req

type MarketingActivityReq struct {
	Uuid uint64 `form:"uuid"` // 活动UUID
}

type MarketingActivityDetailReq struct {
	Uuid uint64 `form:"uuid"` // 活动UUID
}
