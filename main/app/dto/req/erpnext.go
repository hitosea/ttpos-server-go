package req

type ErpnextSiteCompanyReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`
	CompanyName string `form:"company_name" json:"company_name" binding:"omitempty"`
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"omitempty"`
}

type InitShopReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`       // 站点编码
	ShopName    string `form:"shop_name" json:"shop_name" binding:"required"`       // 店铺名称
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"required"` // 公司缩写编码
	ShopUuid    uint64 `form:"shop_uuid" json:"shop_uuid" binding:"required"`       // 店铺UUID
	AdminUuid   uint64 `form:"admin_uuid" json:"admin_uuid" binding:"required"`     // 管理员UUID
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
	SiteCode string `form:"site_code" json:"site_code" binding:"required"` // 站点编码
	Branch   string `form:"branch" json:"branch" binding:"required"`       // 分支名称
}
