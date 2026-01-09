// Package lineman 定义 Lineman API 数据传输对象
package lineman

// BaseResponse Lineman API 通用响应结构
// 所有 Lineman API 响应都包含这些基础字段
type BaseResponse struct {
	Status  string `json:"status"`            // "ok" / "error"
	Code    string `json:"code"`              // "SUCCESS" / 错误码
	Message string `json:"message,omitempty"` // 错误消息
}
