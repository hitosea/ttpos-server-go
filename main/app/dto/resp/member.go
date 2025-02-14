package resp

type MemberLevel struct {
	Uuid       uint64 `json:"uuid"`        // 等级Uuid
	Name       string `json:"name"`        // 等级名称
	Priority   int    `json:"priority"`    // 等级优先级
	CreateTime int64  `json:"create_time"` // 创建时间
}

type MemberLevelList struct {
	List []MemberLevel `json:"list"`
}

type SearchMember struct {
	Uuid     uint64 `json:"uuid"`     // 会员Uuid
	Nickname string `json:"nickname"` // 会员昵称
	Phone    string `json:"phone"`    // 手机
}

type SearchMemberList struct {
	List []SearchMember `json:"list"`
}

type RechargeMember struct {
	Uuid      uint64  `json:"uuid"`       // 会员Uuid
	Nickname  string  `json:"nickname"`   // 会员昵称
	CardName  string  `json:"card_name"`  // 会员卡名称
	LevelName string  `json:"level_name"` // 会员等级
	Balance   float64 `json:"balance"`    // 会员余额
	Points    float64 `json:"points"`     // 会员积分
}
