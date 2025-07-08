// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductReason is the golang structure of table ttpos_sale_order_product_reason for DAO operations like Where/Data.
type SaleOrderProductReason struct {
	g.Meta                `orm:"table:ttpos_sale_order_product_reason, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 自增UUID
	SaleOrderUuid         interface{} // 销售订单ID
	SaleOrderProductUuid  interface{} // 销售订单商品ID，如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0
	ReturnFoodReasonUuid  interface{} // 退菜原因ID
	FreeReasonUuid        interface{} // 免单原因ID
	GiftReasonUuid        interface{} // 赠菜原因ID
	MultiLanguageNameUuid interface{} // 原因-多语言名称ID
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
