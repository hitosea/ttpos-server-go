package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ProductPrintLabel struct {
	LabelID        uint64 `gorm:"primaryKey;autoIncrement;comment:Uuid"`
	LabelName      string `gorm:"default:'';comment:属性名"`
	ShopSupplierID uint   `gorm:"default:0;comment:门店id"`
	Sort           uint   `gorm:"default:0;comment:排序"`
	AppID          uint   `gorm:"default:0;unsigned;comment:应用id"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type ProductPrintLabelRepository interface {
	GetProductPrintLabelList() ([]*ProductPrintLabel, error)
	ConvertProductPrintLabel() error
}

type ProductPrintLabelService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductPrintLabelService) GetProductPrintLabelList() ([]*ProductPrintLabel, error) {
	var productPrintLabels []*ProductPrintLabel
	if err := s.db.Find(&productPrintLabels).Error; err != nil {
		return nil, err
	}
	return productPrintLabels, nil
}

func (s *ProductPrintLabelService) ConvertProductPrintLabel() error {
	productPrintLabels, err := s.GetProductPrintLabelList()
	if err != nil {
		return err
	}
	for _, productPrintLabel := range productPrintLabels {
		fmt.Println(fmt.Sprintf("-------迁移product_print_label: %+v", productPrintLabel))

		id, err := utils.GetID()
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		tag := model.PrinterTag{
			Uuid: productPrintLabel.LabelID,
			Name: productPrintLabel.LabelName,
		}
		_, err = base.NewPrinterTagRepo(s.targetDB).CreatePrinterTag(tag)
		if err != nil {
			return err
		}
	}
	return nil
}
