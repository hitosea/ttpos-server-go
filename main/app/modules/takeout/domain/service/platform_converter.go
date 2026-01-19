package service

import (
	"ttpos-server-go/pkg/context"
)

// IPlatformConverter 平台转换器接口（策略模式）
type IPlatformConverter interface {
	// ConvertToTTPOS 从平台格式转换为 TTPOS 数据
	ConvertToTTPOS(ctx context.Context, platformData interface{}) (interface{}, error)

	// ValidateData 验证平台数据格式
	ValidateData(platformData interface{}) error

	// GetPlatformName 获取平台名称
	GetPlatformName() string
}
