package model

// CompanyArea 门店区域表 ttpos_company_area
type CompanyArea struct {
	BaseModel
	Name                  string `gorm:"column:name;type:varchar(1000);default:'';comment:区域名称（JSON多语言）;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称UUID;NOT NULL" json:"multi_language_name_uuid"`
	Sort                  int    `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序;NOT NULL" json:"sort"`

	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid" json:"multi_language_name,omitempty"`
}
