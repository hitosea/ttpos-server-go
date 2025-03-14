package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Printer 打印机信息表 ttpos_printer
type Printer struct {
	BaseModel
	Name            string `gorm:"column:name;type:varchar(255);comment:打印机名称;NOT NULL" json:"name"`
	PrinterTypeUuid uint64 `gorm:"column:printer_type_uuid;type:bigint(20) unsigned;default:0;comment:打印机类型ID;NOT NULL" json:"printer_type_uuid"`
	ConfigJson      string `gorm:"column:config_json;type:text;comment:打印机json配置" json:"config_json"`
	Copies          uint   `gorm:"column:copies;type:int(11) unsigned;default:0;comment:打印份数;NOT NULL" json:"copies"`
	Sort            uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序;NOT NULL" json:"sort"`

	PrinterType *PrinterType `gorm:"foreignKey:PrinterTypeUuid;references:Uuid"` // 关联 printer_type
}

type PrinterConfigJson struct {
	IP      string `json:"IP"`
	PORT    string `json:"PORT"`
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

// PrinterLog 打印日志表 ttpos_printer_log
type PrinterLog struct {
	BaseModel
	PrinterUuid     uint64 `gorm:"column:printer_uuid;type:bigint(20) unsigned;default:0;comment:打印机id;NOT NULL" json:"printer_uuid"`
	CashierDeviceId string `gorm:"column:cashier_device_id;type:varchar(255);comment:收银机绑定的id;NOT NULL" json:"cashier_device_id"`
	RelatedType     int    `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售账单；1-销售订单, 2-充值订单;NOT NULL" json:"related_type"`
	RelatedUuid     uint64 `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:销售账单、充值订单id;NOT NULL" json:"related_uuid"`
	Data            string `gorm:"column:data;type:varchar(255);comment:打印数据" json:"data"`
	Type            int    `gorm:"column:type;type:int(11);default:0;comment:类型:0系统默认队列,1云上服务下放;NOT NULL" json:"type"`
	DataType        int    `gorm:"column:data_type;type:tinyint(2);default:1;comment:数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单 8-充值单 9-退菜单;NOT NULL" json:"data_type"`
	PrintMethod     int    `gorm:"column:print_method;type:tinyint(2);default:1;comment:打印方式 1文本打印, 2图片打印;NOT NULL" json:"print_method"`
	Num             int    `gorm:"column:num;type:int(11);default:0;comment:打印次数;NOT NULL" json:"num"`
	Status          int    `gorm:"column:status;type:tinyint(2);default:1;comment:状态(0结束,1进行中,2成功);NOT NULL" json:"status"`
	Reason          string `gorm:"column:reason;type:varchar(255);comment:原因" json:"reason"`
	PrinterTime     int64  `gorm:"column:printer_time;type:int(11);default:0;comment:打印时间;NOT NULL" json:"printer_time"`
	FirstExecution  int    `gorm:"column:first_execution;type:tinyint(2);default:0;comment:是否首次执行打印 1-是 0-否;NOT NULL" json:"first_execution"`

	Printer  *Printer  `gorm:"foreignKey:PrinterUuid;references:Uuid"` // 关联 printer
	SaleBill *SaleBill `gorm:"foreignKey:RelatedUuid;references:Uuid"` // 关联 sale_order
}

// 压缩数据
func (model *PrinterLog) CompressData() string {
	data := model.Data
	if data == "" {
		return data
	}
	var result strings.Builder
	zeroCount := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '0' {
			zeroCount++
			// 检查后面是否还有字符
			if i == len(data)-1 || data[i+1] != '0' {
				if zeroCount >= 10 {
					result.WriteString(fmt.Sprintf("-zero%d-", zeroCount))
				} else {
					result.WriteString(strings.Repeat("0", zeroCount))
				}
				zeroCount = 0
			}
		} else {
			if zeroCount > 0 {
				if zeroCount >= 10 {
					result.WriteString(fmt.Sprintf("-zero%d-", zeroCount))
				} else {
					result.WriteString(strings.Repeat("0", zeroCount))
				}
				zeroCount = 0
			}
			result.WriteByte(data[i])
		}
	}
	return result.String()
}

// 还原压缩的数据
func (model *PrinterLog) DecompressData() string {
	data := model.Data
	if data == "" {
		return data
	}
	// 使用正则表达式匹配 "-zeroX-" 格式的字符串，其中X是数字
	re := regexp.MustCompile(`-zero(\d+)-`)
	result := re.ReplaceAllStringFunc(data, func(match string) string {
		// 提取数字部分
		numStr := re.FindStringSubmatch(match)[1]
		// 转换为数字
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return match
		}
		// 返回对应数量的0
		return strings.Repeat("0", num)
	})

	return result
}

// PrinterReadLog 打印读取日志表 ttpos_printer_read_log
type PrinterReadLog struct {
	BaseModel
	LogUuid  uint64 `gorm:"column:log_uuid;type:int(11);default:0;comment:打印uuid" json:"log_uuid"`
	DeviceId string `gorm:"column:device_id;type:varchar(255);comment:设备id" json:"device_id"`
}

// ProductPrinter 产品打印机信息表 ttpos_product_printer
type ProductPrinter struct {
	BaseModel
	Name               string `gorm:"column:name;type:varchar(100);comment:名称.厨显上叫档口;NOT NULL" json:"name"`
	Status             int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,1-open开启 1、0-close关闭;NOT NULL" json:"status"`
	PrintMode          int    `gorm:"column:print_mode;type:tinyint(2);default:0;comment:打印模式,0-payment付款打印 1-kitchen送厨打印;NOT NULL" json:"print_mode"`
	PrintMethod        int    `gorm:"column:print_method;type:tinyint(2);default:0;comment:打印方式,0-order整单打印 1-item按一菜一单打印;NOT NULL" json:"print_method"`
	PrintProductSelect int    `gorm:"column:print_product_select;type:tinyint(2);default:0;comment:打印商品选择,0-category按商品分类 1-tag按打印标签;NOT NULL" json:"print_product_select"`
	PrintModeScene     int    `gorm:"column:print_mode_scene;type:tinyint(2);default:0;comment:打印模式场景,0-merge合并 1-separate分开;NOT NULL" json:"print_mode_scene"`

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
