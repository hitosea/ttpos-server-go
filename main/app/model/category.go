package model

import "fmt"

// 商品类别
type ProductCategory struct {
	Id                    uint   `gorm:"primaryKey;comment:记录唯一标识符" json:"id"`
	Uuid                  uint   `gorm:"default:0;comment:UUID" json:"uuid"`
	Name                  string `gorm:"default:'';comment:名称" json:"name"`
	ParentUuid            uint   `gorm:"default:NULL;comment:父级ID" json:"parent_uuid"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	Status                bool   `gorm:"default:false;comment:状态, true-开启、false-关闭" json:"status"`
	CategoryKey           string `gorm:"default:'';comment:关键字" json:"category_key"`
	OrderBy               uint   `gorm:"default:0;comment:排序" json:"order_by"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            int64  `gorm:"comment:删除时间（时间戳）" json:"delete_time"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:id" json:"multi_language_name"`
}

func GenCategoryKey(uuid, parentId uint) string {
	if parentId == 0 {
		return fmt.Sprintf("/%d", uuid)
	}
	return fmt.Sprintf("/%d/%d", parentId, uuid)
}

// 原料类别
type MaterialCategory struct {
	Id                    uint   `gorm:"primaryKey;comment:记录唯一标识符" json:"id"`
	Uuid                  uint   `gorm:"default:0;comment:UUID" json:"uuid"`
	Name                  string `gorm:"default:'';comment:名称" json:"name"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	Status                bool   `gorm:"default:false;comment:状态, true-开启、false-关闭" json:"status"`
	Level                 uint   `gorm:"default:0;comment:层级" json:"level"`
	ParentUuid            uint   `gorm:"default:NULL;comment:父级ID" json:"parent_uuid"`
	CategoryKey           string `gorm:"default:'';comment:关键字" json:"category_key"`
	OrderBy               uint   `gorm:"default:0;comment:排序" json:"order_by"`
	RefCount              uint   `gorm:"default:0;comment:关联数量" json:"ref_count"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            int64  `gorm:"comment:删除时间（时间戳）" json:"delete_time"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:id" json:"multi_language_name"`
}

// 商品特殊类别
type ProductSpecialCategory struct {
	Id                    uint   `gorm:"primaryKey;comment:记录唯一标识符" json:"id"`
	Uuid                  uint   `gorm:"default:0;comment:UUID" json:"uuid"`
	Status                bool   `gorm:"default:false;comment:状态, true-开启、false-关闭" json:"status"`
	Name                  string `gorm:"default:'';comment:名称" json:"name"`
	MultiLanguageNameUuid uint   `gorm:"default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	OrderBy               uint   `gorm:"default:0;comment:排序" json:"order_by"`
	CreateTime            int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            int64  `gorm:"comment:删除时间（时间戳）" json:"delete_time"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:id" json:"multi_language_name"`
}
