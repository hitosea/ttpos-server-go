// Package constant 定义 ttpos-websocket 模块相关的错误
package constant

import "errors"

var (
	// WebSocket 基础错误

	// ErrActionRequired 动作类型不能为空
	ErrActionRequired = errors.New("动作类型不能为空")

	// ErrClientIDRequired 客户端ID不能为空
	ErrClientIDRequired = errors.New("客户端ID不能为空")

	// LAN 打印机上报相关错误

	// ErrCompanyUUIDRequired 公司UUID不能为空
	ErrCompanyUUIDRequired = errors.New("公司UUID不能为空")

	// ErrDeviceIDRequired 设备ID不能为空
	ErrDeviceIDRequired = errors.New("设备ID不能为空")

	// ErrSourceClientRequired 来源客户端不能为空
	ErrSourceClientRequired = errors.New("来源客户端不能为空")

	// ErrPrintersRequired 打印机列表不能为空
	ErrPrintersRequired = errors.New("打印机列表不能为空")
)
