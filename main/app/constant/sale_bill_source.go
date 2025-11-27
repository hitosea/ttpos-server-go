package constant

import "ttpos-server-go/app/constant/jwt"

// MapJwtSourceToSaleBillSource 将 JWT Source 映射到 SaleBill.source 字段值
// 参数: jwtSource - JWT token 中的 source 值（如 "cashier", "assistant" 等）
// 返回: SaleBill.source 字段值（0-5）
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
	default:
		return SaleBillSourceDefault
	}
}
