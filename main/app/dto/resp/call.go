package resp

import "ttpos-server-go/app/dto"

type UnprocessedCallItem struct {
	CallType int    `json:"call_type"` // 呼叫类型:呼叫类型(1服务员,2结账)
	Uuid     int    `json:"uuid"`      // 呼叫Uuid
	IsSend   int    `json:"is_send"`   // 是否已发送：1-是；0-否
	DeskUuid int    `json:"desk_uuid"` // 桌台Uuid
	DeskNo   string `json:"desk_no"`   // 桌台编号
}

type UnprocessedCallList struct {
	List []UnprocessedCallItem `json:"list"` // 未处理呼叫列表
	Meta dto.PageResponse      `json:"meta"`
}

type AbnormalPrintItem struct {
	Uuid         int    `json:"uuid"`           // 打印日志Uuid
	Reason       string `json:"reason"`         // 异常原因
	PrinterUuid  uint64 `json:"printer_uuid"`   // 打印机Uuid
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单uuid
	CreateTime   int64  `json:"create_time"`    // 创建时间
	PrinterName  string `json:"printer_name"`   // 打印时间
	DeskNo       string `json:"desk_no"`        // 桌台编号
}

type AbnormalPrintList struct {
	List []AbnormalPrintItem `json:"list"` // 异常打印列表
	Meta dto.PageResponse    `json:"meta"`
}

type UnprocessedResp struct {
	UnprocessedCallCount int64 `json:"unprocessed_call_count"` // 未处理呼叫数量
	AbnormalPrintCount   int64 `json:"abnormal_print_count"`   // 异常打印数量
}

type ReprintResp struct {
	PrinterLogUuid uint64 `json:"printer_log_uuid"`
	Data           string `json:"data"`
	PrintMethod    int    `json:"print_method"` // 打印方式 1文本打印, 2图片打印
	PrinterUuid    uint64 `json:"printer_uuid"`
	PrinterTime    int64  `json:"printer_time"`
	PrinterName    string `json:"printer_name"`
	PrinterType    string `json:"printer_type"`
	PrinterConfig  string `json:"printer_config"`
	PrintTimes     int    `json:"print_times"`
}
