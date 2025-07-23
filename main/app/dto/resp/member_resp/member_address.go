package member_resp

import "ttpos-server-go/app/dto"

type MemberAddressResp struct {
	Uuid        uint64 `json:"uuid"`          // 地址UUID
	Name        string `json:"name"`          // 联系人
	Phone       string `json:"phone"`         // 手机号
	Country     string `json:"country"`       // 国家代码
	Address     string `json:"address"`       // 详细地址
	Street      string `json:"street"`        // 街道/门牌号
	IsDefault   int    `json:"is_default"`    // 是否默认
	IsAuthPhone bool   `json:"is_auth_phone"` // 是否认证手机号
}

type MemberAddressListResp struct {
	List []MemberAddressResp `json:"list"`
	Meta dto.PageResponse    `json:"meta"` // 分页信息
}

// MemberAddressDetailResp 地址详情响应
type MemberAddressDetailResp struct {
	Uuid          uint64 `json:"uuid"`           // 地址UUID
	Name          string `json:"name"`           // 联系人
	Gender        int    `json:"gender"`         // 性别 0-女 1-男
	Phone         string `json:"phone"`          // 手机号
	Country       string `json:"country"`        // 国家代码
	Province      string `json:"province"`       // 省份
	City          string `json:"city"`           // 城市
	Area          string `json:"area"`           // 区
	Address       string `json:"address"`        // 详细地址
	Street        string `json:"street"`         // 街道/门牌号
	IsDefault     int    `json:"is_default"`     // 是否默认
	Location      string `json:"location"`       // 位置坐标
	AddressDetail string `json:"address_detail"` // 地址详情（完整地址串）
}
