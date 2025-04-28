package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type Printer struct {
	PrinterID      uint   `gorm:"column:printer_id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:打印机id" json:"printer_id"`
	PrinterName    string `gorm:"column:printer_name;type:varchar(255);comment:打印机名称;NOT NULL" json:"printer_name"`
	PrinterType    string `gorm:"column:printer_type;type:varchar(255);comment:打印机类型;NOT NULL" json:"printer_type"`
	PrinterConfig  string `gorm:"column:printer_config;type:text;comment:打印机配置;NOT NULL" json:"printer_config"`
	PrintTimes     uint   `gorm:"column:print_times;type:smallint(6) unsigned;default:0;comment:打印联数(次数);NOT NULL" json:"print_times"`
	Sort           uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序 (数字越小越靠前);NOT NULL" json:"sort"`
	IsDelete       uint   `gorm:"column:is_delete;type:int(11) unsigned;default:0;comment:是否删除0=显示1=隐藏;NOT NULL" json:"is_delete"`
	ShopSupplierId int    `gorm:"column:shop_supplier_id;type:int(11);comment:商户id;NOT NULL" json:"shop_supplier_id"`
	AppId          uint   `gorm:"column:app_id;type:int(11) unsigned;default:0;comment:小程序id;NOT NULL" json:"app_id"`
	CreateTime     uint   `gorm:"column:create_time;type:int(11) unsigned;default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     uint   `gorm:"column:update_time;type:int(11) unsigned;default:0;comment:更新时间;NOT NULL" json:"update_time"`
}

type PrinterRepository interface {
	GetPrinterList() ([]*Printer, error)
	GetPrinterTypeList() ([]*model.PrinterType, error)
	ConvertPrinter() error
}

func NewPrinterService(db *gorm.DB, targetDB *gorm.DB) PrinterRepository {
	return &PrinterService{db: db, targetDB: targetDB}
}

type PrinterService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *PrinterService) GetPrinterList() ([]*Printer, error) {
	var printers []*Printer
	if err := s.db.Find(&printers).Error; err != nil {
		return nil, err
	}
	return printers, nil
}

func (s *PrinterService) GetPrinterTypeList() ([]*model.PrinterType, error) {
	var printerTypes []*model.PrinterType
	if err := s.targetDB.Find(&printerTypes).Error; err != nil {
		return nil, err
	}
	return printerTypes, nil
}

func (s *PrinterService) ConvertPrinter() error {
	printers, err := s.GetPrinterList()
	printerTypes, err := s.GetPrinterTypeList()
	if err != nil {
		return err
	}
	for _, printer := range printers {

		fmt.Println(fmt.Sprintf("printer: %+v", printer))

		// 档口打印机
		printer := model.Printer{
			BaseModel: model.BaseModel{
				Uuid:       uint64(printer.PrinterID),
				CreateTime: int64(printer.CreateTime),
				UpdateTime: int64(printer.UpdateTime),
				DeleteTime: int64(printer.IsDelete),
			},
			Name:            printer.PrinterName,
			PrinterTypeUuid: s.parsePrinterType(printer.PrinterType, printerTypes),
			ConfigJson:      printer.PrinterConfig,
			Copies:          printer.PrintTimes,
			Sort:            printer.Sort,
		}
		_, err = base.NewPrinterRepo(s.targetDB).CreatePrinter(printer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PrinterService) parsePrinterType(printerTypeString string, printerTypes []*model.PrinterType) uint64 {
	for _, printerType := range printerTypes {
		if printerTypeString == printerType.Key {
			return printerType.Uuid
		}
	}
	return 0
}
