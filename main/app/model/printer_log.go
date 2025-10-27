package model

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"ttpos-server-go/app/constant"

	"github.com/google/uuid"
)

// PrinterLog 打印日志表 ttpos_printer_log
type PrinterLog struct {
	BaseModel
	PrinterUuid        uint64 `gorm:"column:printer_uuid;type:bigint(20) unsigned;default:0;comment:打印机id;NOT NULL" json:"printer_uuid"`
	CashierDeviceId    string `gorm:"column:cashier_device_id;type:varchar(255);comment:收银机绑定的id;NOT NULL" json:"cashier_device_id"`
	RelatedType        int    `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售账单；1-销售订单, 2-充值订单, 3-交班单;NOT NULL" json:"related_type"`
	RelatedUuid        uint64 `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:销售账单、充值订单id;NOT NULL" json:"related_uuid"`
	Data               string `gorm:"column:data;type:varchar(255);comment:打印数据" json:"data"`
	Type               int    `gorm:"column:type;type:int(11);default:0;comment:类型:0系统默认队列,1云上服务下放;NOT NULL" json:"type"`
	DataType           int    `gorm:"column:data_type;type:tinyint(2);default:1;comment:数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单 8-充值单 9-退菜单;NOT NULL" json:"data_type"`
	PrintMethod        int    `gorm:"column:print_method;type:tinyint(2);default:1;comment:打印方式 1文本打印, 2图片打印;NOT NULL" json:"print_method"`
	Num                int    `gorm:"column:num;type:int(11);default:0;comment:打印次数;NOT NULL" json:"num"`
	Status             int    `gorm:"column:status;type:tinyint(2);default:1;comment:状态(0结束,1进行中,2成功);NOT NULL" json:"status"`
	Reason             string `gorm:"column:reason;type:varchar(255);comment:原因" json:"reason"`
	PrinterTime        int64  `gorm:"column:printer_time;type:int(11);default:0;comment:打印时间;NOT NULL" json:"printer_time"`
	FirstExecution     int    `gorm:"column:first_execution;type:tinyint(2);default:0;comment:是否首次执行打印 1-是 0-否;NOT NULL" json:"first_execution"`
	ReadDeviceId       string `gorm:"column:read_device_id;type:varchar(255);comment:读取设备id;NOT NULL" json:"read_device_id"`
	ProductPrinterUuid uint64 `gorm:"column:product_printer_uuid;type:bigint(20) unsigned;default:0;comment:商品打印UUID;NOT NULL" json:"product_printer_uuid"`
	PrinterType        string `gorm:"column:printer_type;type:varchar(50);default:'';comment:打印机类型;NOT NULL" json:"printer_type"`
	PrintingTime       int64  `gorm:"column:printing_time;type:int(11);default:0;comment:打印耗时;NOT NULL" json:"printing_time"`
	Copies             uint   `gorm:"column:copies;type:int(11) unsigned;default:0;comment:打印份数;NOT NULL" json:"print_copies"`

	Printer             *Printer             `gorm:"foreignKey:PrinterUuid;references:Uuid"`        // 关联 printer
	SaleBill            *SaleBill            `gorm:"foreignKey:RelatedUuid;references:Uuid"`        // 关联 sale_order
	SaleOrder           *SaleOrder           `gorm:"foreignKey:RelatedUuid;references:Uuid"`        // 关联 sale_order
	ProductPrinter      *ProductPrinter      `gorm:"foreignKey:ProductPrinterUuid;references:Uuid"` // 关联 sale_order
	MemberRechargeOrder *MemberRechargeOrder `gorm:"foreignKey:RelatedUuid;references:Uuid"`        // 关联 sale_order
	PrinterLogData      *PrinterLogData      `gorm:"foreignKey:LogUuid;references:Uuid"`            // 关联 printer_log_data
}

// GetTradeNo 获取交易号
func (model *BaseModel) GetTradeNo(CompanyUuid uint64) string {
	tradeNo := fmt.Sprintf("%d%d", CompanyUuid, model.Uuid)
	if len(tradeNo) > 32 {
		tradeNo = tradeNo[len(tradeNo)-32:]
	}
	return tradeNo
}

// GetRandomTradeNo 获取交易号
func (model *PrinterLog) GetRandomTradeNo() string {
	return fmt.Sprintf("%d%s", time.Now().Unix(), hex.EncodeToString([]byte(strings.Replace(uuid.New().String(), "-", "", -1)))[:8])
}

// 压缩数据
func (model *PrinterLog) CompressData() string {
	data := model.Data
	if data == "" {
		return data
	}
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression) // 或使用gzip.DefaultCompression
	if err != nil {
		return ""
	}
	_, err = zw.Write([]byte(data))
	if err != nil {
		return ""
	}
	if err := zw.Close(); err != nil {
		return ""
	}
	// 将压缩后的二进制数据转换为Base64编码的字符串
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "GZIP:" + encoded

}

// 还原压缩的数据
func (model *PrinterLog) DecompressData() string {
	if model.Data == "" {
		return ""
	}
	// 检查数据是否是压缩过的
	if !strings.HasPrefix(model.Data, "GZIP:") {
		return model.Data // 如果不是压缩过的数据，直接返回
	}
	// 提取Base64编码的部分
	encoded := strings.TrimPrefix(model.Data, "GZIP:")
	// 解码Base64
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return model.Data // 如果解码失败，返回原始数据
	}
	// 解压数据
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return model.Data
	}
	defer zr.Close()
	//
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return model.Data
	}
	return buf.String()
}

// GetCopies 获取打印份数
func (model *PrinterLog) GetCopies() uint {
	if model.Copies > 0 {
		return model.Copies
	}
	if model.Printer == nil {
		return 1
	}
	if model.Printer.Copies == 0 {
		return model.Printer.Copies
	}
	return 1
}

// 是否收银打印机
func (model *PrinterLog) IsCashierPrinter() bool {
	return slices.Contains([]string{
		constant.PrinterTypeCashierCompax,
		constant.PrinterTypeCashierSunmi,
	}, model.PrinterType) || !slices.Contains([]string{
		constant.PrinterTypeFeiEYun,
		constant.PrinterTypeFeiEYunTag,
		constant.PrinterTypePrintCenter,
		constant.PrinterTypeSunmiLan,
		constant.PrinterTypeSunmiCloud,
		constant.PrinterTypeXPrinterLan,
		constant.PrinterTypeXPrinterWifi,
		constant.PrinterTypeCodesoftLan,
		constant.PrinterTypeCodesoftWifi,
		constant.PrinterTypeGpCloud,
	}, model.PrinterType)
}

// 是否usb打印机
func (model *PrinterLog) IsUsbPrinter() bool {
	if model.Printer == nil {
		return false
	}
	return model.Printer.IsUsb == 1
}

// 计算打印耗时
func (model *PrinterLog) CalculationTime(data string) int64 {
	//
	t := int64(200)
	speed := 200
	//
	if model.PrinterType == constant.PrinterTypeXPrinterWifi {
		speed = 70
	} else if model.PrinterType == constant.PrinterTypeCodesoftWifi {
		speed = 70
	}
	//
	t = int64(math.Ceil(float64(len(data)) / float64(speed)))
	//
	if model.PrinterType == constant.PrinterTypeXPrinterWifi || model.PrinterType == constant.PrinterTypeCodesoftWifi {
		if t < 1200 {
			return 1200
		}
	} else if t < 600 {
		return 600
	}
	return t
}
