package setting

import "ttpos-server-go/app/dto"

// KitchenResp 厨显设置，接口响应
type KitchenResp struct {
	IsOpen          string             `json:"is_open"`          // 是否开启厨显功能 0关闭 1开启
	IsComeDish      string             `json:"is_come_dish"`     // 是否开启来菜提醒 0-关闭 1-开启
	IsCallService   string             `json:"is_call_service"`  // 是否开启顾客呼叫提醒 0-关闭 1-开启
	Server          Server             `json:"server"`           // 厨显服务器连接
	IsWaitColor     string             `json:"is_wait_color"`    // 是否开启等待时长颜色 0-关闭 1-开启
	WaitColor       []string           `json:"wait_color"`       // 时长颜色 10分钟-黄色#ffff00 20分钟-红色#ff0000
	LanguageList    []dto.LanguageItem `json:"language_list"`    // 语言列表
	Language        []string           `json:"language"`         // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage string             `json:"default_language"` // 默认语言
	IsSmartKitchen  string             `json:"is_smart_kitchen"` // 是否开启智能后厨 0-关闭 1-开启
}

// Kitchen 厨显设置
type Kitchen struct {
	KitchenResp
	AdvancedPassword string `json:"advanced_password"` // 高级设置密码
}
