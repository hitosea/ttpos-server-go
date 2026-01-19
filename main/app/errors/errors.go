package errors

import (
	"errors"
	"strings"
	"ttpos-server-go/app/constant"
)

type AppError struct {
	Code    int      // 响应码
	Message string   // 错误信息
	Replace []string // 替换信息
	data    any      // 附加数据
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) GetCode() int {
	return e.Code
}

func (e AppError) GetData() any {
	return e.data
}

func New(messages ...string) AppError {
	msg := strings.Join(messages, " ")
	return AppError{Code: constant.CodeFail, Message: msg}
}

func NewWithCode(code int, message string) AppError {
	return AppError{Code: code, Message: message}
}

func NewWithReplace(message string, replace []string) AppError {
	return AppError{Code: constant.CodeFail, Message: message, Replace: replace}
}

func NewWithCodeAndReplace(code int, message string, replace []string) AppError {
	return AppError{Code: code, Message: message, Replace: replace}
}

func NewWithCodeAndData(code int, data any, message string) AppError {
	return AppError{Code: code, Message: message, data: data}
}

func Is(err error, target AppError) bool {
	return errors.As(err, &target)
}

var (
	ErrInternal            = AppError{Code: constant.CodeSystemError, Message: "系统内部错误"}
	ErrNoDeviceSn          = AppError{Code: constant.CodeParamError, Message: "无法解析到设备SN"}
	ErrMustPlanNotComplete = AppError{Code: constant.CodeOrderCheckProductMust, Message: "必点商品未选择完成，请选择对应商品"}
	ErrProductPriceChanged = AppError{Code: constant.CodeOrderCheckProductPriceChanged, Message: "商品价格信息发生变化，请确认后下单"}
)
