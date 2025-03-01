package resp

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/setting"
)

type CashierBase struct {
	Username     string           `json:"username"`      // 登录账号
	CashierUuid  uint64           `json:"cashier_uuid"`  // 收银员UUID
	DeviceId     string           `json:"device_id"`     // 设备ID
	DeviceRemark string           `json:"device_remark"` // 设备备注
	Cashier      setting.Cashier  `json:"cashier"`       // 收银机设置
	Business     setting.Business `json:"business"`      // 门店业务设置
	Buffet       setting.Buffet   `json:"buffet"`        // 自助餐设置
	Currency     setting.Currency `json:"currency"`      // 货币单位
	Permissions  []*Permission    `json:"permissions"`   // 页面权限
	Company      Company          `json:"company"`       // 商家信息
	Tablet       setting.Tablet   `json:"tablet"`        // 平板端设置
}

type AssistantStaff struct {
	Uuid       uint64   `json:"uuid"`       // 点餐助手员工uuid
	RealName   string   `json:"real_name"`  // 点餐助手员工真实姓名
	Phone      string   `json:"phone"`      // 点餐助手员工手机号
	DeviceId   string   `json:"device_id"`  // 点餐助手设备ID
	Permission []string `json:"permission"` // 点餐助手权限
}

type CashierStaff struct {
	RealName     string `json:"real_name"`     // 收银员真实姓名
	Username     string `json:"username"`      // 收银员账号
	DeviceId     string `json:"device_id"`     // 收银机设备ID
	DeviceRemark string `json:"device_remark"` // 收银机备注
}

type AssistantBase struct {
	CashierStaff   CashierStaff      `json:"cashier_staff"`   // 收银机员工
	AssistantStaff AssistantStaff    `json:"assistant_staff"` // 点餐助手员工
	Company        Company           `json:"company"`         // 商家信息
	Assistant      setting.Assistant `json:"assistant"`       // 点餐助手设置
	Buffet         setting.Buffet    `json:"buffet"`          // 自助餐设置
	Payment        setting.Payment   `json:"payment"`         // 支付设置
	Business       setting.Business  `json:"business"`        // 门店业务设置
	Kitchen        setting.Kitchen   `json:"kitchen"`         // 厨显端设置
	Currency       setting.Currency  `json:"currency"`        // 货币设置
}
type TabletBase struct {
	Username   string `json:"username"`    // 登录账号
	TabletUuid uint64 `json:"tablet_uuid"` // 平板端员工UUID
}

type TabletDeskItem struct {
	Uuid   uint64 `json:"uuid"`    // 桌台uuid
	DeskNo string `json:"desk_no"` // 桌台编号
}

type TabletDeskList struct {
	List []TabletDeskItem `json:"list"`
}

type KitchenBase struct {
	Kitchen setting.Kitchen `json:"kitchen"` // 厨显端设置
	Company Company         `json:"company"` // 商家信息
}

type Company struct {
	Uuid     uint64 `json:"uuid"`      // 商家UUID
	Name     string `json:"name"`      // 商家名称
	TimeZone string `json:"time_zone"` // 时区，形如 Asia/Shanghai
}

type Permission struct {
	ID               int           `json:"id"`
	Uuid             uint64        `json:"uuid"`
	Name             string        `json:"name"`
	Path             string        `json:"path"`
	APIPath          string        `json:"-"`
	ParentUuid       uint64        `json:"parent_id"`
	Sort             int           `json:"-"`
	Icon             string        `json:"-"`
	RedirectName     string        `json:"redirect_name"`
	IsRoute          int           `json:"is_route"`
	IsMenu           int           `json:"is_menu"`
	Alias            string        `json:"alias"`
	IsShow           int           `json:"is_show"`
	PlusCategoryUuid uint64        `json:"-"`
	Remark           string        `json:"-"`
	IsSupplier       int           `json:"-"`
	AppId            int           `json:"-"`
	CreateTime       string        `json:"-"`
	UpdateTime       string        `json:"-"`
	Children         []*Permission `json:"children"`
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
	VersionName  string `json:"version_name"`
	ForcedUpdate int    `json:"forced_update"`
	UpdateLog    string `json:"update_log"`
	DownloadURL  string `json:"download_url"`
}

type LoginResp struct {
	Token               string `json:"token"`
	CashierIsFirstLogin bool   `json:"cashier_is_first_login"`
}

type CashierLoginResp struct {
	Token        string `json:"token"`           // token
	IsFirstLogin bool   `json:" is_first_login"` // 是否首次登录
}

type AcceptOrderSetting struct {
	IsAutoOrder    string `json:"is_auto_order"`    // 是否自动接单：0-否；1-是
	AutoOrderLimit string `json:"auto_order_limit"` // 自动接单金额上限，0.01-100000000
	IsAutoVoice    string `json:"is_auto_voice"`    // 是否开启自动接单语音播报 0-否；1-是
}

type SystemSetting struct {
	IsShowScanSoldOut      int    `json:"is_show_scan_sold_out"`      // 扫码点餐端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	IsShowAssistantSoldOut int    `json:"is_show_assistant_sold_out"` // 助手端点餐助手是否显示售罄商品 0-不显示 1-显示
	MenuShowSoldOut        string `json:"menu_show_sold_out"`         // 电子菜单是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DishCardStyle          string `json:"dish_card_style"`            // 菜品卡片样式 0-无图模式 1-图片模式
	IsShowSoldOut          int    `json:"is_show_sold_out"`           // 平板端是否显示售罄商品 0-关闭（不显示售罄） 1-开启（显示售罄）
	DefaultLanguage        string `json:"default_language"`           // 默认语言
	SecondLanguage         string `json:"second_language"`            // 副屏语言
	DeviceId               string `json:"device_id"`                  // 当前机器ID
	DeviceRemark           string `json:"device_remark"`              // 机器备注
	ClientVersion          string `json:"client_version"`             // 客户端版本
	ServerVersion          string `json:"server_version"`             // 服务端版本
}

type CashierBaseSetting struct {
	AcceptOrder AcceptOrderSetting `json:"accept_order"` // 接单设置
	System      SystemSetting      `json:"system"`       // 系统设置
}
