package model

// ProductPackageGroup 商品套餐组模型 `ttpos_product_package_group`
type ProductPackageGroup struct {
	Id                    uint64 `json:"id" gorm:"primaryKey;autoIncrement;comment:自增ID"`
	Uuid                  uint64 `json:"uuid" gorm:"uniqueIndex:unique_uuid;not null;default:0;comment:商品套餐组ID"`
	Name                  string `json:"name" gorm:"type:text;comment:名称"`
	MultiLanguageNameUuid uint64 `json:"multi_language_name_uuid" gorm:"index:idx_multi_language_name_uuid;not null;default:0;comment:多语言名称ID"`
	ProductPackageUuid    uint64 `json:"product_package_uuid" gorm:"index:idx_product_package_uuid;not null;default:0;comment:商品套餐UUID"`
	CreateTime            int    `json:"create_time" gorm:"not null;default:0;comment:创建时间(时间戳)"`
	UpdateTime            int    `json:"update_time" gorm:"not null;default:0;comment:更新时间(时间戳)"`
	DeleteTime            int    `json:"delete_time" gorm:"not null;default:0;comment:删除时间(时间戳)"`

	ProductPackageGroupItems []ProductPackageGroupItem `gorm:"foreignKey:product_package_group_uuid;references:uuid"` // 商品套餐组商品
	MultiLanguageName        MultiLanguageName         `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`   // 多语言名称
	ProductPackage           *ProductPackage           `gorm:"foreignKey:product_package_uuid;references:uuid"`       // 商品套餐
}
