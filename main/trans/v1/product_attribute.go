package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type ProductAttribute struct {
	ProductAttributeID uint64 `gorm:"primaryKey;autoIncrement;comment:产品属性ID"`
	ProductID          uint   `gorm:"default:0;comment:关联产品ID"`
	GroupAttributeID   uint64 `gorm:"default:0;comment:关联产品属性组ID"`
	AttributeID        uint64 `gorm:"default:0;comment:关联属性ID"`
	DefaultSelect      uint   `gorm:"default:0;comment:默认勾选 0-否 1-是"`
	ShopSupplierID     uint   `gorm:"default:0;comment:店铺ID"`
	AppID              uint   `gorm:"default:0;comment:应用ID"`
	CreateTime         int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime         int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type ProductAttributeRepository interface {
	GetProductAttributeList() ([]*ProductAttribute, error)
	ConvertProductAttribute() error
}

type ProductAttributeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductAttributeService) GetProductAttributeList() ([]*ProductAttribute, error) {
	var productAttributes []*ProductAttribute
	if err := s.db.Find(&productAttributes).Error; err != nil {
		return nil, err
	}
	return productAttributes, nil
}

func (s *ProductAttributeService) ConvertProductAttribute() error {
	var productAttributes []*ProductAttribute
	if err := s.db.Find(&productAttributes).Error; err != nil {
		return err
	}
	for _, productAttribute := range productAttributes {
		fmt.Println(fmt.Sprintf("productAttribute: %+v", productAttribute))
		attribute := model.ProductPackageAttribute{
			BaseModel: model.BaseModel{
				Uuid:       uint64(productAttribute.ProductAttributeID),
				CreateTime: productAttribute.CreateTime,
				UpdateTime: productAttribute.UpdateTime,
			},
			ProductPackageAttributeGroupUuid: productAttribute.GroupAttributeID,
			AttributeUuid:                    productAttribute.AttributeID,
			IsDefaultSelected:                uint(productAttribute.DefaultSelect),
		}
		_, err := base.NewProductPackageAttributeRepo(s.targetDB).CreateProductPackageAttribute(attribute)
		if err != nil {
			return err
		}
	}
	return nil
}
