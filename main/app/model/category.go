package model

import "fmt"

// 商品类别
type ProductCategory struct {
	Id                    uint   `gorm:"primaryKey;not null;comment:记录唯一标识符" json:"id"`
	Uuid                  uint   `gorm:"not null;default:0;comment:UUID" json:"uuid"`
	Name                  string `gorm:"not null;default:'';comment:名称" json:"name"`
	ParentUuid            uint   `gorm:"default:NULL;comment:父级ID" json:"parent_uuid"`
	MultiLanguageNameUuid uint   `gorm:"not null;default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	Status                bool   `gorm:"not null;default:true;comment:状态, true-开启、false-关闭" json:"status"`
	Level                 uint   `gorm:"not null;default:1;comment:层级" json:"level"`
	CategoryKey           string `gorm:"not null;default:'';comment:关键字" json:"category_key"`
	OrderBy               uint   `gorm:"not null;default:0;comment:排序" json:"order_by"`
	CreateTime            uint   `gorm:"not null;default:0;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            uint   `gorm:"not null;default:0;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            uint   `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`

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
	Id                    uint   `gorm:"primaryKey;not null;comment:记录唯一标识符" json:"id"`
	Uuid                  uint   `gorm:"not null;default:0;comment:UUID" json:"uuid"`
	Name                  string `gorm:"not null;default:'';comment:名称" json:"name"`
	MultiLanguageNameUuid uint   `gorm:"not null;default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	Status                bool   `gorm:"not null;default:true;comment:状态, true-开启、false-关闭" json:"status"`
	Level                 uint   `gorm:"not null;default:0;comment:层级" json:"level"`
	ParentUuid            uint   `gorm:"default:NULL;comment:父级ID" json:"parent_uuid"`
	CategoryKey           string `gorm:"not null;default:'';comment:关键字" json:"category_key"`
	OrderBy               uint   `gorm:"not null;default:0;comment:排序" json:"order_by"`
	RefCount              uint   `gorm:"not null;default:0;comment:关联数量" json:"ref_count"`
	CreateTime            uint   `gorm:"not null;default:0;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            uint   `gorm:"not null;default:0;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            uint   `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:id" json:"multi_language_name"`
}

// 商品特殊类别
type ProductSpecialCategory struct {
	Id                    uint   `gorm:"primaryKey;not null;comment:记录唯一标识符" json:"id"`
	Uuid                  uint   `gorm:"not null;default:0;comment:UUID" json:"uuid"`
	Status                bool   `gorm:"not null;default:true;comment:状态, true-开启、false-关闭" json:"status"`
	Name                  string `gorm:"not null;default:'';comment:名称" json:"name"`
	MultiLanguageNameUuid uint   `gorm:"not null;default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	OrderBy               uint   `gorm:"not null;default:0;comment:排序" json:"order_by"`
	CreateTime            uint   `gorm:"not null;default:0;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            uint   `gorm:"not null;default:0;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            uint   `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:id" json:"multi_language_name"`
}
