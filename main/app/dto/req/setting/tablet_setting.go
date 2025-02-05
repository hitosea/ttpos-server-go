package setting

type Tablet struct {
	Carousel           []CarouselItem   `json:"carousel"`
	IsCallService      string           `json:"is_call_service"`
	IsCustomerOrder    string           `json:"is_customer_order"`
	IsVoiceRemind      string           `json:"is_voice_remind"`
	IsShowSoldOut      int              `json:"is_show_sold_out"`
	IsBuffetOrderLimit string           `json:"is_buffet_order_limit"`
	BuffetOrderLimit   BuffetOrderLimit `json:"buffet_order_limit"`
	IsOrderLimit       string           `json:"is_order_limit"`
	OrderLimit         OrderLimit       `json:"order_limit"`
	Server             Server           `json:"server"`
	AdvancedPassword   string           `json:"advanced_password"`
	LanguageList       []LanguageItem   `json:"language_list"`
	Language           []string         `json:"language"`
	DefaultLanguage    string           `json:"default_language"`
	KitchenLanguage    string           `json:"kitchen_language"`
}
