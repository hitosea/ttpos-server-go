package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type PrinterTemplate struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;comment:'打印机模板表唯一标识符'"`
	Name           string `gorm:"type:varchar(255);default:'';comment:'打印名称'"`
	Template       int    `gorm:"type:int;default:1;comment:'模板选择'"`
	ShopSupplierID int    `gorm:"type:int;default:0;comment:'店铺id'"`
	AppID          int    `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type PrinterTemplateRepository interface {
	GetPrinterTemplateList() ([]*PrinterTemplate, error)
	ConvertPrinterTemplate() error
}

func NewPrinterTemplateService(db *gorm.DB, targetDB *gorm.DB) PrinterTemplateRepository {
	return &PrinterTemplateService{db: db, targetDB: targetDB}
}

type PrinterTemplateService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *PrinterTemplateService) GetPrinterTemplateList() ([]*PrinterTemplate, error) {
	var printerTemplates []*PrinterTemplate
	err := s.db.Find(&printerTemplates).Error
	return printerTemplates, err
}

func (s *PrinterTemplateService) ConvertPrinterTemplate() error {
	printerTemplates, err := s.GetPrinterTemplateList()
	if err != nil {
		return err
	}
	for _, printerTemplate := range printerTemplates {
		fmt.Println(fmt.Sprintf("printerTemplate: %+v", printerTemplate))

		printerTemplate := model.PrinterTemplate{
			Uuid:       uint64(printerTemplate.ID),
			Name:       printerTemplate.Name,
			Template:   printerTemplate.Template,
			CreateTime: printerTemplate.CreateTime,
			UpdateTime: printerTemplate.UpdateTime,
		}
		if err := repository.NewPrinterTemplateRepo(s.targetDB).CreatePrinterTemplate(printerTemplate); err != nil {
			return err
		}
	}
	return nil
}
