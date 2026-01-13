package utility

import (
	"github.com/gogf/gf/v2/errors/gerror"
)

// MapStatusToLinemanModifier 将 TTPOS 状态（string）映射为 Lineman 修饰符状态（int）
//
// 映射规则:
//   - "AVAILABLE" → 1 (AVAILABLE)
//   - "UNAVAILABLE" → 3 (SUSPENDED)
//   - "SOLD_OUT_TODAY" → 2 (SOLD_OUT_TODAY)
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
	case "UNAVAILABLE":
		return 3, nil // SUSPENDED
	case "SOLD_OUT_TODAY":
		return 2, nil // SOLD_OUT_TODAY
	case "":
		return 0, gerror.New("available_status 不能为空")
	default:
		return 0, gerror.Newf("不支持的状态: %s", ttposStatus)
	}
}
