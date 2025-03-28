package v1

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type Printer struct {
	PrinterID      uint   `gorm:"primaryKey;autoIncrement;comment:'打印机id'"`
	PrinterName    string `gorm:"default:'';comment:'打印机名称'"`
	PrinterType    string `gorm:"default:'';comment:'打印机类型'"`
	PrinterConfig  string `gorm:"default:'';comment:'打印机配置'"`
	PrintTimes     uint   `gorm:"default:0;comment:'打印联数(次数)'"`
	Sort           uint   `gorm:"default:0;comment:'排序 (数字越小越靠前)'"`
	IsDelete       uint   `gorm:"default:0;comment:'是否删除0=显示1=隐藏'"`
	ShopSupplierID uint   `gorm:"default:0;comment:'商户id'"`
	AppID          uint   `gorm:"default:0;comment:'小程序id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type PrinterRepository interface {
	GetPrinterList() ([]*Printer, error)
	ConvertPrinter() error
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

func (s *PrinterService) ConvertPrinter() error {
	// 创建打印机类型数据
	printerTypes := []model.PrinterType{
		{
			BaseModel: model.BaseModel{
				Uuid: 1,
			},
			Name:       "Codesoft（网口）80mm",
			Key:        "CODESOFT_LAN",
			ConfigJson: `[{"key":"IP","name":"打印机IP"},{"key":"PORT","name":"打印机PORT"}]`,
		},
		{
			BaseModel: model.BaseModel{
				Uuid: 2,
			},
			Name:       "Codesoft（WIFI）80mm",
			Key:        "CODESOFT_WIFI",
			ConfigJson: `[{"key":"IP","name":"打印机IP"},{"key":"PORT","name":"打印机PORT"}]`,
		},
		{
			BaseModel: model.BaseModel{
				Uuid: 3,
			},
			Name:       "商米打印机（云打印）80mm",
			Key:        "SUNMI_CLOUD",
			ConfigJson: `[{"key":"APP_ID","name":"打印机APPID"},{"key":"APP_KEY","name":"打印机APPKEY"},{"key":"SN","name":"打印机SN"}]`,
		},
		{
			BaseModel: model.BaseModel{
				Uuid: 4,
			},
			Name:       "商米打印机（局域网）80mm",
			Key:        "SUNMI_LAN",
			ConfigJson: `[{"key":"IP","name":"打印机IP"},{"key":"SN","name":"打印机SN"}]`,
		},
		{
			BaseModel: model.BaseModel{
				Uuid: 5,
			},
			Name:       "芯烨打印机（有线）80mm",
			Key:        "XPRINTER_LAN",
			ConfigJson: `[{"key":"IP","name":"打印机IP"},{"key":"PORT","name":"打印机PORT"}]`,
		},
		{
			BaseModel: model.BaseModel{
				Uuid: 6,
			},
			Name:       "芯烨打印机（WIFI）80mm",
			Key:        "XPRINTER_WIFI",
			ConfigJson: `[{"key":"IP","name":"打印机IP"},{"key":"PORT","name":"打印机PORT"}]`,
		},
	}

	for _, printerType := range printerTypes {
		_, err := base.NewPrinterTypeRepo(s.targetDB).CreatePrinterType(printerType)
		if err != nil {
			return err
		}
	}
	printers, err := s.GetPrinterList()
	if err != nil {
		return err
	}
	for _, printer := range printers {
		// 档口打印机. 转换商品打印（档口）数据
		printer := model.Printer{
			BaseModel: model.BaseModel{
				Uuid:       uint64(printer.PrinterID),
				CreateTime: printer.CreateTime,
				UpdateTime: printer.UpdateTime,
				DeleteTime: int64(printer.IsDelete),
			},
			Name:            printer.PrinterName,
			PrinterTypeUuid: s.parsePrinterType(printer.PrinterType),
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

func (s *PrinterService) parsePrinterType(printerTypeString string) uint64 {
	switch printerTypeString {
	case "CODESOFT_LAN":
		return 1
	case "CODESOFT_WIFI":
		return 2
	case "SUNMI_CLOUD":
		return 3
	case "SUNMI_LAN":
		return 4
	case "XPRINTER_LAN":
		return 5
	case "XPRINTER_WIFI":
		return 6
	default:
		return 0
	}
}
