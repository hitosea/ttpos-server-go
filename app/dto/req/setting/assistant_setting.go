package setting

// Assistant 点餐助手设置
type Assistant struct {
	Server              Server                    `json:"server"` // 服务器连接
	IsAutoSend          string                    `json:"is_auto_send"`
	IsRemainColor       string                    `json:"is_remain_color"`       // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor         []string                  `json:"remain_color"`          // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	AdvancedPassword    string                    `json:"advanced_password"`     // 高级设置密码
	LockPassword        string                    `json:"lock_password"`         // 锁屏密码
	DefaultMode         string                    `json:"default_mode"`          // 默认模式 0-服务员模式 1-顾客模式
	SupportFunctionList []SupportFunctionListItem `json:"support_function_list"` // 支持功能 （添加会员 人数 调整自助餐 转台 清台 并台 优惠折扣 价格 退菜 备注 结账）
	SupportFunction     []SupportFunctionItem     `json:"support_function"`      // 支持功能 （添加会员-add_member 人数-people 调整自助餐-adjust_buffet 转台-turn_table 清台-clear_table 并台-merge_table 优惠折扣-discount 价格-price 退菜-return_dish 备注-remark 结账-settle 转菜-transfer_dish 赠菜-gift_dish）
	IsAutoLockScreen    string                    `json:"is_auto_lock_screen"`   // 是否开启自动锁屏 0-关闭 1-开启
	AutoLockScreen      string                    `json:"auto_lock_screen"`      // 自动锁屏（秒），默认5分钟
	LanguageList        []LanguageItem            `json:"language_list"`         // 语言列表
	Language            []string                  `json:"language"`              // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage     string                    `json:"default_language"`      // 默认语言
	KitchenLanguage     string                    `json:"kitchen_language"`
}

type SupportFunctionListItem SupportFunctionItem

type SupportFunctionItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
