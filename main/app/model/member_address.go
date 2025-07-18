package model

import (
	"strings"
	"time"
	"ttpos-server-go/app/errors"
)

// 会员地址表 `ttpos_member_address`
type MemberAddress struct {
	BaseModel
	MemberUuid uint64 `gorm:"column:member_uuid;type:bigint(20);not null;comment:会员uuid" json:"member_uuid"`
	Name       string `gorm:"column:name;type:varchar(50);not null;comment:联系人" json:"name"`
	Gender     int    `gorm:"column:gender;type:int(1);not null;comment:性别 0-女 1-男" json:"gender"`
	Phone      string `gorm:"column:phone;type:varchar(20);not null;comment:手机号" json:"phone"`
	Country    string `gorm:"column:country;type:varchar(10);not null;comment:国家代码" json:"country"`
	Province   string `gorm:"column:province;type:varchar(50);not null;comment:省份" json:"province"`
	City       string `gorm:"column:city;type:varchar(50);not null;comment:城市" json:"city"`
	Area       string `gorm:"column:area;type:varchar(50);not null;comment:区" json:"area"`
	Address    string `gorm:"column:address;type:varchar(255);not null;comment:详细地址" json:"address"`
	Street     string `gorm:"column:street;type:varchar(255);not null;comment:街道/门牌号" json:"street"`
	IsDefault  int    `gorm:"column:is_default;type:int(1);not null;comment:是否默认" json:"is_default"`
	Location   string `gorm:"column:location;type:varchar(100);not null;comment:位置坐标" json:"location"` // "纬度,经度"
	AuthPhone  string `gorm:"column:auth_phone;type:varchar(20);not null;comment:认证手机号" json:"auth_phone"`
	AuthTime   int64  `gorm:"column:auth_time;type:int(11);not null;comment:认证时间" json:"auth_time"`

	Member *Member `gorm:"foreignKey:MemberUuid;references:Uuid" json:"member"`
}

// 是否认证
func (model *MemberAddress) IsAuthPhone() bool {
	// 检查手机号是否匹配且认证时间不为0
	if model.AuthPhone != model.Phone || model.AuthTime == 0 {
		return false
	}

	// 获取当前时间戳
	now := time.Now().Unix()

	// 计算24小时的时间戳（24 * 60 * 60 = 86400秒）
	dayInSeconds := int64(86400)

	// 判断认证时间是否在24小时之内
	return (now - model.AuthTime) <= dayInSeconds
}

// 获取位置坐标. 返回纬度,经度
func (model *MemberAddress) GetLocation() (string, string, error) {
	if model.Location == "" {
		return "", "", errors.New("位置坐标为空")
	}
	location := strings.Split(model.Location, ",")
	if len(location) != 2 {
		return "", "", errors.New("位置坐标格式错误")
	}
	return location[0], location[1], nil
}

// 获取地址详情,用于显示在前端的地址长串
func (model *MemberAddress) GetAddressDetail() string {
	return model.Address + " " + model.Street
}
