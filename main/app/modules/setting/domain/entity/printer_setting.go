package entity

import (
	"ttpos-server-go/app/modules/setting/domain/valueobject"
)

// PrinterSetting 打印机设置实体（与旧服务 Printer 结构保持一致）
type PrinterSetting struct {
	CashierOpen        string                     `json:"cashier_open"`         // 是否开启打印
	CashierPrinterID   string                     `json:"cashier_printer_id"`   // 打印机id
	CashierPrinter     []CashierPrinterItem       `json:"cashier_printer"`      // 打印机列表
	LanguageList       []valueobject.LanguageItem `json:"language_list"`        // 语言列表
	LanguageMethod     string                     `json:"language_method"`      // 语言方式（收银） 1-单语言 2-多语言
	DefaultLanguage    string                     `json:"default_language"`     // 打印语言（收银）
	PrintMethod        string                     `json:"print_method"`         // 打印方式（收银） 1-文本打印 2-图片打印
	KitchenLanguage    string                     `json:"kitchen_language"`     // 打印语言（送厨）
	KitchenPrintMethod string                     `json:"kitchen_print_method"` // 打印方式（送厨） 1-文本打印 2-图片打印
	ConsumptionTax     string                     `json:"consumption_tax"`      // 消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示
	BuffetSignOpen     string                     `json:"buffet_sign_open"`     // 自助餐标识设置（默认开启）
	MonetaryUnitOpen   string                     `json:"monetary_unit_open"`   // 货币单位（默认开启）
	CalendarList       []CalendarItem             `json:"calendar_list"`        // 日历列表 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
	PrintList          []PrintItem                `json:"print_list"`           // 打印方式列表 （1-文本打印 2-图片打印 ）
	DefaultCalendar    string                     `json:"default_calendar"`     // 日历类型 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
	Language           []string                   `json:"language"`             // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
}

// CashierPrinterItem 收银机打印机项
type CashierPrinterItem struct {
	Key          string `json:"key"`
	PrinterId    string `json:"printer_id"`
	PrinterUsbId string `json:"printer_usb_id"`
	Sn           string `json:"sn"`
}

// KitchenPrinterItem 厨房打印机项
type KitchenPrinterItem struct {
	Key       string `json:"key"`        // 设备SN
	PrinterId string `json:"printer_id"` // 打印机ID
	Sn        string `json:"sn"`         // 打印机SN
}

// CalendarItem 日历项
type CalendarItem struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// PrintItem 打印项
type PrintItem struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// PrinterInfo 打印机信息
type PrinterInfo struct {
	PrinterType            string // 打印机类型
	PrinterUuid            uint64 // 打印机UUID
	Copies                 uint   // 份数
	PrinterConfig          string // 打印机配置
	IsCashierPrinter       bool   // 是否收银机自带打印机
	IsCashierOpen          bool   // 收银机是否开启
	PrinterCashierDeviceSn string // 收银机设备SN
	IsUsbPrinter           bool   // 是否USB打印机
	PrintMethod            int    // 打印方式
	PrinterSn              string // 打印机SN
	PrinterWidth           int    // 打印机宽度
	EnableStatusCheck      int    // 是否启用状态检查
	EnableSound            int    // 是否启用打印提示音
	PrintSpeed             int    // 打印速度
}

// NewPrinterSetting 创建新的打印机设置
func NewPrinterSetting() *PrinterSetting {
	return &PrinterSetting{
		CashierPrinter: make([]CashierPrinterItem, 0),
		LanguageList:   make([]valueobject.LanguageItem, 0),
		CalendarList:   make([]CalendarItem, 0),
		PrintList:      make([]PrintItem, 0),
		Language:       make([]string, 0),
		CashierOpen:    "0", // 默认关闭
	}
}

// DefaultPrinterSetting 获取默认打印机设置（与旧服务保持一致）
func DefaultPrinterSetting(languageList []valueobject.LanguageItem) *PrinterSetting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	return &PrinterSetting{
		CashierOpen:        "1",                           // 是否开启打印
		CashierPrinterID:   "-1",                          // 打印机id
		CashierPrinter:     make([]CashierPrinterItem, 0), // 打印机列表
		LanguageList:       languageList,                  // 语言列表
		LanguageMethod:     "1",                           // 语言方式（收银） 1-单语言 2-多语言
		DefaultLanguage:    defaultLanguage,               // 打印语言（收银）
		PrintMethod:        "1",                           // 打印方式（收银） 1-文本打印 2-图片打印
		KitchenLanguage:    defaultLanguage,               // 打印语言（送厨）
		KitchenPrintMethod: "1",                           // 打印方式（送厨） 1-文本打印 2-图片打印
		ConsumptionTax:     "1",                           // 消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示
		BuffetSignOpen:     "1",                           // 自助餐标识设置（默认开启）
		MonetaryUnitOpen:   "1",                           // 货币单位（默认开启）
		CalendarList: []CalendarItem{{
			Key:  "1",
			Name: "公历",
		}, {
			Key:  "3",
			Name: "佛历",
		}}, // 日历列表 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
		PrintList: []PrintItem{{
			Key:  "1",
			Name: "文本打印",
		}, {
			Key:  "2",
			Name: "图片打印",
		}}, // 打印方式列表 （1-文本打印 2-图片打印 ）
		DefaultCalendar: "1",                       // 日历类型 （1-公历 2-农历 3-佛历 4-伊斯兰历 5-犹太历 ）
		Language:        []string{defaultLanguage}, // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	}
}
