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
	Uuid uint64 `json:"uuid" binding:"required"` // 打印日志Uuid
}

// CallReq 呼叫请求
type CallReq struct {
	CallType uint8 `json:"call_type" binding:"required,oneof=1 2"` // 呼叫类型1-服务员；2-结账
}
