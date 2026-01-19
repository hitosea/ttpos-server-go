package resp

import (
	"encoding/json"
	"strconv"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/setting"
)

type CashierBase struct {
	Username     string        `json:"username"`      // 登录账号
	CashierUuid  uint64        `json:"cashier_uuid"`  // 收银员UUID
	DeviceId     string        `json:"device_id"`     // 设备ID
	DeviceRemark string        `json:"device_remark"` // 设备备注
	Permissions  []*Permission `json:"permissions"`   // 页面权限

	Buffet     setting.BuffetResp  `json:"buffet"`   // 自助餐设置
	CloudBasic setting.CloudBasic  `json:"cloud"`    // 云端基础信息
	Company    Company             `json:"company"`  // 商家信息
	Currency   setting.Currency    `json:"currency"` // 货币单位
	Business   setting.Business    `json:"business"` // 门店业务设置
	Cashier    setting.CashierResp `json:"cashier"`  // 收银机设置
	Printer    setting.Printer     `json:"printer"`  // 打印机设置

	UpdateTime int64 `json:"update_time"` // 更新时间

	// 关联的门店列表
	CompanyList []*CompanyStaffResp `json:"company_list,omitempty"`
}

type AssistantStaff struct {
	Uuid     uint64 `json:"uuid"`      // 点餐助手员工uuid
	RealName string `json:"real_name"` // 点餐助手员工真实姓名
	Phone    string `json:"phone"`     // 点餐助手员工手机号
	DeviceId string `json:"device_id"` // 点餐助手设备ID
}

type CashierStaff struct {
	RealName     string `json:"real_name"`     // 收银员真实姓名
	Username     string `json:"username"`      // 收银员账号
	DeviceId     string `json:"device_id"`     // 收银机设备ID
	DeviceRemark string `json:"device_remark"` // 收银机备注
}

type AssistantBase struct {
	Permissions    []string              `json:"permissions"`     // 点餐助手权限
	CashierStaff   CashierStaff          `json:"cashier_staff"`   // 收银机员工
	AssistantStaff AssistantStaff        `json:"assistant_staff"` // 点餐助手员工
	Buffet         setting.BuffetResp    `json:"buffet"`          // 自助餐设置
	CloudBasic     setting.CloudBasic    `json:"cloud"`           // 云端基础信息
	Company        Company               `json:"company"`         // 商家信息
	Currency       setting.Currency      `json:"currency"`        // 货币设置
	Business       setting.Business      `json:"business"`        // 门店业务设置
	Assistant      setting.AssistantResp `json:"assistant"`       // 点餐助手设置
	Printer        setting.Printer       `json:"printer"`         // 打印机设置
	Kitchen        setting.KitchenResp   `json:"kitchen"`         // 厨显设置
	ClientVersion  string                `json:"client_version"`  // 客户端版本
	ServerVersion  string                `json:"server_version"`  // 服务端版本

	// 关联的门店列表
	CompanyList []*CompanyStaffResp `json:"company_list,omitempty"`
}

type TabletBase struct {
	RealName      string              `json:"real_name"`      // 员工姓名
	ServerVersion string              `json:"system_version"` // 服务端版本
	ClientVersion string              `json:"client_version"` // 客户端版本
	Buffet        setting.BuffetResp  `json:"buffet"`         // 自助餐设置
	CloudBasic    setting.CloudBasic  `json:"cloud"`          // 云端基础信
	Company       Company             `json:"company"`        // 商家信息
	Currency      setting.Currency    `json:"currency"`       // 货币单位
	Business      setting.Business    `json:"business"`       // 门店业务设置
	Tablet        setting.TabletResp  `json:"tablet"`         // 平板端设置
	Kitchen       setting.KitchenResp `json:"kitchen"`        // 厨显设置

	// 关联的门店列表
	CompanyList []*CompanyStaffResp `json:"company_list,omitempty"`
}

type TabletDeskItem struct {
	Uuid   uint64 `json:"uuid"`    // 桌台uuid
	DeskNo string `json:"desk_no"` // 桌台编号
}

type TabletDeskList struct {
	List []TabletDeskItem `json:"list"`
}

type KitchenBase struct {
	RealName      string              `json:"real_name"`      // 员工真实姓名
	Buffet        setting.BuffetResp  `json:"buffet"`         // 自助餐设置
	CloudBasic    setting.CloudBasic  `json:"cloud"`          // 云端基础信息
	Company       Company             `json:"company"`        // 商家信息
	Currency      setting.Currency    `json:"currency"`       // 货币单位
	Business      setting.Business    `json:"business"`       // 门店业务设置
	Kitchen       setting.KitchenResp `json:"kitchen"`        // 厨显端设置
	ServerVersion string              `json:"system_version"` // 服务端版本
	ClientVersion string              `json:"client_version"` // 客户端版本

	// 关联的门店列表
	CompanyList []*CompanyStaffResp `json:"company_list,omitempty"`
}

type KioskBase struct {
	Username     string            `json:"username"`      // 登录账号
	DeviceId     string            `json:"device_id"`     // 设备ID
	DeviceRemark string            `json:"device_remark"` // 设备备注
	Company      Company           `json:"company"`       // 商家信息
	Currency     setting.Currency  `json:"currency"`      // 货币单位
	Business     setting.Business  `json:"business"`      // 门店业务设置
	Kiosk        setting.KioskResp `json:"kiosk"`         // 自助点餐机设置（包含语言列表、轮播广告）
	UpdateTime   int64             `json:"update_time"`   // 更新时间

	ServerVersion string `json:"system_version"` // 服务端版本
	ClientVersion string `json:"client_version"` // 客户端版本

	// 关联的门店列表
	CompanyList []*CompanyStaffResp `json:"company_list,omitempty"`
}

type ProductPrinter struct {
	Uuid   uint64 `json:"uuid"`   // 商品打印机uuid
	Name   string `json:"name"`   // 商品打印机名称
	Status int    `json:"status"` // 商品打印机状态
}

type ProductPrinterList struct {
	List []ProductPrinter `json:"list"`
}

// NewProductPrinterList 创建新的ProductPrinterList实例，确保List字段为空数组而非nil
func NewProductPrinterList() ProductPrinterList {
	return ProductPrinterList{
		List: make([]ProductPrinter, 0),
	}
}

type Company struct {
	Uuid                 uint64 `json:"uuid"`                    // 商家UUID
	Name                 string `json:"name"`                    // 商家名称
	Logo                 string `json:"logo"`                    // 商家logo
	TimeZone             string `json:"time_zone"`               // 时区，形如 Asia/Shanghai
	ExpireTime           int64  `json:"expire_time"`             // 店铺到期时间，0表示没有过期时间
	IsOpenMember         int    `json:"is_open_member"`          // 是否开启会员功能: 0不开启, 1开启
	IsOpenBuffet         int    `json:"is_open_buffet"`          // 是否开启自助餐功能: 0不开启, 1开启
	IsOpenH5Order        int    `json:"is_open_h5_order"`        // 是否开启扫码接单功能: 0不开启, 1开启
	IsOpenOldOrder       int    `json:"is_open_old_order"`       // 是否开启旧订单功能: 0不开启, 1开启
	IsOpenRider          bool   `json:"is_open_rider"`           // 是否开启外送
	IsEnableErp          bool   `json:"is_show_inventory"`       // 是否显示移动管理端进销存功能
	IsOpenMap            bool   `json:"is_open_map"`             // 是否开启地图
	IsOpenDataManagement bool   `json:"is_open_data_management"` // 是否开启数据管理功能
	IsOpenKiosk          bool   `json:"is_open_kiosk"`           // 是否开启自助点餐机功能
	IsOpenGrabDelivery   bool   `json:"is_open_grab_delivery"`   // 是否开启Grab外卖功能
	StoreCode            string `json:"store_code"`              // 店铺编码，shop端用于显示
}

type Permission struct {
	ID           int           `json:"id"`
	Uuid         uint64        `json:"uuid"`
	Name         string        `json:"name"`
	Path         string        `json:"path"`
	ParentUuid   uint64        `json:"parent_id"`
	Sort         int           `json:"-"`
	RedirectName string        `json:"redirect_name"`
	IsRoute      int           `json:"is_route"`
	IsMenu       int           `json:"is_menu"`
	Alias        string        `json:"alias"`
	IsShow       int           `json:"is_show"`
	CreateTime   string        `json:"-"`
	UpdateTime   string        `json:"-"`
	Children     []*Permission `json:"children"`
}
type PermissionGroup struct {
	List []*Permission `json:"list"`
}

type LanguageResp struct {
	Languages       []string           `json:"languages"`        // 语言列表
	LanguageList    []dto.LanguageItem `json:"language_list"`    // 语言列表
	DefaultLanguage string             `json:"default_language"` // 默认语言
}

type Ads struct {
	List []setting.CarouselItem `json:"list"`
}

type UpdateInfo struct {
	VersionName  string `json:"version_name"`  // 版本名称
	ForcedUpdate int    `json:"forced_update"` // 是否强制更新
	UpdateLog    string `json:"update_log"`    // 更新日志
	DownloadURL  string `json:"download_url"`  // 下载URL
}

type LoginResp struct {
	Token               string `json:"token"`
	RefreshToken        string `json:"refresh_token"` // 刷新token，用于重新获取token
	CashierIsFirstLogin bool   `json:"-"`
	NeedChangePassword  bool   `json:"-"` // 是否需要修改密码，首次登录移动管理端需要修改密码
}
type ShopLoginResp struct {
	Token              string `json:"token"`
	RefreshToken       string `json:"refresh_token"`        // 刷新token，用于重新获取token
	NeedChangePassword bool   `json:"need_change_password"` // 是否需要修改密码，首次登录移动管理端需要修改密码
}

type CashierLoginResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`  // 刷新token，用于重新获取token
	IsFirstLogin bool   `json:"is_first_login"` // 是否首次登录
}

type AcceptOrderSetting struct {
	IsAutoOrder    string `json:"is_auto_order"`    // 是否自动接单：0-否；1-是
	AutoOrderLimit string `json:"auto_order_limit"` // 自动接单金额上限，0.01-100000000
	IsAutoVoice    string `json:"is_auto_voice"`    // 是否开启自动接单语音播报 0-否；1-是
}

type AcceptMemberOrderSetting struct {
	IsAutoMemberOrder      string `json:"is_auto_member_order"`       // 是否自动接单会员订单：0-否；1-是
	AutoMemberOrderLimit   string `json:"auto_member_order_limit"`    // 自动接单会员订单金额上限，0.01-100000000
	IsAutoVoiceMemberOrder string `json:"is_auto_voice_member_order"` // 是否开启自动接单会员订单语音播报 0-否；1-是
}

// CanAutoOrder 是否可以自动接单
func (s *AcceptOrderSetting) CanAutoOrder(amount float64) bool {
	amountFloat, _ := strconv.ParseFloat(s.AutoOrderLimit, 64)
	return s.IsAutoOrder == "1" && amount <= amountFloat
}

type SystemSetting struct {
	IsShowScanSoldOut      int    `json:"is_show_scan_sold_out"`      // 扫码点餐端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsShowAssistantSoldOut int    `json:"is_show_assistant_sold_out"` // 助手端点餐助手是否显示售罄商品 0-不显示 1-显示
	MenuShowSoldOut        int    `json:"menu_show_sold_out"`         // 电子菜单是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	MemberShowSoldOut      int    `json:"member_show_sold_out"`       // 会员端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DishCardStyle          string `json:"dish_card_style"`            // 菜品卡片样式 0-无图模式 1-图片模式
	IsShowSoldOut          int    `json:"is_show_sold_out"`           // 平板端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DefaultLanguage        string `json:"default_language"`           // 默认语言
	SecondLanguage         string `json:"second_language"`            // 副屏语言
	DeviceId               string `json:"device_id"`                  // 当前机器ID
	DeviceRemark           string `json:"device_remark"`              // 机器备注
	ClientVersion          string `json:"client_version"`             // 客户端版本
	ServerVersion          string `json:"server_version"`             // 服务端版本
}

type UsbPrinter struct {
	M_name string `json:"m_name"` // 厂商名称
	Name   string `json:"name"`   // 打印机名称
	Pid    any    `json:"pid"`    // 打印机PID
	Sn     string `json:"sn"`     // 打印机SN
	Vid    any    `json:"vid"`    // 打印机VID
}

type UsbPrinterList struct {
	List       []UsbPrinter `json:"list"`
	SelectedSn string       `json:"selected_sn"`
}

type CashierBaseSetting struct {
	AcceptOrder       AcceptOrderSetting       `json:"accept_order"`        // 接单设置
	AcceptMemberOrder AcceptMemberOrderSetting `json:"accept_member_order"` // 接单会员订单设置
	System            SystemSetting            `json:"system"`              // 系统设置
	UsbPrinter        UsbPrinterList           `json:"usb_printer"`         // 自带打印机设置
}

// ShiftInfo 交班信息
type ShiftInfo struct {
	PreviousShiftCash   float64                 `json:"previous_shift_cash"`   // 上一班遗留备用金
	WithdrawCash        float64                 `json:"withdraw_cash"`         // 中途取出现金
	DepositCash         float64                 `json:"deposit_cash"`          // 中途存入现金
	CurrentCashTotal    float64                 `json:"current_cash_total"`    // 当前钱箱现金总计
	RefundAmount        float64                 `json:"refund_amount"`         // 退款金额
	PaymentMethodIncome PaymentMethodIncomeList `json:"payment_method_income"` // 支付方式收入
}

// PaymentMethodIncomeList 支付方式收入列表
type PaymentMethodIncomeList struct {
	List []PaymentMethodIncome `json:"list"`
}

// PaymentMethodIncome 支付方式收入详情
type PaymentMethodIncome struct {
	Name   string  `json:"name"`   // 支付方式名称
	Amount float64 `json:"amount"` // 收入金额
}

type CashierReportResp struct {
	PreviousShiftCash float64     `json:"previous_shift_cash"` // 上一班遗留备用金
	PrinterData       PrinterData `json:"printer_data"`        // 打印数据
}

// ShiftSubmit 交班提交
type ShiftSubmit struct {
	CashIncome   float64     `json:"cash_income"`    // 本班现金收入
	CashTakenOut float64     `json:"cash_taken_out"` // 本班取出现金
	CashLeft     float64     `json:"cash_left"`      // 本班遗留现金
	DutyNo       string      `json:"duty_no"`        // 班次编号
	PrinterData  PrinterData `json:"printer_data"`   // 打印数据
}

// 实现 BinaryMarshaler
func (ss ShiftSubmit) MarshalBinary() ([]byte, error) {
	return json.Marshal(ss)
}

// TakeoutStatusResp 外卖平台状态响应
// 参考: main/app/modules/takeout/interfaces/response/TakeoutStatusResponse
// 对应接口: /api/v1/shop/takeout/status/info
type TakeoutStatusResp struct {
	Platform string `json:"platform"` // 外卖平台 (grab/lineman等)
	Enabled  bool   `json:"enabled"`  // 是否开启
}

type ShopBase struct {
	Username      string              `json:"username"`       // 登录账号
	RealName      string              `json:"real_name"`      // 姓名
	ProfileUuid   uint64              `json:"profile_uuid"`   // 收银员UUID
	Phone         string              `json:"phone"`          // 登录账号手机号
	DeviceId      string              `json:"device_id"`      // 设备ID
	DeviceRemark  string              `json:"device_remark"`  // 设备备注
	TakeoutStatus []TakeoutStatusResp `json:"takeout_status"` // 外卖平台状态列表
	Permissions   []*Permission       `json:"permissions"`    // 页面权限

	Buffet     setting.BuffetResp `json:"buffet"`      // 自助餐设置
	CloudBasic setting.CloudBasic `json:"cloud"`       // 云端基础信息
	Company    Company            `json:"company"`     // 商家信息
	Currency   setting.Currency   `json:"currency"`    // 货币单位
	Business   setting.Business   `json:"business"`    // 门店业务设置
	Profile    ShopProfile        `json:"profile"`     // 门店信息
	IsOpenTax  bool               `json:"is_open_tax"` // 税率设置: false-关闭 true-开启

	// 是否散户site
	IsTtposSite   bool   `json:"is_ttpos_site"`  // 是否是TTPOS站点(散户site)，能登录新商家后台的必须授权erpnext，除了site_code="1"，其他都是连锁店模式，连锁店模式不能修改 单位和属性
	IsHeadquarter bool   `json:"is_headquarter"` // 是否是总部
	ServerVersion string `json:"server_version"` // 服务端版本
	UpdateTime    int64  `json:"update_time"`    // 更新时间

	HasChildren bool `json:"has_children"` // 有下级公司，则表示为上级公司，用于判断调拨规则是否支持

	// 关联的门店列表
	CompanyList []*CompanyStaffResp `json:"company_list,omitempty"`

	IsSyncing    bool  `json:"is_syncing"`     // 是否erp数据同步中
	LastSyncTime int64 `json:"last_sync_time"` // 上次同步erp数据完成时间

	HasDataPermission bool `json:"has_data_permission"` // 是否有数据管理权限

	AllowedTransferTypes string `json:"allowed_transfer_types"` // 允许的调拨类型 "in"-只允许调入 "out"-只允许调出 "in,out"-都允许
}

type ShopProfile struct {
	CompanyName     string                 `json:"company_name"`     // 公司名称，区别于店铺名称
	Address         string                 `json:"address"`          // 地址
	Coordinates     string                 `json:"coordinates"`      // 经纬度
	IpWhiteList     string                 `json:"ip_white_list"`    // ip白名单
	Phone           string                 `json:"phone"`            // 联系电话
	TaxNumber       string                 `json:"tax_number"`       // 税号
	StoreCode       string                 `json:"store_code"`       // 店铺编码，用于发票打印
	TimeZoneList    []setting.TimeZoneItem `json:"time_zone_list"`   // 时区列表
	DefaultLanguage string                 `json:"default_language"` // 默认语言
	LanguageList    []dto.LanguageItem     `json:"language_list"`    // 语言列表，当前勾选了的语言列表
	Language        []string               `json:"language"`         // 云平台限制商家的可用语言列表
}
