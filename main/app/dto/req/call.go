package req

import "ttpos-server-go/app/dto"

// UnprocessedCallListReq 呼叫列表
type UnprocessedCallListReq struct {
	dto.PageReq // 分页参数
}

// AbnormalPrintListReq 打印异常列表
type AbnormalPrintListReq struct {
	dto.PageReq // 分页参数
}

// ProcessedCallReq 处理呼叫
type ProcessedCallReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}

// PrinterLogReq 打印日志ID
type PrinterLogReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}
