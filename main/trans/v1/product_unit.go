package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ProductUnit struct {
	UnitID         uint64 `gorm:"primaryKey;autoIncrement;comment:Uuid"`
	UnitName       string `gorm:"default:'';comment:属性名"`
	ShopSupplierID uint   `gorm:"default:0;comment:门店id"`
	Sort           uint   `gorm:"default:0;comment:排序"`
	AppID          uint   `gorm:"default:0;unsigned;comment:应用id"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type ProductUnitRepository interface {
	GetProductUnitList() ([]*ProductUnit, error)
	ConvertProductUnit() error
}

func NewProductUnitService(db *gorm.DB, targetDB *gorm.DB) ProductUnitRepository {
	return &ProductUnitService{db: db, targetDB: targetDB}
}

type ProductUnitService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductUnitService) GetProductUnitList() ([]*ProductUnit, error) {
	var productUnits []*ProductUnit
	if err := s.db.Order("create_time asc").Find(&productUnits).Error; err != nil {
		return nil, err
	}
	return productUnits, nil
}

func (s *ProductUnitService) ConvertProductUnit() error {
	productUnits, err := s.GetProductUnitList()
	if err != nil {
		return errors.WithMessage(err)
	}
	// v2.5增加排序字段
	sort := 1
	for _, productUnit := range productUnits {
		fmt.Println(fmt.Sprintf("-------迁移product_unit: %+v", productUnit))
		names := Names{}
		err := names.GetNames(productUnit.UnitName)
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("product_unit_id: %d, product_unit_name: %+v", productUnit.UnitID, names))

		id, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(id)
		fmt.Println(fmt.Sprintf("languageName: %+v", languageName))

		unit := model.ProductUnit{
			BaseModel: model.BaseModel{
				Uuid:       uint64(productUnit.UnitID),
				CreateTime: productUnit.CreateTime,
				UpdateTime: productUnit.UpdateTime,
			},
			Sort:                  sort,
			Name:                  productUnit.UnitName,
			MultiLanguageNameUuid: uint64(id),
			MultiLanguageName:     languageName,
		}
		_, err = base.NewProductUnitRepo(s.targetDB).CreateProductUnit(unit)
		if err != nil {
			return errors.WithMessage(err)
		}
		sort++
	}
	return nil
}
