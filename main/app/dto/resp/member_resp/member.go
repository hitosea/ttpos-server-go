package member_resp

type MemberLoginInfoResp struct {
	CompanyUuid uint64   `json:"company_uuid"` // 集团UUID
	AreaCode    []string `json:"area_code"`    // 地区码数组，如 ["+86", "+1", "+44"]
}

// type MemberLoginInfoResp struct {
// 	MemberUuid uint64 `json:"member_uuid"` // 会员UUID
// 	Nickname   string `json:"nickname"`    // 会员昵称
// 	Phone      string `json:"phone"`       // 手机号
// }
