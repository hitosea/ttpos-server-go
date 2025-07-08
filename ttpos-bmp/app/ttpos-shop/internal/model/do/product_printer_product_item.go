// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPrinterProductItem is the golang structure of table ttpos_product_printer_product_item for DAO operations like Where/Data.
type ProductPrinterProductItem struct {
	g.Meta             `orm:"table:ttpos_product_printer_product_item, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 商品打印机商品关联ID
	ProductPrinterUuid interface{} // 商品打印机ID
	ProductPackageUuid interface{} // 商品包ID
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
