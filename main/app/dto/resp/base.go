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
}

type AssistantBase struct {
	Username      string `json:"username"`       // 登录账号
	AssistantUuid uint64 `json:"assistant_uuid"` // 点餐助手员工UUID
}

type Company struct {
	Uuid uint64 `json:"uuid"` // 商家UUID
	Name string `json:"name"` // 商家名称
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
