package member_resp

import "ttpos-server-go/app/dto"

type MemberMarketingActivityPrizeResp struct {
	Uuid      uint64 `json:"uuid"`       // 奖品UUID
	PrizeName string `json:"prize_name"` // 奖品名称
}

type MemberMarketingActivityInfoResp struct {
	Uuid       uint64                             `json:"uuid"`        // 活动UUID
	LocaleName dto.LocaleResponse                 `json:"locale_name"` // 活动名称
	LocaleDesc dto.LocaleResponse                 `json:"locale_desc"` // 活动描述
	StartTime  int64                              `json:"start_time"`  // 活动开始时间
	EndTime    int64                              `json:"end_time"`    // 活动结束时间
	Status     int                                `json:"status"`      // 活动状态 0-未开始 1-进行中 2-已结束
	Prizes     []MemberMarketingActivityPrizeResp `json:"prizes"`      // 奖品列表
}

type MemberMarketingActivityListsResp struct {
	List []MemberMarketingActivityInfoResp `json:"list"` // 活动列表
}
