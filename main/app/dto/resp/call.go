package resp

import "ttpos-server-go/app/dto"

type UnprocessedCallItem struct {
	CallType int    `json:"call_type"`
	Uuid     int    `json:"uuid"`
	IsSend   int    `json:"is_send"`
	Status   int    `json:"status"`
	DeskUuid int    `json:"desk_uuid"`
	DeskNo   string `json:"desk_no"`
}

type UnprocessedCallList struct {
	List []UnprocessedCallItem `json:"list"`
	Meta dto.PageResponse      `json:"meta"`
}

type AbnormalPrintItem struct {
	Uuid         int    `json:"uuid"`
	Reason       string `json:"reason"`
	PrinterUuid  uint64 `json:"printer_uuid"`
	SaleBillUuid uint64 `json:"sale_bill_uuid"`
	CreateTime   int64  `json:"create_time"`
	PrinterName  string `json:"printer_name"`
	DeskNo       string `json:"desk_no"`
}

type AbnormalPrintList struct {
	List []AbnormalPrintItem `json:"list"`
	Meta dto.PageResponse    `json:"meta"`
}

type UnprocessedResp struct {
	Count uint `json:"count"`
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
