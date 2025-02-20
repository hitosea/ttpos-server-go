package req

import "ttpos-server-go/app/dto"

// CallListReq 呼叫列表
type CallListReq struct {
	dto.PageReq // 分页参数
}

// PrintExceptionReq 打印异常列表
type PrintExceptionReq struct {
	dto.PageReq // 分页参数
}
