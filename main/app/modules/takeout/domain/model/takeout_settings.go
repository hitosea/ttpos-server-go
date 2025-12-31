package model

// TakeoutSettings 外卖平台配置表（多平台）
type TakeoutSettings struct {
	BaseModel
	Platform string `gorm:"column:platform" json:"platform"`

	// 基础配置
	IsEnabled int `gorm:"column:is_enabled" json:"is_enabled"`

	// 自动接单配置
	AutoAccept int     `gorm:"column:auto_accept" json:"auto_accept"`
	MaxAmount  float64 `gorm:"column:max_amount" json:"max_amount"`

	// 平台特定配置（JSON 格式）
	PlatformConfig string `gorm:"column:platform_config;type:text" json:"platform_config"`
}

func (*TakeoutSettings) TableName() string {
	return "ttpos_takeout_settings"
}
