package resp

type License struct {
	AppId             int    `json:"app_id"`
	CN                string `json:"c_n"`
	Name              string `json:"name"`
	Sale              int    `json:"sale"`
	Reserve           int    `json:"reserve"`
	CL                int    `json:"c_l"`
	KL                int    `json:"k_l"`
	TL                int    `json:"t_l"`
	AL                int    `json:"a_l"`
	ZL                int    `json:"z_l"`
	PL                int    `json:"p_l"`
	CSet              int    `json:"c_set"`
	SType             int    `json:"s_type"`
	Level             int    `json:"level"`
	Phone             string `json:"phone"`
	Addr              string `json:"addr"`
	Dm                int    `json:"dm"`
	Domain            string `json:"domain"`
	Day               int    `json:"day"`
	CTime             int    `json:"c_time"`
	Mac               string `json:"_mac"`
	UUID              string `json:"_uuid"`
	Ctime             int    `json:"_ctime"`
	Logo              string `json:"logo"`
	IsOpenMember      int    `json:"is_open_member"`
	IsOpenTablet      int    `json:"is_open_tablet"`
	IsOpenScan        int    `json:"is_open_scan"`
	IsOpenAssistant   int    `json:"is_open_assistant"`
	IsOpenKitchenKds  int    `json:"is_open_kitchen_kds"`
	IsOpenBuffet      int    `json:"is_open_buffet"`
	IsAcceptScanOrder int    `json:"is_accept_scan_order"`
	IsOpenLocalPrint  int    `json:"is_open_local_print"`
}

type Permission struct {
	AccessId       int           `json:"access_id"`
	Name           string        `json:"name"`
	Path           string        `json:"path"`
	APIPath        string        `json:"api_path"`
	ParentId       int           `json:"parent_id"`
	Sort           int           `json:"sort"`
	Icon           string        `json:"icon"`
	RedirectName   string        `json:"redirect_name"`
	IsRoute        int           `json:"is_route"`
	IsMenu         int           `json:"is_menu"`
	Alias          string        `json:"alias"`
	IsShow         int           `json:"is_show"`
	PlusCategoryId int           `json:"plus_category_id"`
	Remark         string        `json:"remark"`
	IsSupplier     int           `json:"is_supplier"`
	AppId          int           `json:"app_id"`
	CreateTime     string        `json:"create_time"`
	UpdateTime     string        `json:"update_time"`
	Children       []*Permission `json:"children"`
}

type User struct {
	ShopUserId     int          `json:"shop_user_id"`
	CashierId      int          `json:"cashier_id"`
	UserName       string       `json:"user_name"`
	RealName       string       `json:"real_name"`
	Account        string       `json:"account"`
	Mobile         interface{}  `json:"mobile"`
	IsSuper        int          `json:"is_super"`
	ShopSupplierId int          `json:"shop_supplier_id"`
	Name           string       `json:"name"`
	TimeZone       string       `json:"time_zone"`
	AppId          int          `json:"app_id"`
	IsFirstLogin   int          `json:"is_first_login"`
	Permission     []Permission `json:"permission"`
}

type App struct {
	ExpireTimeText string `json:"expire_time_text"`
	AppId          int    `json:"app_id"`
	AppName        string `json:"app_name"`
	Logo           int    `json:"logo"`
	IsRecycle      int    `json:"is_recycle"`
	IsChain        int    `json:"is_chain"`
	ExpireTime     int    `json:"expire_time"`
	AuthDay        int    `json:"auth_day"`
	AuthStartTime  string `json:"auth_start_time"`
	Status         int    `json:"status"`
	IsDelete       int    `json:"is_delete"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
}

type Vices struct {
	ViceUnit         string `json:"vice_unit"`
	ViceUnitPosition string `json:"vice_unit_position"`
	UnitRate         string `json:"unit_rate"`
}

type Currency struct {
	Unit         string `json:"unit"`
	IsOpen       string `json:"is_open"`
	UnitPosition string `json:"unit_position"`
	Vices        Vices  `json:"vices"`
}

type OrderMethod struct {
	IsCashierOrder string `json:"is_cashier_order"`
	IsTableOrder   string `json:"is_table_order"`
}

type Cashier struct {
	Carousel               []interface{}  `json:"carousel"`
	IsAutoSend             string         `json:"is_auto_send"`
	OrderMethod            OrderMethod    `json:"order_method"`
	Server                 Server         `json:"server"`
	IsRemainColor          string         `json:"is_remain_color"`
	RemainColor            []string       `json:"remain_color"`
	IsOpenCashierPassword  string         `json:"is_open_cashier_password"`
	LockPassword           string         `json:"lock_password"`
	IsAutoLockScreen       string         `json:"is_auto_lock_screen"`
	AutoLockScreen         string         `json:"auto_lock_screen"`
	IsShowScanSoldOut      string         `json:"is_show_scan_sold_out"`
	IsShowAssistantSoldOut string         `json:"is_show_assistant_sold_out"`
	LanguageList           []LanguageItem `json:"language_list"`
	DefaultLanguage        string         `json:"default_language"`
	IsAutoOrder            string         `json:"is_auto_order"`
	AutoOrderLimit         string         `json:"auto_order_limit"`
	IsAutoVoice            string         `json:"is_auto_voice"`
	MenuShowSoldOut        string         `json:"menu_show_sold_out"`
	KitchenLanguage        string         `json:"kitchen_language"`
	Language               []string       `json:"language"`
}

type Tablet struct {
	IsShowSoldOut string `json:"is_show_sold_out"`
}

type Buffet struct {
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

type Payment struct {
	IsCash    string `json:"is_cash"`
	IsBalance string `json:"is_balance"`
	IsOther   string `json:"is_other"`
}

type Base struct {
	BrandName          string `json:"brand_name"`
	BrandLogo          string `json:"brand_logo"`
	BrandLogoLong      string `json:"brand_logo_long"`
	BrowserLogo        string `json:"browser_logo"`
	BrowserTitle       string `json:"browser_title"`
	ExpirationReminder int    `json:"expiration_reminder"`
}

type CloudBasic struct {
	Base Base `json:"base"`
}

type Business struct {
	CheckoutZeroingMethodList []CheckoutZeroingMethodItem `json:"checkout_zeroing_method_list"`
	ZeroingMethod             string                      `json:"zeroing_method"`
	CheckoutZeroingMethod     string                      `json:"checkout_zeroing_method"`
	GiftMethod                string                      `json:"gift_method"`
	FreeMethod                string                      `json:"free_method"`
	DiscountMethod            string                      `json:"discount_method"`
	QrCode                    string                      `json:"qr_code"`
	NoClearTable              string                      `json:"no_clear_table"`
	IsNeedPassword            string                      `json:"is_need_password"`
	DishCardStyle             string                      `json:"dish_card_style"`
	DishCardStyleTime         string                      `json:"dish_card_style_time"`
	IsInvoice                 string                      `json:"is_invoice"`
}

type BaseInfo struct {
	DeviceId             string     `json:"device_id"`
	License              License    `json:"license"`
	User                 User       `json:"user"`
	App                  App        `json:"app"`
	Currency             Currency   `json:"currency"`
	Cashier              Cashier    `json:"cashier"`
	Tablet               Tablet     `json:"tablet"`
	Buffet               Buffet     `json:"buffet"`
	Payment              Payment    `json:"payment"`
	DeviceRemark         string     `json:"device_remark"`
	LicenseRemainingDays int        `json:"license_remaining_days"`
	CloudBasic           CloudBasic `json:"cloud_basic"`
	Business             Business   `json:"business"`
	NoClearTable         string     `json:"no_clear_table"`
	PrintLangSelect      bool       `json:"print_lang_select"`
}

type CashierLoginResponse2 struct {
	CashierId      int         `json:"cashier_id"`
	UserName       string      `json:"user_name"`
	Account        string      `json:"account"`
	Mobile         interface{} `json:"mobile"`
	ShopSupplierId int         `json:"shop_supplier_id"`
	Name           string      `json:"name"`
	AppId          int         `json:"app_id"`
	Token          string      `json:"token"`
	BaseInfo       BaseInfo    `json:"base_info"`
}

type CashierLoginResponse struct {
	Token string `json:"token"`
}
