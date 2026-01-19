package constant

// 类型：0系统默认队列，1云上服务下放
const (
	PrinterLogTypeDefault = 0 // 系统默认队列
	PrinterLogTypeCloud   = 1 //  云上服务下放
)

// 打印方式 1文本打印, 2图片打印
const (
	PrinterLogPrintMethodText  = 1 // 文本打印
	PrinterLogPrintMethodImage = 2 // 图片打印
)

// 状态(0结束,1进行中,2成功)
const (
	PrinterLogStatusEnd        = 0 // 结束
	PrinterLogStatusInProgress = 1 // 进行中
	PrinterLogStatusSuccess    = 2 // 成功
)

const (
	PrinterLogRelatedSaleOrder     = 0
	PrinterLogRelatedRechargeOrder = 1
)

// 打印类型
const (
	PrinterTypeFeiEYun      = "FEI_E_YUN"     // 飞鹅打印机
	PrinterTypeFeiEYunTag   = "FEI_E_YUN_TAG" // 飞鹅标签打印机
	PrinterTypePrintCenter  = "PRINT_CENTER"  // 365云打印
	PrinterTypeSunmiLan     = "SUNMI_LAN"     // 商米 局域网内打印
	PrinterTypeSunmiCloud   = "SUNMI_CLOUD"   // 商米 云打印
	PrinterTypeXPrinterLan  = "XPRINTER_LAN"  // 芯烨-有线
	PrinterTypeXPrinterWifi = "XPRINTER_WIFI" // 芯烨-WIFI
	PrinterTypeCodesoftLan  = "CODESOFT_LAN"  // Codesoft（网口）80mm
	PrinterTypeCodesoftWifi = "CODESOFT_WIFI" // Codesoft（WIFI）80mm
	PrinterTypeGpCloud      = "GP_CLOUD"      // 佳博 云打印
	// 收银打印机
	PrinterTypeCashierCompax = "CASHIER_COMPAX" // Compax 收银打印机 80mm 自带
	PrinterTypeCashierSunmi  = "CASHIER_SUNMI"  // SUNMI 商米 收银打印机 80mm 自带
	// IMIN打印机
	PrinterTypeCashierImmin = "CASHIER_IMMIN" // IMIN 收银打印机 80mm 自带
)

// 打印类型
const (
	PrinterProductTypeOutMenu  = -2 // 出菜单打印
	PrinterProductTypeBackFood = -1 // 退菜打印
	PrinterProductTypePay      = 0  // 付款打印
	PrinterProductTypeKitchen  = 1  // 送厨打印
)

// 打印模板
const (
	PrinterTemplateHandoverSheet   = 1  // 交班单
	PrinterTemplateBilling         = 2  // 结账单
	PrinterTemplatePreBilling      = 3  // 预结账单
	PrinterTemplateOneDishOneMenu  = 4  // 一菜一单
	PrinterTemplateBusiness        = 5  // 营业数据
	PrinterTemplateEntireOrder     = 6  // 整单打印
	PrinterTemplateInvoice         = 7  // 打印发票
	PrinterTemplateRecharge        = 8  // 充值单
	PrinterTemplateReturnDish      = 9  // 退菜单
	PrinterTemplateOutMenu         = 10 // 出菜单
	PrinterTemplateTakeoutOrder    = 11 // 外送单
	PrinterTemplateTakeoutMerchant = 12 // 外卖商家联
	PrinterTemplateTakeoutCustomer = 13 // 外卖顾客联
	PrinterTemplateTakeoutRefund   = 14 // 外卖退单联
)

// 外卖小票类型
const (
	TakeoutReceiptTypeMerchant = "merchant" // 商家联
	TakeoutReceiptTypeCustomer = "customer" // 顾客联
	TakeoutReceiptTypeRefund   = "refund"   // 退单联
)
