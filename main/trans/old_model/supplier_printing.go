package old_model

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type SupplierPrinting struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;comment:'ID'"`
	Name           string `gorm:"default:'';comment:'名称'"`
	CategoryID     string `gorm:"default:'';comment:'菜品id'"`
	IsOpen         uint   `gorm:"default:0;comment:'1开启0关闭'"`
	ProductType    uint   `gorm:"default:0;comment:'0外卖1店内'"`
	ShopSupplierID uint   `gorm:"default:0;comment:'商户id'"`
	PrintType      uint   `gorm:"default:10;comment:'10付款打印20下单打印'"`
	Type           uint   `gorm:"default:10;comment:'10小票打印20标签打印'"`
	PrintMethod    uint   `gorm:"default:10;comment:'10整单打印20商品分组打印30按标签打印'"`
	PrinterID      string `gorm:"default:'';comment:'打印机ids'"`
	LabelID        string `gorm:"default:'';comment:'标签id'"`
	AreaID         string `gorm:"default:'';comment:'区域id'"`
	IsOpenOneFood  uint   `gorm:"default:0;comment:'是否开启一菜一单 0-关闭 1-开启'"`
	PrintSelect    uint   `gorm:"default:1;comment:'打印选择 1-合并打印 2-分开打印'"`
	ProductMethod  uint   `gorm:"default:1;comment:'商品方式 1-按商品分类 2-按打印标签'"`
	ProductIDs     string `gorm:"default:'';comment:'商品ids'"`
	IsDelete       uint   `gorm:"default:0;comment:'是否删除0，否1是'"`
	AppID          uint   `gorm:"default:0;comment:'小程序id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type SupplierPrintingRepository interface {
	GetSupplierPrintingList() ([]*SupplierPrinting, error)
	ConvertSupplierPrinting() error
}

type SupplierPrintingService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *SupplierPrintingService) GetSupplierPrintingList() ([]*SupplierPrinting, error) {
	var supplierPrintings []*SupplierPrinting
	if err := s.db.Find(&supplierPrintings).Error; err != nil {
		return nil, err
	}
	return supplierPrintings, nil
}

func (s *SupplierPrintingService) ConvertSupplierPrinting() error {
	supplierPrintings, err := s.GetSupplierPrintingList()
	if err != nil {
		return err
	}
	for _, supplierPrinting := range supplierPrintings {
		printType := 0
		if supplierPrinting.PrintType == 20 {
			printType = 1
		}

		printProductSelect := 0
		if supplierPrinting.ProductMethod == 2 {
			printProductSelect = 1
		}
		printMethod := 0
		if supplierPrinting.IsOpenOneFood == 1 {
			printMethod = 1
		}
		printModeScene := 0
		if supplierPrinting.PrintSelect == 2 {
			printModeScene = 1
		}
		productPrinter := model.ProductPrinter{
			UUID:               supplierPrinting.ID,
			Name:               supplierPrinting.Name,
			Status:             supplierPrinting.IsOpen,
			PrintMode:          uint(printType),
			PrintMethod:        uint(printMethod),
			PrintProductSelect: uint(printProductSelect),
			PrintModeScene:     uint(printModeScene),
			CreateTime:         supplierPrinting.CreateTime,
			UpdateTime:         supplierPrinting.UpdateTime,
		}
		_, err = base.NewProductPrinterRepo(s.targetDB).CreateProductPrinter(productPrinter)
		if err != nil {
			return err
		}
	}
	return nil
}
