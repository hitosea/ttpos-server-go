package entity

import (
	"ttpos-server-go/app/modules/setting/domain/valueobject"
)

// LanguageResp 语言响应
type LanguageResp struct {
	Languages       []string                   `json:"languages"`
	LanguageList    []valueobject.LanguageItem `json:"language_list"`
	DefaultLanguage string                     `json:"default_language"`
}

// Ads 广告响应
type Ads struct {
	List []valueobject.CarouselItem `json:"list"`
}

// CashierBaseSetting 收银机基础设置
type CashierBaseSetting struct {
	AcceptOrder       AcceptOrderSetting       `json:"accept_order"`
	AcceptMemberOrder AcceptMemberOrderSetting `json:"accept_member_order"`
	System            SystemSetting            `json:"system"`
	UsbPrinter        UsbPrinterList           `json:"usb_printer"`
}

// AcceptOrderSetting 接单设置
type AcceptOrderSetting struct {
	IsAutoOrder    string `json:"is_auto_order"`
	AutoOrderLimit string `json:"auto_order_limit"`
	IsAutoVoice    string `json:"is_auto_voice"`
}

// AcceptMemberOrderSetting 会员接单设置
type AcceptMemberOrderSetting struct {
	IsAutoMemberOrder      string `json:"is_auto_member_order"`
	AutoMemberOrderLimit   string `json:"auto_member_order_limit"`
	IsAutoVoiceMemberOrder string `json:"is_auto_voice_member_order"`
}

// SystemSetting 系统设置
type SystemSetting struct {
	IsShowScanSoldOut      int    `json:"is_show_scan_sold_out"`
	IsShowAssistantSoldOut int    `json:"is_show_assistant_sold_out"`
	MenuShowSoldOut        string `json:"menu_show_sold_out"`
	MemberShowSoldOut      string `json:"member_show_sold_out"`
	DishCardStyle          string `json:"dish_card_style"`
	IsShowSoldOut          int    `json:"is_show_sold_out"`
	DefaultLanguage        string `json:"default_language"`
	SecondLanguage         string `json:"second_language"`
	DeviceId               string `json:"device_id"`
	DeviceRemark           string `json:"device_remark"`
	ClientVersion          string `json:"client_version"`
	ServerVersion          string `json:"server_version"`
}

// UsbPrinterList USB打印机列表
type UsbPrinterList struct {
	List       []UsbPrinter `json:"list"`
	SelectedSn string       `json:"selected_sn"`
}

// UsbPrinter USB打印机
type UsbPrinter struct {
	Sn   string `json:"sn"`
	Name string `json:"name"`
}

// BuffetResp 自助餐响应
// Buffet 自助餐设置
type Buffet struct {
	IsOpen                   string         `json:"is_open"`                     // 是否开启自助餐 0-关闭 1-开启
	TabletEndTime            string         `json:"tablet_end_time"`             // 平板结束时间提醒（分）
	IsRemainContinue         string         `json:"is_remain_continue"`          // 剩余xx分不可继续点餐开关 0-关闭 1-开启
	RemainContinueTime       string         `json:"remain_continue_time"`        // 剩余xx分不可继续点餐
	RemainContinueNoticeTime string         `json:"remain_continue_notice_time"` // 剩余xx分提醒不可继续点餐
	IsBuyContinue            string         `json:"is_buy_continue"`             // 非自助餐商品到时是否能继续选购 0-关闭 1-开启
	IsAddClock               string         `json:"is_add_clock"`                // 是否开启加钟 0-关闭 1-开启
	IsBuffetDiscount         string         `json:"is_buffet_discount"`          // 是否开启自助餐优惠折扣 0-关闭 1-开启
	AddClock                 []AddClockItem `json:"add_clock"`                   // 名称 - 加钟时间（分）- 价格
}

type BuffetResp struct {
	IsOpen                   string         `json:"is_open"`
	TabletEndTime            int            `json:"tablet_end_time"`
	IsRemainContinue         string         `json:"is_remain_continue"`
	RemainContinueTime       string         `json:"remain_continue_time"`
	RemainContinueNoticeTime string         `json:"remain_continue_notice_time"`
	IsBuyContinue            string         `json:"is_buy_continue"`
	IsAddClock               string         `json:"is_add_clock"`
	IsBuffetDiscount         string         `json:"is_buffet_discount"`
	AddClock                 []AddClockItem `json:"add_clock"`
}

// AddClockItem 加钟项
type AddClockItem struct {
	Time  string `json:"time"`
	Price string `json:"price"`
}

// TabletSetting 平板设置
// TabletSetting 平板端设置（字段顺序与旧服务保持一致）
type TabletSetting struct {
	Carousel           []valueobject.CarouselItem `json:"carousel"`              // 上传后的轮播内容url（图片 + 视频）
	IsCallService      string                     `json:"is_call_service"`       // 是否开启呼叫服务员 0-关闭 1-开启
	IsCustomerOrder    string                     `json:"is_customer_order"`     // 是否开启顾客自助下单 0-关闭 1-开启
	IsVoiceRemind      string                     `json:"is_voice_remind"`       // 是否开启声音提醒 0-关闭 1-开启
	IsShowSoldOut      string                     `json:"is_show_sold_out"`      // 是否显示售罄商品（与旧服务保持一致）
	IsBuffetOrderLimit string                     `json:"is_buffet_order_limit"` // 是否开启自助餐下单限制 0-关闭 1-开启
	BuffetOrderLimit   BuffetOrderLimit           `json:"buffet_order_limit"`    // 自助餐下单限制
	IsOrderLimit       string                     `json:"is_order_limit"`        // 是否开启非自助餐下单限制 0-关闭 1-开启
	OrderLimit         OrderLimit                 `json:"order_limit"`           // 非自助餐下单限制
	Server             Server                     `json:"server"`                // 平板服务器连接
	LanguageList       []valueobject.LanguageItem `json:"language_list"`         // 语言列表
	Language           []string                   `json:"language"`              // 常用语言
	DefaultLanguage    string                     `json:"default_language"`      // 默认语言
	AdvancedPassword   string                     `json:"advanced_password"`     // 高级设置密码
}

// PaymentSetting 支付设置
// PaymentSetting 支付设置（字段顺序与旧服务保持一致，多余字段使用 omitempty）
type PaymentSetting struct {
	IsCash     string `json:"is_cash"`
	IsBalance  string `json:"is_balance"`
	IsOther    string `json:"is_other"`
	IsBankCard string `json:"is_bank_card,omitempty"`
	IsWechat   string `json:"is_wechat,omitempty"`
	IsAlipay   string `json:"is_alipay,omitempty"`
	IsUnionPay string `json:"is_union_pay,omitempty"`
}

// CurrencySetting 货币设置（与旧服务字段顺序保持一致）
type CurrencySetting struct {
	Unit             string `json:"unit"`               // 货币单位，默认泰铢
	PrintUnit        string `json:"print_unit"`         // 货币单位 - 打印专用，默认泰铢
	UnitPosition     string `json:"unit_position"`      // 主货币显示位置 0-金额前 1-金额后
	IsOpen           string `json:"is_open"`            // 副货币单位开关
	ViceUnit         string `json:"vice_unit"`          // 副货币单位
	ViceUnitPosition string `json:"vice_unit_position"` // 副货币显示位置 0-金额前 1-金额后
	UnitRate         string `json:"unit_rate"`          // 单位汇率
}

// KitchenSetting 厨显设置
type KitchenSetting struct {
	IsOpen              string                     `json:"is_open"`                // 是否开启厨显功能 0关闭 1开启
	IsComeDish          string                     `json:"is_come_dish"`           // 是否开启来菜提醒 0-关闭 1-开启
	IsCallService       string                     `json:"is_call_service"`        // 是否开启顾客呼叫提醒 0-关闭 1-开启
	Server              Server                     `json:"server"`                 // 厨显服务器连接
	IsWaitColor         string                     `json:"is_wait_color"`          // 是否开启等待时长颜色 0-关闭 1-开启
	WaitColor           []string                   `json:"wait_color"`             // 时长颜色（旧格式：["red", "yellow"]）
	WaitTimeColorRanges []WaitTimeColorRange       `json:"wait_time_color_ranges"` // 等待时长颜色区间配置（新格式）
	LanguageList        []valueobject.LanguageItem `json:"language_list"`          // 语言列表
	Language            []string                   `json:"language"`               // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage     string                     `json:"default_language"`       // 默认语言
	IsSmartKitchen      string                     `json:"is_smart_kitchen"`       // 是否开启智能后厨 0-关闭 1-开启
	AdvancedPassword    string                     `json:"advanced_password"`      // 高级设置密码
}

// WaitTimeColorRange 等待时间颜色范围
type WaitTimeColorRange struct {
	Minute string `json:"minute"`
	Color  string `json:"color"`
}

// H5Setting H5设置
type H5Setting struct {
	IsCallService      string                     `json:"is_call_service"`       // 是否开启呼叫服务员 0-关闭 1-开启
	IsCustomerOrder    string                     `json:"is_customer_order"`     // 是否开启顾客自助下单 0-关闭 1-开启
	IsVoiceRemind      string                     `json:"is_voice_remind"`       // 是否开启声音提醒 0-关闭 1-开启
	IsShowScanSoldOut  int                        `json:"is_show_scan_sold_out"` // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsBuffetOrderLimit string                     `json:"is_buffet_order_limit"` // 自助餐下单限制 是否开启自助餐下单限制 0-关闭 1-开启
	BuffetOrderLimit   BuffetOrderLimit           `json:"buffet_order_limit"`    // 自助餐下单限制配置
	IsOrderLimit       string                     `json:"is_order_limit"`        // 非自助餐下单限制 是否开启非自助餐下单限制 0-关闭 1-开启
	OrderLimit         OrderLimit                 `json:"order_limit"`           // 非自助餐下单限制配置
	LanguageList       []valueobject.LanguageItem `json:"language_list"`         // 语言列表
	Language           []string                   `json:"language"`              // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage    string                     `json:"default_language"`      // 默认语言
}

// KioskSetting 自助点餐机设置
// KioskSetting 自助点餐机设置（字段顺序与旧服务保持一致）
type KioskSetting struct {
	AdvancedPassword  string                     `json:"advanced_password"`
	CallWaiterEnabled int                        `json:"call_waiter_enabled"`
	LanguageList      []valueobject.LanguageItem `json:"language_list"`
	Language          []string                   `json:"language"`
	DefaultLanguage   string                     `json:"default_language"`
	Carousel          []valueobject.CarouselItem `json:"carousel"`
}

// AssistantSetting 点餐助手设置（与旧服务字段顺序保持一致）
type AssistantSetting struct {
	Server                 Server                     `json:"server"`                     // 服务器连接
	IsAutoSend             string                     `json:"is_auto_send"`               // 是否自动发送
	IsRemainColor          string                     `json:"is_remain_color"`            // 是否开启剩余时长颜色 0-关闭 1-开启
	RemainColor            []string                   `json:"remain_color"`               // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
	DefaultMode            string                     `json:"default_mode"`               // 默认模式 0-服务员模式 1-顾客模式
	IsAutoLockScreen       string                     `json:"is_auto_lock_screen"`        // 是否开启自动锁屏 0-关闭 1-开启
	IsCheckOrder           string                     `json:"is_check_order"`             // 下单校验高级密码 0-关闭 1-开启
	AutoLockScreen         string                     `json:"auto_lock_screen"`           // 自动锁屏（秒），默认5分钟
	LanguageList           []valueobject.LanguageItem `json:"language_list"`              // 语言列表
	Language               []string                   `json:"language"`                   // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
	DefaultLanguage        string                     `json:"default_language"`           // 默认语言
	IsShowAssistantSoldOut int                        `json:"is_show_assistant_sold_out"` // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	AdvancedPassword       string                     `json:"advanced_password"`          // 高级设置密码
	LockPassword           string                     `json:"lock_password"`              // 锁屏密码
}

// PointsSetting 积分设置（与旧服务字段顺序保持一致）
type PointsSetting struct {
	DeductionOrder     string             `json:"deduction_order"`      // 扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
	DeductRatioMain    string             `json:"deduct_ratio_main"`    // 主账户扣款比例0-100
	DeductRatioGift    string             `json:"deduct_ratio_gift"`    // 赠送账户扣款比例0-100
	PointsName         string             `json:"points_name"`          // 积分名称自定义
	IsShoppingGift     string             `json:"is_shopping_gift"`     // 是否开启购物送积分
	GiftRatio          string             `json:"gift_ratio"`           // 积分赠送比例
	IsShoppingDiscount string             `json:"is_shopping_discount"` // 是否开启积分抵现
	Discount           DiscountItem       `json:"discount"`             // 抵现设置
	Describe           string             `json:"describe"`             // 积分规则说明
	DeductOrder        string             `json:"deduct_order"`         // 扣款顺序（与 deduction_order 相同，兼容旧服务）
	ShoppingGiftRules  []ShoppingGiftRule `json:"shopping_gift_rules"`  // 购物赠礼规则
	Exchange           PointsExchange     `json:"exchange"`             // 积分兑换设置
}

// PointsExchange 积分兑换设置
type PointsExchange struct {
	OpenPointsExchange string `json:"open_points_exchange"` // 是否开启积分兑换
	PointsExchangeRate string `json:"points_exchange_rate"` // 积分兑换比例
	AutoPointsExchange string `json:"auto_points_exchange"` // 是否自动兑换
}

// ShoppingGiftRule 购物赠礼规则（与旧服务字段保持一致）
type ShoppingGiftRule struct {
	Type                     string            `json:"type"`                       // 类型
	IsOpen                   string            `json:"is_open"`                    // 是否开启
	IsMemberLevelRelated     string            `json:"is_member_level_related"`    // 是否按会员等级
	Value                    string            `json:"value"`                      // 值
	PaymentAmountRequirement string            `json:"payment_amount_requirement"` // 付款金额要求
	MealType                 []string          `json:"meal_type"`                  // 就餐类型
	BalancePaymentGetPoints  string            `json:"balance_payment_get_points"` // 余额支付获取积分
	RefundReturnPoints       string            `json:"refund_return_points"`       // 退款返还积分
	MemberLevels             []MemberLevelItem `json:"member_levels"`              // 会员等级列表
}

// MemberLevelItem 会员等级项
type MemberLevelItem struct {
	Uuid  uint64 `json:"uuid"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DiscountItem 抵现项
type DiscountItem struct {
	DiscountRatio  string `json:"discount_ratio"`   // 抵现比例
	FullOrderPrice string `json:"full_order_price"` // 满单金额
	MaxMoneyRatio  string `json:"max_money_ratio"`  // 最大抵现金额比例
}

// OrderMethod 用餐方式
type OrderMethod struct {
	IsCashierOrder string `json:"is_cashier_order"` // 收银用餐 0-关闭 1-开启
	IsTableOrder   string `json:"is_table_order"`   // 桌台用餐 0-关闭 1-开启
}

// Server 服务器连接
type Server struct {
	IP   string `json:"ip"`   // IP地址
	Port string `json:"port"` // 端口号
}

// CloudBasicSetting 云端基础设置（字段顺序与旧服务保持一致）
type CloudBasicSetting struct {
	BrandLogo          string `json:"brand_logo"`          // 品牌LOGO： (正方型)
	BrandLogoLong      string `json:"brand_logo_long"`     // 品牌LOGO： (长方型)
	BrandName          string `json:"brand_name"`          // 品牌名称
	BrowserLogo        string `json:"browser_logo"`        // 浏览器title： （LOGO）
	BrowserTitle       string `json:"browser_title"`       // 浏览器title： (品牌名称)
	ExpirationReminder int    `json:"expiration_reminder"` // 准备到期提醒，根据填写的时间提醒商家剩余可用时间
}

// ServiceCharge 服务费
type ServiceCharge struct {
	IsOpen              string  `json:"is_open"`                // 是否开启 0关闭 1开启
	ChargeType          string  `json:"charge_type"`            // 服务费类型 1-固定金额 2-百分比
	ServiceCharge       string  `json:"service_charge"`         // 服务费金额
	ServiceChargeRate   string  `json:"service_charge_rate"`    // 服务费率
	IsOpenTax           string  `json:"is_open_tax"`            // 税费 1-收取税费 0-不收取税费
	ApplyScope          string  `json:"apply_scope"`            // 适用范围 1-全部 2-部分
	ApplyScopeOrdering  string  `json:"apply_scope_ordering"`   // 适用范围-点餐 0-关闭 1-开启
	ApplyScopeTable     string  `json:"apply_scope_table"`      // 适用范围-桌台 0-关闭 1-开启
	ApplyScopeTableList []int64 `json:"apply_scope_table_list"` // 适用范围-桌台id列表
	ServiceFeeBase      string  `json:"service_fee_base"`       // 服务费计算基准 0-商品惠后价 1-商品价格合计
}

// TaxRate 税率
type TaxRate struct {
	IsOpen         string               `json:"is_open"`          // 是否开启 0关闭 1开启
	CalcType       string               `json:"calc_type"`        // 计算类型 商品已含税价-1 商品未含税价-2
	AddTaxCategory []AddTaxCategoryItem `json:"add_tax_category"` // 增值税分类
}

// AddTaxCategoryItem 附加税类别项
type AddTaxCategoryItem struct {
	ID      int    `json:"id"`       // 税类id
	Name    string `json:"name"`     // 名称
	TaxRate string `json:"tax_rate"` // 税率
	Action  string `json:"action"`   // 操作结果 delete-删除 edit-编辑 add-新增
}

// UpdateStoreSetting 更新门店设置请求
type UpdateStoreSetting struct {
	Name        string                     `json:"name"`
	CompanyName string                     `json:"company_name"`
	Address     string                     `json:"address"`
	Phone       string                     `json:"phone"`
	StoreCode   string                     `json:"store_code"`
	TimeZone    string                     `json:"time_zone"`
	Coordinates string                     `json:"coordinates"`
	LogoUrl     string                     `json:"logo_url"`
	AvatarUrl   string                     `json:"avatar_url"`
	Language    []valueobject.LanguageItem `json:"language"`
}

// UpdateBusinessSetting 更新业务设置请求
type UpdateBusinessSetting struct {
	DishCardStyle              string   `json:"dish_card_style"`
	DeliveryPriceRatio         string   `json:"delivery_price_ratio"`
	DiscountAuthorizedStaffIds []uint64 `json:"discount_authorized_staff_ids"`
	RefundAuthorizedStaffIds   []uint64 `json:"refund_authorized_staff_ids"`
}

// SaveCashierSettingReq 保存收银机设置请求
type SaveCashierSettingReq struct {
	Carousel                []valueobject.CarouselItem `json:"carousel"`
	NoOrderCarouselInterval string                     `json:"no_order_carousel_interval"`
	OrderDisplayMode        string                     `json:"order_display_mode"`
	OrderCarousel           []valueobject.CarouselItem `json:"order_carousel"`
	OrderCarouselInterval   string                     `json:"order_carousel_interval"`
}

// SaveKioskSettingReq 保存自助点餐机设置请求
type SaveKioskSettingReq struct {
	AdvancedPassword  string                     `json:"advanced_password"`
	CallWaiterEnabled int                        `json:"call_waiter_enabled"`
	Language          []string                   `json:"language"`
	DefaultLanguage   string                     `json:"default_language"`
	Carousel          []valueobject.CarouselItem `json:"carousel"`
}

// SaveKitchenSettingReq 保存厨显设置请求
type SaveKitchenSettingReq struct {
	IsWaitColor         string               `json:"is_wait_color"`
	WaitTimeColorRanges []WaitTimeColorRange `json:"wait_time_color_ranges"`
}

// UpdateSystemSetting 更新系统设置请求
type UpdateSystemSetting struct {
	IsShowAssistantSoldOut *int   `json:"is_show_assistant_sold_out"`
	IsShowScanSoldOut      *int   `json:"is_show_scan_sold_out"`
	MenuShowSoldOut        *int   `json:"menu_show_sold_out"`
	IsShowSoldOut          *int   `json:"is_show_sold_out"`
	DishCardStyle          string `json:"dish_card_style"`
	DeviceRemark           string `json:"device_remark"`
}

// UpdateAcceptOrderSetting 更新接单设置请求
type UpdateAcceptOrderSetting struct {
	IsAutoOrder    string `json:"is_auto_order"`
	AutoOrderLimit string `json:"auto_order_limit"`
}

// UpdateAcceptMemberOrderSetting 更新会员接单设置请求
type UpdateAcceptMemberOrderSetting struct {
	IsAutoMemberOrder    string `json:"is_auto_member_order"`
	AutoMemberOrderLimit string `json:"auto_member_order_limit"`
}

// ShopBusinessSetting 商店业务设置响应
// ShopBusinessSetting 商户业务设置（使用嵌入式结构体，与旧服务 JSON 结构保持一致）
type ShopBusinessSetting struct {
	BusinessSetting                                 // 嵌入式，JSON 字段会被平铺
	FreeReasonCount                          int    `json:"free_reason_count"`
	ReturnFoodReasonCount                    int    `json:"return_food_reason_count"`
	OrderRemarkCount                         int    `json:"order_remark_count"`
	OrderItemRemarkCount                     int    `json:"order_item_remark_count"`
	HeadquarterRequiredParentCompanyApproval string `json:"headquarter_required_parent_company_approval"`
	HeadquarterViaParentCompanyWarehouse     string `json:"headquarter_via_parent_company_warehouse"`
	OrderSourceCount                         int    `json:"order_source_count"`
	NationalityCount                         int    `json:"nationality_count"`
}

// PaymentMethodListResp 支付方式列表响应
type PaymentMethodListResp struct {
	List []PaymentMethod `json:"list"`
}

// PaymentMethod 支付方式
type PaymentMethod struct {
	Uuid        uint64 `json:"uuid"`
	Name        string `json:"name"`
	PaymentName string `json:"payment_name"`
}

// DataManageSetting 数据管理设置
type DataManageSetting struct {
	IsEnableDataManage bool `json:"is_enable_data_manage"`
}

// CompanySetting 公司设置
type CompanySetting struct {
	IsOpenMember       int
	IsOpenBuffet       int
	IsHeadquarter      int
	HeadquarterUuid    uint64
	ErpnextSiteCode    string
	ErpnextCompanyAbbr string
	ErpnextBranchName  string
	DeliveryStatus     int
	TimeZone           string
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	VersionName  string `json:"version_name"`
	ForcedUpdate int    `json:"forced_update"`
	UpdateLog    string `json:"update_log"`
	DownloadURL  string `json:"download_url"`
}

// DefaultH5Setting 获取默认H5设置
func DefaultH5Setting(languageList []valueobject.LanguageItem) *H5Setting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	return &H5Setting{
		IsCallService:      "1", // 是否开启呼叫服务员 0-关闭 1-开启
		IsCustomerOrder:    "1", // 是否开启顾客自助下单 0-关闭 1-开启
		IsVoiceRemind:      "1", // 是否开启声音提醒 0-关闭 1-开启
		IsShowScanSoldOut:  0,   // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
		IsBuffetOrderLimit: "0", // 自助餐下单限制 是否开启自助餐下单限制 0-关闭 1-开启
		BuffetOrderLimit: BuffetOrderLimit{
			IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
			LimitTime:   "0", // 时间限制（分钟）
			IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
			LimitNum:    "0", // 数量限制
		},
		IsOrderLimit: "0", // 非自助餐下单限制 是否开启非自助餐下单限制 0-关闭 1-开启
		OrderLimit: OrderLimit{
			IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
			LimitTime:   "0", // 时间限制（分钟）
			IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
			LimitNum:    "0", // 数量限制
		},
		LanguageList:    languageList,              // 语言列表
		Language:        []string{defaultLanguage}, // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
		DefaultLanguage: defaultLanguage,           // 默认语言
	}
}

// DefaultKioskSetting 获取默认自助点餐机设置
func DefaultKioskSetting(languageList []valueobject.LanguageItem) *KioskSetting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	// 默认勾选所有语言
	commonLanguages := make([]string, 0, len(languageList))
	for _, lang := range languageList {
		commonLanguages = append(commonLanguages, lang.Name)
	}

	return &KioskSetting{
		AdvancedPassword:  "666888",                            // 高级密码（默认666888）
		CallWaiterEnabled: 1,                                   // 呼叫服务员开关（默认1-开启）
		LanguageList:      languageList,                        // 语言列表
		Language:          commonLanguages,                     // 已设置的语言（默认所有语言）
		DefaultLanguage:   defaultLanguage,                     // 默认语言（默认语言1）
		Carousel:          make([]valueobject.CarouselItem, 0), // 轮播内容（默认空数组）
	}
}

// DefaultAssistantSetting 获取默认点餐助手设置（与旧服务保持一致）
func DefaultAssistantSetting(languageList []valueobject.LanguageItem, serverIP, serverPort string) *AssistantSetting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	return &AssistantSetting{
		Server:                 Server{IP: serverIP, Port: serverPort}, // 服务器连接
		IsAutoSend:             "",                                     // 与旧服务保持一致
		IsRemainColor:          "1",                                    // 是否开启剩余时长颜色 0-关闭 1-开启
		RemainColor:            []string{"#E50028", "#F2A000"},         // 剩余时长颜色 10分钟-红色(#E50028) 20分钟-黄色(#F2A000)
		DefaultMode:            "0",                                    // 默认模式 0-服务员模式 1-顾客模式
		IsAutoLockScreen:       "1",                                    // 是否开启自动锁屏 0-关闭 1-开启
		IsCheckOrder:           "0",                                    // 下单校验高级密码 0-关闭 1-开启
		AutoLockScreen:         "300",                                  // 自动锁屏（秒），默认5分钟
		LanguageList:           languageList,                           // 语言列表
		Language:               []string{defaultLanguage},              // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
		DefaultLanguage:        defaultLanguage,                        // 默认语言
		IsShowAssistantSoldOut: 1,                                      // 是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
		AdvancedPassword:       "666888",                               // 高级设置密码
		LockPassword:           "666888",                               // 锁屏密码
	}
}

// DefaultTabletSetting 获取默认平板端设置
func DefaultTabletSetting(languageList []valueobject.LanguageItem) *TabletSetting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	return &TabletSetting{
		Carousel:           make([]valueobject.CarouselItem, 0), // 上传后的轮播内容url（图片 + 视频）
		IsCallService:      "1",                                 // 是否开启呼叫服务员 0-关闭 1-开启
		IsCustomerOrder:    "1",                                 // 是否开启顾客自助下单 0-关闭 1-开启
		IsVoiceRemind:      "1",                                 // 是否开启声音提醒 0-关闭 1-开启
		IsBuffetOrderLimit: "0",                                 // 是否开启自助餐下单限制 0-关闭 1-开启
		BuffetOrderLimit: BuffetOrderLimit{
			IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
			LimitTime:   "0", // 时间限制（分钟）
			IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
			LimitNum:    "0", // 数量限制
		},
		IsOrderLimit: "0", // 是否开启非自助餐下单限制 0-关闭 1-开启
		OrderLimit: OrderLimit{
			IsLimitTime: "1", // 是否开启时间限制 0-关闭 1-开启
			LimitTime:   "0", // 时间限制（分钟）
			IsLimitNum:  "1", // 是否开启数量限制 0-关闭 1-开启
			LimitNum:    "0", // 数量限制
		},
		Server: Server{
			IP:   "127.0.0.1", // 默认IP
			Port: "8080",      // 默认端口
		}, // 平板服务器连接
		LanguageList:     languageList,              // 语言列表
		Language:         []string{defaultLanguage}, // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
		DefaultLanguage:  defaultLanguage,           // 默认语言
		AdvancedPassword: "666888",                  // 高级设置密码
	}
}

// DefaultKitchenSetting 获取默认厨显端设置
func DefaultKitchenSetting(languageList []valueobject.LanguageItem) *KitchenSetting {
	var defaultLanguage = "en"
	if len(languageList) > 0 {
		defaultLanguage = languageList[0].Name
	}

	return &KitchenSetting{
		IsOpen:        "1", // 是否开启厨显功能 0关闭 1开启
		IsComeDish:    "1", // 是否开启来菜提醒 0-关闭 1-开启
		IsCallService: "1", // 是否开启顾客呼叫提醒 0-关闭 1-开启
		Server: Server{
			IP:   "127.0.0.1", // 默认IP
			Port: "8080",      // 默认端口
		}, // 厨显服务器连接
		IsWaitColor: "0",               // 是否开启等待时长颜色 0-关闭 1-开启
		WaitColor:   make([]string, 0), // 时长颜色（旧格式：["red", "yellow"]）
		WaitTimeColorRanges: []WaitTimeColorRange{ // 等待时长颜色区间配置（新格式）
			{Minute: "0", Color: "#100A05"},  // 第一区间：0分钟（黑色）
			{Minute: "10", Color: "#FFBE00"}, // 第二区间：10分钟（黄色）
			{Minute: "20", Color: "#E50028"}, // 第三区间：20分钟（红色）
		},
		LanguageList:     languageList,              // 语言列表
		Language:         []string{defaultLanguage}, // 常用语言 泰语、英语、中文、繁体 'th', 'en', 'zh', 'zhtw'
		DefaultLanguage:  defaultLanguage,           // 默认语言
		IsSmartKitchen:   "0",                       // 是否开启智能后厨 0-关闭 1-开启
		AdvancedPassword: "666888",                  // 高级设置密码
	}
}

// DefaultPaymentSetting 获取默认支付方式设置
func DefaultPaymentSetting() *PaymentSetting {
	return &PaymentSetting{
		IsCash:    "0", // 是否开启现金支付 0-关闭 1-开启
		IsBalance: "0", // 是否开启余额支付 0-关闭 1-开启
		IsOther:   "1", // 是否开启其他方式支付 0-关闭 1-开启
	}
}

// DefaultCurrencySetting 获取默认货币单位设置
func DefaultCurrencySetting() *CurrencySetting {
	return &CurrencySetting{
		Unit:             "฿", // 货币单位，默认泰铢
		PrintUnit:        "฿", // 货币单位 - 打印专用，默认泰铢
		UnitPosition:     "0", // 主货币显示位置 0-金额前 1-金额后
		IsOpen:           "0", // 副货币单位开关
		ViceUnit:         "",  // 副货币单位（与旧服务保持一致）
		ViceUnitPosition: "0", // 副货币显示位置 0-金额前 1-金额后
		UnitRate:         "",  // 单位汇率（与旧服务保持一致）
	}
}

// DefaultBuffetSetting 获取默认自助餐设置
func DefaultBuffetSetting() *Buffet {
	return &Buffet{
		IsOpen:                   "1",                     // 是否开启自助餐 0-关闭 1-开启
		TabletEndTime:            "5",                     // 平板结束时间提醒（分）（与旧服务保持一致）
		IsRemainContinue:         "0",                     // 剩余xx分不可继续点餐开关 0-关闭 1-开启（与旧服务保持一致）
		RemainContinueTime:       "10",                    // 剩余xx分不可继续点餐
		RemainContinueNoticeTime: "10",                    // 剩余xx分提醒不可继续点餐（与旧服务保持一致）
		IsBuyContinue:            "0",                     // 非自助餐商品到时是否能继续选购 0-关闭 1-开启
		IsAddClock:               "0",                     // 是否开启加钟 0-关闭 1-开启（与旧服务保持一致）
		IsBuffetDiscount:         "",                      // 是否开启自助餐优惠折扣（与旧服务保持一致，默认为空）
		AddClock:                 make([]AddClockItem, 0), // 加钟设置
	}
}

// DefaultServiceCharge 获取默认服务费设置
func DefaultServiceCharge() *ServiceCharge {
	return &ServiceCharge{
		IsOpen:              "0",              // 是否开启 0关闭 1开启
		ChargeType:          "1",              // 服务费类型 1-固定金额 2-百分比
		IsOpenTax:           "0",              // 税费 1-收取税费 0-不收取税费
		ApplyScope:          "1",              // 适用范围 1-全部 2-部分
		ApplyScopeOrdering:  "0",              // 适用范围-点餐 0-关闭 1-开启
		ApplyScopeTable:     "0",              // 适用范围-桌台 0-关闭 1-开启
		ApplyScopeTableList: make([]int64, 0), // 适用范围-桌台id列表
	}
}

// DefaultTaxRate 获取默认税率设置
func DefaultTaxRate() *TaxRate {
	return &TaxRate{
		IsOpen:         "0",                           // 是否开启 0关闭 1开启
		CalcType:       "1",                           // 计算类型 商品已含税价-1 商品未含税价-2
		AddTaxCategory: make([]AddTaxCategoryItem, 0), // 增值税分类
	}
}

// DefaultPointsSetting 获取默认积分设置（与旧服务保持一致）
func DefaultPointsSetting() *PointsSetting {
	return &PointsSetting{
		DeductionOrder:     "1",   // 扣款顺序 1-先主账户后赠送账户 2-先赠送账户后主账户 3-按比例
		DeductRatioMain:    "100", // 主账户扣款比例0-100
		DeductRatioGift:    "0",   // 赠送账户扣款比例0-100
		PointsName:         "积分",  // 积分名称自定义
		IsShoppingGift:     "0",   // 是否开启购物送积分
		GiftRatio:          "100", // 积分赠送比例
		IsShoppingDiscount: "0",
		Discount: DiscountItem{
			DiscountRatio:  "0.01",
			FullOrderPrice: "100.00",
			MaxMoneyRatio:  "10",
		},
		Describe: "a) 积分不可兑现、不可转让,仅可在本平台使用;\n" +
			"b) 您在本平台参加特定活动也可使用积分,详细使用规则以具体活动时的规则为准;\n" +
			"c) 积分的数值精确到个位(小数点后全部舍弃,不进行四舍五入)\n" +
			"d) 买家在完成该笔交易(订单状态为\"已签收\")后才能得到此笔交易的相应积分,如购买商品参加店铺其他优惠,则优惠的金额部分不享受积分获取;",
		DeductOrder: "", // 与旧服务保持一致
		ShoppingGiftRules: []ShoppingGiftRule{
			{
				Type:                     "payment_amount",
				IsOpen:                   "0",
				IsMemberLevelRelated:     "0",
				Value:                    "",
				PaymentAmountRequirement: "",
				MealType:                 []string{},
				BalancePaymentGetPoints:  "1",
				RefundReturnPoints:       "0",
				MemberLevels:             []MemberLevelItem{},
			},
			{
				Type:                     "desk",
				IsOpen:                   "0",
				IsMemberLevelRelated:     "0",
				Value:                    "",
				PaymentAmountRequirement: "",
				MealType:                 []string{"buffet"},
				BalancePaymentGetPoints:  "1",
				RefundReturnPoints:       "0",
				MemberLevels:             []MemberLevelItem{},
			},
		},
		Exchange: PointsExchange{
			OpenPointsExchange: "0",
			PointsExchangeRate: "",
			AutoPointsExchange: "0",
		},
	}
}

// DefaultCloudBasicSetting 获取默认云端基础信息
func DefaultCloudBasicSetting() *CloudBasicSetting {
	return &CloudBasicSetting{
		BrandName:          "TTPOS",
		BrandLogo:          "",
		BrandLogoLong:      "",
		BrowserLogo:        "",
		BrowserTitle:       "TTPOS",
		ExpirationReminder: 30,
	}
}
