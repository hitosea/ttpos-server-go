package setting

import (
	"ttpos-server-go/app/dto"
)

// Printer 小票打印机设置
type Printer struct {
	CashierOpen        string               `json:"cashier_open"`         // 是否开启打印
	CashierPrinterID   string               `json:"cashier_printer_id"`   // 打印机id
	CashierPrinter     []CashierPrinterItem `json:"cashier_printer"`      // 打印机列表
	LanguageList       []dto.LanguageItem   `json:"language_list"`        // 语言列表
	LanguageMethod     string               `json:"language_method"`      // 语言方式（收银） 1-单语言 2-多语言
	DefaultLanguage    string               `json:"default_language"`     // 打印语言（收银）
	PrintMethod        string               `json:"print_method"`         // 打印方式（收银） 1-文本打印 2-图片打印
	KitchenLanguage    string               `json:"kitchen_language"`     // 打印语言（送厨）
	KitchenPrintMethod string               `json:"kitchen_print_method"` // 打印方式（送厨） 1-文本打印 2-图片打印
	ConsumptionTax     string               `json:"consumption_tax"`      // 消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示
	BuffetSignOpen     string               `json:"buffet_sign_open"`     // 自助餐标识设置（默认开启）
	MonetaryUnitOpen   string               `json:"monetary_unit_open"`   // 货币单位（默认开启）
	CalendarList       []CalendarItem       `json:"calendar_list"`        // 日历列表 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
	PrintList          []PrintItem          `json:"print_list"`           // 打印方式列表 （1-文本打印 2-图片打印 ）
	DefaultCalendar    string               `json:"default_calendar"`     // 日历类型 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
	Language           []string             `json:"language"`             // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	EnableCustomCopies string               `json:"enable_custom_copies"` // 是否启用自定义打印联数 "0"-关闭 "1"-开启
	CheckoutSlipCopies *int                 `json:"checkout_slip_copies"` // 结账单打印联数 0-10 nil表示未设置
}

type CalendarItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type PrintItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type CashierPrinterItem struct {
	Key          string `json:"key"`            // 收银机设备ID
	PrinterId    any    `json:"printer_id"`     // 收银机设备ID（32位字符串），或者printer表的Uuid uint64 20个字符
	PrinterUsbId string `json:"printer_usb_id"` // 收银机设备SN
	Sn           string `json:"sn"`             // 打印机SN
}

type PrinterInfo struct {
	PrinterUuid            uint64 // 0 或者 model.Printer 的Uuid
	PrinterType            string // 打印机类型
	Copies                 uint   // 打印份数
	PrinterConfig          string // 打印机设置
	IsCashierPrinter       bool   // 是否收银机自带打印机
	IsCashierOpen          bool   // 是否开启打印
	PrinterCashierDeviceSn string // 执行打印的收银机设备ID
	IsUsbPrinter           bool   // 是否usb打印机
	PrintMethod            int    // 打印方式 1文本打印, 2图片打印
	PrinterSn              string // 打印机SN
	PrinterWidth           int    // 打印机宽度mm
	EnableStatusCheck      int    // 是否启用状态检查
	EnableSound            int    // 是否启用打印提示音
	PrintSpeed             int    // 打印速度 1-流畅(不分片打印) 2-稳定(分片大包打印) 3-兼容(分片小包打印)
}

// 是否启用
func (p *PrinterInfo) IsEnableSound() bool {
	return p.EnableSound == 1
}

// 是否启用状态检查
func (p *PrinterInfo) IsEnableStatusCheck() bool {
	return p.EnableStatusCheck == 1
}

// PrintSettingResp 打印设置响应（仅包含自定义打印联数相关字段）
type PrintSettingResp struct {
	EnableCustomCopies string `json:"enable_custom_copies"` // 是否启用自定义打印联数 "0"-关闭 "1"-开启
	CheckoutSlipCopies *int   `json:"checkout_slip_copies"` // 结账单打印联数 0-10 nil表示未设置
}
