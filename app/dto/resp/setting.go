package resp

type BuffetSettingValue struct {
	IsOpen                   string        `json:"is_open"`
	TabletEndTime            string        `json:"tablet_end_time"`
	IsRemainContinue         string        `json:"is_remain_continue"`
	RemainContinueTime       string        `json:"remain_continue_time"`
	RemainContinueNoticeTime string        `json:"remain_continue_notice_time"`
	IsBuyContinue            string        `json:"is_buy_continue"`
	IsAddClock               string        `json:"is_add_clock"`
	IsBuffetDiscount         string        `json:"is_buffet_discount"`
	AddClock                 []interface{} `json:"add_clock"`
}

type PaymentSettingValue struct {
	IsCash    string `json:"is_cash"`
	IsBalance string `json:"is_balance"`
	IsOther   string `json:"is_other"`
}

type BusinessSettingValue struct {
	ZeroingMethodList         []ZeroingMethodItem         `json:"zeroing_method_list"`
	CheckoutZeroingMethodList []CheckoutZeroingMethodItem `json:"checkout_zeroing_method_list"`
	ZeroingMethod             string                      `json:"zeroing_method"`
	CheckoutZeroingMethod     string                      `json:"checkout_zeroing_method"`
	GiftMethodList            []GiftMethodItem            `json:"gift_method_list"`
	GiftMethod                string                      `json:"gift_method"`
	FreeMethodList            []FreeMethodItem            `json:"free_method_list"`
	FreeMethod                string                      `json:"free_method"`
	DiscountMethod            string                      `json:"discount_method"`
	QrCode                    string                      `json:"qr_code"`
	NoClearTable              string                      `json:"no_clear_table"`
	IsNeedPassword            string                      `json:"is_need_password"`
	DishCardStyle             string                      `json:"dish_card_style"`
	DishCardStyleTime         string                      `json:"dish_card_style_time"`
	IsInvoice                 string                      `json:"is_invoice"`
}

type ZeroingMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type GiftMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type FreeMethodItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type AssistantSettingValue struct {
	Server                 Server                `json:"server"`
	IsRemainColor          string                `json:"is_remain_color"`
	RemainColor            []string              `json:"remain_color"`
	AdvancedPassword       string                `json:"advanced_password"`
	LockPassword           string                `json:"lock_password"`
	DefaultMode            string                `json:"default_mode"`
	SupportFunctionList    []SupportFunctionItem `json:"support_function_list"`
	SupportFunction        []interface{}         `json:"support_function"`
	IsAutoLockScreen       string                `json:"is_auto_lock_screen"`
	AutoLockScreen         string                `json:"auto_lock_screen"`
	LanguageList           []interface{}         `json:"language_list"`
	Language               []string              `json:"language"`
	DefaultLanguage        string                `json:"default_language"`
	IsShowAssistantSoldOut int                   `json:"is_show_assistant_sold_out"`
	KitchenLanguage        string                `json:"kitchen_language"`
}

type SupportFunctionItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type KitchenSettingValue struct {
	IsOpen           string        `json:"is_open"`
	IsComeDish       string        `json:"is_come_dish"`
	IsCallService    string        `json:"is_call_service"`
	Server           Server        `json:"server"`
	AdvancedPassword string        `json:"advanced_password"`
	IsWaitColor      string        `json:"is_wait_color"`
	WaitColor        []interface{} `json:"wait_color"`
	LanguageList     []interface{} `json:"language_list"`
	DefaultLanguage  string        `json:"default_language"`
	KitchenLanguage  string        `json:"kitchen_language"`
	Language         []string      `json:"language"`
}

type H5SettingValue struct {
	IsCallService      string          `json:"is_call_service"`
	IsCustomerOrder    string          `json:"is_customer_order"`
	IsVoiceRemind      string          `json:"is_voice_remind"`
	IsShowSoldOut      string          `json:"is_show_sold_out"`
	IsBuffetOrderLimit string          `json:"is_buffet_order_limit"`
	BuffetOrderLimit   BuffetOrderItem `json:"buffet_order_limit"`
	IsOrderLimit       string          `json:"is_order_limit"`
	OrderLimit         OrderLimit      `json:"order_limit"`
	LanguageList       []interface{}   `json:"language_list"`
	Language           []string        `json:"language"`
	DefaultLanguage    string          `json:"default_language"`
	IsShowScanSoldOut  int             `json:"is_show_scan_sold_out"`
	KitchenLanguage    string          `json:"kitchen_language"`
}

type TabletSettingValue struct {
	Carousel           []interface{}   `json:"carousel"`
	IsCallService      string          `json:"is_call_service"`
	IsCustomerOrder    string          `json:"is_customer_order"`
	IsVoiceRemind      string          `json:"is_voice_remind"`
	IsShowSoldOut      int             `json:"is_show_sold_out"`
	IsBuffetOrderLimit string          `json:"is_buffet_order_limit"`
	BuffetOrderLimit   BuffetOrderItem `json:"buffet_order_limit"`
	IsOrderLimit       string          `json:"is_order_limit"`
	OrderLimit         OrderLimit      `json:"order_limit"`
	Server             Server          `json:"server"`
	AdvancedPassword   string          `json:"advanced_password"`
	LanguageList       []LanguageItem  `json:"language_list"`
	DefaultLanguage    string          `json:"default_language"`
	KitchenLanguage    string          `json:"kitchen_language"`
	Language           []string        `json:"language"`
}

type CashierSettingValue struct {
	Carousel               []interface{}  `json:"carousel"`
	IsAutoSend             string         `json:"is_auto_send"`
	OrderMethod            OrderMethod    `json:"order_method"`
	Server                 Server         `json:"server"`
	IsRemainColor          string         `json:"is_remain_color"`
	RemainColor            []string       `json:"remain_color"`
	AdvancedPassword       string         `json:"advanced_password"`
	IsOpenCashierPassword  string         `json:"is_open_cashier_password"`
	CashierPassword        string         `json:"cashier_password"`
	LockPassword           string         `json:"lock_password"`
	IsAutoLockScreen       string         `json:"is_auto_lock_screen"`
	AutoLockScreen         string         `json:"auto_lock_screen"`
	IsShowScanSoldOut      int            `json:"is_show_scan_sold_out"`
	IsShowAssistantSoldOut int            `json:"is_show_assistant_sold_out"`
	LanguageList           []LanguageItem `json:"language_list"`
	DefaultLanguage        string         `json:"default_language"`
	IsAutoOrder            string         `json:"is_auto_order"`
	AutoOrderLimit         string         `json:"auto_order_limit"`
	IsAutoVoice            string         `json:"is_auto_voice"`
	MenuShowSoldOut        string         `json:"menu_show_sold_out"`
	KitchenLanguage        string         `json:"kitchen_language"`
	Language               []string       `json:"language"`
}

type ChargeSettingValue struct {
	IsOpen            string `json:"is_open"`
	ChargeType        string `json:"charge_type"`
	ServiceCharge     string `json:"service_charge"`
	ServiceChargeRate string `json:"service_charge_rate"`
	IsOpenTax         string `json:"is_open_tax"`
}

type TaxRateSettingValue struct {
	IsOpen         string        `json:"is_open"`
	TaxRate        string        `json:"tax_rate"`
	CalcType       string        `json:"calc_type"`
	AddTaxCategory []interface{} `json:"add_tax_category"`
}

type CurrencySettingValue struct {
	Unit             string `json:"unit"`
	PrintUnit        string `json:"print_unit"`
	UnitPosition     string `json:"unit_position"`
	IsOpen           string `json:"is_open"`
	ViceUnit         string `json:"vice_unit"`
	ViceUnitPosition string `json:"vice_unit_position"`
	UnitRate         string `json:"unit_rate"`
}

type BalanceSettingValue struct {
	IsOpen   string `json:"is_open"`
	IsPlan   string `json:"is_plan"`
	MinMoney int    `json:"min_money"`
	Describe string `json:"describe"`
}

type SysConfigSettingValue struct {
	ShopName     string `json:"shop_name"`
	ShopBgImg    string `json:"shop_bg_img"`
	ShopLogoImg  string `json:"shop_logo_img"`
	CashierName  string `json:"cashier_name"`
	CashierBgImg string `json:"cashier_bg_img"`
}

type SysAdminConfigSettingValue struct {
	BrandName          string `json:"brand_name"`
	BrandLogo          string `json:"brand_logo"`
	BrandLogoLong      string `json:"brand_logo_long"`
	BrowserLogo        string `json:"browser_logo"`
	BrowserTitle       string `json:"browser_title"`
	ExpirationReminder int    `json:"expiration_reminder"`
}

type PointsSettingValue struct {
	DeductionOrder     string   `json:"deduction_order"`
	DeductRatioMain    string   `json:"deduct_ratio_main"`
	DeductRatioGift    string   `json:"deduct_ratio_gift"`
	PointsName         string   `json:"points_name"`
	IsShoppingGift     string   `json:"is_shopping_gift"`
	GiftRatio          string   `json:"gift_ratio"`
	IsShoppingDiscount string   `json:"is_shopping_discount"`
	Discount           Discount `json:"discount"`
	Describe           string   `json:"describe"`
}

type Discount struct {
	DiscountRatio  string `json:"discount_ratio"`
	FullOrderPrice string `json:"full_order_price"`
	MaxMoneyRatio  string `json:"max_money_ratio"`
}

type RechargeSettingValue struct {
	IsEntrance  string `json:"is_entrance"`
	IsCustom    string `json:"is_custom"`
	IsMatchPlan string `json:"is_match_plan"`
	Describe    string `json:"describe"`
}

type PrinterSettingValue struct {
	CashierOpen        string         `json:"cashier_open"`
	CashierPrinterID   string         `json:"cashier_printer_id"`
	CashierPrinter     []interface{}  `json:"cashier_printer"`
	LanguageList       []LanguageItem `json:"language_list"`
	LanguageMethod     string         `json:"language_method"`
	DefaultLanguage    string         `json:"default_language"`
	PrintMethod        int            `json:"print_method"`
	KitchenLanguage    string         `json:"kitchen_language"`
	KitchenPrintMethod int            `json:"kitchen_print_method"`
	ConsumptionTax     string         `json:"consumption_tax"`
	BuffetSignOpen     string         `json:"buffet_sign_open"`
	MonetaryUnitOpen   string         `json:"monetary_unit_open"`
	CalendarList       []CalendarItem `json:"calendar_list"`
	PrintList          []PrintItem    `json:"print_list"`
	DefaultCalendar    string         `json:"default_calendar"`
	Language           []string       `json:"language"`
}
type CalendarItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type PrintItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type StorageSettingValue struct {
	Default string `json:"default"`
	Engine  Engine `json:"engine"`
}

type Qiniu struct {
	Bucket    string `json:"bucket"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Domain    string `json:"domain"`
}

type Aliyun struct {
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Domain          string `json:"domain"`
}

type Qcloud struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Domain    string `json:"domain"`
}

type Google struct {
	CredentialsFile  string `json:"credentials_file"`
	Bucket           string `json:"bucket"`
	UploadsCatalogue string `json:"uploads_catalogue"`
	Domain           string `json:"domain"`
}

type Engine struct {
	Local  []interface{} `json:"local"`
	Qiniu  Qiniu         `json:"qiniu"`
	Aliyun Aliyun        `json:"aliyun"`
	Qcloud Qcloud        `json:"qcloud"`
	Google Google        `json:"google"`
}
