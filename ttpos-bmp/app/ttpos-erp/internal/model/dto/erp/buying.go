// Package erp 提供ERP系统相关的数据传输对象
// 包含供应商、采购订单、销售订单、付款计划等业务实体
package erp

// Supplier 结构体，表示供应商信息
type Supplier struct {
	// 基础字段
	Name       string `json:"name,omitempty" validate:"required"` // 名称
	Owner      string `json:"owner,omitempty"`                    // 拥有者
	Creation   string `json:"creation,omitempty"`                 // 创建时间
	Modified   string `json:"modified,omitempty"`                 // 修改时间
	ModifiedBy string `json:"modified_by,omitempty"`              // 修改人
	Docstatus  int    `json:"docstatus,omitempty"`                // 单据状态
	Idx        int    `json:"idx,omitempty"`                      // 索引
	Doctype    string `json:"doctype,omitempty"`                  // 单据类型

	// 供应商信息
	SupplierName      string `json:"supplier_name,omitempty" validate:"required"` // 供应商名称
	Country           string `json:"country,omitempty"`                           // 国家
	SupplierType      string `json:"supplier_type,omitempty"`                     // 供应商类型
	Language          string `json:"language,omitempty"`                          // 语言
	RepresentsCompany string `json:"represents_company,omitempty"`                // 代表公司

	// 编码规则
	NamingSeries string `json:"naming_series,omitempty"` // 编码规则

	// 状态标识
	IsTransporter      bool `json:"is_transporter,omitempty"`       // 是否承运商
	IsInternalSupplier bool `json:"is_internal_supplier,omitempty"` // 是否内部供应商
	IsFrozen           bool `json:"is_frozen,omitempty"`            // 是否冻结
	Disabled           bool `json:"disabled"`                       // 是否禁用
	OnHold             bool `json:"on_hold,omitempty"`              // 是否暂停

	// 警告和阻止设置
	WarnRfqs    bool `json:"warn_rfqs,omitempty"`    // 询价单警告
	WarnPos     bool `json:"warn_pos,omitempty"`     // 采购订单警告
	PreventRfqs bool `json:"prevent_rfqs,omitempty"` // 阻止询价单
	PreventPos  bool `json:"prevent_pos,omitempty"`  // 阻止采购订单

	// 暂停类型
	HoldType string `json:"hold_type,omitempty"` // 暂停类型

	// 权限设置
	AllowPurchaseInvoiceCreationWithoutPurchaseOrder   bool `json:"allow_purchase_invoice_creation_without_purchase_order,omitempty"`   // 允许无采购订单创建采购发票
	AllowPurchaseInvoiceCreationWithoutPurchaseReceipt bool `json:"allow_purchase_invoice_creation_without_purchase_receipt,omitempty"` // 允许无收货单创建采购发票

	// 关联数据
	Accounts    []interface{}           `json:"accounts,omitempty"`     // 账户信息
	PortalUsers []interface{}           `json:"portal_users,omitempty"` // 门户用户
	Companies   []AllowedToTransactWith `json:"companies,omitempty"`    // 允许交易的公司

	Company string `json:"custom_company,omitempty"` //所属公司
	Branch  string `json:"custom_branch,omitempty"`  //所属分支

	CustomPermissionRule []PermissionRule `json:"custom_permission_rule,omitempty"` //自定权限清单 多选表格

	CustomAliasName string `json:"custom_aliasname,omitempty"` // 别名名称 ，对外显示的名称
	SupplierGroup   string `json:"supplier_group,omitempty"`   // 供应商组

}

// AllowedToTransactWith 结构体，表示允许交易的公司
type AllowedToTransactWith struct {
	// 基础字段
	Name       string `json:"name,omitempty"`        // 名称
	Owner      string `json:"owner,omitempty"`       // 拥有者
	Creation   string `json:"creation,omitempty"`    // 创建时间
	Modified   string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy string `json:"modified_by,omitempty"` // 修改人
	Docstatus  int    `json:"docstatus,omitempty"`   // 单据状态
	Idx        int    `json:"idx,omitempty"`         // 索引
	Doctype    string `json:"doctype,omitempty"`     // 单据类型

	// 公司关联
	Company     string `json:"company,omitempty" validate:"required"` // 公司
	Parent      string `json:"parent,omitempty"`                      // 父级
	Parentfield string `json:"parentfield,omitempty"`                 // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`                  // 父级类型
}

// PaymentSchedule 结构体，表示付款计划
type PaymentSchedule struct {
	// 基础字段
	Name       string `json:"name,omitempty" validate:"required"` // 名称
	Owner      string `json:"owner,omitempty"`                    // 拥有者
	Creation   string `json:"creation,omitempty"`                 // 创建时间
	Modified   string `json:"modified,omitempty"`                 // 修改时间
	ModifiedBy string `json:"modified_by,omitempty"`              // 修改人
	Docstatus  int    `json:"docstatus,omitempty"`                // 单据状态
	Idx        int    `json:"idx,omitempty"`                      // 索引
	Doctype    string `json:"doctype,omitempty"`                  // 单据类型

	// 付款信息
	DueDate          string  `json:"due_date,omitempty" validate:"required"` // 到期日期
	InvoicePortion   float64 `json:"invoice_portion,omitempty"`              // 发票比例
	Discount         float64 `json:"discount,omitempty"`                     // 折扣
	PaymentAmount    float64 `json:"payment_amount,omitempty"`               // 付款金额
	Outstanding      float64 `json:"outstanding,omitempty"`                  // 未付金额
	PaidAmount       float64 `json:"paid_amount,omitempty"`                  // 已付金额
	DiscountedAmount float64 `json:"discounted_amount,omitempty"`            // 折扣金额

	// 基础货币信息
	BasePaymentAmount float64 `json:"base_payment_amount,omitempty"` // 基础付款金额
	BaseOutstanding   float64 `json:"base_outstanding,omitempty"`    // 基础未付金额
	BasePaidAmount    float64 `json:"base_paid_amount,omitempty"`    // 基础已付金额

	// 父级关联
	Parent      string `json:"parent,omitempty"`      // 父级
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
}

// PurchaseOrderItem 结构体，表示采购订单项目
type PurchaseOrderItem struct {
	// 基础字段
	Name       string `json:"name,omitempty" validate:"required"` // 名称
	Owner      string `json:"owner,omitempty"`                    // 拥有者
	Creation   string `json:"creation,omitempty"`                 // 创建时间
	Modified   string `json:"modified,omitempty"`                 // 修改时间
	ModifiedBy string `json:"modified_by,omitempty"`              // 修改人
	Docstatus  int    `json:"docstatus,omitempty"`                // 单据状态
	Idx        int    `json:"idx,omitempty"`                      // 索引
	Doctype    string `json:"doctype,omitempty"`                  // 单据类型

	// 项目基本信息
	ItemCode    string `json:"item_code,omitempty" validate:"required"` // 项目编码
	ItemName    string `json:"item_name,omitempty" validate:"required"` // 项目名称
	ItemGroup   string `json:"item_group,omitempty"`                    // 项目组
	Description string `json:"description,omitempty"`                   // 描述
	Image       string `json:"image,omitempty"`                         // 图片

	// 数量和单位
	Qty                   float64 `json:"qty,omitempty" validate:"required,gt=0"` // 数量
	StockUom              string  `json:"stock_uom,omitempty"`                    // 库存单位
	Uom                   string  `json:"uom,omitempty"`                          // 单位
	ConversionFactor      float64 `json:"conversion_factor,omitempty"`            // 转换因子
	StockQty              float64 `json:"stock_qty,omitempty"`                    // 库存数量
	FgItemQty             float64 `json:"fg_item_qty,omitempty"`                  // 成品项目数量
	SubcontractedQuantity float64 `json:"subcontracted_quantity,omitempty"`       // 分包数量
	ActualQty             float64 `json:"actual_qty,omitempty"`                   // 实际数量
	ReceivedQty           float64 `json:"received_qty,omitempty"`                 // 已收数量
	ReturnedQty           float64 `json:"returned_qty,omitempty"`                 // 退回数量

	// 日期相关
	ScheduleDate string `json:"schedule_date,omitempty"` // 计划日期

	// 价格相关
	PriceListRate     float64 `json:"price_list_rate,omitempty"`      // 价目表价格
	LastPurchaseRate  float64 `json:"last_purchase_rate,omitempty"`   // 最后采购价格
	BasePriceListRate float64 `json:"base_price_list_rate,omitempty"` // 基础价目表价格
	Rate              float64 `json:"rate,omitempty"`                 // 价格
	BaseRate          float64 `json:"base_rate,omitempty"`            // 基础价格
	StockUomRate      float64 `json:"stock_uom_rate,omitempty"`       // 库存单位价格
	NetRate           float64 `json:"net_rate,omitempty"`             // 净价格
	BaseNetRate       float64 `json:"base_net_rate,omitempty"`        // 基础净价格

	// 金额相关
	Amount        float64 `json:"amount,omitempty"`          // 金额
	BaseAmount    float64 `json:"base_amount,omitempty"`     // 基础金额
	NetAmount     float64 `json:"net_amount,omitempty"`      // 净金额
	BaseNetAmount float64 `json:"base_net_amount,omitempty"` // 基础净金额
	BilledAmt     float64 `json:"billed_amt,omitempty"`      // 已开票金额

	// 折扣相关
	DiscountPercentage        float64 `json:"discount_percentage,omitempty"`         // 折扣百分比
	DiscountAmount            float64 `json:"discount_amount,omitempty"`             // 折扣金额
	DistributedDiscountAmount float64 `json:"distributed_discount_amount,omitempty"` // 分配折扣金额

	// 利润率相关
	MarginType         string  `json:"margin_type,omitempty"`           // 利润率类型
	MarginRateOrAmount float64 `json:"margin_rate_or_amount,omitempty"` // 利润率或金额
	RateWithMargin     float64 `json:"rate_with_margin,omitempty"`      // 含利润率价格
	BaseRateWithMargin float64 `json:"base_rate_with_margin,omitempty"` // 基础含利润率价格

	// 仓库和库存
	Warehouse           string  `json:"warehouse,omitempty"`             // 仓库
	CompanyTotalStock   float64 `json:"company_total_stock,omitempty"`   // 公司总库存
	MaterialRequest     string  `json:"material_request,omitempty"`      // 物料请求
	MaterialRequestItem string  `json:"material_request_item,omitempty"` // 物料请求项目

	// 状态标识
	DeliveredBySupplier  bool `json:"delivered_by_supplier,omitempty"`  // 供应商交付
	AgainstBlanketOrder  bool `json:"against_blanket_order,omitempty"`  // 针对总括订单
	IsFreeItem           bool `json:"is_free_item,omitempty"`           // 是否免费项目
	ApplyTds             bool `json:"apply_tds,omitempty"`              // 应用TDS
	IncludeExplodedItems bool `json:"include_exploded_items,omitempty"` // 包含分解项目
	IsFixedAsset         bool `json:"is_fixed_asset,omitempty"`         // 是否固定资产
	PageBreak            bool `json:"page_break,omitempty"`             // 分页符

	// 其他信息
	BlanketOrderRate float64 `json:"blanket_order_rate,omitempty"` // 总括订单价格
	ExpenseAccount   string  `json:"expense_account,omitempty"`    // 费用账户
	WeightPerUnit    float64 `json:"weight_per_unit,omitempty"`    // 单位重量
	TotalWeight      float64 `json:"total_weight,omitempty"`       // 总重量
	CostCenter       string  `json:"cost_center,omitempty"`        // 成本中心
	ItemTaxRate      string  `json:"item_tax_rate,omitempty"`      // 项目税率

	// 父级关联
	Parent      string `json:"parent,omitempty"`      // 父级
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
}

// PurchaseOrder 结构体，表示采购订单
type PurchaseOrder struct {
	// 基础字段
	Name       string `json:"name,omitempty" validate:"required"` // 名称
	Owner      string `json:"owner,omitempty"`                    // 拥有者
	Creation   string `json:"creation,omitempty"`                 // 创建时间
	Modified   string `json:"modified,omitempty"`                 // 修改时间
	ModifiedBy string `json:"modified_by,omitempty"`              // 修改人
	Docstatus  int    `json:"docstatus,omitempty"`                // 单据状态
	Idx        int    `json:"idx,omitempty"`                      // 索引
	Doctype    string `json:"doctype,omitempty"`                  // 单据类型

	// 基本信息
	Title        string `json:"title,omitempty" validate:"required"`    // 标题
	NamingSeries string `json:"naming_series,omitempty"`                // 编码规则
	Supplier     string `json:"supplier,omitempty" validate:"required"` // 供应商
	SupplierName string `json:"supplier_name,omitempty"`                // 供应商名称
	Company      string `json:"company,omitempty" validate:"required"`  // 公司

	// 日期信息
	TransactionDate string `json:"transaction_date,omitempty" validate:"required"` // 交易日期
	ScheduleDate    string `json:"schedule_date,omitempty"`                        // 计划日期

	// 状态标识
	IsSubcontracted            bool   `json:"is_subcontracted,omitempty"`              // 是否分包
	HasUnitPriceItems          bool   `json:"has_unit_price_items,omitempty"`          // 是否有单位价格项目
	ApplyTds                   bool   `json:"apply_tds,omitempty"`                     // 应用TDS
	IgnorePricingRule          bool   `json:"ignore_pricing_rule,omitempty"`           // 忽略定价规则
	GroupSameItems             bool   `json:"group_same_items,omitempty"`              // 分组相同项目
	IsInternalSupplier         bool   `json:"is_internal_supplier,omitempty"`          // 是否内部供应商
	InterCompanyOrderReference string `json:"inter_company_order_reference,omitempty"` // 跨公司订单引用
	IsOldSubcontractingFlow    bool   `json:"is_old_subcontracting_flow,omitempty"`    // 是否旧分包流程
	DisableRoundedTotal        bool   `json:"disable_rounded_total,omitempty"`         // 禁用舍入总额

	// 货币和汇率
	Currency          string  `json:"currency,omitempty"`            // 货币
	ConversionRate    float64 `json:"conversion_rate,omitempty"`     // 转换汇率
	BuyingPriceList   string  `json:"buying_price_list,omitempty"`   // 采购价目表
	PriceListCurrency string  `json:"price_list_currency,omitempty"` // 价目表货币
	PlcConversionRate float64 `json:"plc_conversion_rate,omitempty"` // PLC转换汇率

	// 仓库设置
	SetWarehouse string `json:"set_warehouse,omitempty"` // 设置仓库

	// 数量和重量
	TotalQty       float64 `json:"total_qty,omitempty"`        // 总数量
	TotalNetWeight float64 `json:"total_net_weight,omitempty"` // 总净重

	// 金额信息
	BaseTotal                   float64 `json:"base_total,omitempty"`                      // 基础总额
	BaseNetTotal                float64 `json:"base_net_total,omitempty"`                  // 基础净总额
	Total                       float64 `json:"total,omitempty"`                           // 总额
	NetTotal                    float64 `json:"net_total,omitempty"`                       // 净总额
	TaxWithholdingNetTotal      float64 `json:"tax_withholding_net_total,omitempty"`       // 代扣税净总额
	BaseTaxWithholdingNetTotal  float64 `json:"base_tax_withholding_net_total,omitempty"`  // 基础代扣税净总额
	BaseTaxesAndChargesAdded    float64 `json:"base_taxes_and_charges_added,omitempty"`    // 基础税费和费用增加
	BaseTaxesAndChargesDeducted float64 `json:"base_taxes_and_charges_deducted,omitempty"` // 基础税费和费用扣除
	BaseTotalTaxesAndCharges    float64 `json:"base_total_taxes_and_charges,omitempty"`    // 基础税费和费用总额
	TaxesAndChargesAdded        float64 `json:"taxes_and_charges_added,omitempty"`         // 税费和费用增加
	TaxesAndChargesDeducted     float64 `json:"taxes_and_charges_deducted,omitempty"`      // 税费和费用扣除
	TotalTaxesAndCharges        float64 `json:"total_taxes_and_charges,omitempty"`         // 税费和费用总额
	BaseGrandTotal              float64 `json:"base_grand_total,omitempty"`                // 基础总金额
	BaseRoundingAdjustment      float64 `json:"base_rounding_adjustment,omitempty"`        // 基础舍入调整
	BaseRoundedTotal            float64 `json:"base_rounded_total,omitempty"`              // 基础舍入总额
	GrandTotal                  float64 `json:"grand_total,omitempty"`                     // 总金额
	RoundingAdjustment          float64 `json:"rounding_adjustment,omitempty"`             // 舍入调整
	RoundedTotal                float64 `json:"rounded_total,omitempty"`                   // 舍入总额
	AdvancePaid                 float64 `json:"advance_paid,omitempty"`                    // 预付款

	// 折扣相关
	ApplyDiscountOn              string  `json:"apply_discount_on,omitempty"`              // 应用折扣于
	BaseDiscountAmount           float64 `json:"base_discount_amount,omitempty"`           // 基础折扣金额
	AdditionalDiscountPercentage float64 `json:"additional_discount_percentage,omitempty"` // 额外折扣百分比
	DiscountAmount               float64 `json:"discount_amount,omitempty"`                // 折扣金额

	// 地址信息
	ShippingAddress        string `json:"shipping_address,omitempty"`         // 送货地址
	ShippingAddressDisplay string `json:"shipping_address_display,omitempty"` // 送货地址显示
	BillingAddress         string `json:"billing_address,omitempty"`          // 账单地址
	BillingAddressDisplay  string `json:"billing_address_display,omitempty"`  // 账单地址显示

	// 状态和百分比
	Status      string  `json:"status,omitempty"`       // 状态
	PerBilled   float64 `json:"per_billed,omitempty"`   // 已开票百分比
	PerReceived float64 `json:"per_received,omitempty"` // 已收货百分比

	// 其他信息
	Language             string `json:"language,omitempty"`               // 语言
	RepresentsCompany    string `json:"represents_company,omitempty"`     // 代表公司
	PartyAccountCurrency string `json:"party_account_currency,omitempty"` // 对方账户货币
	TaxCategory          string `json:"tax_category,omitempty"`           // 税类别

	// 金额大写
	BaseInWords string `json:"base_in_words,omitempty"` // 基础金额大写
	InWords     string `json:"in_words,omitempty"`      // 金额大写

	// 关联数据
	PricingRules    []interface{}        `json:"pricing_rules,omitempty"`    // 定价规则
	SuppliedItems   []interface{}        `json:"supplied_items,omitempty"`   // 供应项目
	PaymentSchedule []PaymentSchedule    `json:"payment_schedule,omitempty"` // 付款计划
	Items           []*PurchaseOrderItem `json:"items,omitempty"`            // 项目
	Taxes           []interface{}        `json:"taxes,omitempty"`            // 税费

}

// SaleOrderItem 结构体，表示销售订单项目
type SaleOrderItem struct {
	// 基础字段
	Owner     string `json:"owner,omitempty"`     // 拥有者
	Docstatus int    `json:"docstatus,omitempty"` // 单据状态
	Idx       int    `json:"idx,omitempty"`       // 索引
	Doctype   string `json:"doctype,omitempty"`   // 单据类型

	// 项目基本信息
	ItemCode    string `json:"item_code,omitempty"`   // 项目编码
	ItemName    string `json:"item_name,omitempty"`   // 项目名称
	ItemGroup   string `json:"item_group,omitempty"`  // 项目组
	Description string `json:"description,omitempty"` // 描述
	Image       string `json:"image,omitempty"`       // 图片

	// 数量和单位
	Qty              float64 `json:"qty,omitempty"`                // 数量
	StockUom         string  `json:"stock_uom,omitempty"`          // 库存单位
	Uom              string  `json:"uom,omitempty"`                // 单位
	ConversionFactor float64 `json:"conversion_factor,omitempty"`  // 转换因子
	StockQty         float64 `json:"stock_qty,omitempty"`          // 库存数量
	StockReservedQty float64 `json:"stock_reserved_qty,omitempty"` // 库存预留数量

	DeliveryDate string `json:"delivery_date,omitempty"` //送货时间

	// 价格相关
	PriceListRate     float64 `json:"price_list_rate,omitempty"`      // 价目表价格
	BasePriceListRate float64 `json:"base_price_list_rate,omitempty"` // 基础价目表价格
	Rate              float64 `json:"rate,omitempty"`                 // 价格
	BaseRate          float64 `json:"base_rate,omitempty"`            // 基础价格
	StockUomRate      float64 `json:"stock_uom_rate,omitempty"`       // 库存单位价格
	NetRate           float64 `json:"net_rate,omitempty"`             // 净价格
	BaseNetRate       float64 `json:"base_net_rate,omitempty"`        // 基础净价格

	// 金额相关
	Amount        float64 `json:"amount,omitempty"`          // 金额
	BaseAmount    float64 `json:"base_amount,omitempty"`     // 基础金额
	NetAmount     float64 `json:"net_amount,omitempty"`      // 净金额
	BaseNetAmount float64 `json:"base_net_amount,omitempty"` // 基础净金额
	BilledAmt     float64 `json:"billed_amt,omitempty"`      // 已开票金额

	// 折扣相关
	DiscountPercentage        float64 `json:"discount_percentage,omitempty"`         // 折扣百分比
	DiscountAmount            float64 `json:"discount_amount,omitempty"`             // 折扣金额
	DistributedDiscountAmount float64 `json:"distributed_discount_amount,omitempty"` // 分配折扣金额

	// 利润率相关
	MarginType         string  `json:"margin_type,omitempty"`           // 利润率类型
	MarginRateOrAmount float64 `json:"margin_rate_or_amount,omitempty"` // 利润率或金额
	RateWithMargin     float64 `json:"rate_with_margin,omitempty"`      // 含利润率价格
	BaseRateWithMargin float64 `json:"base_rate_with_margin,omitempty"` // 基础含利润率价格

	// 状态标识
	EnsureDeliveryBasedOnProducedSerialNo bool `json:"ensure_delivery_based_on_produced_serial_no,omitempty"` // 确保基于生产的序列号交付
	IsStockItem                           bool `json:"is_stock_item,omitempty"`                               // 是否库存项目
	ReserveStock                          bool `json:"reserve_stock,omitempty"`                               // 预留库存
	IsFreeItem                            bool `json:"is_free_item,omitempty"`                                // 是否免费项目
	GrantCommission                       bool `json:"grant_commission,omitempty"`                            // 授予佣金
	DeliveredBySupplier                   bool `json:"delivered_by_supplier,omitempty"`                       // 供应商交付
	AgainstBlanketOrder                   bool `json:"against_blanket_order,omitempty"`                       // 针对总括订单
	PageBreak                             bool `json:"page_break,omitempty"`                                  // 分页符

	// 数量和库存相关
	ActualQty         float64 `json:"actual_qty,omitempty"`          // 实际数量
	CompanyTotalStock float64 `json:"company_total_stock,omitempty"` // 公司总库存
	ProjectedQty      float64 `json:"projected_qty,omitempty"`       // 预计数量
	OrderedQty        float64 `json:"ordered_qty,omitempty"`         // 已订购数量
	PlannedQty        float64 `json:"planned_qty,omitempty"`         // 计划数量
	ProductionPlanQty float64 `json:"production_plan_qty,omitempty"` // 生产计划数量
	WorkOrderQty      float64 `json:"work_order_qty,omitempty"`      // 工作订单数量
	DeliveredQty      float64 `json:"delivered_qty,omitempty"`       // 已交付数量
	ProducedQty       float64 `json:"produced_qty,omitempty"`        // 已生产数量
	ReturnedQty       float64 `json:"returned_qty,omitempty"`        // 退回数量
	PickedQty         float64 `json:"picked_qty,omitempty"`          // 已拣货数量

	// 其他信息
	ValuationRate       float64 `json:"valuation_rate,omitempty"`        // 评估价格
	GrossProfit         float64 `json:"gross_profit,omitempty"`          // 毛利润
	WeightPerUnit       float64 `json:"weight_per_unit,omitempty"`       // 单位重量
	TotalWeight         float64 `json:"total_weight,omitempty"`          // 总重量
	BlanketOrderRate    float64 `json:"blanket_order_rate,omitempty"`    // 总括订单价格
	ItemTaxRate         string  `json:"item_tax_rate,omitempty"`         // 项目税率
	TransactionDate     string  `json:"transaction_date,omitempty"`      // 交易日期
	MaterialRequest     string  `json:"material_request,omitempty"`      // 物料请求
	PurchaseOrder       string  `json:"purchase_order,omitempty"`        // 采购订单
	MaterialRequestItem string  `json:"material_request_item,omitempty"` // 物料请求项目
	PurchaseOrderItem   string  `json:"purchase_order_item,omitempty"`   // 采购订单项目
	CostCenter          string  `json:"cost_center,omitempty"`           // 成本中心

	// 父级关联
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
}

// SaleOrder 结构体，表示销售订单
type SaleOrder struct {
	// 基础字段
	Owner     string `json:"owner,omitempty"`     // 拥有者
	Docstatus int    `json:"docstatus,omitempty"` // 单据状态
	Idx       int    `json:"idx,omitempty"`       // 索引
	Doctype   string `json:"doctype,omitempty"`   // 单据类型
	Name      string `json:"name,omitempty"`      // 名称

	// 基本信息
	Title        string `json:"title,omitempty"`         // 标题
	NamingSeries string `json:"naming_series,omitempty"` // 编码规则
	Customer     string `json:"customer,omitempty"`      // 客户
	CustomerName string `json:"customer_name,omitempty"` // 客户名称
	OrderType    string `json:"order_type,omitempty"`    // 订单类型
	Company      string `json:"company,omitempty"`       // 公司

	SetWarehouse string `json:"set_warehouse,omitempty"` // 设置来源仓库

	// 日期信息
	TransactionDate string `json:"transaction_date,omitempty"` // 交易日期
	DeliveryDate    string `json:"delivery_date,omitempty"`    //送货时间

	// 状态标识
	SkipDeliveryNote    bool `json:"skip_delivery_note,omitempty"`    // 跳过送货单
	HasUnitPriceItems   bool `json:"has_unit_price_items,omitempty"`  // 是否有单位价格项目
	IgnorePricingRule   bool `json:"ignore_pricing_rule,omitempty"`   // 忽略定价规则
	ReserveStock        bool `json:"reserve_stock,omitempty"`         // 预留库存
	GroupSameItems      bool `json:"group_same_items,omitempty"`      // 分组相同项目
	IsInternalCustomer  bool `json:"is_internal_customer,omitempty"`  // 是否内部客户
	DisableRoundedTotal bool `json:"disable_rounded_total,omitempty"` // 禁用舍入总额

	// 货币和汇率
	Currency          string  `json:"currency,omitempty"`            // 货币
	ConversionRate    float64 `json:"conversion_rate,omitempty"`     // 转换汇率
	SellingPriceList  string  `json:"selling_price_list,omitempty"`  // 销售价目表
	PriceListCurrency string  `json:"price_list_currency,omitempty"` // 价目表货币
	PlcConversionRate float64 `json:"plc_conversion_rate,omitempty"` // PLC转换汇率

	// 数量和重量
	TotalQty       float64 `json:"total_qty,omitempty"`        // 总数量
	TotalNetWeight float64 `json:"total_net_weight,omitempty"` // 总净重

	// 金额信息
	BaseTotal                float64 `json:"base_total,omitempty"`                   // 基础总额
	BaseNetTotal             float64 `json:"base_net_total,omitempty"`               // 基础净总额
	Total                    float64 `json:"total,omitempty"`                        // 总额
	NetTotal                 float64 `json:"net_total,omitempty"`                    // 净总额
	TaxCategory              string  `json:"tax_category,omitempty"`                 // 税类别
	BaseTotalTaxesAndCharges float64 `json:"base_total_taxes_and_charges,omitempty"` // 基础税费和费用总额
	TotalTaxesAndCharges     float64 `json:"total_taxes_and_charges,omitempty"`      // 税费和费用总额
	BaseGrandTotal           float64 `json:"base_grand_total,omitempty"`             // 基础总金额
	BaseRoundingAdjustment   float64 `json:"base_rounding_adjustment,omitempty"`     // 基础舍入调整
	BaseRoundedTotal         float64 `json:"base_rounded_total,omitempty"`           // 基础舍入总额
	GrandTotal               float64 `json:"grand_total,omitempty"`                  // 总金额
	RoundingAdjustment       float64 `json:"rounding_adjustment,omitempty"`          // 舍入调整
	RoundedTotal             float64 `json:"rounded_total,omitempty"`                // 舍入总额
	AdvancePaid              float64 `json:"advance_paid,omitempty"`                 // 预付款

	// 折扣相关
	ApplyDiscountOn              string  `json:"apply_discount_on,omitempty"`              // 应用折扣于
	BaseDiscountAmount           float64 `json:"base_discount_amount,omitempty"`           // 基础折扣金额
	AdditionalDiscountPercentage float64 `json:"additional_discount_percentage,omitempty"` // 额外折扣百分比
	DiscountAmount               float64 `json:"discount_amount,omitempty"`                // 折扣金额

	// 状态信息
	Status         string `json:"status,omitempty"`          // 状态
	DeliveryStatus string `json:"delivery_status,omitempty"` // 交付状态
	BillingStatus  string `json:"billing_status,omitempty"`  // 开票状态

	// 百分比信息
	PerDelivered float64 `json:"per_delivered,omitempty"` // 已交付百分比
	PerBilled    float64 `json:"per_billed,omitempty"`    // 已开票百分比
	PerPicked    float64 `json:"per_picked,omitempty"`    // 已拣货百分比

	// 佣金和忠诚度
	AmountEligibleForCommission float64 `json:"amount_eligible_for_commission,omitempty"` // 符合佣金条件的金额
	CommissionRate              float64 `json:"commission_rate,omitempty"`                // 佣金率
	TotalCommission             float64 `json:"total_commission,omitempty"`               // 总佣金
	LoyaltyPoints               float64 `json:"loyalty_points,omitempty"`                 // 忠诚度积分
	LoyaltyAmount               float64 `json:"loyalty_amount,omitempty"`                 // 忠诚度金额

	// 其他信息
	Language                   string `json:"language,omitempty"`                      // 语言
	RepresentsCompany          string `json:"represents_company,omitempty"`            // 代表公司
	InterCompanyOrderReference string `json:"inter_company_order_reference,omitempty"` // 公司间订单引用

	// 金额大写
	BaseInWords string `json:"base_in_words,omitempty"` // 基础金额大写
	InWords     string `json:"in_words,omitempty"`      // 金额大写

	// 关联数据
	PricingRules    []interface{}      `json:"pricing_rules,omitempty"`    // 定价规则
	SalesTeam       []interface{}      `json:"sales_team,omitempty"`       // 销售团队
	PaymentSchedule []*PaymentSchedule `json:"payment_schedule,omitempty"` // 付款计划
	Items           []*SaleOrderItem   `json:"items,omitempty"`            // 项目
	Taxes           []interface{}      `json:"taxes,omitempty"`            // 税费
	PackedItems     []interface{}      `json:"packed_items,omitempty"`     // 包装项目
}

// PurchaseReceiptItem 结构体，表示采购收货单项目
type PurchaseReceiptItem struct {
	// 基础字段
	Owner     string `json:"owner,omitempty"`     // 拥有者
	Docstatus int    `json:"docstatus,omitempty"` // 单据状态
	Idx       int    `json:"idx,omitempty"`       // 索引
	Doctype   string `json:"doctype,omitempty"`   // 单据类型

	// 项目基本信息
	ItemCode    string `json:"item_code,omitempty"`   // 项目编码
	ItemName    string `json:"item_name,omitempty"`   // 项目名称
	ItemGroup   string `json:"item_group,omitempty"`  // 项目组
	Description string `json:"description,omitempty"` // 描述
	Image       string `json:"image,omitempty"`       // 图片

	// 数量和单位
	Qty              float64 `json:"qty,omitempty"`                // 数量
	StockUom         string  `json:"stock_uom,omitempty"`          // 库存单位
	Uom              string  `json:"uom,omitempty"`                // 单位
	ConversionFactor float64 `json:"conversion_factor,omitempty"`  // 转换因子
	StockQty         float64 `json:"stock_qty,omitempty"`          // 库存数量
	ReceivedQty      float64 `json:"received_qty,omitempty"`       // 已收数量
	RejectedQty      float64 `json:"rejected_qty,omitempty"`       // 拒绝数量
	ReturnedQty      float64 `json:"returned_qty,omitempty"`       // 退回数量
	ReceivedStockQty float64 `json:"received_stock_qty,omitempty"` // 已收库存数量

	// 样本相关
	RetainSample   bool    `json:"retain_sample,omitempty"`   // 保留样本
	SampleQuantity float64 `json:"sample_quantity,omitempty"` // 样本数量

	// 价格相关
	PriceListRate     float64 `json:"price_list_rate,omitempty"`      // 价目表价格
	BasePriceListRate float64 `json:"base_price_list_rate,omitempty"` // 基础价目表价格
	Rate              float64 `json:"rate,omitempty"`                 // 价格
	BaseRate          float64 `json:"base_rate,omitempty"`            // 基础价格
	StockUomRate      float64 `json:"stock_uom_rate,omitempty"`       // 库存单位价格
	NetRate           float64 `json:"net_rate,omitempty"`             // 净价格
	BaseNetRate       float64 `json:"base_net_rate,omitempty"`        // 基础净价格
	ValuationRate     float64 `json:"valuation_rate,omitempty"`       // 评估价格
	SalesIncomingRate float64 `json:"sales_incoming_rate,omitempty"`  // 销售收入价格

	// 金额相关
	Amount        float64 `json:"amount,omitempty"`          // 金额
	BaseAmount    float64 `json:"base_amount,omitempty"`     // 基础金额
	NetAmount     float64 `json:"net_amount,omitempty"`      // 净金额
	BaseNetAmount float64 `json:"base_net_amount,omitempty"` // 基础净金额
	BilledAmt     float64 `json:"billed_amt,omitempty"`      // 已开票金额

	// 折扣相关
	DiscountPercentage        float64 `json:"discount_percentage,omitempty"`         // 折扣百分比
	DiscountAmount            float64 `json:"discount_amount,omitempty"`             // 折扣金额
	DistributedDiscountAmount float64 `json:"distributed_discount_amount,omitempty"` // 分配折扣金额

	// 利润率相关
	MarginType         string  `json:"margin_type,omitempty"`           // 利润率类型
	MarginRateOrAmount float64 `json:"margin_rate_or_amount,omitempty"` // 利润率或金额
	RateWithMargin     float64 `json:"rate_with_margin,omitempty"`      // 含利润率价格
	BaseRateWithMargin float64 `json:"base_rate_with_margin,omitempty"` // 基础含利润率价格

	// 税费相关
	ItemTaxAmount float64 `json:"item_tax_amount,omitempty"` // 项目税费金额

	// 成本相关
	RmSuppCost                          float64 `json:"rm_supp_cost,omitempty"`                            // 原材料供应成本
	LandedCostVoucherAmount             float64 `json:"landed_cost_voucher_amount,omitempty"`              // 到岸成本凭证金额
	AmountDifferenceWithPurchaseInvoice float64 `json:"amount_difference_with_purchase_invoice,omitempty"` // 与采购发票的金额差异

	// 状态标识
	HasItemScanned                 bool `json:"has_item_scanned,omitempty"`                   // 是否有项目扫描
	IsFreeItem                     bool `json:"is_free_item,omitempty"`                       // 是否免费项目
	ApplyTds                       bool `json:"apply_tds,omitempty"`                          // 应用TDS
	AllowZeroValuationRate         bool `json:"allow_zero_valuation_rate,omitempty"`          // 允许零评估价格
	ReturnQtyFromRejectedWarehouse bool `json:"return_qty_from_rejected_warehouse,omitempty"` // 从拒绝仓库退回数量
	IsFixedAsset                   bool `json:"is_fixed_asset,omitempty"`                     // 是否固定资产
	UseSerialBatchFields           bool `json:"use_serial_batch_fields,omitempty"`            // 使用序列批次字段
	IncludeExplodedItems           bool `json:"include_exploded_items,omitempty"`             // 包含分解项目
	PageBreak                      bool `json:"page_break,omitempty"`                         // 分页符

	// 日期相关
	ScheduleDate string `json:"schedule_date,omitempty"` // 计划日期

	// 仓库和库存
	Warehouse string `json:"warehouse,omitempty"` // 仓库

	// 关联信息
	MaterialRequest     string `json:"material_request,omitempty"`      // 物料请求
	PurchaseOrder       string `json:"purchase_order,omitempty"`        // 采购订单
	MaterialRequestItem string `json:"material_request_item,omitempty"` // 物料请求项目
	PurchaseOrderItem   string `json:"purchase_order_item,omitempty"`   // 采购订单项目

	// 其他信息
	WeightPerUnit  float64 `json:"weight_per_unit,omitempty"` // 单位重量
	TotalWeight    float64 `json:"total_weight,omitempty"`    // 总重量
	ExpenseAccount string  `json:"expense_account,omitempty"` // 费用账户
	ItemTaxRate    string  `json:"item_tax_rate,omitempty"`   // 项目税率
	CostCenter     string  `json:"cost_center,omitempty"`     // 成本中心

	// 父级关联
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
}

// PurchaseReceipt 结构体，表示采购收货单
type PurchaseReceipt struct {
	// 基础字段
	Owner     string `json:"owner,omitempty"`     // 拥有者
	Docstatus int    `json:"docstatus,omitempty"` // 单据状态
	Idx       int    `json:"idx,omitempty"`       // 索引
	Doctype   string `json:"doctype,omitempty"`   // 单据类型
	Name      string `json:"name,omitempty"`      // 名称

	// 基本信息
	NamingSeries string `json:"naming_series,omitempty"` // 编码规则
	Supplier     string `json:"supplier,omitempty"`      // 供应商
	SupplierName string `json:"supplier_name,omitempty"` // 供应商名称
	Company      string `json:"company,omitempty"`       // 公司

	// 日期和时间
	PostingDate    string `json:"posting_date,omitempty"`     // 过账日期
	PostingTime    string `json:"posting_time,omitempty"`     // 过账时间
	SetPostingTime bool   `json:"set_posting_time,omitempty"` // 设置过账时间

	// 状态标识
	ApplyPutawayRule        bool `json:"apply_putaway_rule,omitempty"`         // 应用上架规则
	IsReturn                bool `json:"is_return,omitempty"`                  // 是否退货
	IgnorePricingRule       bool `json:"ignore_pricing_rule,omitempty"`        // 忽略定价规则
	IsSubcontracted         bool `json:"is_subcontracted,omitempty"`           // 是否分包
	GroupSameItems          bool `json:"group_same_items,omitempty"`           // 分组相同项目
	IsInternalSupplier      bool `json:"is_internal_supplier,omitempty"`       // 是否内部供应商
	IsOldSubcontractingFlow bool `json:"is_old_subcontracting_flow,omitempty"` // 是否旧分包流程
	DisableRoundedTotal     bool `json:"disable_rounded_total,omitempty"`      // 禁用舍入总额

	// 货币和汇率
	Currency          string  `json:"currency,omitempty"`            // 货币
	ConversionRate    float64 `json:"conversion_rate,omitempty"`     // 转换汇率
	BuyingPriceList   string  `json:"buying_price_list,omitempty"`   // 采购价目表
	PriceListCurrency string  `json:"price_list_currency,omitempty"` // 价目表货币
	PlcConversionRate float64 `json:"plc_conversion_rate,omitempty"` // PLC转换汇率

	// 数量和重量
	TotalQty       float64 `json:"total_qty,omitempty"`        // 总数量
	TotalNetWeight float64 `json:"total_net_weight,omitempty"` // 总净重

	// 金额信息
	BaseTotal                   float64 `json:"base_total,omitempty"`                      // 基础总额
	BaseNetTotal                float64 `json:"base_net_total,omitempty"`                  // 基础净总额
	Total                       float64 `json:"total,omitempty"`                           // 总额
	NetTotal                    float64 `json:"net_total,omitempty"`                       // 净总额
	TaxWithholdingNetTotal      float64 `json:"tax_withholding_net_total,omitempty"`       // 代扣税净总额
	BaseTaxWithholdingNetTotal  float64 `json:"base_tax_withholding_net_total,omitempty"`  // 基础代扣税净总额
	TaxCategory                 string  `json:"tax_category,omitempty"`                    // 税类别
	BaseTaxesAndChargesAdded    float64 `json:"base_taxes_and_charges_added,omitempty"`    // 基础税费和费用增加
	BaseTaxesAndChargesDeducted float64 `json:"base_taxes_and_charges_deducted,omitempty"` // 基础税费和费用扣除
	BaseTotalTaxesAndCharges    float64 `json:"base_total_taxes_and_charges,omitempty"`    // 基础税费和费用总额
	TaxesAndChargesAdded        float64 `json:"taxes_and_charges_added,omitempty"`         // 税费和费用增加
	TaxesAndChargesDeducted     float64 `json:"taxes_and_charges_deducted,omitempty"`      // 税费和费用扣除
	TotalTaxesAndCharges        float64 `json:"total_taxes_and_charges,omitempty"`         // 税费和费用总额
	BaseGrandTotal              float64 `json:"base_grand_total,omitempty"`                // 基础总金额
	BaseRoundingAdjustment      float64 `json:"base_rounding_adjustment,omitempty"`        // 基础舍入调整
	BaseRoundedTotal            float64 `json:"base_rounded_total,omitempty"`              // 基础舍入总额
	GrandTotal                  float64 `json:"grand_total,omitempty"`                     // 总金额
	RoundingAdjustment          float64 `json:"rounding_adjustment,omitempty"`             // 舍入调整
	RoundedTotal                float64 `json:"rounded_total,omitempty"`                   // 舍入总额

	// 折扣相关
	ApplyDiscountOn              string  `json:"apply_discount_on,omitempty"`              // 应用折扣于
	BaseDiscountAmount           float64 `json:"base_discount_amount,omitempty"`           // 基础折扣金额
	AdditionalDiscountPercentage float64 `json:"additional_discount_percentage,omitempty"` // 额外折扣百分比
	DiscountAmount               float64 `json:"discount_amount,omitempty"`                // 折扣金额

	// 状态信息
	Status string `json:"status,omitempty"` // 状态

	// 百分比信息
	PerBilled   float64 `json:"per_billed,omitempty"`   // 已开票百分比
	PerReturned float64 `json:"per_returned,omitempty"` // 已退货百分比

	// 其他信息
	Language          string `json:"language,omitempty"`           // 语言
	RepresentsCompany string `json:"represents_company,omitempty"` // 代表公司

	// 金额大写
	BaseInWords string `json:"base_in_words,omitempty"` // 基础金额大写
	InWords     string `json:"in_words,omitempty"`      // 金额大写

	// 关联数据
	PricingRules  []interface{}          `json:"pricing_rules,omitempty"`  // 定价规则
	SuppliedItems []interface{}          `json:"supplied_items,omitempty"` // 供应项目
	Items         []*PurchaseReceiptItem `json:"items,omitempty"`          // 项目
	Taxes         []interface{}          `json:"taxes,omitempty"`          // 税费
}
