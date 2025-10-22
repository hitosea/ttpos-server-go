package resp

import (
	"ttpos-server-go/app/dto"
)

type PrinterData struct {
	Uuid              uint64 `json:"uuid"`                // 打印日志Uuid
	Data              string `json:"data"`                // 打印数据
	PrintMethod       int    `json:"print_method"`        // 打印方式 1文本打印, 2图片打印'
	Copies            uint   `json:"copies"`              // 打印机.份数
	PrinterType       string `json:"printer_type"`        // 打印机.类型 打印机类型 (SUNMI_LAN:商米打印机, SUNMI_CLOUD:商米打印机-云, XPRINTER_LAN:芯烨-有线 , XPRINTER_WIFI:芯烨-WIFI , CASHIER:收银机自带打印机)
	PrinterConfig     string `json:"printer_config"`      // 打印机.配置
	IsCashierPrinter  bool   `json:"is_cashier_printer"`  // 是否是收银机自带打印机
	IsUsbPrinter      bool   `json:"is_usb_printer"`      // 是否是usb打印机
	PrintingTime      int64  `json:"printing_time"`       // 打印耗时
	EnableStatusCheck int    `json:"enable_status_check"` // 是否启用状态检查
}

type PrinterDataList struct {
	List []PrinterData `json:"list"` // 打印数据列表
}

type PrinterLogData struct {
	Uuid         uint64 `json:"uuid"`           // 打印日志Uuid
	SerialNo     string `json:"serial_no"`      // 桌台号或者序号(如果有)
	OrderNo      string `json:"order_no"`       // 订单号
	PrinterName  string `json:"printer_name"`   // 打印机名称
	RuleName     string `json:"rule_name"`      // 商品打印规则名称
	DataTypeName string `json:"data_type_name"` // 数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单;
	CreateTime   int64  `json:"create_time"`    // 日志创建时间戳
	Status       int    `json:"status"`         // 状态(0结束,1进行中,2成功)
	StatusText   string `json:"status_text"`    // 状态文本
	PrinterTime  int64  `json:"printer_time"`   // 最后打印时间
	Reason       string `json:"reason"`         // 原因
}

type PrinterListPaginationResp struct {
	List []PrinterLogData `json:"list"` // 打印数据列表
	Meta dto.PageResponse `json:"meta"` // Meta信息
}

type PrinterInfo struct {
	PrinterType   string // 打印机类型
	PrinterConfig string // 打印机设置
	PrintCopies   uint   // 打印次数
}

type PrinterBase struct {
	Uuid uint64 `json:"uuid"` // 打印日志Uuid
	Name string `json:"name"` // 打印机名称
}

type PrinterBaseResp struct {
	PrinterList  []PrinterBase `json:"printer_list"`      // 打印机列表
	PrinterTypes []PrinterBase `json:"printer_type_list"` // 打印类型列表
}

type PrinterReportResp struct {
	OldPrinterName string `json:"old_printer_name"` // 旧打印机名称
	OldPrinterSn   string `json:"old_printer_sn"`   // 旧打印机SN
	NewPrinterName string `json:"new_printer_name"` // 新打印机名称
	NewPrinterSn   string `json:"new_printer_sn"`   // 新打印机SN
}

type PrintMenu struct {
	ID         uint64             `json:"id"`          // 模版ID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 模版名称
}

type PrintMenuGroup struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分组名称
	GroupType  int                `json:"group_type"`  // 分组类型
	List       []PrintMenu        `json:"list"`        // 模版列表
}

type PrintMenuListResp struct {
	List []PrintMenuGroup `json:"list"` // 模版分组列表
}

type PrintMenuDetail struct {
	ID      uint64 `json:"id"`       // 模板ID
	Name    string `json:"name"`     // 模板名称
	IsUse   bool   `json:"is_use"`   // 是否USB模板
	TempImg string `json:"temp_img"` // 模板图片路径
	TmpUuid uint64 `json:"tmp_uuid"` // 临时模板UUID
}

type PrintMenuDetailResp struct {
	DefaultTpl      PrintMenuDetail   `json:"default_tpl"`        // 默认模板
	AdvReceiptTpls  []PrintMenuDetail `json:"adv_receipt_tpls"`   // 高级模板列表
	IsAdvReceiptTpl bool              `json:"is_adv_receipt_tpl"` // 是否启用高级模板
}

type ConfigInfo struct {
	Name  string `json:"name"`  // 模块名称
	Value string `json:"value"` // 模块数据
}

type ConfigInfoGroup struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分组名称
	List       []ConfigInfo       `json:"list"`        // 分组列表
}

// ConfigInfoResp
type ConfigInfoResp struct {
	ConfigList   []ConfigInfoGroup `json:"config_list"`   // 模版配置分组列表
	TemplateJson string            `json:"template_json"` // 模版 (JSON字符串)
	TemplateName string            `json:"template_name"` // 模版名称
	TemplateData string            `json:"template_data"` // 模版数据(JSON字符串)
}
