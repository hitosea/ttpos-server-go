package model

import "encoding/json"

// Setting 设置表 ttpos_setting
type Setting struct {
	Key        string `gorm:"column:key;type:varchar(30);comment:设置项标示;NOT NULL" json:"key"`
	Describe   string `gorm:"column:describe;type:varchar(255);comment:设置项描述;NOT NULL" json:"describe"`
	Values     string `gorm:"column:values;type:mediumtext;comment:设置内容（json格式）;NOT NULL" json:"values"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;type:int(10);comment:'创建时间(时间戳)'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:'更新时间(时间戳)'"`
	DeleteTime int64  `gorm:"column:delete_time;type:int(10);default:0;comment:'删除时间(时间戳)'"`
}

const (
	SettingKeyPurchaseBrandDailyLimit = "purchase_brand_daily_limit"
)

func (s *Setting) GetPurchaseBrandDailyLimit() int {
	if s.Values == "" {
		return -1
	}
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(s.Values), &config); err != nil {
		return -1
	}
	if limitVal, ok := config["limit"]; ok {
		switch v := limitVal.(type) {
		case float64:
			return int(v)
		case int:
			return int(v)
		default:
			return -1
		}
	}
	return -1
}
