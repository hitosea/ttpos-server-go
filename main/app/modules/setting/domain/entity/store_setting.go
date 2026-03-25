package entity

import (
	"strings"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
)

// StoreSetting 门店设置实体
type StoreSetting struct {
	Name          string                     `json:"name"`           // 门店名称
	AvatarURL     string                     `json:"avatar_url"`     // 默认头像
	LogoURL       string                     `json:"logo_url"`       // 门店logo
	ZeroingMethod string                     `json:"zeroing_method"` // 抹零方式: 0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
	IPWhiteList   string                     `json:"ip_white_list"`  // ip白名单
	TimeZone      string                     `json:"time_zone"`      // 时区
	NoClearTable  string                     `json:"no_clear_table"` // 结账后不清台 0-清台 1-不清台
	TimeZoneList  []valueobject.TimeZoneItem `json:"time_zone_list"` // 时区列表
	Company       string                     `json:"company"`        // 公司名称
	Address       string                     `json:"address"`        // 地址
	Phone         string                     `json:"phone"`          // 联系电话
	TaxNumber     string                     `json:"tax_number"`     // 税号
	StoreCode     string                     `json:"store_code"`     // 店铺编码，用于发票打印
	ChainNumber   string                     `json:"chain_number"`   // 连锁编号
	Language      []valueobject.LanguageItem `json:"language"`       // 系统语言
	AuthLanguage  string                     `json:"auth_language"`  // 授权语言
	Coordinates   string                     `json:"coordinates"`    // 经纬度
	Latitude      string                     `json:"-"`              // 纬度
	Longitude     string                     `json:"-"`              // 经度
}

// NewStoreSetting 创建新的门店设置
func NewStoreSetting(name, company, address, phone string) *StoreSetting {
	return &StoreSetting{
		Name:          name,
		Company:       company,
		Address:       address,
		Phone:         phone,
		ZeroingMethod: "0", // 默认不抹零
		NoClearTable:  "0", // 默认清台
		TimeZoneList:  make([]valueobject.TimeZoneItem, 0),
		Language:      make([]valueobject.LanguageItem, 0),
	}
}

// DefaultStoreSetting 获取默认门店设置
func DefaultStoreSetting() *StoreSetting {
	return &StoreSetting{
		Name:          "Shop",                     // 商城名称
		AvatarURL:     "image/user/avatarUrl.png", //默认头像
		LogoURL:       "image/diy/logo.png",       //商城logo
		ZeroingMethod: "0",                        // 抹零方式，0-不抹零 1-抹分 2-抹角 3-四舍五入到角 4-四舍五入到元
		NoClearTable:  "0",                        // 结账后不清台 0-清台 1-不清台
		TimeZoneList: []valueobject.TimeZoneItem{
			{
				Name:  "（UTC+09:00）东京",
				Key:   "Asia/Tokyo",
				Value: "（UTC+09:00）东京",
			}, {
				Name:  "（UTC+08:00）北京，重庆，香港特别行政区，乌鲁木齐",
				Key:   "Asia/Shanghai",
				Value: "（UTC+08:00）北京，重庆，香港特别行政区，乌鲁木齐",
			}, {
				Name:  "（UTC+07:00）曼谷，河内，雅加达",
				Key:   "Asia/Bangkok",
				Value: "（UTC+07:00）曼谷，河内，雅加达",
			}, {
				Name:  "（UTC+07:00）胡志明市，河内",
				Key:   "Asia/Ho_Chi_Minh",
				Value: "（UTC+07:00）胡志明市，河内",
			}, {
				Name:  "（UTC+06:30）内比都，仰光",
				Key:   "Asia/Yangon",
				Value: "（UTC+06:30）内比都，仰光",
			}, {
				Name:  "（UTC+3:00）安卡拉",
				Key:   "Europe/Istanbul",
				Value: "（UTC+3:00）安卡拉",
			}, {
				Name:  "（UTC+2:00）斯德哥尔摩",
				Key:   "Europe/Stockholm",
				Value: "（UTC+2:00）斯德哥尔摩",
			},
		}, // 时区列表
		Language: make([]valueobject.LanguageItem, 0), // 语言列表
	}
}

// UpdateBasicInfo 更新基本信息
func (s *StoreSetting) UpdateBasicInfo(name, company, address, phone, storeCode string) {
	s.Name = name
	s.Company = company
	s.Address = address
	s.Phone = phone
	s.StoreCode = storeCode
}

// UpdateLogo 更新logo
func (s *StoreSetting) UpdateLogo(logoURL string) {
	s.LogoURL = logoURL
}

// UpdateAvatar 更新头像
func (s *StoreSetting) UpdateAvatar(avatarURL string) {
	s.AvatarURL = avatarURL
}

// UpdateCoordinates 更新经纬度
func (s *StoreSetting) UpdateCoordinates(coordinates string) {
	s.Coordinates = coordinates
	// 解析经纬度
	latLng := strings.Split(coordinates, ",")
	if len(latLng) == 2 {
		s.Latitude = latLng[0]
		s.Longitude = latLng[1]
	}
}

// AddLanguage 添加语言
func (s *StoreSetting) AddLanguage(language valueobject.LanguageItem) {
	s.Language = append(s.Language, language)
}

// RemoveLanguage 移除语言
func (s *StoreSetting) RemoveLanguage(languageName string) {
	for i, lang := range s.Language {
		if lang.Name == languageName {
			s.Language = append(s.Language[:i], s.Language[i+1:]...)
			break
		}
	}
}

// SetLanguages 设置语言列表
func (s *StoreSetting) SetLanguages(languages []valueobject.LanguageItem) {
	s.Language = languages
}
