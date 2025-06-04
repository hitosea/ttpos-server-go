package member_req

type MemberMarketingActivityReq struct {
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid"` // 集团ID
}
