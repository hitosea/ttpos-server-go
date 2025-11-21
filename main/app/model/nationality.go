package model

// Nationality 国籍 ttpos_nationality
type Nationality struct {
	BaseModel
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称UUID;NOT NULL" json:"multi_language_name_uuid"`
	Sort                  int    `gorm:"column:sort;type:int(11);default:0;comment:排序;NOT NULL" json:"sort"`
	Status                int    `gorm:"column:status;type:tinyint(1);default:1;comment:状态：1-启用；0-禁用;NOT NULL" json:"status"`

	// 关联多语言名称表
	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid" json:"multi_language_name,omitempty"`
}

// TableName 指定表名
func (Nationality) TableName() string {
	return "ttpos_nationality"
}
