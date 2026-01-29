package utility

import (
	"github.com/gogf/gf/v2/errors/gerror"
)

// MapStatusToLineman 将 TTPOS 状态映射为 Lineman 菜单状态（string）
//
// 映射规则:
//   - "AVAILABLE" → "AVAILABLE"
//   - "UNAVAILABLE", "HIDE" → "SUSPENDED"
//   - "UNAVAILABLETODAY" → "SOLD_OUT_TODAY"
//   - 其他值 → 返回错误
//
// 参考: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633
func MapStatusToLineman(ttposStatus string) (string, error) {
	switch ttposStatus {
	case "AVAILABLE":
		return "AVAILABLE", nil
	case "UNAVAILABLE", "HIDE":
		return "SUSPENDED", nil
	case "UNAVAILABLETODAY":
		return "SOLD_OUT_TODAY", nil
	default:
		return "", gerror.Newf("不支持的状态: %s", ttposStatus)
	}
}

// MapStatusToLinemanModifier 将 TTPOS 状态（string）映射为 Lineman 修饰符状态（int）
//
// 映射规则:
//   - "AVAILABLE" → 1 (AVAILABLE)
//   - "UNAVAILABLE", "HIDE" → 3 (SUSPENDED)
//   - "UNAVAILABLETODAY" → 2 (SOLD_OUT_TODAY)
//   - 其他值 → 返回错误
//
// 参数:
//   - ttposStatus: TTPOS 状态（string）
//
// 返回:
//   - int: Lineman 状态（1, 2, 3）
//   - error: 错误信息（不支持的状态）
//
// 参考: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=1934684079#gid=1934684079
func MapStatusToLinemanModifier(ttposStatus string) (int, error) {
	switch ttposStatus {
	case "AVAILABLE":
		return 1, nil // AVAILABLE
	case "UNAVAILABLE", "HIDE":
		return 3, nil // SUSPENDED
	case "UNAVAILABLETODAY":
		return 2, nil // SOLD_OUT_TODAY
	case "":
		return 0, gerror.New("available_status 不能为空")
	default:
		return 0, gerror.Newf("不支持的状态: %s", ttposStatus)
	}
}
