package selling

// SalesOrder 销售订单主结构
// 用于表示销售订单的完整信息
type SalesOrder struct {
	Name                         string            `json:"name,omitempty"`                           // 销售订单名称
	Owner                        string            `json:"owner,omitempty"`                          // 所有者
	Creation                     string            `json:"creation,omitempty"`                       // 创建时间
	Modified                     string            `json:"modified,omitempty"`                       // 修改时间
	ModifiedBy                   string            `json:"modified_by,omitempty"`                    // 修改者
	Docstatus                    int               `json:"docstatus,omitempty"`                      // 文档状态：0-已保存, 1-已提交, 2-已取消
	Idx                          int               `json:"idx,omitempty"`                            // 索引
	Title                        string            `json:"title,omitempty"`                          // 标题
	NamingSeries                 string            `json:"naming_series,omitempty"`                  // 命名系列
	Customer                     string            `json:"customer,omitempty"`                       // 客户
	CustomerName                 string            `json:"customer_name,omitempty"`                  // 客户名称
	OrderType                    string            `json:"order_type,omitempty"`                     // 订单类型
	TransactionDate              string            `json:"transaction_date,omitempty"`               // 交易日期
	DeliveryDate                 string            `json:"delivery_date,omitempty"`                  // 交付日期
	Company                      string            `json:"company,omitempty"`                        // 公司
	SkipDeliveryNote             int               `json:"skip_delivery_note,omitempty"`             // 跳过交付单
	HasUnitPriceItems            int               `json:"has_unit_price_items,omitempty"`           // 是否有单价商品
	Currency                     string            `json:"currency,omitempty"`                       // 货币
	ConversionRate               float64           `json:"conversion_rate,omitempty"`                // 汇率
	SellingPriceList             string            `json:"selling_price_list,omitempty"`             // 销售价格表
	PriceListCurrency            string            `json:"price_list_currency,omitempty"`            // 价格表货币
	PlcConversionRate            float64           `json:"plc_conversion_rate,omitempty"`            // 价格表汇率
	IgnorePricingRule            int               `json:"ignore_pricing_rule,omitempty"`            // 忽略定价规则
	SetWarehouse                 string            `json:"set_warehouse,omitempty"`                  // 设置仓库
	ReserveStock                 int               `json:"reserve_stock,omitempty"`                  // 预留库存
	TotalQty                     float64           `json:"total_qty,omitempty"`                      // 总数量
	TotalNetWeight               float64           `json:"total_net_weight,omitempty"`               // 总净重
	BaseTotal                    float64           `json:"base_total,omitempty"`                     // 基础总计
	BaseNetTotal                 float64           `json:"base_net_total,omitempty"`                 // 基础净总计
	Total                        float64           `json:"total,omitempty"`                          // 总计
	NetTotal                     float64           `json:"net_total,omitempty"`                      // 净总计
	TaxCategory                  string            `json:"tax_category,omitempty"`                   // 税费类别
	BaseTotalTaxesAndCharges     float64           `json:"base_total_taxes_and_charges,omitempty"`   // 基础税费和费用总计
	TotalTaxesAndCharges         float64           `json:"total_taxes_and_charges,omitempty"`        // 税费和费用总计
	BaseGrandTotal               float64           `json:"base_grand_total,omitempty"`               // 基础总金额
	BaseRoundingAdjustment       float64           `json:"base_rounding_adjustment,omitempty"`       // 基础四舍五入调整
	BaseRoundedTotal             float64           `json:"base_rounded_total,omitempty"`             // 基础四舍五入总计
	BaseInWords                  string            `json:"base_in_words,omitempty"`                  // 基础金额大写
	GrandTotal                   float64           `json:"grand_total,omitempty"`                    // 总金额
	RoundingAdjustment           float64           `json:"rounding_adjustment,omitempty"`            // 四舍五入调整
	RoundedTotal                 float64           `json:"rounded_total,omitempty"`                  // 四舍五入总计
	InWords                      string            `json:"in_words,omitempty"`                       // 金额大写
	AdvancePaid                  float64           `json:"advance_paid,omitempty"`                   // 预付款
	DisableRoundedTotal          int               `json:"disable_rounded_total,omitempty"`          // 禁用四舍五入总计
	ApplyDiscountOn              string            `json:"apply_discount_on,omitempty"`              // 折扣应用位置
	BaseDiscountAmount           float64           `json:"base_discount_amount,omitempty"`           // 基础折扣金额
	AdditionalDiscountPercentage float64           `json:"additional_discount_percentage,omitempty"` // 额外折扣百分比
	DiscountAmount               float64           `json:"discount_amount,omitempty"`                // 折扣金额
	Status                       string            `json:"status,omitempty"`                         // 状态
	DeliveryStatus               string            `json:"delivery_status,omitempty"`                // 交付状态
	PerDelivered                 float64           `json:"per_delivered,omitempty"`                  // 已交付百分比
	PerBilled                    float64           `json:"per_billed,omitempty"`                     // 已开票百分比
	PerPicked                    float64           `json:"per_picked,omitempty"`                     // 已拣货百分比
	BillingStatus                string            `json:"billing_status,omitempty"`                 // 开票状态
	AmountEligibleForCommission  float64           `json:"amount_eligible_for_commission,omitempty"` // 符合佣金条件的金额
	CommissionRate               float64           `json:"commission_rate,omitempty"`                // 佣金率
	TotalCommission              float64           `json:"total_commission,omitempty"`               // 总佣金
	LoyaltyPoints                int               `json:"loyalty_points,omitempty"`                 // 忠诚度积分
	LoyaltyAmount                float64           `json:"loyalty_amount,omitempty"`                 // 忠诚度金额
	GroupSameItems               int               `json:"group_same_items,omitempty"`               // 分组相同商品
	Language                     string            `json:"language,omitempty"`                       // 语言
	IsInternalCustomer           int               `json:"is_internal_customer,omitempty"`           // 是否内部客户
	RepresentsCompany            string            `json:"represents_company,omitempty"`             // 代表公司
	PartyAccountCurrency         string            `json:"party_account_currency,omitempty"`         // 对方账户货币
	Doctype                      string            `json:"doctype,omitempty"`                        // 文档类型
	Items                        []SalesOrderItem  `json:"items,omitempty"`                          // 商品项目列表
	PricingRules                 []interface{}     `json:"pricing_rules,omitempty"`                  // 定价规则列表
	PackedItems                  []interface{}     `json:"packed_items,omitempty"`                   // 包装商品列表
	SalesTeam                    []interface{}     `json:"sales_team,omitempty"`                     // 销售团队列表
	PaymentSchedule              []PaymentSchedule `json:"payment_schedule,omitempty"`               // 付款计划列表
	Taxes                        []interface{}     `json:"taxes,omitempty"`                          // 税费列表
}

// SalesOrderItem 销售订单商品项目结构
// 用于表示销售订单中的商品项目信息
type SalesOrderItem struct {
	Name                                  string  `json:"name,omitempty"`                                        // 项目名称
	Owner                                 string  `json:"owner,omitempty"`                                       // 所有者
	Creation                              string  `json:"creation,omitempty"`                                    // 创建时间
	Modified                              string  `json:"modified,omitempty"`                                    // 修改时间
	ModifiedBy                            string  `json:"modified_by,omitempty"`                                 // 修改者
	Docstatus                             int     `json:"docstatus,omitempty"`                                   // 文档状态
	Idx                                   int     `json:"idx,omitempty"`                                         // 索引
	ItemCode                              string  `json:"item_code,omitempty"`                                   // 商品编码
	EnsureDeliveryBasedOnProducedSerialNo int     `json:"ensure_delivery_based_on_produced_serial_no,omitempty"` // 确保基于生产序列号的交付
	IsStockItem                           int     `json:"is_stock_item,omitempty"`                               // 是否库存商品
	ReserveStock                          int     `json:"reserve_stock,omitempty"`                               // 预留库存
	DeliveryDate                          string  `json:"delivery_date,omitempty"`                               // 交付日期
	ItemName                              string  `json:"item_name,omitempty"`                                   // 商品名称
	Description                           string  `json:"description,omitempty"`                                 // 描述
	ItemGroup                             string  `json:"item_group,omitempty"`                                  // 商品组
	Image                                 string  `json:"image,omitempty"`                                       // 图片
	Qty                                   float64 `json:"qty,omitempty"`                                         // 数量
	StockUom                              string  `json:"stock_uom,omitempty"`                                   // 库存单位
	Uom                                   string  `json:"uom,omitempty"`                                         // 单位
	ConversionFactor                      float64 `json:"conversion_factor,omitempty"`                           // 转换因子
	StockQty                              float64 `json:"stock_qty,omitempty"`                                   // 库存数量
	StockReservedQty                      float64 `json:"stock_reserved_qty,omitempty"`                          // 库存预留数量
	PriceListRate                         float64 `json:"price_list_rate,omitempty"`                             // 价格表费率
	BasePriceListRate                     float64 `json:"base_price_list_rate,omitempty"`                        // 基础价格表费率
	MarginType                            string  `json:"margin_type,omitempty"`                                 // 利润率类型
	MarginRateOrAmount                    float64 `json:"margin_rate_or_amount,omitempty"`                       // 利润率或金额
	RateWithMargin                        float64 `json:"rate_with_margin,omitempty"`                            // 含利润率费率
	DiscountPercentage                    float64 `json:"discount_percentage,omitempty"`                         // 折扣百分比
	DiscountAmount                        float64 `json:"discount_amount,omitempty"`                             // 折扣金额
	DistributedDiscountAmount             float64 `json:"distributed_discount_amount,omitempty"`                 // 分配折扣金额
	BaseRateWithMargin                    float64 `json:"base_rate_with_margin,omitempty"`                       // 基础含利润率费率
	Rate                                  float64 `json:"rate,omitempty"`                                        // 费率
	Amount                                float64 `json:"amount,omitempty"`                                      // 金额
	BaseRate                              float64 `json:"base_rate,omitempty"`                                   // 基础费率
	BaseAmount                            float64 `json:"base_amount,omitempty"`                                 // 基础金额
	PricingRules                          string  `json:"pricing_rules,omitempty"`                               // 定价规则
	StockUomRate                          float64 `json:"stock_uom_rate,omitempty"`                              // 库存单位费率
	IsFreeItem                            int     `json:"is_free_item,omitempty"`                                // 是否免费商品
	GrantCommission                       int     `json:"grant_commission,omitempty"`                            // 授予佣金
	NetRate                               float64 `json:"net_rate,omitempty"`                                    // 净费率
	NetAmount                             float64 `json:"net_amount,omitempty"`                                  // 净金额
	BaseNetRate                           float64 `json:"base_net_rate,omitempty"`                               // 基础净费率
	BaseNetAmount                         float64 `json:"base_net_amount,omitempty"`                             // 基础净金额
	BilledAmt                             float64 `json:"billed_amt,omitempty"`                                  // 已开票金额
	ValuationRate                         float64 `json:"valuation_rate,omitempty"`                              // 估值费率
	GrossProfit                           float64 `json:"gross_profit,omitempty"`                                // 毛利
	DeliveredBySupplier                   int     `json:"delivered_by_supplier,omitempty"`                       // 由供应商交付
	WeightPerUnit                         float64 `json:"weight_per_unit,omitempty"`                             // 单位重量
	TotalWeight                           float64 `json:"total_weight,omitempty"`                                // 总重量
	Warehouse                             string  `json:"warehouse,omitempty"`                                   // 仓库
	AgainstBlanketOrder                   int     `json:"against_blanket_order,omitempty"`                       // 针对一揽子订单
	BlanketOrderRate                      float64 `json:"blanket_order_rate,omitempty"`                          // 一揽子订单费率
	ActualQty                             float64 `json:"actual_qty,omitempty"`                                  // 实际数量
	CompanyTotalStock                     float64 `json:"company_total_stock,omitempty"`                         // 公司总库存
	ProjectedQty                          float64 `json:"projected_qty,omitempty"`                               // 预计数量
	OrderedQty                            float64 `json:"ordered_qty,omitempty"`                                 // 已订购数量
	PlannedQty                            float64 `json:"planned_qty,omitempty"`                                 // 已计划数量
	ProductionPlanQty                     float64 `json:"production_plan_qty,omitempty"`                         // 生产计划数量
	WorkOrderQty                          float64 `json:"work_order_qty,omitempty"`                              // 工单数量
	DeliveredQty                          float64 `json:"delivered_qty,omitempty"`                               // 已交付数量
	ProducedQty                           float64 `json:"produced_qty,omitempty"`                                // 已生产数量
	ReturnedQty                           float64 `json:"returned_qty,omitempty"`                                // 已退货数量
	PickedQty                             float64 `json:"picked_qty,omitempty"`                                  // 已拣货数量
	PageBreak                             int     `json:"page_break,omitempty"`                                  // 分页符
	ItemTaxRate                           string  `json:"item_tax_rate,omitempty"`                               // 商品税费率
	TransactionDate                       string  `json:"transaction_date,omitempty"`                            // 交易日期
	CostCenter                            string  `json:"cost_center,omitempty"`                                 // 成本中心
	Parent                                string  `json:"parent,omitempty"`                                      // 父级
	Parentfield                           string  `json:"parentfield,omitempty"`                                 // 父级字段
	Parenttype                            string  `json:"parenttype,omitempty"`                                  // 父级类型
	Doctype                               string  `json:"doctype,omitempty"`                                     // 文档类型
}

// PaymentSchedule 付款计划结构
// 用于表示销售订单中的付款计划信息
type PaymentSchedule struct {
	Name              string  `json:"name,omitempty"`                // 名称
	Owner             string  `json:"owner,omitempty"`               // 所有者
	Creation          string  `json:"creation,omitempty"`            // 创建时间
	Modified          string  `json:"modified,omitempty"`            // 修改时间
	ModifiedBy        string  `json:"modified_by,omitempty"`         // 修改者
	Docstatus         int     `json:"docstatus,omitempty"`           // 文档状态
	Idx               int     `json:"idx,omitempty"`                 // 索引
	DueDate           string  `json:"due_date,omitempty"`            // 到期日期
	InvoicePortion    float64 `json:"invoice_portion,omitempty"`     // 发票部分
	Discount          float64 `json:"discount,omitempty"`            // 折扣
	PaymentAmount     float64 `json:"payment_amount,omitempty"`      // 付款金额
	Outstanding       float64 `json:"outstanding,omitempty"`         // 未结金额
	PaidAmount        float64 `json:"paid_amount,omitempty"`         // 已付金额
	DiscountedAmount  float64 `json:"discounted_amount,omitempty"`   // 折扣金额
	BasePaymentAmount float64 `json:"base_payment_amount,omitempty"` // 基础付款金额
	BaseOutstanding   float64 `json:"base_outstanding,omitempty"`    // 基础未结金额
	BasePaidAmount    float64 `json:"base_paid_amount,omitempty"`    // 基础已付金额
	Parent            string  `json:"parent,omitempty"`              // 父级
	Parentfield       string  `json:"parentfield,omitempty"`         // 父级字段
	Parenttype        string  `json:"parenttype,omitempty"`          // 父级类型
	Doctype           string  `json:"doctype,omitempty"`             // 文档类型
}

// SalesOrderListReq 销售订单列表查询请求
type SalesOrderListReq struct {
	Name                 string `json:"name,omitempty"`                   // 销售订单名称（模糊查询）
	Customer             string `json:"customer,omitempty"`               // 客户
	CustomerName         string `json:"customer_name,omitempty"`          // 客户名称（模糊查询）
	Company              string `json:"company,omitempty"`                // 公司
	CompanyAbbr          string `json:"company_abbr,omitempty"`           // 公司缩写
	OrderType            string `json:"order_type,omitempty"`             // 订单类型
	Status               string `json:"status,omitempty"`                 // 订单状态
	DeliveryStatus       string `json:"delivery_status,omitempty"`        // 交付状态
	BillingStatus        string `json:"billing_status,omitempty"`         // 开票状态
	TransactionDateStart string `json:"transaction_date_start,omitempty"` // 交易日期开始
	TransactionDateEnd   string `json:"transaction_date_end,omitempty"`   // 交易日期结束
	Docstatus            int    `json:"docstatus,omitempty"`              // 文档状态：0-草稿, 1-已提交, 2-已取消
	IsInternalCustomer   int    `json:"is_internal_customer,omitempty"`   // 是否内部客户
	Limit                int    `json:"limit,omitempty"`                  // 每页数量
	LimitStart           int    `json:"limit_start,omitempty"`            // 起始位置
}
