package setting

import "ttpos-server-go/app/dto"

// KioskResp 自助点餐机设置，接口响应
type KioskResp struct {
	AdvancedPassword  string             `json:"advanced_password"`   // 高级密码（默认666888）
	CallWaiterEnabled int                `json:"call_waiter_enabled"` // 呼叫服务员开关（默认1-开启）
	LanguageList      []dto.LanguageItem `json:"language_list"`       // 语言列表，当前勾选了的语言列表
	Language          []string           `json:"language"`            // 已设置的语言（常用语言）
	DefaultLanguage   string             `json:"default_language"`    // 默认语言（默认语言1）
	Carousel          []CarouselItem     `json:"carousel"`            // 轮播内容（图片+视频，最多10张图片+5个视频，总共最多15个，支持排序）
}

// Kiosk 自助点餐机设置（内部使用，包含敏感字段）
type Kiosk struct {
	KioskResp
	// 可扩展其他内部字段
}
