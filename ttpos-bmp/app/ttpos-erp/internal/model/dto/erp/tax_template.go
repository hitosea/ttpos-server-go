// Package erp 提供ERP系统相关的数据传输对象
// 包含税费模板相关结构体
package erp

// PurchaseTaxesAndChargesTemplate 采购税费模板
// 参考: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/purchase_taxes_and_charges_template/purchase_taxes_and_charges_template.json
type PurchaseTaxesAndChargesTemplate struct {
	Name        string                     `json:"name,omitempty"`         // 模板名称（主键）
	Title       string                     `json:"title,omitempty"`        // 标题 (必填)
	IsDefault   int                        `json:"is_default,omitempty"`   // 是否默认 (0/1)
	Disabled    int                        `json:"disabled,omitempty"`     // 是否禁用 (0/1)
	Company     string                     `json:"company,omitempty"`      // 所属公司 (必填, Link to Company)
	TaxCategory string                     `json:"tax_category,omitempty"` // 税类别 (Link to Tax Category)
	Taxes       []*PurchaseTaxesAndCharges `json:"taxes,omitempty"`        // 税费明细 (Table: Purchase Taxes and Charges)
}

// SalesTaxesAndChargesTemplate 销售税费模板
// 参考: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/sales_taxes_and_charges_template/sales_taxes_and_charges_template.json
type SalesTaxesAndChargesTemplate struct {
	Name        string                  `json:"name,omitempty"`         // 模板名称（主键）
	Title       string                  `json:"title,omitempty"`        // 标题 (必填)
	IsDefault   int                     `json:"is_default,omitempty"`   // 是否默认 (0/1)
	Disabled    int                     `json:"disabled,omitempty"`     // 是否禁用 (0/1)
	Company     string                  `json:"company,omitempty"`      // 所属公司 (必填, Link to Company)
	TaxCategory string                  `json:"tax_category,omitempty"` // 税类别 (Link to Tax Category)
	Taxes       []*SalesTaxesAndCharges `json:"taxes,omitempty"`        // 税费明细 (Table: Sales Taxes and Charges)
}

// PurchaseTaxesAndCharges 采购税费明细（taxes 子表）
// 参考: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/purchase_taxes_and_charges/purchase_taxes_and_charges.json
type PurchaseTaxesAndCharges struct {
	// 基础字段
	Name string `json:"name,omitempty"` // 行名称
	Idx  int    `json:"idx,omitempty"`  // 行索引

	// 税费类型
	Category     string `json:"category,omitempty"`       // 类别: "Total", "Valuation", "Valuation and Total" (必填, 默认 "Total")
	AddDeductTax string `json:"add_deduct_tax,omitempty"` // 增/减: "Add", "Deduct" (必填, 默认 "Add")
	ChargeType   string `json:"charge_type,omitempty"`    // 计费类型 (必填, 默认 "On Net Total")
	// ChargeType 可选值:
	//   - "Actual": 实际金额
	//   - "On Net Total": 基于净额
	//   - "On Previous Row Amount": 基于前一行金额
	//   - "On Previous Row Total": 基于前一行总计
	//   - "On Item Quantity": 基于商品数量

	RowId string `json:"row_id,omitempty"` // 参考行号 (当 ChargeType 涉及 Previous Row 时使用)

	// 会计科目
	AccountHead string `json:"account_head,omitempty"` // 会计科目 (必填, Link to Account)
	Description string `json:"description,omitempty"`  // 描述 (必填)
	CostCenter  string `json:"cost_center,omitempty"`  // 成本中心 (Link to Cost Center)

	// 税率和金额
	Rate                         float64 `json:"rate,omitempty"`                             // 税率
	TaxAmount                    float64 `json:"tax_amount,omitempty"`                       // 税额
	Total                        float64 `json:"total,omitempty"`                            // 总计
	TaxAmountAfterDiscountAmount float64 `json:"tax_amount_after_discount_amount,omitempty"` // 折扣后税额

	// 基础货币金额
	BaseTaxAmount                    float64 `json:"base_tax_amount,omitempty"`                       // 基础税额
	BaseTotal                        float64 `json:"base_total,omitempty"`                            // 基础总计
	BaseTaxAmountAfterDiscountAmount float64 `json:"base_tax_amount_after_discount_amount,omitempty"` // 基础折扣后税额

	// 打印相关
	IncludedInPrintRate  int `json:"included_in_print_rate,omitempty"`  // 是否包含在打印价格中 (0/1)
	IncludedInPaidAmount int `json:"included_in_paid_amount,omitempty"` // 是否包含在已付金额中 (0/1)

	// 明细
	ItemWiseTaxDetail string `json:"item_wise_tax_detail,omitempty"` // 按商品税费明细 (JSON Text)

	// 关联字段
	Parent      string `json:"parent,omitempty"`      // 父文档
	Parentfield string `json:"parentfield,omitempty"` // 父字段名
	Parenttype  string `json:"parenttype,omitempty"`  // 父文档类型
	Doctype     string `json:"doctype,omitempty"`     // 文档类型
}

// SalesTaxesAndCharges 销售税费明细
// 参考: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/sales_taxes_and_charges/sales_taxes_and_charges.json
type SalesTaxesAndCharges struct {
	// 基础字段
	Name string `json:"name,omitempty"` // 行名称
	Idx  int    `json:"idx,omitempty"`  // 行索引

	// 税费类型
	ChargeType string `json:"charge_type,omitempty"` // 计费类型 (必填, 默认 "On Net Total")
	RowId      string `json:"row_id,omitempty"`      // 参考行号

	// 会计科目
	AccountHead string `json:"account_head,omitempty"` // 会计科目 (必填, Link to Account)
	Description string `json:"description,omitempty"`  // 描述 (必填)
	CostCenter  string `json:"cost_center,omitempty"`  // 成本中心 (Link to Cost Center)

	// 税率和金额
	Rate                         float64 `json:"rate,omitempty"`                             // 税率
	TaxAmount                    float64 `json:"tax_amount,omitempty"`                       // 税额
	Total                        float64 `json:"total,omitempty"`                            // 总计
	TaxAmountAfterDiscountAmount float64 `json:"tax_amount_after_discount_amount,omitempty"` // 折扣后税额

	// 基础货币金额
	BaseTaxAmount                    float64 `json:"base_tax_amount,omitempty"`                       // 基础税额
	BaseTotal                        float64 `json:"base_total,omitempty"`                            // 基础总计
	BaseTaxAmountAfterDiscountAmount float64 `json:"base_tax_amount_after_discount_amount,omitempty"` // 基础折扣后税额

	// 打印相关
	IncludedInPrintRate int `json:"included_in_print_rate,omitempty"` // 是否包含在打印价格中 (0/1)

	// 明细
	ItemWiseTaxDetail string `json:"item_wise_tax_detail,omitempty"` // 按商品税费明细 (JSON Text)

	// 关联字段
	Parent      string `json:"parent,omitempty"`      // 父文档
	Parentfield string `json:"parentfield,omitempty"` // 父字段名
	Parenttype  string `json:"parenttype,omitempty"`  // 父文档类型
	Doctype     string `json:"doctype,omitempty"`     // 文档类型
}
