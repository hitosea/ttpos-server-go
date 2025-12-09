package setting

import "ttpos-server-go/app/dto"

// WaitTimeColorRange 等待时长颜色区间
type WaitTimeColorRange struct {
	Minute string `json:"minute"` // 时间阈值（分钟，字符串类型以兼容 PHP）
	Color  string `json:"color"`  // 颜色值（RGB 格式，统一使用 #xxxxxx 格式）
	// 颜色值限定：黑色 #100A05，黄色 #FFBE00，红色 #E50028
}

// KitchenResp 厨显设置，接口响应
type KitchenResp struct {
	IsOpen              string               `json:"is_open"`                // 是否开启厨显功能 0关闭 1开启
	IsComeDish          string               `json:"is_come_dish"`           // 是否开启来菜提醒 0-关闭 1-开启
	IsCallService       string               `json:"is_call_service"`        // 是否开启顾客呼叫提醒 0-关闭 1-开启
	Server              Server               `json:"server"`                 // 厨显服务器连接
	IsWaitColor         string               `json:"is_wait_color"`          // 是否开启等待时长颜色 0-关闭 1-开启
	WaitColor           []string             `json:"wait_color"`             // 时长颜色（旧格式，保持兼容：["red", "yellow"]）
	WaitTimeColorRanges []WaitTimeColorRange `json:"wait_time_color_ranges"` // 等待时长颜色区间配置（新格式）
	LanguageList        []dto.LanguageItem   `json:"language_list"`          // 语言列表
	Language            []string             `json:"language"`               // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage     string               `json:"default_language"`       // 默认语言
	IsSmartKitchen      string               `json:"is_smart_kitchen"`       // 是否开启智能后厨 0-关闭 1-开启
}

// Kitchen 厨显设置
type Kitchen struct {
	KitchenResp
	AdvancedPassword string `json:"advanced_password"` // 高级设置密码
}
