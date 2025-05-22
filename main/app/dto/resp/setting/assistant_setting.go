package setting

import "ttpos-server-go/app/dto"

// AssistantResp 点餐助手设置，接口响应
type AssistantResp struct {
	Server                 Server             `json:"server"` // 服务器连接
	IsAutoSend             string             `json:"is_auto_send"`
	IsRemainColor          string             `json:"is_remain_color"`            // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor            []string           `json:"remain_color"`               // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	DefaultMode            string             `json:"default_mode"`               // 默认模式 0-服务员模式 1-顾客模式
	IsAutoLockScreen       string             `json:"is_auto_lock_screen"`        // 是否开启自动锁屏 0-关闭 1-开启
	IsCheckOrder           string             `json:"is_check_order"`             // 下单校验高级密码 0-关闭 1-开启。默认关闭
	AutoLockScreen         string             `json:"auto_lock_screen"`           // 自动锁屏（秒），默认5分钟
	LanguageList           []dto.LanguageItem `json:"language_list"`              // 语言列表
	Language               []string           `json:"language"`                   // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage        string             `json:"default_language"`           // 默认语言
	IsShowAssistantSoldOut int                `json:"is_show_assistant_sold_out"` // 是否显示售罄
}

// IsCheckOrderPassword 是否开启下单校验高级密码
func (resp *AssistantResp) IsCheckOrderPassword() bool {
	return resp.IsCheckOrder == "1"
}

// Assistant 点餐助手设置
type Assistant struct {
	AssistantResp
	AdvancedPassword string `json:"advanced_password"` // 高级设置密码
	LockPassword     string `json:"lock_password"`     // 锁屏密码
}
