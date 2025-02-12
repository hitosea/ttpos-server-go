package resp

type MemberLevel struct {
	Uuid       uint64 `json:"uuid"`        // 等级Uuid
	Name       string `json:"name"`        // 等级名称
	Priority   int    `json:"priority"`    // 等级优先级
	CreateTime int64  `json:"create_time"` // 创建时间
}

type SearchMember struct {
	Uuid     uint64 `json:"uuid"`     // 会员Uuid
	Nickname string `json:"nickname"` // 会员昵称
	Phone    string `json:"phone"`    // 手机
}
