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

type AssistantBase struct {
	Username      string `json:"username"`       // 登录账号
	AssistantUuid uint64 `json:"assistant_uuid"` // 点餐助手员工UUID
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
	Username    string `json:"username"`     // 登录账号
	KitchenUuid uint64 `json:"kitchen_uuid"` // 厨显端员工UUID
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
