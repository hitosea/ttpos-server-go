package setting

type Assistant struct {
	Server              Server                    `json:"server"`
	IsAutoSend          string                    `json:"is_auto_send"`
	IsRemainColor       string                    `json:"is_remain_color"`
	RemainColor         []string                  `json:"remain_color"`
	AdvancedPassword    string                    `json:"advanced_password"`
	LockPassword        string                    `json:"lock_password"`
	DefaultMode         string                    `json:"default_mode"`
	SupportFunctionList []SupportFunctionListItem `json:"support_function_list"`
	SupportFunction     []SupportFunctionItem     `json:"support_function"`
	IsAutoLockScreen    string                    `json:"is_auto_lock_screen"`
	AutoLockScreen      string                    `json:"auto_lock_screen"`
	LanguageList        []LanguageItem            `json:"language_list"`
	Language            []string                  `json:"language"`
	DefaultLanguage     string                    `json:"default_language"`
	KitchenLanguage     string                    `json:"kitchen_language"`
}

type SupportFunctionListItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type SupportFunctionItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
