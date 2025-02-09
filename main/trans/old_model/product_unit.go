package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

type ProductUnit struct {
	UnitID         uint   `gorm:"primaryKey;autoIncrement;comment:Uuid"`
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

type ProductUnitService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductUnitService) GetProductUnitList() ([]*ProductUnit, error) {
	var productUnits []*ProductUnit
	if err := s.db.Find(&productUnits).Error; err != nil {
		return nil, err
	}
	return productUnits, nil
}

func (s *ProductUnitService) ConvertProductUnit() error {
	productUnits, err := s.GetProductUnitList()
	if err != nil {
		return err
	}
	for _, productUnit := range productUnits {
		fmt.Println(fmt.Sprintf("-------迁移product_unit: %+v", productUnit))
		names := Names{}
		err := names.GetNames(productUnit.UnitName)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("product_unit_id: %d, product_unit_name: %+v", productUnit.UnitID, names))

		id, err := database.GetID()
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(uint(id))
		fmt.Println(fmt.Sprintf("languageName: %+v", languageName))

		unit := model.ProductUnit{
			Uuid:                  productUnit.UnitID,
			Name:                  names.Zh,
			MultiLanguageNameUuid: uint(id),
			CreateTime:            0,
			UpdateTime:            0,
			DeleteTime:            0,
			MultiLanguageName:     languageName,
		}
		_, err = base.NewProductUnitRepo(s.targetDB).CreateProductUnit(unit)
		if err != nil {
			return err
		}
	}
	return nil
}
