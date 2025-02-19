package resp

import "ttpos-server-go/app/dto"

type CallItem struct {
	CallType   int    `json:"call_type"`
	CreateTime string `json:"create_time"`
	Id         int    `json:"id"`
	IsSend     int    `json:"is_send"`
	Status     int    `json:"status"`
	TableId    int    `json:"table_id"`
	TableNo    string `json:"table_no"`
}

type CallList struct {
	List []CallItem       `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

type PrintExceptionItem struct {
	Id          int         `json:"id"`
	Reason      string      `json:"reason"`
	PrinterId   int         `json:"printer_id"`
	OrderId     int         `json:"order_id"`
	CreateTime  string      `json:"create_time"`
	PrinterName string      `json:"printer_name"`
	No          interface{} `json:"no"`
}

type PrintExceptionList struct {
	List []PrintExceptionItem `json:"list"`
	Meta dto.PageResponse     `json:"meta"`
}

type UnprocessedResp struct {
	Count uint `json:"count"`
}
