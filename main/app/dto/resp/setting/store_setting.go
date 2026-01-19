package setting

import "ttpos-server-go/app/dto"

// Store 商城设置
type Store struct {
	Name          string             `json:"name"`           // 商城名称
	AvatarURL     string             `json:"avatar_url"`     // 默认头像
	LogoURL       string             `json:"logo_url"`       // 商城logo
	ZeroingMethod string             `json:"zeroing_method"` // 抹零方式: 0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
	IPWhiteList   string             `json:"ip_white_list"`  // ip白名单
	TimeZone      string             `json:"time_zone"`      // 时区
	NoClearTable  string             `json:"no_clear_table"` // 结账后不清台 0-清台 1-不清台
	TimeZoneList  []TimeZoneItem     `json:"time_zone_list"` // 时区列表
	Company       string             `json:"company"`        // 公司名称
	Address       string             `json:"address"`        // 地址
	Phone         string             `json:"phone"`          // 联系电话
	TaxNumber     string             `json:"tax_number"`     // 税号
	StoreCode     string             `json:"store_code"`     // 店铺编码，用于发票打印
	ChainNumber   string             `json:"chain_number"`   // 连锁编号
	Language      []dto.LanguageItem `json:"language"`       // 系统语言
	AuthLanguage  string             `json:"auth_language"`  // 授权语言
	Coordinates   string             `json:"coordinates"`    // 经纬度
	Latitude      string             `json:"-"`              // 纬度
	Longitude     string             `json:"-"`              // 经度
}

type TimeZoneItem struct {
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
