package setting

type H5 struct {
	IsCallService      string           `json:"is_call_service"`
	IsCustomerOrder    string           `json:"is_customer_order"`
	IsVoiceRemind      string           `json:"is_voice_remind"`
	IsShowSoldOut      string           `json:"is_show_sold_out"`
	IsBuffetOrderLimit string           `json:"is_buffet_order_limit"`
	BuffetOrderLimit   BuffetOrderLimit `json:"buffet_order_limit"`
	IsOrderLimit       string           `json:"is_order_limit"`
	OrderLimit         OrderLimit       `json:"order_limit"`
	LanguageList       []LanguageItem   `json:"language_list"`
	Language           []string         `json:"language"`
	DefaultLanguage    string           `json:"default_language"`
	KitchenLanguage    string           `json:"kitchen_language"`
}
