package v1

import (
	"fmt"
	"strconv"
	"strings"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/utils"

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

func NewSupplierPrintingService(db *gorm.DB, targetDB *gorm.DB) SupplierPrintingRepository {
	return &SupplierPrintingService{
		db:       db,
		targetDB: targetDB,
	}
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
		if supplierPrinting.PrintMethod == 40 {
			printMethod = 1
		}
		printModeScene := 0
		if supplierPrinting.PrintSelect == 2 {
			printModeScene = 1
		}
		// 档口产品关联ID列表
		productIDList, err := s.parseIdList(supplierPrinting.ProductIDs)
		if err != nil {
			return err
		}
		for _, productID := range productIDList {
			id, err := utils.GetID()
			if err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("id: %d", id))
			// 转换档口产品关联数据
			productPrinterProductItem := model.ProductPrinterProductItem{
				BaseModel: model.BaseModel{
					Uuid:       uint64(id),
					CreateTime: supplierPrinting.CreateTime,
					UpdateTime: supplierPrinting.UpdateTime,
					DeleteTime: int64(supplierPrinting.IsDelete),
				},
				ProductPrinterUuid: uint64(supplierPrinting.ID),
				ProductPackageUuid: uint64(productID),
			}
			_, err = base.NewProductPrinterProductItemRepo(s.targetDB).CreateProductPrinterProductItem(productPrinterProductItem)
			if err != nil {
				return err
			}
		}

		// 档口打印机ID列表
		printerIDList, err := s.parseIdList(supplierPrinting.PrinterID)
		if err != nil {
			return err
		}
		for _, printerID := range printerIDList {
			id, err := utils.GetID()
			if err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("id: %d", id))
			// 转换档口打印机数据
			productPrinterItem := model.ProductPrinterItem{
				BaseModel: model.BaseModel{
					Uuid:       uint64(id),
					CreateTime: supplierPrinting.CreateTime,
					UpdateTime: supplierPrinting.UpdateTime,
					DeleteTime: int64(supplierPrinting.IsDelete),
				},
				ProductPrinterUuid: uint64(supplierPrinting.ID),
				PrinterUuid:        uint64(printerID),
			}
			_, err = base.NewProductPrinterItemRepo(s.targetDB).CreateProductPrinterItem(productPrinterItem)
			if err != nil {
				return err
			}
		}
		// 档口区域ID列表
		regionIDList, err := s.parseIdList(supplierPrinting.AreaID)
		if err != nil {
			return err
		}
		for _, regionID := range regionIDList {
			id, err := utils.GetID()
			if err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("id: %d", id))
			// 转换档口区域数据
			productPrinterRegion := model.ProductPrinterRegion{
				BaseModel: model.BaseModel{
					Uuid:       uint64(id),
					CreateTime: supplierPrinting.CreateTime,
					UpdateTime: supplierPrinting.UpdateTime,
				},
				ProductPrinterUuid: uint64(supplierPrinting.ID),
				DeskRegionUuid:     uint64(regionID),
			}
			_, err = base.NewProductPrinterRegionRepo(s.targetDB).CreateProductPrinterRegion(productPrinterRegion)
			if err != nil {
				return err
			}
		}
		// 档口打印机. 转换商品打印（档口）数据
		productPrinter := model.ProductPrinter{
			BaseModel: model.BaseModel{
				Uuid:       uint64(supplierPrinting.ID),
				CreateTime: supplierPrinting.CreateTime,
				UpdateTime: supplierPrinting.UpdateTime,
			},
			Name:               supplierPrinting.Name,
			Status:             int(supplierPrinting.IsOpen),
			PrintMode:          int(printType),
			PrintMethod:        int(printMethod),
			PrintProductSelect: int(printProductSelect),
			PrintModeScene:     int(printModeScene),
		}
		_, err = base.NewProductPrinterRepo(s.targetDB).CreateProductPrinter(productPrinter)
		if err != nil {
			return err
		}
	}
	return nil
}

// 将字符串数据转为数组
func (s *SupplierPrintingService) parseIdList(regionArrayString string) ([]uint, error) {
	if regionArrayString == "" {
		return []uint{}, nil
	}
	if regionArrayString == "\"\"" {
		return []uint{}, nil
	}
	regionIDs := []uint{}
	regionArrayString = strings.ReplaceAll(regionArrayString, `"`, "")
	regionArrayString = strings.Trim(regionArrayString, "[")
	regionArrayString = strings.Trim(regionArrayString, "]")
	regionArray := strings.Split(regionArrayString, ",")
	for _, region := range regionArray {
		regionID, err := strconv.ParseUint(region, 10, 64)
		if err != nil {
			return nil, err
		}
		regionIDs = append(regionIDs, uint(regionID))
	}
	return regionIDs, nil
}
