package setting

type Cashier struct {
	Carousel               []CarouselItem  `json:"carousel"`
	IsAutoSend             string          `json:"is_auto_send"`
	OrderMethod            OrderMethodItem `json:"order_method"`
	Server                 Server          `json:"server"`
	IsRemainColor          string          `json:"is_remain_color"`
	RemainColor            []string        `json:"remain_color"`
	AdvancedPassword       string          `json:"advanced_password"`
	IsOpenCashierPassword  string          `json:"is_open_cashier_password"`
	CashierPassword        string          `json:"cashier_password"`
	LockPassword           string          `json:"lock_password"`
	IsAutoLockScreen       string          `json:"is_auto_lock_screen"`
	AutoLockScreen         string          `json:"auto_lock_screen"`
	IsShowScanSoldOut      int             `json:"is_show_scan_sold_out"`
	IsShowAssistantSoldOut int             `json:"is_show_assistant_sold_out"`
	LanguageList           []LanguageItem  `json:"language_list"`
	Language               []string        `json:"language"`
	DefaultLanguage        string          `json:"default_language"`
	IsAutoOrder            string          `json:"is_auto_order"`
	AutoOrderLimit         string          `json:"auto_order_limit"`
	IsAutoVoice            string          `json:"is_auto_voice"`
	MenuShowSoldOut        string          `json:"menu_show_sold_out"`
	KitchenLanguage        string          `json:"kitchen_language"`
}

type OrderMethodItem struct {
	IsCashierOrder string `json:"is_cashier_order"`
	IsTableOrder   string `json:"is_table_order"`
}
