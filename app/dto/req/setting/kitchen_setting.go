package setting

type Kitchen struct {
	IsOpen           string         `json:"is_open"`
	IsComeDish       string         `json:"is_come_dish"`
	IsCallService    string         `json:"is_call_service"`
	Server           Server         `json:"server"`
	AdvancedPassword string         `json:"advanced_password"`
	IsWaitColor      string         `json:"is_wait_color"`
	WaitColor        []string       `json:"wait_color"`
	LanguageList     []LanguageItem `json:"language_list"`
	Language         []string       `json:"language"`
	DefaultLanguage  string         `json:"default_language"`
	KitchenLanguage  string         `json:"kitchen_language"`
}
