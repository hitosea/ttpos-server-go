package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

// 产品属性组表 `jjjfood_product_attribute_group`
type ProductAttributeGroup struct {
	GroupAttributeID       uint64 `gorm:"primaryKey;autoIncrement;comment:产品属性组ID"`
	ProductID              uint64 `gorm:"default:0;comment:关联产品ID"`
	AttributeID            uint   `gorm:"default:0;comment:关联属性ID"`
	AttributeRequired      uint   `gorm:"default:0;comment:属性组是否必填 0-否 1-是"`
	AttributeOpenMaxSelect uint   `gorm:"default:0;comment:属性组最开启最多可选数量 0-否 1-是"`
	AttributeMaxSelect     uint   `gorm:"default:0;comment:属性组最多可选数量"`
	AttributeMinSelect     uint   `gorm:"default:0;comment:属性组最少可选数量"`
	ShopSupplierID         uint   `gorm:"default:0;comment:店铺ID"`
	AppID                  uint   `gorm:"default:0;comment:应用ID"`
	CreateTime             int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime             int64  `gorm:"autoUpdateTime;comment:更新时间"`

	ProductAttributes []*ProductAttribute `gorm:"foreignKey:GroupAttributeID;references:GroupAttributeID"`
}

type ProductAttributeGroupRepository interface {
	GetProductAttributeGroupList() ([]*ProductAttributeGroup, error)
	ConvertProductAttributeGroup() error
}

type ProductAttributeGroupService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductAttributeGroupService) GetProductAttributeGroupList() ([]*ProductAttributeGroup, error) {
	var productAttributeGroups []*ProductAttributeGroup
	if err := s.db.Find(&productAttributeGroups).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return productAttributeGroups, nil
}

func (s *ProductAttributeGroupService) ConvertProductAttributeGroup() error {
	var productAttributeGroups []*ProductAttributeGroup
	if err := s.db.Find(&productAttributeGroups).Error; err != nil {
		return errors.WithMessage(err)
	}
	for _, productAttributeGroup := range productAttributeGroups {
		fmt.Println(fmt.Sprintf("productAttributeGroup: %+v", productAttributeGroup))
		group := model.ProductPackageAttributeGroup{
			BaseModel: model.BaseModel{
				Uuid:       uint64(productAttributeGroup.GroupAttributeID),
				CreateTime: productAttributeGroup.CreateTime,
				UpdateTime: productAttributeGroup.UpdateTime,
			},
			IsMust:                    productAttributeGroup.AttributeRequired,
			MaxSelection:              productAttributeGroup.AttributeMaxSelect,
			MinSelection:              productAttributeGroup.AttributeMinSelect,
			ProductPackageUuid:        productAttributeGroup.ProductID,
			ProductAttributeGroupUuid: uint64(productAttributeGroup.AttributeID),
		}
		_, err := base.NewProductPackageAttributeGroupRepo(s.targetDB).CreateProductPackageAttributeGroup(group)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
