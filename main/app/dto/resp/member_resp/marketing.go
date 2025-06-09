package member_resp

type MemberMarketingActivityResp struct {
	Name       string `json:"name"`       // 活动名称
	Desc       string `json:"desc"`       // 活动描述
	QrCodeCode string `json:"qr_code"`    // 活动二维码-base64
	StartTime  int64  `json:"start_time"` // 活动开始时间
	EndTime    int64  `json:"end_time"`   // 活动结束时间
	IsInvalid  int    `json:"is_invalid"` // 活动状态 0-未失效 1-已失效
}

type MemberInfoResp struct {
	Uuid     uint64 `json:"uuid"`     // 会员UUID
	Nickname string `json:"nickname"` // 会员昵称
	Phone    string `json:"phone"`    // 会员手机号
}

type CompanyInfoResp struct {
	Uuid uint64 `json:"uuid"` // 公司UUID
	Name string `json:"name"` // 公司名称
	Logo string `json:"logo"` // 公司logo
}

type MemberMarketingActivityListResp struct {
	List       []MemberMarketingActivityResp `json:"list"`        // 活动列表
	MemberInfo MemberInfoResp                `json:"member_info"` // 会员信息
	Company    CompanyInfoResp               `json:"company"`     // 公司信息
}
