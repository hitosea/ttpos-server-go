package member_req

type MemberLoginInfoReq struct {
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid"` // 集团ID
}

type MemberSendCodeReq struct {
	Phone       string `json:"phone" form:"phone"`               // 手机号
	AreaCode    string `json:"area_code" form:"area_code"`       // 区号
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid"` // 集团ID
}
