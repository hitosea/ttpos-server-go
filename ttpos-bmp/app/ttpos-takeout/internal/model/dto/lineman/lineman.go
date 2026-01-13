package lineman

import "github.com/golang-jwt/jwt/v5"

// LinemanTokenClaims LINE MAN Token 的 JWT Claims
type LinemanTokenClaims struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

// BaseResponse Lineman API 通用响应结构
// 所有 Lineman API 响应都包含这些基础字段
type BaseResponse struct {
	Status  string `json:"status"`            // "ok" / "error"
	Code    string `json:"code"`              // "SUCCESS" / 错误码
	Message string `json:"message,omitempty"` // 错误消息
}
