package member_resp

import "ttpos-server-go/app/dto"

type MemberLoginInfoResp struct {
	CompanyUuid  uint64             `json:"company_uuid"`  // 集团UUID
	CompanyName  string             `json:"company_name"`  // 集团名称
	Logo         string             `json:"logo"`          // 集团logo
	AreaCode     []string           `json:"area_code"`     // 地区码数组，如 ["+86", "+1", "+44"]
	LanguageList []dto.LanguageItem `json:"language_list"` // 语言列表
}
