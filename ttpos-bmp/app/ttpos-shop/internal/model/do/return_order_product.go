// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnOrderProduct is the golang structure of table ttpos_return_order_product for DAO operations like Where/Data.
type ReturnOrderProduct struct {
	g.Meta               `orm:"table:ttpos_return_order_product, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 退货单商品唯一标识符
	SaleOrderUuid        interface{} // 销售订单ID
	SaleOrderProductUuid interface{} // 销售订单商品表ID
	ReturnOrderUuid      interface{} // 退货单ID
	ProductType          interface{} // 商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct
	ProductPackageUuid   interface{} // 商品包ID
	ProductName          interface{} // 商品名称
	ProductPrice         interface{} // 商品单价
	TaxRate              interface{} // 税率,根据结账时税率计算
	Num                  interface{} // 商品数量,退货的商品数量
	ProductDiscount      interface{} // 商品折扣
	ProductTotalAmount   interface{} // 商品总金额（退款总金额）
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
