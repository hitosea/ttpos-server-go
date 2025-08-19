package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type Attribute struct {
	AttributeID    uint64 `gorm:"column:attribute_id;primaryKey;autoIncrement;comment:'属性ID'"`
	ParentID       uint64 `gorm:"column:parent_id;default:0;comment:'父级ID'"`
	AttributeName  string `gorm:"column:attribute_name;type:text;not null;default:'';comment:'属性名称'"`
	ShopSupplierID int    `gorm:"column:shop_supplier_id;default:0;comment:'店铺ID'"`
	AppID          int    `gorm:"column:app_id;default:0;comment:'应用ID'"`
	CreateTime     int64  `gorm:"column:create_time;not null;default:0;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"column:update_time;not null;default:0;comment:'更新时间'"`
}

type AttributeInterface interface {
	GetAttributeList() ([]Attribute, error)
	ConvertAttribute() error
}

func NewAttributeService(db *gorm.DB, targetDB *gorm.DB) AttributeInterface {
	return &AttributeRepository{db: db, targetDB: targetDB}
}

type AttributeRepository struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (r *AttributeRepository) GetAttributeList() ([]Attribute, error) {
	var attributes []Attribute
	err := r.db.Order("create_time asc").Find(&attributes).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return attributes, nil
}

func (r *AttributeRepository) ConvertAttribute() error {
	attributes, err := r.GetAttributeList()
	if err != nil {
		return errors.WithMessage(err)
	}

	groupSort := 1
	sort := 1

	for _, attribute := range attributes {
		fmt.Println(fmt.Sprintf("-------迁移attribute: %+v", attribute))
		names := Names{}
		err := names.GetNames(attribute.AttributeName)
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("attribute_id: %d, attribute_name: %+v", attribute.AttributeID, names))

		id, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(id)
		if attribute.ParentID == 0 {
			// 创建商品属性组
			attributeGroup := model.ProductAttributeGroup{
				BaseModel:             model.BaseModel{Uuid: attribute.AttributeID},
				Name:                  attribute.AttributeName,
				MultiLanguageNameUuid: id,
				MultiLanguageName:     languageName,
				Sort:                  groupSort,
			}
			fmt.Println(fmt.Sprintf("attribute_group: %+v", attributeGroup))
			_, err = base.NewProductAttributeGroupRepo(r.targetDB).CreateProductAttributeGroup(attributeGroup)
			if err != nil {
				return errors.WithMessage(err)
			}
			groupSort++
		} else {
			// 创建商品属性
			productAttribute := model.ProductAttribute{
				BaseModel:          model.BaseModel{Uuid: attribute.AttributeID},
				Name:               attribute.AttributeName,
				AttributeGroupUuid: attribute.ParentID,
				MultiLanguageName:  languageName,
				Sort:               sort,
			}
			_, err = base.NewProductAttributeRepo(r.targetDB).CreateProductAttribute(productAttribute)
			if err != nil {
				return errors.WithMessage(err)
			}
			sort++
		}
	}
	return nil
}
