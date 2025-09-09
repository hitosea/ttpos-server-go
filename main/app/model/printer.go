package model

import (
	"encoding/json"
	"strings"
)

// Printer 打印机信息表 ttpos_printer
type Printer struct {
	BaseModel
	Name              string `gorm:"column:name;type:varchar(255);comment:打印机名称;NOT NULL" json:"name"`
	PrinterTypeUuid   uint64 `gorm:"column:printer_type_uuid;type:bigint(20) unsigned;default:0;comment:打印机类型ID;NOT NULL" json:"printer_type_uuid"`
	ConfigJson        string `gorm:"column:config_json;type:text;comment:打印机json配置" json:"config_json"`
	Sn                string `gorm:"column:sn;type:varchar(255);comment:打印机SN" json:"sn"`
	Copies            uint   `gorm:"column:copies;type:int(11) unsigned;default:0;comment:打印份数;NOT NULL" json:"copies"`
	Sort              uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序;NOT NULL" json:"sort"`
	IsUsb             int    `gorm:"column:is_usb;type:tinyint(1);default:0;comment:是否usb;NOT NULL" json:"is_usb"`
	Status            int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,1-在线 0-离线;NOT NULL" json:"status"`
	LastHeartbeatTime uint   `gorm:"column:last_heartbeat_time;type:int(10) unsigned;default:0;comment:最后心跳时间;NOT NULL" json:"last_heartbeat_time"`
	SourceDeviceSn    string `gorm:"column:source_device_sn;type:varchar(255);comment:源设备SN;NOT NULL" json:"source_device_sn"`
	PrintMethod       int    `gorm:"column:print_method;type:tinyint(2);default:0;comment:打印方式,1-文本打印,2-图片打印;NOT NULL" json:"print_method"`
	Width             int    `gorm:"column:width;type:int(11) unsigned;default:80;comment:纸张宽度（mm）;NOT NULL" json:"width"`
	EnableStatusCheck int    `gorm:"column:enable_status_check;type:tinyint(1);default:1;comment:是否启用状态检查,0-关闭,1-开启;NOT NULL" json:"enable_status_check"`

	PrinterType *PrinterType `gorm:"foreignKey:PrinterTypeUuid;references:Uuid"` // 关联 printer_type
}

// 是否usb打印机
func (model *Printer) IsUsbPrinter() bool {
	if model == nil {
		return false
	}
	return model.IsUsb == 1
}

type PrinterConfigJson struct {
	IP   string `json:"IP"`
	PORT string `json:"PORT"`
	// 商米打印机信息
	APP_ID  string `json:"APP_ID"`
	APP_KEY string `json:"APP_KEY"`
	SN      string `json:"SN"`
}

// GetConfigJson 获取打印机配置
func (model *Printer) GetConfigJson() PrinterConfigJson {
	// 如果为空则返回空 map
	if model.ConfigJson == "" {
		return PrinterConfigJson{}
	}
	// 处理JSON字符串
	// 1. 去除所有反斜杠
	cleanedJson := strings.ReplaceAll(model.ConfigJson, "\\", "")
	// 2. 去除外层引号 - 这是关键步骤
	cleanedJson = strings.TrimPrefix(cleanedJson, "\"")
	cleanedJson = strings.TrimSuffix(cleanedJson, "\"")
	// 解析打印机配置JSON
	var cfg PrinterConfigJson
	err := json.Unmarshal([]byte(cleanedJson), &cfg)
	if err != nil {
		return PrinterConfigJson{}
	}
	// 返回配置
	return cfg
}

// PrinterType 打印机类型信息表 ttpos_printer_type
type PrinterType struct {
	BaseModel
	Name                  string `gorm:"column:name;type:varchar(255);comment:打印机类型名称;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称ID;NOT NULL" json:"multi_language_name_uuid"`
	Key                   string `gorm:"column:key;type:varchar(255);comment:打印机类型key;NOT NULL" json:"key"`
	ConfigJson            string `gorm:"column:config_json;type:text;comment:打印机类型json配置,描述需要填写的字段" json:"config_json"`

	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}

// ProductPrinter 产品打印机信息表 ttpos_product_printer
type ProductPrinter struct {
	BaseModel
	Name               string `gorm:"column:name;type:varchar(100);comment:名称.厨显上叫档口;NOT NULL" json:"name"`
	Status             int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,1-open开启 1、0-close关闭;NOT NULL" json:"status"`
	PrintMode          int    `gorm:"column:print_mode;type:tinyint(2);default:0;comment:打印模式,0-payment付款打印 1-kitchen送厨打印;NOT NULL" json:"print_mode"`
	PrintMethod        int    `gorm:"column:print_method;type:tinyint(2);default:0;comment:打印方式,-1-全选 0-order整单打印 1-item按一菜一单打印;NOT NULL" json:"print_method"`
	PrintProductSelect int    `gorm:"column:print_product_select;type:tinyint(2);default:0;comment:打印商品选择,0-category按商品分类 1-tag按打印标签;NOT NULL" json:"print_product_select"`
	PrintModeScene     int    `gorm:"column:print_mode_scene;type:tinyint(2);default:0;comment:打印模式场景,0-merge合并 1-separate分开;NOT NULL" json:"print_mode_scene"`
	Copies             uint   `gorm:"column:copies;type:int(11) unsigned;default:1;comment:打印份数;NOT NULL" json:"copies"`

	// 关联模型
	ProductPrinterRegions      []*ProductPrinterRegion      `gorm:"foreignKey:ProductPrinterUuid;references:Uuid"`
	ProductPrinterItems        []*ProductPrinterItem        `gorm:"foreignKey:ProductPrinterUuid;references:Uuid"`
	ProductPrinterProductItems []*ProductPrinterProductItem `gorm:"foreignKey:ProductPrinterUuid;references:Uuid"`
}

// GetPrinterRegionIds 获取打印机关联的区域ID列表
func (model *ProductPrinter) GetPrinterRegionUuids() []uint64 {
	var regionIds []uint64
	for _, printerRegion := range model.ProductPrinterRegions {
		regionIds = append(regionIds, printerRegion.DeskRegionUuid)
	}
	return regionIds
}

// GetPrinterProductIds 获取打印机关联的产品ID列表
func (model *ProductPrinter) GetPrinterProductIds() []uint64 {
	var productIds []uint64
	for _, printerProductItem := range model.ProductPrinterProductItems {
		productIds = append(productIds, printerProductItem.ProductPackageUuid)
	}
	return productIds
}

// ProductPrinterRegion 产品打印机区域信息表 ttpos_product_printer_region
type ProductPrinterRegion struct {
	BaseModel
	ProductPrinterUuid uint64 `gorm:"default:0;column:product_printer_uuid;comment:'产品打印机ID'"`
	DeskRegionUuid     uint64 `gorm:"default:0;column:desk_region_uuid;comment:'桌台区域ID'"`
}

// ProductPrinterItem 商品打印（档口）打印机信息表 ttpos_product_printer_item
type ProductPrinterItem struct {
	BaseModel
	ProductPrinterUuid uint64   `gorm:"default:0;column:product_printer_uuid;comment:'商品打印（档口）ID'"`
	PrinterUuid        uint64   `gorm:"default:0;column:printer_uuid;comment:'打印机ID'"`
	Printer            *Printer `gorm:"foreignKey:PrinterUuid;references:Uuid"`
}

// ProductPrinterProductItem 产品打印机产品明细信息表 ttpos_product_printer_product_item
type ProductPrinterProductItem struct {
	BaseModel
	ProductPrinterUuid uint64 `gorm:"default:0;column:product_printer_uuid;comment:'产品打印机ID'"`
	ProductPackageUuid uint64 `gorm:"default:0;column:product_package_uuid;comment:'产品包ID'"`
}
