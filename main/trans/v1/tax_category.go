package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TaxCategory struct {
	ID         uint    `gorm:"primaryKey;autoIncrement;not null"`
	Name       string  `gorm:"type:varchar(2000);default:'';comment:'名称'"`
	TaxRate    float64 `gorm:"type:decimal(10,2);default:0.00;comment:'税率'"`
	AppID      int     `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime int     `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime int     `gorm:"type:int;not null;default:0;comment:'更新时间'"`
}

func (s *TaxCategory) GetTaxRate() float64 {
	rate := decimal.NewFromFloat(s.TaxRate).Div(decimal.NewFromInt(100)).Round(2).InexactFloat64()
	return rate
}

type TaxCategoryRepository interface {
	GetTaxCategoryList() ([]*TaxCategory, error)
	ConvertTaxCategory() error
}

func NewTaxCategoryService(db *gorm.DB, targetDB *gorm.DB) TaxCategoryRepository {
	return &TaxCategoryService{
		db:       db,
		targetDB: targetDB,
	}
}

type TaxCategoryService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *TaxCategoryService) GetTaxCategoryList() ([]*TaxCategory, error) {
	var taxCategories []*TaxCategory
	err := s.db.Find(&taxCategories).Error
	return taxCategories, err
}

func (s *TaxCategoryService) ConvertTaxCategory() error {
	taxCategories, err := s.GetTaxCategoryList()
	if err != nil {
		return err
	}
	for _, taxCategory := range taxCategories {
		fmt.Println(fmt.Sprintf("taxCategory: %+v", taxCategory))
		tax := model.Tax{
			BaseModel: model.BaseModel{
				Uuid:       uint64(taxCategory.ID),
				CreateTime: int64(taxCategory.CreateTime),
				UpdateTime: int64(taxCategory.UpdateTime),
			},
			Name:    taxCategory.Name,
			TaxRate: taxCategory.GetTaxRate(),
		}
		err := repository.NewTaxRepo(s.targetDB).CreateTax(tax)
		if err != nil {
			return err
		}
	}
	return nil
}
