package resp

type ErpnextSiteCompanyResp struct {
	List []ErpnextSiteCompany `json:"list"`
}

type ErpnextSiteCompany struct {
	CompanyName string `json:"company_name"`
	CompanyAbbr string `json:"company_abbr"`
}
