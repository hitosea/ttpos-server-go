package errors

import "ttpos-server-go/app/constant"

type AppError struct {
	Code    int      // 响应码
	Message string   // 错误信息
	Replace []string // 替换信息
	Cause   string   // 错误原因
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) GetCode() int {
	return e.Code
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewWithReplace(code int, message string, replace []string) *AppError {
	return &AppError{Code: code, Message: message, Replace: replace}
}

func NewWithCause(code int, message string, cause string) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

func NewWithAll(code int, message string, replace []string, cause string) *AppError {
	return &AppError{Code: code, Message: message, Replace: replace, Cause: cause}
}

var (
	ErrUserNotFound       = &AppError{Code: constant.CodeFail, Message: "user not found"}
	ErrInvalidCredentials = &AppError{Code: 401, Message: "invalid credentials"}
)
