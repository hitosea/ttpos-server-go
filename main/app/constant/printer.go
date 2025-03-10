package constant

// 类型：0系统默认队列，1云上服务下放
const (
	PrinterLogTypeDefault = 0 // 系统默认队列
	PrinterLogTypeCloud   = 1 //  云上服务下放
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
	PrinterTypeFeiEYun       = "FEI_E_YUN"      // 飞鹅打印机
	PrinterTypeFeiEYunTag    = "FEI_E_YUN_TAG"  // 飞鹅标签打印机
	PrinterTypePrintCenter   = "PRINT_CENTER"   // 365云打印
	PrinterTypeSunmiLan      = "SUNMI_LAN"      // 商米 局域网内打印
	PrinterTypeSunmiCloud    = "SUNMI_CLOUD"    // 商米 云打印
	PrinterTypeXPrinterLan   = "XPRINTER_LAN"   // 芯烨-有线
	PrinterTypeXPrinterWifi  = "XPRINTER_WIFI"  // 芯烨-WIFI
	PrinterTypeCashierCompax = "CASHIER_COMPAX" // Compax 收银打印机 80mm 自带
	PrinterTypeCashierSunmi  = "CASHIER_SUNMI"  // SUNMI 商米 收银打印机 80mm 自带
	PrinterTypeCodesoftLan   = "CODESOFT_LAN"   // Codesoft（网口）80mm
	PrinterTypeCodesoftWifi  = "CODESOFT_WIFI"  //Codesoft（WIFI）80mm
)

// 打印类型
const (
	PrinterProductTypeBackFood = -1 // 退菜打印
	PrinterProductTypePay      = 0  // 付款打印
	PrinterProductTypeKitchen  = 1  // 送厨打印
)
