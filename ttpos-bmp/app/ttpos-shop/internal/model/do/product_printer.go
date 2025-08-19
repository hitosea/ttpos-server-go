// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPrinter is the golang structure of table ttpos_product_printer for DAO operations like Where/Data.
type ProductPrinter struct {
	g.Meta             `orm:"table:ttpos_product_printer, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 商品打印机ID
	Name               interface{} // 名称.厨显上叫档口
	Status             interface{} // 状态,1-open开启 1、0-close关闭
	PrintMode          interface{} // 打印模式,0-payment付款打印 1-kitchen送厨打印
	PrintMethod        interface{} // 打印方式,0-order整单打印 1-item按一菜一单打印
	PrintProductSelect interface{} // 打印商品选择,0-category按商品分类 1-tag按打印标签
	PrintModeScene     interface{} // 打印模式场景,0-merge合并 1-separate分开
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
