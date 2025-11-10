package erp

// Company 公司信息
type Company struct {
	Name                                        string         `json:"name,omitempty"`                                              // 公司名称
	Owner                                       string         `json:"owner,omitempty"`                                             // 所有者
	Creation                                    string         `json:"creation,omitempty"`                                          // 创建时间
	Modified                                    string         `json:"modified,omitempty"`                                          // 修改时间
	ModifiedBy                                  string         `json:"modified_by,omitempty"`                                       // 修改人
	Docstatus                                   int            `json:"docstatus,omitempty"`                                         // 文档状态
	Idx                                         int            `json:"idx,omitempty"`                                               // 索引
	CompanyName                                 string         `json:"company_name,omitempty"`                                      // 公司名称
	Abbr                                        string         `json:"abbr,omitempty"`                                              // 公司简称
	DefaultCurrency                             string         `json:"default_currency,omitempty"`                                  // 默认货币
	Country                                     string         `json:"country,omitempty"`                                           // 国家
	IsGroup                                     int            `json:"is_group,omitempty"`                                          // 是否为分组
	Lft                                         int            `json:"lft,omitempty"`                                               // 左值（嵌套集模型）
	Rgt                                         int            `json:"rgt,omitempty"`                                               // 右值（嵌套集模型）
	OldParent                                   string         `json:"old_parent,omitempty"`                                        // 旧父级
	ParentCompany                               string         `json:"parent_company,omitempty"`                                    // 父公司
	CreateChartOfAccountsBasedOn                string         `json:"create_chart_of_accounts_based_on,omitempty"`                 // 创建会计科目表基于
	ChartOfAccounts                             string         `json:"chart_of_accounts,omitempty"`                                 // 会计科目表
	DefaultCashAccount                          string         `json:"default_cash_account,omitempty"`                              // 默认现金账户
	DefaultReceivableAccount                    string         `json:"default_receivable_account,omitempty"`                        // 默认应收账户
	DefaultPayableAccount                       string         `json:"default_payable_account,omitempty"`                           // 默认应付账户
	WriteOffAccount                             string         `json:"write_off_account,omitempty"`                                 // 核销账户
	AllowAccountCreationAgainstChildCompany     int            `json:"allow_account_creation_against_child_company,omitempty"`      // 允许为子公司创建账户
	DefaultExpenseAccount                       string         `json:"default_expense_account,omitempty"`                           // 默认费用账户
	DefaultIncomeAccount                        string         `json:"default_income_account,omitempty"`                            // 默认收入账户
	CostCenter                                  string         `json:"cost_center,omitempty"`                                       // 成本中心
	ExchangeGainLossAccount                     string         `json:"exchange_gain_loss_account,omitempty"`                        // 汇兑损益账户
	RoundOffAccount                             string         `json:"round_off_account,omitempty"`                                 // 舍入账户
	RoundOffCostCenter                          string         `json:"round_off_cost_center,omitempty"`                             // 舍入成本中心
	BookAdvancePaymentsInSeparatePartyAccount   int            `json:"book_advance_payments_in_separate_party_account,omitempty"`   // 在单独的客户账户中记录预付款
	ReconcileOnAdvancePaymentDate               int            `json:"reconcile_on_advance_payment_date,omitempty"`                 // 在预付款日期核对
	ReconciliationTakesEffectOn                 string         `json:"reconciliation_takes_effect_on,omitempty"`                    // 核对生效于
	AutoExchangeRateRevaluation                 int            `json:"auto_exchange_rate_revaluation,omitempty"`                    // 自动汇率重估
	AutoErrFrequency                            string         `json:"auto_err_frequency,omitempty"`                                // 自动错误频率
	SubmitErrJv                                 int            `json:"submit_err_jv,omitempty"`                                     // 提交错误凭证
	AccumulatedDepreciationAccount              string         `json:"accumulated_depreciation_account,omitempty"`                  // 累计折旧账户
	DepreciationExpenseAccount                  string         `json:"depreciation_expense_account,omitempty"`                      // 折旧费用账户
	ExpensesIncludedInAssetValuation            string         `json:"expenses_included_in_asset_valuation,omitempty"`              // 资产估值中包含的费用
	DisposalAccount                             string         `json:"disposal_account,omitempty"`                                  // 处置账户
	DepreciationCostCenter                      string         `json:"depreciation_cost_center,omitempty"`                          // 折旧成本中心
	CapitalWorkInProgressAccount                string         `json:"capital_work_in_progress_account,omitempty"`                  // 在建工程账户
	AssetReceivedButNotBilled                   string         `json:"asset_received_but_not_billed,omitempty"`                     // 已收资产但未开票
	SalesMonthlyHistory                         string         `json:"sales_monthly_history,omitempty"`                             // 月度销售历史
	MonthlySalesTarget                          float64        `json:"monthly_sales_target,omitempty"`                              // 月度销售目标
	TotalMonthlySales                           float64        `json:"total_monthly_sales,omitempty"`                               // 月度总销售额
	CreditLimit                                 float64        `json:"credit_limit,omitempty"`                                      // 信用额度
	DefaultEmployeeAdvanceAccount               string         `json:"default_employee_advance_account,omitempty"`                  // 默认员工预支账户
	DefaultPayrollPayableAccount                string         `json:"default_payroll_payable_account,omitempty"`                   // 默认工资应付账户
	TransactionsAnnualHistory                   string         `json:"transactions_annual_history,omitempty"`                       // 年度交易历史
	EnablePerpetualInventory                    int            `json:"enable_perpetual_inventory,omitempty"`                        // 启用永续盘存
	EnableProvisionalAccountingForNonStockItems int            `json:"enable_provisional_accounting_for_non_stock_items,omitempty"` // 启用非库存物品的临时会计
	DefaultInventoryAccount                     string         `json:"default_inventory_account,omitempty"`                         // 默认库存账户
	StockAdjustmentAccount                      string         `json:"stock_adjustment_account,omitempty"`                          // 库存调整账户
	StockReceivedButNotBilled                   string         `json:"stock_received_but_not_billed,omitempty"`                     // 已收货但未开票
	ExpensesIncludedInValuation                 string         `json:"expenses_included_in_valuation,omitempty"`                    // 估值中包含的费用
	Doctype                                     string         `json:"doctype,omitempty"`                                           // 文档类型
	Onload                                      *CompanyOnload `json:"__onload,omitempty"`                                          // 加载时数据
}

// CompanyOnload 公司加载时数据
type CompanyOnload struct {
	AddrList    []interface{} `json:"addr_list,omitempty"`    // 地址列表
	ContactList []interface{} `json:"contact_list,omitempty"` // 联系人列表
}
