package member_req

import (
	"errors"
	"slices"
	"ttpos-server-go/app/dto"
)

// MemberAddressSearchReq 搜索地址请求
type MemberAddressSearchReq struct {
	Text      string  `form:"text"`      // 搜索文本
	Latitude  float64 `form:"latitude"`  // 纬度
	Longitude float64 `form:"longitude"` // 经度
}

type MemberAddressListReq struct {
	dto.PageReq // 分页参数
}

type MemberAddressAddReq struct {
	Name        string `json:"name"`         // 联系人
	Phone       string `json:"phone"`        // 手机号
	Address     string `json:"address"`      // 详细地址
	Street      string `json:"street"`       // 街道/门牌号
	IsDefault   bool   `json:"is_default"`   // 是否默认
	Location    string `json:"location"`     // 位置坐标
	PhonePrefix string `json:"phone_prefix"` // 地区编码，仅支持+86、+66
}

func (r *MemberAddressAddReq) Validate() error {
	if r.Name == "" {
		return errors.New("联系人不能为空")
	}
	if r.Phone == "" {
		return errors.New("手机号不能为空")
	}
	if len(r.Phone) != 11 && len(r.Phone) != 10 {
		return errors.New("手机号不正确")
	}
	if r.Address == "" {
		return errors.New("详细地址不能为空")
	}
	if r.Street == "" {
		return errors.New("街道/门牌号不能为空")
	}
	if !slices.Contains([]string{"+86", "+66"}, r.PhonePrefix) {
		return errors.New("地区编码不正确")
	}
	return nil
}

type MemberAddressUpdateReq struct {
	Uuid        uint64 `json:"uuid"`         // 地址UUID
	Name        string `json:"name"`         // 联系人
	Phone       string `json:"phone"`        // 手机号
	Address     string `json:"address"`      // 详细地址
	Street      string `json:"street"`       // 街道/门牌号
	IsDefault   bool   `json:"is_default"`   // 是否默认
	Location    string `json:"location"`     // 位置坐标
	PhonePrefix string `json:"phone_prefix"` // 地区编码，仅支持+86、+66
}

func (r *MemberAddressUpdateReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("地址UUID不能为空")
	}
	if r.Name == "" {
		return errors.New("联系人不能为空")
	}
	if r.Phone == "" {
		return errors.New("手机号不能为空")
	}
	if len(r.Phone) != 11 && len(r.Phone) != 10 {
		return errors.New("手机号不正确")
	}
	if r.Address == "" {
		return errors.New("详细地址不能为空")
	}
	if r.Street == "" {
		return errors.New("街道/门牌号不能为空")
	}
	if !slices.Contains([]string{"+86", "+66"}, r.PhonePrefix) {
		return errors.New("地区编码不正确")
	}
	return nil
}

type MemberAddressDeleteReq struct {
	Uuid uint64 `json:"uuid"` // 地址UUID
}

type MemberAddressAuthReq struct {
	Uuid          uint64 `json:"uuid"`           // 地址UUID
	Code          string `json:"code"`           // 验证码
	IsRegister    bool   `json:"is_register"`    // 是否注册
	ReferrerPhone string `json:"referrer_phone"` // 推荐人手机号
}

func (r *MemberAddressAuthReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("地址UUID不能为空")
	}
	if r.Code == "" {
		return errors.New("验证码不能为空")
	}
	return nil
}

// MemberAddressDetailReq 获取地址详情请求
type MemberAddressDetailReq struct {
	Uuid uint64 `form:"uuid" json:"uuid"` // 地址UUID
}

func (r *MemberAddressDetailReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("地址UUID不能为空")
	}
	return nil
}
