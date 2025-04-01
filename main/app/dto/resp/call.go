package resp

import "ttpos-server-go/app/dto"

type UnprocessedCallItem struct {
	CallType uint8  `json:"call_type"` // 呼叫类型:呼叫类型(1服务员,2结账)
	Uuid     uint64 `json:"uuid"`      // 呼叫Uuid
	IsSend   int    `json:"is_send"`   // 是否已发送：1-是；0-否
	DeskUuid uint64 `json:"desk_uuid"` // 桌台Uuid
	DeskNo   string `json:"desk_no"`   // 桌台编号
}

type UnprocessedCallList struct {
	List []UnprocessedCallItem `json:"list"` // 未处理呼叫列表
	Meta dto.PageResponse      `json:"meta"`
}

type AbnormalPrintItem struct {
	Uuid         uint64 `json:"uuid"`           // 打印日志Uuid
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

type UnprocessedCallItemForNotice struct {
	CallType uint8  `json:"call_type"` // 呼叫类型:呼叫类型(1服务员,2结账)
	Uuid     uint64 `json:"uuid"`      // 呼叫Uuid
	DeskNo   string `json:"desk_no"`   // 桌台编号
	CallText string `json:"call_text"` // 呼叫语音文字
}

type UnprocessedCall struct {
	List []UnprocessedCallItemForNotice `json:"list"`
}

type AbnormalPrintItemForNotice struct {
	Uuid   uint64 `json:"uuid"`    // 打印日志Uuid
	Reason string `json:"reason"`  // 异常原因
	DeskNo string `json:"desk_no"` // 桌台编号
}
type UnprocessedAbnormalPrint struct {
	List []AbnormalPrintItemForNotice `json:"list"`
}

type UnprocessedH5OrderItem struct {
	Uuid         uint64 `json:"uuid"`           // h5订单Uuid
	DeskNo       string `json:"desk_no"`        // 桌台编号
	Status       uint   `json:"status"`         // 订单状态, 0-未下单 1-未接单 2-已接单 3-已拒单
	IsAutoAccept bool   `json:"is_auto_accept"` // 是否自动接单
}

type UnprocessedH5Order struct {
	List []UnprocessedH5OrderItem `json:"list"`
}

type UnprocessedListResp struct {
	Call          UnprocessedCall          `json:"call"`           // 未处理呼叫列表，最新十条
	AbnormalPrint UnprocessedAbnormalPrint `json:"abnormal_print"` // 未处理的异常打印，最新十条
	H5Order       UnprocessedH5Order       `json:"h5_order"`       // 未处理的H5订单，最新十条
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
