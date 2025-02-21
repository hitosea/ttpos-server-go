package constant

// 类型：0系统默认队列，1云上服务下放
const (
	PrinterLogTypeDefault = 0
	PrinterLogTypeCloud   = 1
)

// 数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单 8-充值单 9-退菜单
const (
	PrinterLogDataTypePreBilling     = 1 // 预结账单
	PrinterLogDataTypeBilling        = 2 // 结账单
	PrinterLogDataTypeOneDishOneMenu = 3 // 一菜一单
	PrinterLogDataTypeEntireOrder    = 4 // 整单打印
	PrinterLogDataTypeInvoice        = 5 // 打印发票
	PrinterLogDataTypeBusiness       = 6 // 打印营业数据
	PrinterLogDataTypeShiftHandover  = 7 // 打印交班单
	PrinterLogDataTypeRecharge       = 8 // 充值单
	PrinterLogDataTypeReturnDish     = 9 // 退菜单
)

// 打印方式 1文本打印, 2图片打印
const (
	PrinterLogPintMethodText  = 1 // 文本打印
	PrinterLogPintMethodImage = 2 // 图片打印
)

// 状态(0结束,1进行中,2成功)
const (
	PrinterLogStatusEnd        = 0 // 结束
	PrinterLogStatusInProgress = 1 // 进行中
	PrinterLogStatusSuccess    = 2 // 成功
)
