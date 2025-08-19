// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPrinterItem is the golang structure of table ttpos_product_printer_item for DAO operations like Where/Data.
type ProductPrinterItem struct {
	g.Meta             `orm:"table:ttpos_product_printer_item, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 商品打印(档口)打印机ID
	ProductPrinterUuid interface{} // 商品打印(档口)ID
	PrinterUuid        interface{} // 打印机ID
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
