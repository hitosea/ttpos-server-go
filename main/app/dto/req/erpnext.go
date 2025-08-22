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

type SaveUomReq struct {
	SiteCode          string `form:"site_code" json:"site_code" binding:"required"`                       // 站点编码
	CompanyAbbr       string `form:"company_abbr" json:"company_abbr" binding:"required"`                 // 公司缩写编码
	Branch            string `form:"branch" json:"branch" binding:"required"`                             // 分支名称
	UomName           string `form:"uom_name" json:"uom_name" binding:"required"`                         // 单位名称
	AliasName         string `form:"alias_name" json:"alias_name" binding:"required"`                     // 单位别名
	MustBeWholeNumber bool   `form:"must_be_whole_number" json:"must_be_whole_number" binding:"required"` // 是否必须为整数
}

type SaveAttributeValueReq struct {
	AttributeValue string `form:"attribute_value" json:"attribute_value" binding:"required"` // 属性值
	Abbr           string `form:"abbr" json:"abbr" binding:"required"`                       // 属性别名
}

type SaveAttributeReq struct {
	SiteCode           string                  `form:"site_code" json:"site_code" binding:"required"`                       // 站点编码
	CompanyAbbr        string                  `form:"company_abbr" json:"company_abbr" binding:"required"`                 // 公司缩写编码
	Branch             string                  `form:"branch" json:"branch" binding:"required"`                             // 分支名称
	AttributeName      string                  `form:"attribute_name" json:"attribute_name" binding:"required"`             // 属性名称
	AliasName          string                  `form:"alias_name" json:"alias_name" binding:"required"`                     // 属性别名
	AttributeValueList []SaveAttributeValueReq `form:"attribute_value_list" json:"attribute_value_list" binding:"required"` // 属性值列表
}

type OpenPosEntryReq struct {
	SiteCode           string               `form:"site_code" json:"site_code" binding:"required"`                         // 站点编码
	PosProfileName     string               `form:"pos_profile_name" json:"pos_profile_name" binding:"required"`           // Pos Profile名称
	CashierEmail       string               `form:"cashier_email" json:"cashier_email" binding:"required"`                 // 收银员邮箱
	CompanyAbbr        string               `form:"company_abbr" json:"company_abbr" binding:"required"`                   // 公司缩写编码
	PeriodStartDate    int64                `form:"period_start_date" json:"period_start_date" binding:"required"`         // 开始日期
	OpenPosEntryDetail []OpenPosEntryDetail `form:"open_pos_entry_detail" json:"open_pos_entry_detail" binding:"required"` // 开账详情
	Branch             string               `form:"branch" json:"branch" binding:"required"`                               // 分支名称
}

type OpenPosEntryDetail struct {
	ModeOfPayment string  `form:"mode_of_payment" json:"mode_of_payment" binding:"required"` // 支付方式
	OpeningAmount float64 `form:"opening_amount" json:"opening_amount" binding:"required"`   // 开账金额
}
