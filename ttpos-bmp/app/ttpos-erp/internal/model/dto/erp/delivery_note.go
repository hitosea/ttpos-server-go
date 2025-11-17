package erp

// DeliveryNote 送货单/交货单
type DeliveryNote struct {
	// 基础字段
	Owner     string `json:"owner,omitempty"`     // 拥有者
	Docstatus int    `json:"docstatus,omitempty"` // 单据状态
	Idx       int    `json:"idx,omitempty"`       // 索引
	Doctype   string `json:"doctype,omitempty"`   // 单据类型
	Name      string `json:"name,omitempty"`      // 名称

	// 基本信息
	NamingSeries  string `json:"naming_series,omitempty"`  // 编码规则
	Customer      string `json:"customer,omitempty"`       // 客户
	CustomerName  string `json:"customer_name,omitempty"`  // 客户名称
	CustomerGroup string `json:"customer_group,omitempty"` // 客户组
	Company       string `json:"company,omitempty"`        // 公司

	// 日期和时间
	PostingDate    string `json:"posting_date,omitempty"`     // 过账日期
	PostingTime    string `json:"posting_time,omitempty"`     // 过账时间
	SetPostingTime int    `json:"set_posting_time,omitempty"` // 设置过账时间

	// 状态标识
	IsReturn            int `json:"is_return,omitempty"`             // 是否退货
	IssueCreditNote     int `json:"issue_credit_note,omitempty"`     // 发出贷项通知单
	IgnorePricingRule   int `json:"ignore_pricing_rule,omitempty"`   // 忽略定价规则
	GroupSameItems      int `json:"group_same_items,omitempty"`      // 分组相同项目
	IsInternalCustomer  int `json:"is_internal_customer,omitempty"`  // 是否内部客户
	DisableRoundedTotal int `json:"disable_rounded_total,omitempty"` // 禁用舍入总额
	PrintWithoutAmount  int `json:"print_without_amount,omitempty"`  // 打印不含金额

	// 货币和汇率
	Currency          string  `json:"currency,omitempty"`            // 货币
	ConversionRate    float64 `json:"conversion_rate,omitempty"`     // 转换汇率
	SellingPriceList  string  `json:"selling_price_list,omitempty"`  // 销售价目表
	PriceListCurrency string  `json:"price_list_currency,omitempty"` // 价目表货币
	PlcConversionRate float64 `json:"plc_conversion_rate,omitempty"` // PLC转换汇率

	// 仓库设置
	SetWarehouse       string `json:"set_warehouse,omitempty"`        // 设置仓库
	SetTargetWarehouse string `json:"set_target_warehouse,omitempty"` // 设置目标仓库

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

	// 折扣相关
	ApplyDiscountOn              string  `json:"apply_discount_on,omitempty"`              // 应用折扣于
	BaseDiscountAmount           float64 `json:"base_discount_amount,omitempty"`           // 基础折扣金额
	AdditionalDiscountPercentage float64 `json:"additional_discount_percentage,omitempty"` // 额外折扣百分比
	DiscountAmount               float64 `json:"discount_amount,omitempty"`                // 折扣金额

	// 状态和百分比
	Status       string  `json:"status,omitempty"`        // 状态
	PerBilled    float64 `json:"per_billed,omitempty"`    // 已开票百分比
	PerInstalled float64 `json:"per_installed,omitempty"` // 已安装百分比
	PerReturned  float64 `json:"per_returned,omitempty"`  // 已退货百分比

	// 物流相关
	LrDate string `json:"lr_date,omitempty"` // 运输收据日期
	PoNo   string `json:"po_no,omitempty"`   // 采购订单号

	// 佣金相关
	AmountEligibleForCommission float64 `json:"amount_eligible_for_commission,omitempty"` // 符合佣金条件的金额
	CommissionRate              float64 `json:"commission_rate,omitempty"`                // 佣金率
	TotalCommission             float64 `json:"total_commission,omitempty"`               // 总佣金

	// 其他信息
	Language          string `json:"language,omitempty"`           // 语言
	RepresentsCompany string `json:"represents_company,omitempty"` // 代表公司

	// 金额大写
	BaseInWords string `json:"base_in_words,omitempty"` // 基础金额大写
	InWords     string `json:"in_words,omitempty"`      // 金额大写

	// 关联数据
	Items        []*DeliveryNoteItem `json:"items,omitempty"`         // 项目
	PricingRules []interface{}       `json:"pricing_rules,omitempty"` // 定价规则
	PackedItems  []interface{}       `json:"packed_items,omitempty"`  // 包装项目
	SalesTeam    []interface{}       `json:"sales_team,omitempty"`    // 销售团队
	Taxes        []interface{}       `json:"taxes,omitempty"`         // 税费
}

// DeliveryNoteItem 送货单明细项
type DeliveryNoteItem struct {
	// 基础字段
	Owner     string `json:"owner,omitempty"`     // 拥有者
	Docstatus int    `json:"docstatus,omitempty"` // 单据状态
	Idx       int    `json:"idx,omitempty"`       // 索引
	Doctype   string `json:"doctype,omitempty"`   // 单据类型

	// 项目基本信息
	HasItemScanned int    `json:"has_item_scanned,omitempty"` // 是否已扫描物品
	ItemCode       string `json:"item_code,omitempty"`        // 项目编码
	ItemName       string `json:"item_name,omitempty"`        // 项目名称
	ItemGroup      string `json:"item_group,omitempty"`       // 项目组
	Description    string `json:"description,omitempty"`      // 描述
	Image          string `json:"image,omitempty"`            // 图片

	// 数量和单位
	Qty              float64 `json:"qty,omitempty"`               // 数量
	StockUom         string  `json:"stock_uom,omitempty"`         // 库存单位
	Uom              string  `json:"uom,omitempty"`               // 单位
	ConversionFactor float64 `json:"conversion_factor,omitempty"` // 转换因子
	StockQty         float64 `json:"stock_qty,omitempty"`         // 库存数量
	ReturnedQty      float64 `json:"returned_qty,omitempty"`      // 退回数量

	// 价格相关
	PriceListRate     float64 `json:"price_list_rate,omitempty"`      // 价目表价格
	BasePriceListRate float64 `json:"base_price_list_rate,omitempty"` // 基础价目表价格
	Rate              float64 `json:"rate,omitempty"`                 // 价格
	BaseRate          float64 `json:"base_rate,omitempty"`            // 基础价格
	StockUomRate      float64 `json:"stock_uom_rate,omitempty"`       // 库存单位价格
	NetRate           float64 `json:"net_rate,omitempty"`             // 净价格
	BaseNetRate       float64 `json:"base_net_rate,omitempty"`        // 基础净价格
	IncomingRate      float64 `json:"incoming_rate,omitempty"`        // 进货价格

	// 金额相关
	Amount        float64 `json:"amount,omitempty"`          // 金额
	BaseAmount    float64 `json:"base_amount,omitempty"`     // 基础金额
	NetAmount     float64 `json:"net_amount,omitempty"`      // 净金额
	BaseNetAmount float64 `json:"base_net_amount,omitempty"` // 基础净金额
	BilledAmt     float64 `json:"billed_amt,omitempty"`      // 已开票金额

	// 折扣和利润率相关
	MarginType                string  `json:"margin_type,omitempty"`                 // 利润率类型
	MarginRateOrAmount        float64 `json:"margin_rate_or_amount,omitempty"`       // 利润率或金额
	RateWithMargin            float64 `json:"rate_with_margin,omitempty"`            // 含利润率价格
	BaseRateWithMargin        float64 `json:"base_rate_with_margin,omitempty"`       // 基础含利润率价格
	DiscountPercentage        float64 `json:"discount_percentage,omitempty"`         // 折扣百分比
	DiscountAmount            float64 `json:"discount_amount,omitempty"`             // 折扣金额
	DistributedDiscountAmount float64 `json:"distributed_discount_amount,omitempty"` // 分配折扣金额

	// 状态标识
	IsFreeItem             int `json:"is_free_item,omitempty"`              // 是否免费项目
	GrantCommission        int `json:"grant_commission,omitempty"`          // 授予佣金
	AllowZeroValuationRate int `json:"allow_zero_valuation_rate,omitempty"` // 允许零评估价格
	UseSerialBatchFields   int `json:"use_serial_batch_fields,omitempty"`   // 使用序列批次字段
	PageBreak              int `json:"page_break,omitempty"`                // 分页符

	// 库存相关
	ActualQty         float64 `json:"actual_qty,omitempty"`          // 实际数量
	ActualBatchQty    float64 `json:"actual_batch_qty,omitempty"`    // 实际批次数量
	CompanyTotalStock float64 `json:"company_total_stock,omitempty"` // 公司总库存
	InstalledQty      float64 `json:"installed_qty,omitempty"`       // 已安装数量
	PackedQty         float64 `json:"packed_qty,omitempty"`          // 已包装数量
	ReceivedQty       float64 `json:"received_qty,omitempty"`        // 已收数量

	// 重量相关
	WeightPerUnit float64 `json:"weight_per_unit,omitempty"` // 单位重量
	TotalWeight   float64 `json:"total_weight,omitempty"`    // 总重量

	// 仓库和账户
	Warehouse      string `json:"warehouse,omitempty"`       // 仓库
	ExpenseAccount string `json:"expense_account,omitempty"` // 费用账户
	CostCenter     string `json:"cost_center,omitempty"`     // 成本中心

	// 关联信息
	AgainstSalesOrder   string `json:"against_sales_order,omitempty"`   // 针对销售订单
	SoDetail            string `json:"so_detail,omitempty"`             // 销售订单详情
	MaterialRequest     string `json:"material_request,omitempty"`      // 物料请求
	PurchaseOrder       string `json:"purchase_order,omitempty"`        // 采购订单
	PurchaseOrderItem   string `json:"purchase_order_item,omitempty"`   // 采购订单项目
	MaterialRequestItem string `json:"material_request_item,omitempty"` // 物料请求项目

	// 税费相关
	ItemTaxRate string `json:"item_tax_rate,omitempty"` // 项目税率

	// 父级关联
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
}
