package req

import (
	"ttpos-server-go/app/dto"
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
