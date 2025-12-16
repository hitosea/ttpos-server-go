package req

import (
	"errors"
	"regexp"
	"slices"
	"ttpos-server-go/app/dto/resp/setting"
	errs "ttpos-server-go/app/errors"
)

// SaveKioskSettingReq 保存自助点餐机设置请求
type SaveKioskSettingReq struct {
	AdvancedPassword  string                 `json:"advanced_password"`   // 高级密码（4-8位整数，默认666888）
	CallWaiterEnabled int                    `json:"call_waiter_enabled"` // 呼叫服务员开关（0-关闭，1-开启，默认1）
	Language          []string               `json:"language"`            // 已设置的语言（常用语言，JSON数组，默认所有语言）
	DefaultLanguage   string                 `json:"default_language"`    // 默认语言（默认语言1）
	Carousel          []setting.CarouselItem `json:"carousel"`            // 轮播内容（图片+视频，最多10张图片+5个视频，总共最多15个，支持排序）
}

// Validate 验证自助点餐机设置请求参数
func (r *SaveKioskSettingReq) Validate() error {
	// 高级密码格式校验：4-8位整数
	if r.AdvancedPassword != "" {
		matched, _ := regexp.MatchString(`^[0-9]{4,8}$`, r.AdvancedPassword)
		if !matched {
			return errs.WithMessage(errors.New("高级密码必须为4-8位整数"))
		}
	}

	// 呼叫服务员开关校验
	if r.CallWaiterEnabled != 0 && r.CallWaiterEnabled != 1 {
		return errs.WithMessage(errors.New("呼叫服务员开关值无效"))
	}

	// 语言和默认语言校验
	if len(r.Language) > 0 && r.DefaultLanguage != "" {
		if !slices.Contains(r.Language, r.DefaultLanguage) {
			return errs.WithMessage(errors.New("默认语言必须在常用语言列表中"))
		}
	}

	// 轮播内容数量限制：最多15个（10张图片+5个视频）
	if len(r.Carousel) > 15 {
		return errs.WithMessage(errors.New("轮播内容最多15个"))
	}

	// 统计图片和视频数量
	imageCount := 0
	videoCount := 0
	for _, item := range r.Carousel {
		if item.Type == "image" {
			imageCount++
		} else if item.Type == "video" {
			videoCount++
		}
	}

	// 图片数量限制：最多10张
	if imageCount > 10 {
		return errs.WithMessage(errors.New("轮播图片最多10张"))
	}

	// 视频数量限制：最多5个
	if videoCount > 5 {
		return errs.WithMessage(errors.New("轮播视频最多5个"))
	}

	return nil
}
