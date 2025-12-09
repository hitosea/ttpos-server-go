package req

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"ttpos-server-go/app/dto/resp/setting"
	errs "ttpos-server-go/app/errors"
)

// WaitTimeColorRange 等待时长颜色区间
type WaitTimeColorRange struct {
	Minute string `json:"minute" binding:"required"` // 时间阈值（分钟，字符串类型以兼容 PHP）
	Color  string `json:"color" binding:"required"`  // 颜色值（RGB 格式，统一使用 #xxxxxx 格式）
	// 颜色值限定：黑色 #100A05，黄色 #FFBE00，红色 #E50028
}

// SaveKitchenSettingReq 保存厨显设置请求
type SaveKitchenSettingReq struct {
	IsWaitColor         string               `json:"is_wait_color" binding:"required,oneof=0 1"`     // 是否开启等待时长颜色 0-关闭 1-开启
	WaitTimeColorRanges []WaitTimeColorRange `json:"wait_time_color_ranges" binding:"required,dive"` // 等待时长颜色区间配置
}

// Validate 验证厨显设置请求参数
func (r *SaveKitchenSettingReq) Validate() error {
	// 验证等待时长颜色区间数量
	if len(r.WaitTimeColorRanges) != 3 {
		return errs.WithMessage(errors.New("必须配置3个时间区间"))
	}

	// 解析并验证第一区间必须为0分钟
	minute0, err := strconv.Atoi(r.WaitTimeColorRanges[0].Minute)
	if err != nil || minute0 != 0 {
		return errs.WithMessage(errors.New("第一区间必须为0分钟"))
	}

	// 解析并验证第二区间范围：1-60分钟
	minute1, err := strconv.Atoi(r.WaitTimeColorRanges[1].Minute)
	if err != nil || minute1 < 1 || minute1 > 60 {
		return errs.WithMessage(errors.New("第二区间必须在1-60分钟之间"))
	}

	// 解析并验证第三区间范围：必须大于第二区间，且≤60分钟
	minute2, err := strconv.Atoi(r.WaitTimeColorRanges[2].Minute)
	if err != nil || minute2 <= minute1 || minute2 > 60 {
		return errs.WithMessage(errors.New("第三区间起点必须大于第二区间，且不超过60分钟"))
	}

	// 验证颜色格式（RGB 格式，限定颜色值）
	validColors := []string{"#100A05", "#FFBE00", "#E50028"}
	for i, rangeItem := range r.WaitTimeColorRanges {
		if !isValidColor(rangeItem.Color, validColors) {
			return errs.WithMessage(errors.New("颜色格式不正确，请使用 RGB 格式（如 #100A05），限定颜色：黑色 #100A05，黄色 #FFBE00，红色 #E50028"))
		}
		// 第一区间必须是黑色
		if i == 0 && strings.ToUpper(rangeItem.Color) != "#100A05" {
			return errs.WithMessage(errors.New("第一区间必须使用黑色 #100A05"))
		}
	}

	return nil
}

// isValidColor 验证颜色格式
func isValidColor(color string, validColors []string) bool {
	// RGB 格式：以 # 开头，6位十六进制
	if !strings.HasPrefix(color, "#") {
		return false
	}
	matched, _ := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, color)
	if !matched {
		return false
	}
	// 限定颜色值
	colorUpper := strings.ToUpper(color)
	for _, validColor := range validColors {
		if colorUpper == validColor {
			return true
		}
	}
	return false
}

// ToSettingWaitTimeColorRanges 转换为 setting.WaitTimeColorRange 格式
func (r *SaveKitchenSettingReq) ToSettingWaitTimeColorRanges() []setting.WaitTimeColorRange {
	result := make([]setting.WaitTimeColorRange, len(r.WaitTimeColorRanges))
	for i, item := range r.WaitTimeColorRanges {
		result[i] = setting.WaitTimeColorRange{
			Minute: item.Minute,                 // 直接传递字符串
			Color:  strings.ToUpper(item.Color), // 统一转换为大写
		}
	}
	return result
}
