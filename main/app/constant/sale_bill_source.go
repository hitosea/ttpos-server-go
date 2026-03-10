package constant

import (
	"strings"
	"ttpos-server-go/app/constant/jwt"
)

// MapJwtSourceToSaleBillSource 将 JWT Source 映射到 SaleBill.source 字段值
// 参数: jwtSource - JWT token 中的 source 值（如 "cashier", "assistant", "kiosk" 等）
// 返回: SaleBill.source 字段值（0-6）
// 注意: SaleBillSource 常量已在 device.go 中定义
func MapJwtSourceToSaleBillSource(jwtSource string) uint {
	switch jwtSource {
	case jwt.SourceCashier:
		return SaleBillSourceCashier
	case jwt.SourceAssistant:
		return SaleBillSourceAssistant
	case jwt.SourceTablet:
		return SaleBillSourceTablet
	case jwt.SourceH5:
		return SaleBillSourceH5
	case jwt.SourceMember:
		return SaleBillSourceMember
	case jwt.SourceKiosk:
		return SaleBillSourceKiosk
	default:
		return SaleBillSourceDefault
	}
}

// NormalizeClientVersion 标准化客户端版本号格式
// 功能: 去除前缀 "v" 或 "V"，统一为 x.y.z 格式
// 参数: version - 原始版本号（如 "2.10.0", "v2.10.0", "V2.10.0"）
// 返回: 标准化后的版本号（如 "2.10.0"），如果为空则返回空字符串
func NormalizeClientVersion(version string) string {
	if version == "" {
		return ""
	}
	// 去除前后空格，再去除前缀 "v" 或 "V"
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	return version
}
