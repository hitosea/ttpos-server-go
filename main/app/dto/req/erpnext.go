package req

type ErpnextSiteCompanyReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`
	CompanyName string `form:"company_name" json:"company_name" binding:"omitempty"`
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"omitempty"`
}
