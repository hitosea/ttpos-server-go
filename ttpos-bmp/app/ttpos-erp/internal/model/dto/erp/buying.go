package erp

// Supplier 结构体，表示供应商信息
type Supplier struct {
	Name                                               string                  `json:"name,omitempty"`                                                     // 名称
	Owner                                              string                  `json:"owner,omitempty"`                                                    // 拥有者
	Creation                                           string                  `json:"creation,omitempty"`                                                 // 创建时间
	Modified                                           string                  `json:"modified,omitempty"`                                                 // 修改时间
	ModifiedBy                                         string                  `json:"modified_by,omitempty"`                                              // 修改人
	Docstatus                                          int                     `json:"docstatus,omitempty"`                                                // 单据状态
	Idx                                                int                     `json:"idx,omitempty"`                                                      // 索引
	NamingSeries                                       string                  `json:"naming_series,omitempty"`                                            // 编码规则
	SupplierName                                       string                  `json:"supplier_name,omitempty"`                                            // 供应商名称
	Country                                            string                  `json:"country,omitempty"`                                                  // 国家
	SupplierType                                       string                  `json:"supplier_type,omitempty"`                                            // 供应商类型
	IsTransporter                                      int                     `json:"is_transporter,omitempty"`                                           // 是否承运商
	IsInternalSupplier                                 int                     `json:"is_internal_supplier,omitempty"`                                     // 是否内部供应商
	RepresentsCompany                                  string                  `json:"represents_company,omitempty"`                                       // 代表公司
	Language                                           string                  `json:"language,omitempty"`                                                 // 语言
	AllowPurchaseInvoiceCreationWithoutPurchaseOrder   int                     `json:"allow_purchase_invoice_creation_without_purchase_order,omitempty"`   // 允许无采购订单创建采购发票
	AllowPurchaseInvoiceCreationWithoutPurchaseReceipt int                     `json:"allow_purchase_invoice_creation_without_purchase_receipt,omitempty"` // 允许无收货单创建采购发票
	IsFrozen                                           int                     `json:"is_frozen,omitempty"`                                                // 是否冻结
	Disabled                                           int                     `json:"disabled,omitempty"`                                                 // 是否禁用
	WarnRfqs                                           int                     `json:"warn_rfqs,omitempty"`                                                // 询价单警告
	WarnPos                                            int                     `json:"warn_pos,omitempty"`                                                 // 采购订单警告
	PreventRfqs                                        int                     `json:"prevent_rfqs,omitempty"`                                             // 阻止询价单
	PreventPos                                         int                     `json:"prevent_pos,omitempty"`                                              // 阻止采购订单
	OnHold                                             int                     `json:"on_hold,omitempty"`                                                  // 是否暂停
	HoldType                                           string                  `json:"hold_type,omitempty"`                                                // 暂停类型
	Doctype                                            string                  `json:"doctype,omitempty"`                                                  // 单据类型
	Accounts                                           []interface{}           `json:"accounts,omitempty"`                                                 // 账户信息
	PortalUsers                                        []interface{}           `json:"portal_users,omitempty"`                                             // 门户用户
	Companies                                          []AllowedToTransactWith `json:"companies,omitempty"`                                                // 允许交易的公司
}

// AllowedToTransactWith 结构体，表示允许交易的公司
type AllowedToTransactWith struct {
	Name        string `json:"name,omitempty"`        // 名称
	Owner       string `json:"owner,omitempty"`       // 拥有者
	Creation    string `json:"creation,omitempty"`    // 创建时间
	Modified    string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy  string `json:"modified_by,omitempty"` // 修改人
	Docstatus   int    `json:"docstatus,omitempty"`   // 单据状态
	Idx         int    `json:"idx,omitempty"`         // 索引
	Company     string `json:"company,omitempty"`     // 公司
	Parent      string `json:"parent,omitempty"`      // 父级
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
	Doctype     string `json:"doctype,omitempty"`     // 单据类型
}
