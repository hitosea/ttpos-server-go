package req

import (
	"errors"
	"slices"
	"strconv"

	"ttpos-server-go/app/dto/resp/setting"
	errs "ttpos-server-go/app/errors"
)

// SaveCashierSettingReq 保存收银机设置请求
type SaveCashierSettingReq struct {
	Carousel                []setting.CarouselItem `json:"carousel"`                   // 轮播内容（已存在，最多15个）
	NoOrderCarouselInterval string                 `json:"no_order_carousel_interval"` // 未点餐时轮播间隔(秒)，范围10-120，默认10
	OrderDisplayMode        string                 `json:"order_display_mode"`         // 点餐时展示模式 carousel/order/order_carousel，默认carousel
	OrderCarousel           []setting.CarouselItem `json:"order_carousel"`             // 点餐时轮播内容（最多15个）
	OrderCarouselInterval   string                 `json:"order_carousel_interval"`    // 点餐时轮播间隔(秒)，范围10-120，默认10
}

// Validate 验证收银机设置请求参数
func (r *SaveCashierSettingReq) Validate() error {
	// 轮播内容数量限制：最多15个
	if len(r.Carousel) > 15 {
		return errs.WithMessage(errors.New("轮播内容最多15个"))
	}

	// 点餐时轮播内容数量限制：最多15个
	if len(r.OrderCarousel) > 15 {
		return errs.WithMessage(errors.New("点餐时轮播内容最多15个"))
	}

	// 参数验证：未点餐时轮播间隔
	if r.NoOrderCarouselInterval == "" || r.NoOrderCarouselInterval == "0" {
		// 默认值
		r.NoOrderCarouselInterval = "10"
	} else {
		interval, err := strconv.Atoi(r.NoOrderCarouselInterval)
		if err != nil {
			return errs.WithMessage(errors.New("未点餐时轮播间隔格式错误"))
		}
		if interval < 10 || interval > 120 {
			return errs.WithMessage(errors.New("未点餐时轮播间隔必须在10-120秒之间"))
		}
	}

	// 参数验证：点餐时展示模式
	if r.OrderDisplayMode != "" {
		validModes := []string{"carousel", "order", "order_carousel"}
		if !slices.Contains(validModes, r.OrderDisplayMode) {
			return errs.WithMessage(errors.New("点餐时展示模式无效"))
		}
	} else {
		// 默认值
		r.OrderDisplayMode = "carousel"
	}

	// 参数验证：点餐时轮播间隔
	if r.OrderCarouselInterval == "" || r.OrderCarouselInterval == "0" {
		// 默认值
		r.OrderCarouselInterval = "10"
	} else {
		interval, err := strconv.Atoi(r.OrderCarouselInterval)
		if err != nil {
			return errs.WithMessage(errors.New("点餐时轮播间隔格式错误"))
		}
		if interval < 10 || interval > 120 {
			return errs.WithMessage(errors.New("点餐时轮播间隔必须在10-120秒之间"))
		}
	}

	return nil
}
