// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// H5OrderProduct is the golang structure of table ttpos_h5_order_product for DAO operations like Where/Data.
type H5OrderProduct struct {
	g.Meta               `orm:"table:ttpos_h5_order_product, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 扫码订单商品uuid
	Name                 interface{} // 商品名称.接单和拒单后从sale_order_product表获取，不再改变
	Price                interface{} // 最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变
	SalePrice            interface{} // 销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变
	Num                  interface{} // 最终商品数量.接单和拒单后从sale_order_product表获取，不再改变
	AttributeText        interface{} // 商品属性文本。接单和拒单后从sale_order_product表获取，不再改变
	Remark               interface{} // 备注。接单和拒单后从sale_order_product表获取，不再改变
	SaleOrderProductUuid interface{} // 销售订单商品uuid
	H5OrderUuid          interface{} // 扫码订单uuid
	SaleBillUuid         interface{} // 销售账单uuid
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
