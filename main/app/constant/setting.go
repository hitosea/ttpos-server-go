package constant

const (
	SettingPrinter        = "printer"          // 小票打印机设置
	SettingStore          = "store"            // 商城设置
	SettingRecharge       = "recharge"         // 充值设置
	SettingPoints         = "points"           // 积分设置
	SettingSysAdminConfig = "sys_admin_config" // 系统配置
	SettingSysConfig      = "sys_config"       // 系统配置
	SettingBalance        = "balance"          // 充值设置
	SettingCurrency       = "currency"         // 门店-货币单位
	SettingTaxRate        = "tax_rate"         // 门店-税率管理
	SettingServiceCharge  = "service_charge"   // 门店-服务费
	SettingPayment        = "payment"          // 门店-支付方式
	SettingBusiness       = "business"         // 门店-业务设置
	SettingCashier        = "cashier"          // 各端-收银机设置
	SettingTablet         = "tablet"           // 各端-平板端设置
	SettingH5             = "h5"               // 各端-扫码H5设置
	SettingKitchen        = "kitchen"          // 各端-厨显设置
	SettingAssistant      = "assistant"        // 各端-点餐助手设置
	SettingBuffet         = "buffet"           // 自助餐-自助餐设置
	SettingPointsRule     = "points_rule"      // 积分规则设置
	SettingCloudBasic     = "cloud_basic"      // 云端-基础信息
	SettingBatchColor     = "batch_color"      // 分批类型颜色设置
)

const (
	PasswordTypeCashBox  = "cash_box" // 收银机钱箱密码
	PasswordTypeAdvanced = "advanced" // 收银机高级密码
	PasswordTypeLock     = "lock"     // 收银机锁屏密码
)

const (
	AppTypeCashier   = 1 // 收银端
	AppTypeTablet    = 2 // 平板端
	AppTypeKitchen   = 3 // 厨显端
	AppTypeShop      = 4 // 商家后台端
	AppTypeAssistant = 5 // 点餐助手端
)

// 结账是否自动清台
const (
	AutoClearTable    = "0" // 自动清台
	NotAutoClearTable = "1" // 不自动清台
)

// 积分赠送类型
const (
	// 积分赠送类型
	RuleTypePaymentAmount = 0 // 按付款金额比例赠送
	RuleTypeDesk          = 1 // 按桌台人数赠送
)

const (
	KdsModeDefault     uint = 0 // 默认，传菜模式
	KdsModeMake        uint = 1 // 制作模式
	KdsModeMakeAndSend uint = 2 // 制作+传菜模式
)
