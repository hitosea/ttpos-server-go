package errors

import "ttpos-server-go/app/constant"

type AppError struct {
	Code    int         // 响应码
	Message string      // 错误信息
	Replace []string    // 替换信息
	data    interface{} // 附加数据
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) GetCode() int {
	return e.Code
}

func New(message string) *AppError {
	return &AppError{Code: constant.CodeFail, Message: message}
}

func NewWithCode(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewWithReplace(message string, replace []string) *AppError {
	return &AppError{Code: constant.CodeFail, Message: message, Replace: replace}
}

func NewWithCodeAndReplace(code int, message string, replace []string) *AppError {
	return &AppError{Code: code, Message: message, Replace: replace}
}

func NewWithCodeAndData(code int, data interface{}, message string) *AppError {
	return &AppError{Code: code, Message: message, data: data}
}

var (
	ErrInternal   = &AppError{Code: constant.CodeSystemError, Message: "系统内部错误"}
	ErrNoDeviceSn = &AppError{Code: constant.CodeParamError, Message: "无法解析到设备SN"}
)
