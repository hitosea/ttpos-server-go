package req

import "ttpos-bmp/app/ttpos-erp/api/selling"

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
	SiteCode        string `form:"site_code" json:"site_code" binding:"required"`       // 站点编码
	CompanyAbbr     string `form:"company_abbr" json:"company_abbr" binding:"required"` // 公司缩写编码
	CompanyUuid     uint64 `form:"company_uuid" json:"company_uuid" binding:"required"` // 公司UUID
	HeadquarterAbbr string `form:"headquarter_abbr" json:"headquarter_abbr"`            // 总部简称
}

type GetUomListReq struct {
	SiteCode       string `form:"site_code" json:"site_code" binding:"required"`               // 站点编码
	Branch         string `form:"branch" json:"branch" binding:"required"`                     // 分支名称
	CompanyAbbr    string `form:"company_abbr" json:"company_abbr" binding:"required"`         // 公司缩写编码
	UomName        string `form:"uom_name" json:"uom_name" binding:"required"`                 // 单位名称
	AliasName      string `form:"alias_name" json:"alias_name" binding:"required"`             // 单位别名
	SubCompanyAbbr string `form:"sub_company_abbr" json:"sub_company_abbr" binding:"required"` // 连锁子店公司简称
}

type GetAttributeListReq struct {
	SiteCode       string `form:"site_code" json:"site_code" binding:"required"`               // 站点编码
	Branch         string `form:"branch" json:"branch" binding:"required"`                     // 分支名称
	CompanyAbbr    string `form:"company_abbr" json:"company_abbr" binding:"required"`         // 公司缩写编码
	AttributeName  string `form:"attribute_name" json:"attribute_name" binding:"required"`     // 属性名称
	AliasName      string `form:"alias_name" json:"alias_name" binding:"required"`             // 属性别名
	SubCompanyAbbr string `form:"sub_company_abbr" json:"sub_company_abbr" binding:"required"` // 连锁子店公司简称
}

type SaveErpFlavorReq struct {
	SiteCode       string                  `form:"site_code" json:"site_code" binding:"required"`               // 站点编码
	GroupName      string                  `form:"group_name" json:"group_name" binding:"required"`             // 规格组名称 companyabbr-Specifications
	GroupAliasName string                  `form:"group_alias_name" json:"group_alias_name" binding:"required"` // 规格组别名
	Branch         string                  `form:"branch" json:"branch" binding:"required"`                     // 分支名称
	CompanyAbbr    string                  `form:"company_abbr" json:"company_abbr" binding:"required"`         // 公司缩写编码
	ValueList      []SaveErpFlavorValueReq `form:"value_list" json:"value_list" binding:"required"`             // 规格值列表
}

type SaveErpFlavorValueReq struct {
	ValueName      string `form:"value_name" json:"value_name" binding:"required"`             // 规格值名称
	ValueAliasName string `form:"value_alias_name" json:"value_alias_name" binding:"required"` // 规格值别名
}

type GetErpFlavorListReq struct {
	SiteCode      string `form:"site_code" json:"site_code" binding:"required"`           // 站点编码
	Branch        string `form:"branch" json:"branch" binding:"required"`                 // 分支名称
	CompanyAbbr   string `form:"company_abbr" json:"company_abbr" binding:"required"`     // 公司缩写编码
	AttributeName string `form:"attribute_name" json:"attribute_name" binding:"required"` // 属性名称
	AliasName     string `form:"alias_name" json:"alias_name" binding:"required"`         // 属性别名
}

type SyncUomAndAttributeReq struct {
	SiteCode string `form:"site_code" json:"site_code" binding:"required"` // 站点编码
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

type DeleteUomReq struct {
	SiteCode    string `form:"site_code" json:"site_code" binding:"required"`       // 站点编码
	CompanyAbbr string `form:"company_abbr" json:"company_abbr" binding:"required"` // 公司缩写编码
	Branch      string `form:"branch" json:"branch" binding:"required"`             // 分支名称
	UomName     string `form:"uom_name" json:"uom_name" binding:"required"`         // 单位名称
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

type ClosePosEntryReq struct {
	SiteCode            string                `form:"site_code" json:"site_code" binding:"required"`                           // 站点编码
	PosProfileName      string                `form:"pos_profile_name" json:"pos_profile_name" binding:"required"`             // Pos Profile名称
	PosOpenEntryName    string                `form:"pos_open_entry_name" json:"pos_open_entry_name" binding:"required"`       // 开账名称
	PeriodEndDate       int64                 `form:"period_end_date" json:"period_end_date" binding:"required"`               // 结束日期
	ClosePosEntryDetail []ClosePosEntryDetail `form:"close_pos_entry_detail" json:"close_pos_entry_detail" binding:"required"` // 关账详情
}

type ClosePosEntryDetail struct {
	ModeOfPayment string  `form:"mode_of_payment" json:"mode_of_payment" binding:"required"` // 支付方式
	OpeningAmount float64 `form:"opening_amount" json:"opening_amount" binding:"required"`   // 开账金额
	ClosingAmount float64 `form:"closing_amount" json:"closing_amount" binding:"required"`   // 关账金额
}

type SavePosInvoiceReq struct {
	SiteCode         string                       `form:"site_code" json:"site_code" binding:"required"`                     // 站点编码
	OrderNo          string                       `form:"order_no" json:"order_no" binding:"required"`                       // 订单号
	OpenPosEntryName string                       `form:"open_pos_entry_name" json:"open_pos_entry_name" binding:"required"` // 开账名称
	PostingDatetime  int64                        `form:"posting_datetime" json:"posting_datetime" binding:"required"`       // 过账日期时间.订单结账时间
	CustomerUuid     string                       `form:"customer_uuid" json:"customer_uuid" binding:"required"`             // 会员uuid
	Items            []*selling.PosInvoiceItem    `form:"items" json:"items" binding:"required"`                             // 商品项目列表 必填
	MaterialItems    []*selling.PosInvoiceItem    `form:"material_items" json:"material_items" binding:"required"`           // 原材料项目列表 必填
	Taxes            []*selling.PosInvoiceTax     `form:"taxes" json:"taxes" binding:"required"`                             // 税费列表 必填
	Payments         []*selling.PosInvoicePayment `form:"payments" json:"payments" binding:"required"`                       // 付款列表 必填

	AmendedProductsInvoiceName string `form:"amended_products_invoice_name" json:"amended_products_invoice_name" binding:"required"` // 反结账后，重新结账时填写原商品发票名称
	AmendedMaterialInvoiceName string `form:"amended_material_invoice_name" json:"amended_material_invoice_name" binding:"required"` // 反结账后，重新结账时填写原材料发票名称

}

type CancelPosInvoiceReq struct {
	ProductsInvoiceName string `form:"products_invoice_name" json:"products_invoice_name" ` // 商品销售发票  同步模式填
	MaterialInvoiceName string `form:"material_invoice_name" json:"material_invoice_name" ` // 材料销售发票 同步模式填
	OpenPosEntryName    string `form:"open_pos_entry_name" json:"open_pos_entry_name"`      // 开账名称,异步模式必填
	OrderNo             string `form:"order_no" json:"order_no"`                            // 销售订单号,异步模式必填

}

type ReturnPosInvoiceReq struct {
	SiteCode         string                       `form:"site_code" json:"site_code" binding:"required"`                     // 站点编码
	OrderNo          string                       `form:"order_no" json:"order_no" binding:"required"`                       // 销售订单号
	OpenPosEntryName string                       `form:"open_pos_entry_name" json:"open_pos_entry_name" binding:"required"` // 开账名称
	PostingDatetime  int64                        `form:"posting_datetime" json:"posting_datetime" binding:"required"`       // 过账日期时间
	CompanyAbbr      string                       `form:"company_abbr" json:"company_abbr" binding:"required"`               // 公司缩写编码
	InvoiceName      string                       `form:"invoice_name" json:"invoice_name" binding:"required"`               // 销售发票
	Items            []*selling.PosInvoiceItem    `form:"items" json:"items" binding:"required"`                             // 商品项目列表 必填
	Taxes            []*selling.PosInvoiceTax     `form:"taxes" json:"taxes" binding:"required"`                             // 税费列表 必填
	Payments         []*selling.PosInvoicePayment `form:"payments" json:"payments" binding:"required"`                       // 付款列表 必填
}

type GetPaymentMethodListReq struct {
	CompanyUuid uint64 `form:"company_uuid" json:"company_uuid" binding:"required"` // 公司UUID
}

type AddPaymentMethodReq struct {
	CompanyUuid        uint64   `json:"company_uuid" binding:"required"`    // 公司UUID
	ErpnextPayment     string   `json:"erpnext_payment" binding:"required"` // 商家支付方式
	Name               string   `json:"name" binding:"required"`            // 支付方式名称
	Fee                *float64 `json:"fee" binding:"required"`             // 手续费
	Sort               *int     `json:"sort" binding:"required"`            // 排序
	Status             *int     `json:"status" binding:"required"`          // 状态: 0-禁用 1-启用
	CheckoutShow       []string `json:"checkout_show"`                      // 结账显示，可选cashier、assistant
	MemberRechargeShow []string `json:"member_recharge_show"`               // 会员充值显示，可选cashier
}

type CreateSupplierReq struct {
	SiteCode     string
	SupplierName string // 供应商名称
	CompanyAbbr  string // 公司缩写编码
	Branch       string // 分支
	Disabled     bool   // 是否禁用
}

type UpdateSupplierReq struct {
	CreateSupplierReq
	Name string // 供应商erp_code
}

type DeleteSupplierReq struct {
	Name     string // 供应商erp_code
	SiteCode string // 站点编码
}

type CreateErpnextWarehouseReq struct {
	SiteCode      string // 站点编码
	WarehouseName string // 仓库名称
	AliasName     string // 别名
	CompanyAbbr   string // 公司简称
	Branch        string // 分支
	Disabled      bool   // 是否禁用
	WarehouseType string // 仓库类型
}

type UpdateErpnextWarehouseReq struct {
	CreateErpnextWarehouseReq
	Name string // erp_Code
}

type DeleteErpnextWarehouseReq struct {
	SiteCode string // 站点编码
	Name     string // erp_code
}

type GetErpnextWarehouseListReq struct {
	SiteCode       string // 站点编码
	CompanyAbbr    string
	Branch         string
	SubCompanyAbbr string // 连锁子店公司简称
}

type GetErpnextSupplierListReq struct {
	SiteCode       string // 站点编码
	CompanyAbbr    string
	Branch         string
	SubCompanyAbbr string // 连锁子店公司简称
}

type GetHeadquarterMaterialListReq struct {
}

type GetErpSauceListReq struct {
	SiteCode       string // 站点编码
	CompanyAbbr    string
	Branch         string
	SubCompanyAbbr string // 连锁子店公司简称
}
