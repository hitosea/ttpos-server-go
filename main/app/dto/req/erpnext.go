package req

type ErpnextSiteCompanyReq struct {
	SiteCode      string `form:"site_code" json:"site_code" binding:"required"`
	CompanyName   string `form:"company_name" json:"company_name" binding:"omitempty"`     // 筛选公司名称
	CompanyAbbr   string `form:"company_abbr" json:"company_abbr" binding:"omitempty"`     // 筛选公司缩写编码
	ParentCompany string `form:"parent_company" json:"parent_company" binding:"omitempty"` // 筛选父公司名称
}

type ErpnextSitePosProfileReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"required"` // 公司缩写编码
}

type InitShopReq struct {
	SiteCode           string `form:"site_code" json:"site_code" binding:"required"`               // 站点编码
	CompanyAbbr        string `form:"company_abbr" json:"company_abbr" binding:"required"`         // 公司缩写编码
	CompanyUuid        uint64 `form:"company_uuid" json:"company_uuid" binding:"required"`         // 公司UUID
	DefaultCompanyAbbr string `form:"default_company_abbr" json:"default_company_abbr"`            // 默认公司缩写编码，用于同步单位和属性
	PosProfileName     string `form:"pos_profile_name" json:"pos_profile_name" binding:"required"` // Pos Profile名称
}

type GetUomListReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`       // 站点编码
	Branch      string `form:"branch" json:"branch" binding:"required"`             // 分支名称
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"required"` // 公司缩写编码
	UomName     string `form:"uom_name" json:"uom_name" binding:"required"`         // 单位名称
	AliasName   string `form:"alias_name" json:"alias_name" binding:"required"`     // 单位别名
}

type GetAttributeListReq struct {
	SiteCode      string `form:"site_code" json:"site_code" binding:"required"`           // 站点编码
	Branch        string `form:"branch" json:"branch" binding:"required"`                 // 分支名称
	CompanyAbbr   string `form:"company_abbr" json:"company_abbr" binding:"required"`     // 公司缩写编码
	AttributeName string `form:"attribute_name" json:"attribute_name" binding:"required"` // 属性名称
	AliasName     string `form:"alias_name" json:"alias_name" binding:"required"`         // 属性别名
}

type SyncUomAndAttributeReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`       // 站点编码
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"required"` // 初始化同步单位和属性时，默认公司缩写编码
}

type GetPosProfileListReq struct {
	SiteCode       string `form:"site_code" json:"site_code" binding:"required"`               // 站点编码
	Company        string `form:"company" json:"company" binding:"required"`                   // 公司名称
	CompanyAbbr    string `form:"company_abbr" json:"company_abbr" binding:"required"`         // 公司缩写编码
	PosProfileName string `form:"pos_profile_name" json:"pos_profile_name" binding:"required"` // Pos Profile名称
}

type ErpnextSiteAddLianLianPaymentReq struct {
	CompanyUuid uint64 `json:"company_uuid" binding:"required"` // 公司UUID
}
