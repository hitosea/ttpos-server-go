package erp

// POSProfile 结构体定义
// 用于表示POS配置文件的完整信息
type POSProfile struct {
	Name                          string             `json:"name,omitempty"`                               // POS配置文件名称
	Owner                         string             `json:"owner,omitempty"`                              // 所有者
	Creation                      string             `json:"creation,omitempty"`                           // 创建时间
	Modified                      string             `json:"modified,omitempty"`                           // 修改时间
	ModifiedBy                    string             `json:"modified_by,omitempty"`                        // 修改者
	Docstatus                     int                `json:"docstatus,omitempty"`                          // 文档状态
	Idx                           int                `json:"idx,omitempty"`                                // 索引
	Company                       string             `json:"company,omitempty"`                            // 公司
	Country                       string             `json:"country,omitempty"`                            // 国家
	Disabled                      int                `json:"disabled,omitempty"`                           // 是否禁用
	Warehouse                     string             `json:"warehouse,omitempty"`                          // 仓库
	HideImages                    int                `json:"hide_images,omitempty"`                        // 是否隐藏图片
	HideUnavailableItems          int                `json:"hide_unavailable_items,omitempty"`             // 是否隐藏不可用商品
	AutoAddItemToCart             int                `json:"auto_add_item_to_cart,omitempty"`              // 是否自动添加商品到购物车
	ValidateStockOnSave           int                `json:"validate_stock_on_save,omitempty"`             // 保存时是否验证库存
	PrintReceiptOnOrderComplete   int                `json:"print_receipt_on_order_complete,omitempty"`    // 订单完成时是否打印收据
	UpdateStock                   int                `json:"update_stock,omitempty"`                       // 是否更新库存
	IgnorePricingRule             int                `json:"ignore_pricing_rule,omitempty"`                // 是否忽略定价规则
	AllowRateChange               int                `json:"allow_rate_change,omitempty"`                  // 是否允许汇率变更
	AllowDiscountChange           int                `json:"allow_discount_change,omitempty"`              // 是否允许折扣变更
	DisableGrandTotalToDefaultMop int                `json:"disable_grand_total_to_default_mop,omitempty"` // 是否禁用总计到默认支付方式
	AllowPartialPayment           int                `json:"allow_partial_payment,omitempty"`              // 是否允许部分支付
	SellingPriceList              string             `json:"selling_price_list,omitempty"`                 // 销售价格表
	Currency                      string             `json:"currency,omitempty"`                           // 货币
	WriteOffAccount               string             `json:"write_off_account,omitempty"`                  // 核销账户
	WriteOffCostCenter            string             `json:"write_off_cost_center,omitempty"`              // 核销成本中心
	WriteOffLimit                 float64            `json:"write_off_limit,omitempty"`                    // 核销限额
	DisableRoundedTotal           int                `json:"disable_rounded_total,omitempty"`              // 是否禁用四舍五入总计
	ApplyDiscountOn               string             `json:"apply_discount_on,omitempty"`                  // 折扣应用位置
	Doctype                       string             `json:"doctype,omitempty"`                            // 文档类型
	CustomerGroups                []interface{}      `json:"customer_groups,omitempty"`                    // 客户组列表
	ApplicableForUsers            []POSProfileUser   `json:"applicable_for_users,omitempty"`               // 适用用户列表
	ItemGroups                    []interface{}      `json:"item_groups,omitempty"`                        // 商品组列表
	Payments                      []POSPaymentMethod `json:"payments,omitempty"`                           // 支付方式列表
	Branch                        string             `json:"branch,omitempty"`                             // 分公司
}

// POSProfileUser 结构体定义
// 用于表示POS配置文件中的用户信息
type POSProfileUser struct {
	Name        string `json:"name,omitempty"`        // 名称
	Owner       string `json:"owner,omitempty"`       // 所有者
	Creation    string `json:"creation,omitempty"`    // 创建时间
	Modified    string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy  string `json:"modified_by,omitempty"` // 修改者
	Docstatus   int    `json:"docstatus,omitempty"`   // 文档状态
	Idx         int    `json:"idx,omitempty"`         // 索引
	Default     int    `json:"default,omitempty"`     // 是否默认
	User        string `json:"user,omitempty"`        // 用户
	Parent      string `json:"parent,omitempty"`      // 父级
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
	Doctype     string `json:"doctype,omitempty"`     // 文档类型
}

// POSPaymentMethod 结构体定义
// 用于表示POS配置文件中的支付方式信息
type POSPaymentMethod struct {
	Name           string `json:"name,omitempty"`             // 名称
	Owner          string `json:"owner,omitempty"`            // 所有者
	Creation       string `json:"creation,omitempty"`         // 创建时间
	Modified       string `json:"modified,omitempty"`         // 修改时间
	ModifiedBy     string `json:"modified_by,omitempty"`      // 修改者
	Docstatus      int    `json:"docstatus,omitempty"`        // 文档状态
	Idx            int    `json:"idx,omitempty"`              // 索引
	Default        int    `json:"default,omitempty"`          // 是否默认
	AllowInReturns int    `json:"allow_in_returns,omitempty"` // 是否允许退货
	ModeOfPayment  string `json:"mode_of_payment,omitempty"`  // 支付方式
	Parent         string `json:"parent,omitempty"`           // 父级
	Parentfield    string `json:"parentfield,omitempty"`      // 父级字段
	Parenttype     string `json:"parenttype,omitempty"`       // 父级类型
	Doctype        string `json:"doctype,omitempty"`          // 文档类型
}

// ModeOfPayment 结构体定义
// 用于表示支付方式的完整信息
type ModeOfPayment struct {
	Name          string                 `json:"name,omitempty"`            // 支付方式名称
	Owner         string                 `json:"owner,omitempty"`           // 所有者
	Creation      string                 `json:"creation,omitempty"`        // 创建时间
	Modified      string                 `json:"modified,omitempty"`        // 修改时间
	ModifiedBy    string                 `json:"modified_by,omitempty"`     // 修改者
	Docstatus     int                    `json:"docstatus,omitempty"`       // 文档状态
	Idx           int                    `json:"idx,omitempty"`             // 索引
	ModeOfPayment string                 `json:"mode_of_payment,omitempty"` // 支付方式
	Enabled       int                    `json:"enabled,omitempty"`         // 是否启用
	Type          string                 `json:"type,omitempty"`            // 类型
	CustomBranch  string                 `json:"custom_branch,omitempty"`   // 自定义分支
	Doctype       string                 `json:"doctype,omitempty"`         // 文档类型
	Accounts      []ModeOfPaymentAccount `json:"accounts,omitempty"`        // 账户列表
}

// ModeOfPaymentAccount 结构体定义
// 用于表示支付方式中的账户信息
type ModeOfPaymentAccount struct {
	Name           string `json:"name,omitempty"`            // 名称
	Owner          string `json:"owner,omitempty"`           // 所有者
	Creation       string `json:"creation,omitempty"`        // 创建时间
	Modified       string `json:"modified,omitempty"`        // 修改时间
	ModifiedBy     string `json:"modified_by,omitempty"`     // 修改者
	Docstatus      int    `json:"docstatus,omitempty"`       // 文档状态
	Idx            int    `json:"idx,omitempty"`             // 索引
	Company        string `json:"company,omitempty"`         // 公司
	DefaultAccount string `json:"default_account,omitempty"` // 默认账户
	Parent         string `json:"parent,omitempty"`          // 父级
	Parentfield    string `json:"parentfield,omitempty"`     // 父级字段
	Parenttype     string `json:"parenttype,omitempty"`      // 父级类型
	Doctype        string `json:"doctype,omitempty"`         // 文档类型
}

// POSOpeningEntry POS开单条目主结构
type POSOpeningEntry struct {
	Name            string                  `json:"name,omitempty"`              // POS开单条目名称
	Owner           string                  `json:"owner,omitempty"`             // 所有者
	Creation        string                  `json:"creation,omitempty"`          // 创建时间
	Modified        string                  `json:"modified,omitempty"`          // 修改时间
	ModifiedBy      string                  `json:"modified_by,omitempty"`       // 修改者
	Docstatus       int                     `json:"docstatus,omitempty"`         // 文档状态
	Idx             int                     `json:"idx,omitempty"`               // 索引
	PeriodStartDate string                  `json:"period_start_date,omitempty"` // 期间开始日期
	Status          string                  `json:"status,omitempty"`            // 状态
	PostingDate     string                  `json:"posting_date,omitempty"`      // 过账日期
	SetPostingDate  int                     `json:"set_posting_date,omitempty"`  // 设置过账日期
	Company         string                  `json:"company,omitempty"`           // 公司
	PosProfile      string                  `json:"pos_profile,omitempty"`       // POS配置文件
	User            string                  `json:"user,omitempty"`              // 用户
	Doctype         string                  `json:"doctype,omitempty"`           // 文档类型
	BalanceDetails  []POSOpeningEntryDetail `json:"balance_details,omitempty"`   // 余额详情列表
}

// POSOpeningEntryDetail POS开单条目详情结构
type POSOpeningEntryDetail struct {
	Name          string  `json:"name,omitempty"`            // 详情名称
	Owner         string  `json:"owner,omitempty"`           // 所有者
	Creation      string  `json:"creation,omitempty"`        // 创建时间
	Modified      string  `json:"modified,omitempty"`        // 修改时间
	ModifiedBy    string  `json:"modified_by,omitempty"`     // 修改者
	Docstatus     int     `json:"docstatus,omitempty"`       // 文档状态
	Idx           int     `json:"idx,omitempty"`             // 索引
	ModeOfPayment string  `json:"mode_of_payment,omitempty"` // 支付方式
	OpeningAmount float64 `json:"opening_amount"`            // 开盘金额
	Parent        string  `json:"parent,omitempty"`          // 父级
	Parentfield   string  `json:"parentfield,omitempty"`     // 父级字段
	Parenttype    string  `json:"parenttype,omitempty"`      // 父级类型
	Doctype       string  `json:"doctype,omitempty"`         // 文档类型
}

// POSCloseEntry POS关单条目主结构
type POSCloseEntry struct {
	Name                  string                     `json:"name,omitempty"`                   // POS关单条目名称
	Owner                 string                     `json:"owner,omitempty"`                  // 所有者
	Creation              string                     `json:"creation,omitempty"`               // 创建时间
	Modified              string                     `json:"modified,omitempty"`               // 修改时间
	ModifiedBy            string                     `json:"modified_by,omitempty"`            // 修改者
	Docstatus             int                        `json:"docstatus,omitempty"`              // 文档状态
	Idx                   int                        `json:"idx,omitempty"`                    // 索引
	PeriodStartDate       string                     `json:"period_start_date,omitempty"`      // 期间开始日期
	PeriodEndDate         string                     `json:"period_end_date,omitempty"`        // 期间结束日期
	PostingDate           string                     `json:"posting_date,omitempty"`           // 过账日期
	PostingTime           string                     `json:"posting_time,omitempty"`           // 过账时间
	PosOpeningEntry       string                     `json:"pos_opening_entry,omitempty"`      // POS开单条目
	Status                string                     `json:"status,omitempty"`                 // 状态
	Company               string                     `json:"company,omitempty"`                // 公司
	PosProfile            string                     `json:"pos_profile,omitempty"`            // POS配置文件
	User                  string                     `json:"user,omitempty"`                   // 用户
	GrandTotal            float64                    `json:"grand_total,omitempty"`            // 总计金额
	NetTotal              float64                    `json:"net_total,omitempty"`              // 净额
	TotalQuantity         int                        `json:"total_quantity,omitempty"`         // 总数量
	ErrorMessage          string                     `json:"error_message,omitempty"`          // 错误消息
	Doctype               string                     `json:"doctype,omitempty"`                // 文档类型
	PosTransactions       []POSTransaction           `json:"pos_transactions,omitempty"`       // POS交易列表
	Taxes                 []interface{}              `json:"taxes,omitempty"`                  // 税费列表
	PaymentReconciliation []POSPaymentReconciliation `json:"payment_reconciliation,omitempty"` // 支付对账列表
}

// POSTransaction POS交易结构
type POSTransaction struct {
	Name          string  `json:"name,omitempty"`           // 交易名称
	Owner         string  `json:"owner,omitempty"`          // 所有者
	Creation      string  `json:"creation,omitempty"`       // 创建时间
	Modified      string  `json:"modified,omitempty"`       // 修改时间
	ModifiedBy    string  `json:"modified_by,omitempty"`    // 修改者
	Docstatus     int     `json:"docstatus,omitempty"`      // 文档状态
	Idx           int     `json:"idx,omitempty"`            // 索引
	PosInvoice    string  `json:"pos_invoice,omitempty"`    // POS发票
	PostingDate   string  `json:"posting_date,omitempty"`   // 过账日期
	Customer      string  `json:"customer,omitempty"`       // 客户
	GrandTotal    float64 `json:"grand_total,omitempty"`    // 总计金额
	IsReturn      int     `json:"is_return,omitempty"`      // 是否退货
	ReturnAgainst string  `json:"return_against,omitempty"` // 退货对应
	Parent        string  `json:"parent,omitempty"`         // 父级
	Parentfield   string  `json:"parentfield,omitempty"`    // 父级字段
	Parenttype    string  `json:"parenttype,omitempty"`     // 父级类型
	Doctype       string  `json:"doctype,omitempty"`        // 文档类型
}

// POSPaymentReconciliation POS支付对账结构
type POSPaymentReconciliation struct {
	Name           string  `json:"name,omitempty"`            // 对账名称
	Owner          string  `json:"owner,omitempty"`           // 所有者
	Creation       string  `json:"creation,omitempty"`        // 创建时间
	Modified       string  `json:"modified,omitempty"`        // 修改时间
	ModifiedBy     string  `json:"modified_by,omitempty"`     // 修改者
	Docstatus      int     `json:"docstatus,omitempty"`       // 文档状态
	Idx            int     `json:"idx,omitempty"`             // 索引
	ModeOfPayment  string  `json:"mode_of_payment,omitempty"` // 支付方式
	OpeningAmount  float64 `json:"opening_amount"`            // 开盘金额
	ExpectedAmount float64 `json:"expected_amount,omitempty"` // 预期金额
	ClosingAmount  float64 `json:"closing_amount"`            // 收盘金额
	Difference     float64 `json:"difference,omitempty"`      // 差异金额
	Parent         string  `json:"parent,omitempty"`          // 父级
	Parentfield    string  `json:"parentfield,omitempty"`     // 父级字段
	Parenttype     string  `json:"parenttype,omitempty"`      // 父级类型
	Doctype        string  `json:"doctype,omitempty"`         // 文档类型
}

// SalesInvoiceDoc 销售发票完整文档结构（ERPNext 响应）
type SalesInvoiceDoc struct {
	Name                                   string                `json:"name,omitempty"`                                       // 销售发票名称
	Owner                                  string                `json:"owner,omitempty"`                                      // 所有者
	Creation                               string                `json:"creation,omitempty"`                                   // 创建时间
	Modified                               string                `json:"modified,omitempty"`                                   // 修改时间
	ModifiedBy                             string                `json:"modified_by,omitempty"`                                // 修改者
	Docstatus                              int                   `json:"docstatus,omitempty"`                                  // 文档状态
	Idx                                    int                   `json:"idx,omitempty"`                                        // 索引
	Title                                  string                `json:"title,omitempty"`                                      // 标题
	NamingSeries                           string                `json:"naming_series,omitempty"`                              // 命名系列
	Customer                               string                `json:"customer,omitempty"`                                   // 客户
	CustomerName                           string                `json:"customer_name,omitempty"`                              // 客户名称
	Company                                string                `json:"company,omitempty"`                                    // 公司
	PostingDate                            string                `json:"posting_date,omitempty"`                               // 过账日期
	PostingTime                            string                `json:"posting_time,omitempty"`                               // 过账时间
	SetPostingTime                         int                   `json:"set_posting_time,omitempty"`                           // 设置过账时间
	DueDate                                string                `json:"due_date,omitempty"`                                   // 到期日期
	IsPos                                  int                   `json:"is_pos,omitempty"`                                     // 是否POS
	PosProfile                             string                `json:"pos_profile,omitempty"`                                // POS配置文件
	IsConsolidated                         int                   `json:"is_consolidated,omitempty"`                            // 是否合并
	IsReturn                               int                   `json:"is_return,omitempty"`                                  // 是否退货
	UpdateOutstandingForSelf               int                   `json:"update_outstanding_for_self,omitempty"`                // 更新自身未结金额
	UpdateBilledAmountInSalesOrder         int                   `json:"update_billed_amount_in_sales_order,omitempty"`        // 更新销售订单中的已开票金额
	UpdateBilledAmountInDeliveryNote       int                   `json:"update_billed_amount_in_delivery_note,omitempty"`      // 更新送货单中的已开票金额
	IsDebitNote                            int                   `json:"is_debit_note,omitempty"`                              // 是否借记单
	Currency                               string                `json:"currency,omitempty"`                                   // 货币
	ConversionRate                         float64               `json:"conversion_rate,omitempty"`                            // 汇率
	SellingPriceList                       string                `json:"selling_price_list,omitempty"`                         // 销售价格表
	PriceListCurrency                      string                `json:"price_list_currency,omitempty"`                        // 价格表货币
	PlcConversionRate                      float64               `json:"plc_conversion_rate,omitempty"`                        // 价格表汇率
	IgnorePricingRule                      int                   `json:"ignore_pricing_rule,omitempty"`                        // 忽略定价规则
	UpdateStock                            int                   `json:"update_stock,omitempty"`                               // 更新库存
	SetWarehouse                           string                `json:"set_warehouse,omitempty"`                              // 设置仓库
	TotalQty                               int                   `json:"total_qty,omitempty"`                                  // 总数量
	TotalNetWeight                         float64               `json:"total_net_weight,omitempty"`                           // 总净重
	BaseTotal                              float64               `json:"base_total,omitempty"`                                 // 基础总计
	BaseNetTotal                           float64               `json:"base_net_total,omitempty"`                             // 基础净额
	Total                                  float64               `json:"total,omitempty"`                                      // 总计
	NetTotal                               float64               `json:"net_total,omitempty"`                                  // 净额
	TaxCategory                            string                `json:"tax_category,omitempty"`                               // 税费类别
	BaseTotalTaxesAndCharges               float64               `json:"base_total_taxes_and_charges,omitempty"`               // 基础税费和费用总计
	TotalTaxesAndCharges                   float64               `json:"total_taxes_and_charges,omitempty"`                    // 税费和费用总计
	BaseGrandTotal                         float64               `json:"base_grand_total,omitempty"`                           // 基础总计
	BaseRoundingAdjustment                 float64               `json:"base_rounding_adjustment,omitempty"`                   // 基础四舍五入调整
	BaseRoundedTotal                       float64               `json:"base_rounded_total,omitempty"`                         // 基础四舍五入总计
	BaseInWords                            string                `json:"base_in_words,omitempty"`                              // 基础金额大写
	GrandTotal                             float64               `json:"grand_total,omitempty"`                                // 总计
	RoundingAdjustment                     float64               `json:"rounding_adjustment,omitempty"`                        // 四舍五入调整
	UseCompanyRoundoffCostCenter           int                   `json:"use_company_roundoff_cost_center,omitempty"`           // 使用公司四舍五入成本中心
	RoundedTotal                           float64               `json:"rounded_total,omitempty"`                              // 四舍五入总计
	InWords                                string                `json:"in_words,omitempty"`                                   // 金额大写
	TotalAdvance                           float64               `json:"total_advance,omitempty"`                              // 预付款总计
	OutstandingAmount                      float64               `json:"outstanding_amount,omitempty"`                         // 未结金额
	DisableRoundedTotal                    int                   `json:"disable_rounded_total,omitempty"`                      // 禁用四舍五入总计
	ApplyDiscountOn                        string                `json:"apply_discount_on,omitempty"`                          // 折扣应用位置
	BaseDiscountAmount                     float64               `json:"base_discount_amount,omitempty"`                       // 基础折扣金额
	IsCashOrNonTradeDiscount               int                   `json:"is_cash_or_non_trade_discount,omitempty"`              // 是否现金或非贸易折扣
	AdditionalDiscountPercentage           float64               `json:"additional_discount_percentage,omitempty"`             // 额外折扣百分比
	DiscountAmount                         float64               `json:"discount_amount,omitempty"`                            // 折扣金额
	TotalBillingHours                      float64               `json:"total_billing_hours,omitempty"`                        // 计费小时总计
	TotalBillingAmount                     float64               `json:"total_billing_amount,omitempty"`                       // 计费金额总计
	BasePaidAmount                         float64               `json:"base_paid_amount,omitempty"`                           // 基础已付金额
	PaidAmount                             float64               `json:"paid_amount,omitempty"`                                // 已付金额
	BaseChangeAmount                       float64               `json:"base_change_amount,omitempty"`                         // 基础找零金额
	ChangeAmount                           float64               `json:"change_amount,omitempty"`                              // 找零金额
	AccountForChangeAmount                 string                `json:"account_for_change_amount,omitempty"`                  // 找零账户
	AllocateAdvancesAutomatically          int                   `json:"allocate_advances_automatically,omitempty"`            // 自动分配预付款
	OnlyIncludeAllocatedPayments           int                   `json:"only_include_allocated_payments,omitempty"`            // 仅包含已分配付款
	WriteOffAmount                         float64               `json:"write_off_amount,omitempty"`                           // 核销金额
	BaseWriteOffAmount                     float64               `json:"base_write_off_amount,omitempty"`                      // 基础核销金额
	WriteOffOutstandingAmountAutomatically int                   `json:"write_off_outstanding_amount_automatically,omitempty"` // 自动核销未结金额
	WriteOffAccount                        string                `json:"write_off_account,omitempty"`                          // 核销账户
	WriteOffCostCenter                     string                `json:"write_off_cost_center,omitempty"`                      // 核销成本中心
	RedeemLoyaltyPoints                    int                   `json:"redeem_loyalty_points,omitempty"`                      // 兑换忠诚度积分
	LoyaltyPoints                          int                   `json:"loyalty_points,omitempty"`                             // 忠诚度积分
	LoyaltyAmount                          float64               `json:"loyalty_amount,omitempty"`                             // 忠诚度金额
	IgnoreDefaultPaymentTermsTemplate      int                   `json:"ignore_default_payment_terms_template,omitempty"`      // 忽略默认付款条件模板
	PaymentTermsTemplate                   string                `json:"payment_terms_template,omitempty"`                     // 付款条件模板
	PoNo                                   string                `json:"po_no,omitempty"`                                      // 采购订单号
	DebitTo                                string                `json:"debit_to,omitempty"`                                   // 借记到
	PartyAccountCurrency                   string                `json:"party_account_currency,omitempty"`                     // 对方账户货币
	IsOpening                              string                `json:"is_opening,omitempty"`                                 // 是否期初
	AgainstIncomeAccount                   string                `json:"against_income_account,omitempty"`                     // 对应收入账户
	AmountEligibleForCommission            float64               `json:"amount_eligible_for_commission,omitempty"`             // 符合佣金条件的金额
	CommissionRate                         float64               `json:"commission_rate,omitempty"`                            // 佣金率
	TotalCommission                        float64               `json:"total_commission,omitempty"`                           // 佣金总计
	GroupSameItems                         int                   `json:"group_same_items,omitempty"`                           // 分组相同商品
	Language                               string                `json:"language,omitempty"`                                   // 语言
	Status                                 string                `json:"status,omitempty"`                                     // 状态
	CustomerGroup                          string                `json:"customer_group,omitempty"`                             // 客户组
	IsInternalCustomer                     int                   `json:"is_internal_customer,omitempty"`                       // 是否内部客户
	IsDiscounted                           int                   `json:"is_discounted,omitempty"`                              // 是否已折扣
	Remarks                                string                `json:"remarks,omitempty"`                                    // 备注
	Doctype                                string                `json:"doctype,omitempty"`                                    // 文档类型
	PricingRules                           []interface{}         `json:"pricing_rules,omitempty"`                              // 定价规则
	Advances                               []interface{}         `json:"advances,omitempty"`                                   // 预付款
	SalesTeam                              []interface{}         `json:"sales_team,omitempty"`                                 // 销售团队
	PaymentSchedule                        []interface{}         `json:"payment_schedule,omitempty"`                           // 付款计划
	Items                                  []SalesInvoiceItem    `json:"items,omitempty"`                                      // 商品项目
	Taxes                                  []interface{}         `json:"taxes,omitempty"`                                      // 税费
	Timesheets                             []interface{}         `json:"timesheets,omitempty"`                                 // 工时表
	Payments                               []SalesInvoicePayment `json:"payments,omitempty"`                                   // 付款
	PackedItems                            []interface{}         `json:"packed_items,omitempty"`                               // 包装商品
}

// SalesInvoiceItemDoc 销售发票商品项目完整结构（ERPNext 响应）
type SalesInvoiceItemDoc struct {
	Name                      string  `json:"name,omitempty"`                        // 项目名称
	Owner                     string  `json:"owner,omitempty"`                       // 所有者
	Creation                  string  `json:"creation,omitempty"`                    // 创建时间
	Modified                  string  `json:"modified,omitempty"`                    // 修改时间
	ModifiedBy                string  `json:"modified_by,omitempty"`                 // 修改者
	Docstatus                 int     `json:"docstatus,omitempty"`                   // 文档状态
	Idx                       int     `json:"idx,omitempty"`                         // 索引
	HasItemScanned            int     `json:"has_item_scanned,omitempty"`            // 是否已扫描商品
	ItemCode                  string  `json:"item_code,omitempty"`                   // 商品编码
	ItemName                  string  `json:"item_name,omitempty"`                   // 商品名称
	Description               string  `json:"description,omitempty"`                 // 描述
	ItemGroup                 string  `json:"item_group,omitempty"`                  // 商品组
	Image                     string  `json:"image,omitempty"`                       // 图片
	Qty                       int     `json:"qty,omitempty"`                         // 数量
	StockUom                  string  `json:"stock_uom,omitempty"`                   // 库存单位
	Uom                       string  `json:"uom,omitempty"`                         // 单位
	ConversionFactor          float64 `json:"conversion_factor,omitempty"`           // 转换因子
	StockQty                  float64 `json:"stock_qty,omitempty"`                   // 库存数量
	PriceListRate             float64 `json:"price_list_rate,omitempty"`             // 价格表费率
	BasePriceListRate         float64 `json:"base_price_list_rate,omitempty"`        // 基础价格表费率
	MarginType                string  `json:"margin_type,omitempty"`                 // 利润率类型
	MarginRateOrAmount        float64 `json:"margin_rate_or_amount,omitempty"`       // 利润率或金额
	RateWithMargin            float64 `json:"rate_with_margin,omitempty"`            // 含利润率费率
	DiscountPercentage        float64 `json:"discount_percentage,omitempty"`         // 折扣百分比
	DiscountAmount            float64 `json:"discount_amount,omitempty"`             // 折扣金额
	DistributedDiscountAmount float64 `json:"distributed_discount_amount,omitempty"` // 分配折扣金额
	BaseRateWithMargin        float64 `json:"base_rate_with_margin,omitempty"`       // 基础含利润率费率
	Rate                      float64 `json:"rate,omitempty"`                        // 费率
	Amount                    float64 `json:"amount,omitempty"`                      // 金额
	BaseRate                  float64 `json:"base_rate,omitempty"`                   // 基础费率
	BaseAmount                float64 `json:"base_amount,omitempty"`                 // 基础金额
	StockUomRate              float64 `json:"stock_uom_rate,omitempty"`              // 库存单位费率
	IsFreeItem                int     `json:"is_free_item,omitempty"`                // 是否免费商品
	GrantCommission           int     `json:"grant_commission,omitempty"`            // 授予佣金
	NetRate                   float64 `json:"net_rate,omitempty"`                    // 净费率
	NetAmount                 float64 `json:"net_amount,omitempty"`                  // 净金额
	BaseNetRate               float64 `json:"base_net_rate,omitempty"`               // 基础净费率
	BaseNetAmount             float64 `json:"base_net_amount,omitempty"`             // 基础净金额
	DeliveredBySupplier       int     `json:"delivered_by_supplier,omitempty"`       // 供应商交付
	IncomeAccount             string  `json:"income_account,omitempty"`              // 收入账户
	IsFixedAsset              int     `json:"is_fixed_asset,omitempty"`              // 是否固定资产
	ExpenseAccount            string  `json:"expense_account,omitempty"`             // 费用账户
	EnableDeferredRevenue     int     `json:"enable_deferred_revenue,omitempty"`     // 启用递延收入
	WeightPerUnit             float64 `json:"weight_per_unit,omitempty"`             // 单位重量
	TotalWeight               float64 `json:"total_weight,omitempty"`                // 总重量
	Warehouse                 string  `json:"warehouse,omitempty"`                   // 仓库
	UseSerialBatchFields      int     `json:"use_serial_batch_fields,omitempty"`     // 使用序列批号字段
	AllowZeroValuationRate    int     `json:"allow_zero_valuation_rate,omitempty"`   // 允许零估值费率
	IncomingRate              float64 `json:"incoming_rate,omitempty"`               // 入库费率
	ItemTaxRate               string  `json:"item_tax_rate,omitempty"`               // 商品税费率
	ActualBatchQty            int     `json:"actual_batch_qty,omitempty"`            // 实际批号数量
	ActualQty                 int     `json:"actual_qty,omitempty"`                  // 实际数量
	CompanyTotalStock         int     `json:"company_total_stock,omitempty"`         // 公司总库存
	DeliveredQty              int     `json:"delivered_qty,omitempty"`               // 已交付数量
	PosInvoice                string  `json:"pos_invoice,omitempty"`                 // POS发票
	PosInvoiceItem            string  `json:"pos_invoice_item,omitempty"`            // POS发票项目
	CostCenter                string  `json:"cost_center,omitempty"`                 // 成本中心
	PageBreak                 int     `json:"page_break,omitempty"`                  // 分页
	Parent                    string  `json:"parent,omitempty"`                      // 父级
	Parentfield               string  `json:"parentfield,omitempty"`                 // 父级字段
	Parenttype                string  `json:"parenttype,omitempty"`                  // 父级类型
	Doctype                   string  `json:"doctype,omitempty"`                     // 文档类型
}

// SalesInvoicePayment 销售发票付款结构
type SalesInvoicePayment struct {
	Name          string  `json:"name,omitempty"`            // 付款名称
	Owner         string  `json:"owner,omitempty"`           // 所有者
	Creation      string  `json:"creation,omitempty"`        // 创建时间
	Modified      string  `json:"modified,omitempty"`        // 修改时间
	ModifiedBy    string  `json:"modified_by,omitempty"`     // 修改者
	Docstatus     int     `json:"docstatus,omitempty"`       // 文档状态
	Idx           int     `json:"idx,omitempty"`             // 索引
	Default       int     `json:"default,omitempty"`         // 是否默认
	ModeOfPayment string  `json:"mode_of_payment,omitempty"` // 支付方式
	Amount        float64 `json:"amount,omitempty"`          // 金额
	Account       string  `json:"account,omitempty"`         // 账户
	Type          string  `json:"type,omitempty"`            // 类型
	BaseAmount    float64 `json:"base_amount,omitempty"`     // 基础金额
	Parent        string  `json:"parent,omitempty"`          // 父级
	Parentfield   string  `json:"parentfield,omitempty"`     // 父级字段
	Parenttype    string  `json:"parenttype,omitempty"`      // 父级类型
	Doctype       string  `json:"doctype,omitempty"`         // 文档类型
}

// POSInvoice 结构体定义
// 用于表示POS发票的完整信息
type POSInvoice struct {
	AmendedFrom                            string              `json:"amended_from,omitempty"`                               //  修订发票
	Name                                   string              `json:"name,omitempty"`                                       // 发票名称
	Owner                                  string              `json:"owner,omitempty"`                                      // 所有者
	Creation                               string              `json:"creation,omitempty"`                                   // 创建时间
	Modified                               string              `json:"modified,omitempty"`                                   // 修改时间
	ModifiedBy                             string              `json:"modified_by,omitempty"`                                // 修改者
	Docstatus                              int                 `json:"docstatus,omitempty"`                                  // 文档状态
	Idx                                    int                 `json:"idx,omitempty"`                                        // 索引
	Title                                  string              `json:"title,omitempty"`                                      // 标题
	NamingSeries                           string              `json:"naming_series,omitempty"`                              // 命名系列
	Customer                               string              `json:"customer,omitempty"`                                   // 客户
	CustomerName                           string              `json:"customer_name,omitempty"`                              // 客户名称
	PosProfile                             string              `json:"pos_profile,omitempty"`                                // POS配置文件
	IsPos                                  int                 `json:"is_pos,omitempty"`                                     // 是否为POS
	IsReturn                               int                 `json:"is_return,omitempty"`                                  // 是否为退货
	UpdateBilledAmountInSalesOrder         int                 `json:"update_billed_amount_in_sales_order,omitempty"`        // 是否更新销售订单中的已开票金额
	UpdateBilledAmountInDeliveryNote       int                 `json:"update_billed_amount_in_delivery_note,omitempty"`      // 是否更新送货单中的已开票金额
	Company                                string              `json:"company,omitempty"`                                    // 公司
	PostingDate                            string              `json:"posting_date,omitempty"`                               // 过账日期
	PostingTime                            string              `json:"posting_time,omitempty"`                               // 过账时间
	SetPostingTime                         int                 `json:"set_posting_time,omitempty"`                           // 是否设置过账时间
	DueDate                                string              `json:"due_date,omitempty"`                                   // 到期日期
	Currency                               string              `json:"currency,omitempty"`                                   // 货币
	ConversionRate                         float64             `json:"conversion_rate,omitempty"`                            // 汇率
	SellingPriceList                       string              `json:"selling_price_list,omitempty"`                         // 销售价格表
	PriceListCurrency                      string              `json:"price_list_currency,omitempty"`                        // 价格表货币
	PlcConversionRate                      float64             `json:"plc_conversion_rate,omitempty"`                        // 价格表汇率
	IgnorePricingRule                      int                 `json:"ignore_pricing_rule,omitempty"`                        // 是否忽略定价规则
	SetWarehouse                           string              `json:"set_warehouse,omitempty"`                              // 设置仓库
	UpdateStock                            int32               `json:"update_stock,omitempty"`                               // 是否更新库存
	TotalBillingAmount                     float64             `json:"total_billing_amount,omitempty"`                       // 总开票金额
	TotalQty                               int                 `json:"total_qty,omitempty"`                                  // 总数量
	BaseTotal                              float64             `json:"base_total,omitempty"`                                 // 基础总计
	BaseNetTotal                           float64             `json:"base_net_total,omitempty"`                             // 基础净总计
	Total                                  float64             `json:"total,omitempty"`                                      // 总计
	NetTotal                               float64             `json:"net_total,omitempty"`                                  // 净总计
	TotalNetWeight                         float64             `json:"total_net_weight,omitempty"`                           // 总净重
	TaxesAndCharges                        string              `json:"taxes_and_charges,omitempty"`                          // 税费和费用
	TaxCategory                            string              `json:"tax_category,omitempty"`                               // 税费类别
	BaseTotalTaxesAndCharges               float64             `json:"base_total_taxes_and_charges,omitempty"`               // 基础税费和费用总计
	TotalTaxesAndCharges                   float64             `json:"total_taxes_and_charges,omitempty"`                    // 税费和费用总计
	LoyaltyPoints                          int                 `json:"loyalty_points,omitempty"`                             // 忠诚度积分
	LoyaltyAmount                          float64             `json:"loyalty_amount,omitempty"`                             // 忠诚度金额
	RedeemLoyaltyPoints                    int                 `json:"redeem_loyalty_points,omitempty"`                      // 兑换忠诚度积分
	ApplyDiscountOn                        string              `json:"apply_discount_on,omitempty"`                          // 折扣应用位置
	BaseDiscountAmount                     float64             `json:"base_discount_amount,omitempty"`                       // 基础折扣金额
	AdditionalDiscountPercentage           float64             `json:"additional_discount_percentage,omitempty"`             // 额外折扣百分比
	DiscountAmount                         float64             `json:"discount_amount,omitempty"`                            // 折扣金额
	BaseGrandTotal                         float64             `json:"base_grand_total,omitempty"`                           // 基础总金额
	BaseRoundingAdjustment                 float64             `json:"base_rounding_adjustment,omitempty"`                   // 基础四舍五入调整
	BaseRoundedTotal                       float64             `json:"base_rounded_total,omitempty"`                         // 基础四舍五入总计
	BaseInWords                            string              `json:"base_in_words,omitempty"`                              // 基础金额大写
	GrandTotal                             float64             `json:"grand_total,omitempty"`                                // 总金额
	RoundingAdjustment                     float64             `json:"rounding_adjustment,omitempty"`                        // 四舍五入调整
	RoundedTotal                           float64             `json:"rounded_total,omitempty"`                              // 四舍五入总计
	InWords                                string              `json:"in_words,omitempty"`                                   // 金额大写
	TotalAdvance                           float64             `json:"total_advance,omitempty"`                              // 总预付款
	OutstandingAmount                      float64             `json:"outstanding_amount,omitempty"`                         // 未付金额
	AllocateAdvancesAutomatically          int                 `json:"allocate_advances_automatically,omitempty"`            // 是否自动分配预付款
	BasePaidAmount                         float64             `json:"base_paid_amount,omitempty"`                           // 基础已付金额
	PaidAmount                             float64             `json:"paid_amount,omitempty"`                                // 已付金额
	BaseChangeAmount                       float64             `json:"base_change_amount,omitempty"`                         // 基础找零金额
	ChangeAmount                           float64             `json:"change_amount,omitempty"`                              // 找零金额
	AccountForChangeAmount                 string              `json:"account_for_change_amount,omitempty"`                  // 找零账户
	WriteOffAmount                         float64             `json:"write_off_amount,omitempty"`                           // 核销金额
	BaseWriteOffAmount                     float64             `json:"base_write_off_amount,omitempty"`                      // 基础核销金额
	WriteOffOutstandingAmountAutomatically int                 `json:"write_off_outstanding_amount_automatically,omitempty"` // 是否自动核销未付金额
	WriteOffAccount                        string              `json:"write_off_account,omitempty"`                          // 核销账户
	WriteOffCostCenter                     string              `json:"write_off_cost_center,omitempty"`                      // 核销成本中心
	GroupSameItems                         int                 `json:"group_same_items,omitempty"`                           // 是否分组相同商品
	Language                               string              `json:"language,omitempty"`                                   // 语言
	IsDiscounted                           int                 `json:"is_discounted,omitempty"`                              // 是否有折扣
	Status                                 string              `json:"status,omitempty"`                                     // 状态
	DebitTo                                string              `json:"debit_to,omitempty"`                                   // 借记到
	PartyAccountCurrency                   string              `json:"party_account_currency,omitempty"`                     // 对方账户货币
	IsOpening                              string              `json:"is_opening,omitempty"`                                 // 是否为期初
	Remarks                                string              `json:"remarks,omitempty"`                                    // 备注
	AmountEligibleForCommission            float64             `json:"amount_eligible_for_commission,omitempty"`             // 符合佣金条件的金额
	CommissionRate                         float64             `json:"commission_rate,omitempty"`                            // 佣金率
	TotalCommission                        float64             `json:"total_commission,omitempty"`                           // 总佣金
	Doctype                                string              `json:"doctype,omitempty"`                                    // 文档类型
	PricingRules                           []interface{}       `json:"pricing_rules,omitempty"`                              // 定价规则
	Advances                               []interface{}       `json:"advances,omitempty"`                                   // 预付款
	SalesTeam                              []interface{}       `json:"sales_team,omitempty"`                                 // 销售团队
	PaymentSchedule                        []interface{}       `json:"payment_schedule,omitempty"`                           // 付款计划
	Items                                  []POSInvoiceItem    `json:"items,omitempty"`                                      // 商品项目
	Taxes                                  []POSInvoiceTax     `json:"taxes,omitempty"`                                      // 税费
	Timesheets                             []interface{}       `json:"timesheets,omitempty"`                                 // 工时表
	Payments                               []POSInvoicePayment `json:"payments,omitempty"`                                   // 付款
	PackedItems                            []interface{}       `json:"packed_items,omitempty"`                               // 包装商品
	ReturnAgainst                          string              `json:"return_against,omitempty"`                             // 退款销售订单
	CustomerOrder                          string              `json:"po_no,omitempty"`                                      // 客户订单号, ttpos 订单

	CustomerUUID          string `json:"custom_customer_uuid,omitempty"`     // 客户UUID 自定义字段
	CustomPosOpeningEntry string `json:"custom_pos_opening_entry,omitempty"` // 自定义POS开帐分录
	CustomTakeoutOrderNo  string `json:"custom_takeout_order_no,omitempty"`  // 外卖订单号
	CustomTakeoutProvider string `json:"custom_takeout_provider,omitempty"`  // 外卖平台提供商
}

// POSInvoiceItem 结构体定义
// 用于表示POS发票中的商品项目信息
type POSInvoiceItem struct {
	Name                      string  `json:"name,omitempty"`                        // 名称
	Owner                     string  `json:"owner,omitempty"`                       // 所有者
	Creation                  string  `json:"creation,omitempty"`                    // 创建时间
	Modified                  string  `json:"modified,omitempty"`                    // 修改时间
	ModifiedBy                string  `json:"modified_by,omitempty"`                 // 修改者
	Docstatus                 int     `json:"docstatus,omitempty"`                   // 文档状态
	Idx                       int     `json:"idx,omitempty"`                         // 索引
	HasItemScanned            int     `json:"has_item_scanned,omitempty"`            // 是否已扫描商品
	ItemCode                  string  `json:"item_code,omitempty"`                   // 商品编码
	ItemName                  string  `json:"item_name,omitempty"`                   // 商品名称
	Description               string  `json:"description,omitempty"`                 // 描述
	ItemGroup                 string  `json:"item_group,omitempty"`                  // 商品组
	Image                     string  `json:"image,omitempty"`                       // 图片
	Qty                       float64 `json:"qty,omitempty"`                         // 数量
	StockUom                  string  `json:"stock_uom,omitempty"`                   // 库存单位
	Uom                       string  `json:"uom,omitempty"`                         // 单位
	ConversionFactor          float64 `json:"conversion_factor,omitempty"`           // 转换因子
	StockQty                  int     `json:"stock_qty,omitempty"`                   // 库存数量
	PriceListRate             float64 `json:"price_list_rate,omitempty"`             // 价格表费率
	BasePriceListRate         float64 `json:"base_price_list_rate,omitempty"`        // 基础价格表费率
	MarginType                string  `json:"margin_type,omitempty"`                 // 利润率类型
	MarginRateOrAmount        float64 `json:"margin_rate_or_amount,omitempty"`       // 利润率或金额
	RateWithMargin            float64 `json:"rate_with_margin,omitempty"`            // 含利润率费率
	DiscountPercentage        float64 `json:"discount_percentage,omitempty"`         // 折扣百分比
	DiscountAmount            float64 `json:"discount_amount,omitempty"`             // 折扣金额
	DistributedDiscountAmount float64 `json:"distributed_discount_amount,omitempty"` // 分配折扣金额
	BaseRateWithMargin        float64 `json:"base_rate_with_margin,omitempty"`       // 基础含利润率费率
	Rate                      float64 `json:"rate"`                                  // 费率, 可以为0
	Amount                    float64 `json:"amount"`                                // 金额，可以为0
	BaseRate                  float64 `json:"base_rate"`                             // 基础费率，可以为0
	BaseAmount                float64 `json:"base_amount"`                           // 基础金额，可以为0
	PricingRules              string  `json:"pricing_rules,omitempty"`               // 定价规则
	IsFreeItem                bool    `json:"is_free_item,omitempty"`                // 是否为免费商品
	GrantCommission           int     `json:"grant_commission,omitempty"`            // 是否授予佣金
	NetRate                   float64 `json:"net_rate,omitempty"`                    // 净费率
	NetAmount                 float64 `json:"net_amount,omitempty"`                  // 净金额
	BaseNetRate               float64 `json:"base_net_rate,omitempty"`               // 基础净费率
	BaseNetAmount             float64 `json:"base_net_amount,omitempty"`             // 基础净金额
	DeliveredBySupplier       int     `json:"delivered_by_supplier,omitempty"`       // 是否由供应商交付
	IncomeAccount             string  `json:"income_account,omitempty"`              // 收入账户
	IsFixedAsset              int     `json:"is_fixed_asset,omitempty"`              // 是否为固定资产
	ExpenseAccount            string  `json:"expense_account,omitempty"`             // 费用账户
	EnableDeferredRevenue     int     `json:"enable_deferred_revenue,omitempty"`     // 是否启用递延收入
	WeightPerUnit             float64 `json:"weight_per_unit,omitempty"`             // 每单位重量
	TotalWeight               float64 `json:"total_weight,omitempty"`                // 总重量
	Warehouse                 string  `json:"warehouse,omitempty"`                   // 仓库
	UseSerialBatchFields      int     `json:"use_serial_batch_fields,omitempty"`     // 是否使用序列批号字段
	AllowZeroValuationRate    int     `json:"allow_zero_valuation_rate,omitempty"`   // 是否允许零估值费率
	ItemTaxRate               string  `json:"item_tax_rate,omitempty"`               // 商品税费率
	ActualBatchQty            int     `json:"actual_batch_qty,omitempty"`            // 实际批号数量
	ActualQty                 int     `json:"actual_qty,omitempty"`                  // 实际数量
	DeliveredQty              int     `json:"delivered_qty,omitempty"`               // 已交付数量
	CostCenter                string  `json:"cost_center,omitempty"`                 // 成本中心
	PageBreak                 int     `json:"page_break,omitempty"`                  // 分页符
	Parent                    string  `json:"parent,omitempty"`                      // 父级
	Parentfield               string  `json:"parentfield,omitempty"`                 // 父级字段
	Parenttype                string  `json:"parenttype,omitempty"`                  // 父级类型
	Doctype                   string  `json:"doctype,omitempty"`                     // 文档类型
}

// POSInvoicePayment 结构体定义
// 用于表示POS发票中的付款信息
type POSInvoicePayment struct {
	Name          string  `json:"name,omitempty"`            // 名称
	Owner         string  `json:"owner,omitempty"`           // 所有者
	Creation      string  `json:"creation,omitempty"`        // 创建时间
	Modified      string  `json:"modified,omitempty"`        // 修改时间
	ModifiedBy    string  `json:"modified_by,omitempty"`     // 修改者
	Docstatus     int     `json:"docstatus,omitempty"`       // 文档状态
	Idx           int     `json:"idx,omitempty"`             // 索引
	Default       int     `json:"default,omitempty"`         // 是否默认
	ModeOfPayment string  `json:"mode_of_payment,omitempty"` // 支付方式
	Amount        float64 `json:"amount"`                    // 金额，可以为0
	Account       string  `json:"account,omitempty"`         // 账户
	Type          string  `json:"type,omitempty"`            // 类型
	BaseAmount    float64 `json:"base_amount,omitempty"`     // 基础金额
	Parent        string  `json:"parent,omitempty"`          // 父级
	Parentfield   string  `json:"parentfield,omitempty"`     // 父级字段
	Parenttype    string  `json:"parenttype,omitempty"`      // 父级类型
	Doctype       string  `json:"doctype,omitempty"`         // 文档类型
}

// POSInvoiceTax 结构体定义
// 用于表示POS发票中的税金信息，及其他费用
type POSInvoiceTax struct {
	ChargeType  string  `json:"charge_type,omitempty"`  // 计费类型
	AccountHead string  `json:"account_head,omitempty"` // 会计科目
	Rate        float64 `json:"rate"`                   // 税率, 可以为0
	TaxAmount   float64 `json:"tax_amount"`             // 税费金额, 可以为0
	Description string  `json:"description,omitempty"`  // 描述
}
