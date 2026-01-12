package constant

// 限购配置状态
const (
	PurchaseQuotaConfigStatusDisabled = 0 // 禁用
	PurchaseQuotaConfigStatusEnabled  = 1 // 启用
)

// 限购周期类型
const (
	PurchaseQuotaPeriodTypeDaily     = 0 // 按天（默认）
	PurchaseQuotaPeriodTypeMonthly   = 1 // 月度
	PurchaseQuotaPeriodTypeQuarterly = 2 // 季度（预留）
	PurchaseQuotaPeriodTypeYearly    = 3 // 年度（预留）
)

// 限购超限策略
const (
	PurchaseQuotaStrictModeStrict = 1 // 严格拒绝
	PurchaseQuotaStrictModeSoft   = 2 // 柔性提醒（预留）
)

// 限购配置来源
const (
	PurchaseQuotaConfigSourceShop        = 1 // 门店自配
	PurchaseQuotaConfigSourceHeadquarter = 2 // 总部下发
)
