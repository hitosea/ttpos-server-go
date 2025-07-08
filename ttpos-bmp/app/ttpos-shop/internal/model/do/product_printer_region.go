// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPrinterRegion is the golang structure of table ttpos_product_printer_region for DAO operations like Where/Data.
type ProductPrinterRegion struct {
	g.Meta             `orm:"table:ttpos_product_printer_region, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 商品打印机区域ID
	ProductPrinterUuid interface{} // 商品打印机ID
	DeskRegionUuid     interface{} // 桌台区域ID
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
