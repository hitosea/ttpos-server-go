package member_resp

import "ttpos-server-go/app/dto"

type MemberResp struct {
	Id        uint    `json:"id"`         // 会员ID
	Uuid      uint64  `json:"uuid"`       // 会员UUID
	Nickname  string  `json:"nickname"`   // 会员昵称
	Phone     string  `json:"phone"`      // 会员手机号
	Point     float64 `json:"point"`      // 会员积分
	Balance   float64 `json:"balance"`    // 会员余额
	IsVisitor bool    `json:"is_visitor"` // 是否游客
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
	Member              MemberResp         `json:"member"`                  // 会员信息
	Company             CompanyResp        `json:"company"`                 // 公司信息
	AreaCode            []string           `json:"area_code"`               // 区号列表
	LanguageList        []dto.LanguageItem `json:"language_list"`           // 语言列表
	IsMemberShowSoldOut int                `json:"is_member_show_sold_out"` // 是否显示售罄商品
}
