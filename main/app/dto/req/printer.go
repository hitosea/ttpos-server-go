package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// PrinterListReq 打印机列表查询
type PrinterListReq struct {
	dto.PageReq           // 分页参数
	QueryStartTime uint   `form:"query_start_time"`     // 查询开始时间戳
	QueryEndTime   uint   `form:"query_end_time"`       // 查询结束时间戳
	Status         int    `form:"status,default=-1"`    // 状态, -1=全都、0=失败, 1=成功, 2=补打成功, 3=补打失败
	DataType       int    `form:"data_type,default=-1"` // 数据类型 (打印类型), -1=全都、.....
	PrinterUuid    uint64 `form:"printer_uuid"`         // 打印机UUID
}

// PrinterPrintReq 打印请求
type PrinterPrintReq struct {
	Uuid uint64 `json:"uuid"` // 打印日志uuid
}

type PrinterReportReq struct {
	Uuid   uint64 `json:"uuid"`   // 打印uuid
	Status int    `json:"status"` // 状态 0=失败, 1=成功
	Reason string `json:"reason"` // 失败原因
}

type PrinterReportReqs struct {
	Data []PrinterReportReq `json:"data"` // 打印日志uuid
}

// Validate 验证参数
func (req PrinterReportReqs) Validate() error {
	if len(req.Data) == 0 {
		return errors.New("打印报告数据不能为空")
	}
	for _, report := range req.Data {
		if report.Status != 0 && report.Status != 1 {
			return errors.New("打印状态错误")
		}
		if report.Status == 0 && report.Reason == "" {
			return errors.New("打印失败时必须提供失败原因")
		}
	}
	return nil
}

// Uuids 获取所有打印UUID
func (req PrinterReportReqs) Uuids() []uint64 {
	uuids := make([]uint64, 0, len(req.Data))
	for _, report := range req.Data {
		uuids = append(uuids, report.Uuid)
	}
	return uuids
}

type UsbPrinterReportPrinterData struct {
	M_name string `json:"m_name"` // 厂商名称
	Name   string `json:"name"`   // 打印机名称
	Pid    any    `json:"pid"`    // 打印机PID
	Sn     string `json:"sn"`     // 打印机SN
	Vid    any    `json:"vid"`    // 打印机VID
}

// UsbPrinterReportReq usb 打印机上报
type UsbPrinterReportReq struct {
	List       []UsbPrinterReportPrinterData `json:"list"`        // 所有在线的USB打印机数据
	SelectedSn string                        `json:"selected_sn"` // 选择的打印机SN
}

// CreatePrinterCustomizeReq 创建打印机定制请求
type CreatePrinterCustomizeReq struct {
	TemplateId uint64 `json:"template_id"` // 模板ID
	Name       string `json:"name"`        // 打印机定制名称
	Data       string `json:"data"`        // 打印机定制数据
}

// EditPrinterCustomizeReq 编辑打印机定制请求
type EditPrinterCustomizeReq struct {
	CustomizeUuid uint64 `json:"customize_uuid"` // 打印机定制UUID
	Name          string `json:"name"`           // 模板名称
	Data          string `json:"data"`           // 定制数据
}

// GetConfigInfoReq 获取配置信息请求
type GetConfigInfoReq struct {
	TemplateId    uint64 `json:"template_id"`    // 模板ID
	CustomizeUuid uint64 `json:"customize_uuid"` // 定制UUID, 编辑的时候使用，新增的时候 传0
	IsAdv         int    `json:"is_adv"`         // 是否高级模版 0=否, 1=是
}
