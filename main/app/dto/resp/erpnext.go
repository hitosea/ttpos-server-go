package resp

type ErpnextSiteCompanyResp struct {
	List []ErpnextSiteCompany `json:"list"`
}

type ErpnextSiteCompany struct {
	CompanyName   string               `json:"company_name"`
	CompanyAbbr   string               `json:"company_abbr"`
	ParentCompany string               `json:"parent_company"`
	IsUsed        bool                 `json:"is_used"`  // 是否已被使用
	Children      []ErpnextSiteCompany `json:"children"` // 子公司列表，用于树形结构
}

type InitShopResp struct {
	BranchName string `json:"branch_name"`
	AdminEmail string `json:"admin_email"`
}

type GetUomListResp struct {
	List []UomInfo `json:"list"`
}

type UomInfo struct {
	UomName           string `json:"uom_name"`
	AliasName         string `json:"alias_name"`
	MustBeWholeNumber bool   `json:"must_be_whole_number"` // 是否必须为整数
	CompanyAbbr       string `json:"company_abbr"`
	Branch            string `json:"branch"`
}

type GetAttributeListResp struct {
	List []AttributeInfo `json:"list"`
}

type AttributeInfo struct {
	AttributeName      string               `json:"attribute_name"`
	AliasName          string               `json:"alias_name"`           // 属性别名
	CompanyAbbr        string               `json:"company_abbr"`         // 公司缩写编码
	Branch             string               `json:"branch"`               // 分支名称
	AttributeValueList []AttributeValueInfo `json:"attribute_value_list"` // 属性值列表
}

type AttributeValueInfo struct {
	AttributeValue string `json:"attribute_value"` // 属性值
	Abbr           string `json:"abbr"`            // 属性值缩写
}

type GetErpFlavorListResp struct {
	List []ErpFlavorInfo `json:"list"`
}

type ErpFlavorInfo struct {
	AttributeName      string               `json:"attribute_name"`
	AliasName          string               `json:"alias_name"`           // 属性别名
	CompanyAbbr        string               `json:"company_abbr"`         // 公司缩写编码
	Branch             string               `json:"branch"`               // 分支名称
	AttributeValueList []ErpFlavorValueInfo `json:"attribute_value_list"` // 属性值列表
}

type ErpFlavorValueInfo struct {
	AttributeValue string `json:"attribute_value"` // 属性值
	Abbr           string `json:"abbr"`            // 属性值缩写
}

type GetPosInvoiceErrorResp struct {
	ErrorScene string `json:"error_scene"` // 错误场景
	ItemCode   string `json:"item_code"`   // 物品编码
}

type GetPosProfileListResp struct {
	List []PosProfileInfo `json:"list"`
}

type PosProfileInfo struct {
	Name      string `json:"name"`
	Company   string `json:"company"`
	Branch    string `json:"branch"`
	Warehouse string `json:"warehouse"`
}

type GetPaymentMethodListResp struct {
	List []PaymentMethodInfo `json:"list"`
}

type PaymentMethodInfo struct {
	Name      string `json:"name"`       // 支付名称
	IsAddable bool   `json:"is_addable"` // 是否可添加
}
