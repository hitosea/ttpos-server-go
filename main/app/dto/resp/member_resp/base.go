package member_resp

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/setting"
)

type UserResp struct {
	Id        uint    `json:"id"`         // 会员ID
	Uuid      uint64  `json:"uuid"`       // 会员UUID
	Nickname  string  `json:"nickname"`   // 会员昵称
	Phone     string  `json:"phone"`      // 会员手机号
	Point     float64 `json:"point"`      // 会员积分
	Balance   float64 `json:"balance"`    // 会员余额
	IsVisitor bool    `json:"is_visitor"` // 是否游客
}

type MemberResp struct {
	LanguageList         []dto.LanguageItem `json:"language_list"`            // 语言列表
	IsMemberShowSoldOut  bool               `json:"is_member_show_sold_out"`  // 是否显示售罄商品
	IsOpenRider          bool               `json:"is_open_rider"`            // 是否开启外送
	IsOpenStoreScanOrder bool               `json:"is_open_store_scan_order"` // 是否开启门店扫码点餐
	IsStoreResting       bool               `json:"is_store_resting"`         // 商家是否休息中
	AreaCode             []string           `json:"area_code"`                // 区号列表
	Language             []string           `json:"language"`                 // 常用语言
	DefaultLanguage      string             `json:"default_language"`         // 默认语言
}

type CompanyResp struct {
	Uuid         uint64 `json:"uuid"`          // 公司UUID
	Name         string `json:"name"`          // 公司名称
	Logo         string `json:"logo"`          // 公司logo
	Address      string `json:"address"`       // 公司地址
	LinkPhone    string `json:"link_phone"`    // 公司联系电话
	OpeningHours string `json:"opening_hours"` // 公司营业时间
}

type MemberBaseInfoResp struct {
	User          UserResp         `json:"user"`           // 会员信息
	Member        MemberResp       `json:"member"`         // 会员端配置信息
	Company       CompanyResp      `json:"company"`        // 公司信息
	Currency      setting.Currency `json:"currency"`       // 货币单位
	TemplateStyle int              `json:"template_style"` // 模板样式：1/2/3
}
