// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderBuffetDelayProduct is the golang structure of table ttpos_sale_order_buffet_delay_product for DAO operations like Where/Data.
type SaleOrderBuffetDelayProduct struct {
	g.Meta          `orm:"table:ttpos_sale_order_buffet_delay_product, do:true"`
	Id              interface{} // 自增ID
	Uuid            interface{} // 自助餐加钟价格ID
	SaleOrderUuid   interface{} // 销售订单ID
	BuffetDelayUuid interface{} // 自助餐加钟价格ID
	Name            interface{} // 自助餐加钟商品名称，下单时固定不受后台改变
	Num             interface{} // 数量
	Price           interface{} // 价格,下单时固定不受后台改变，结账时再检查是否改变
	DelayTime       interface{} // 加钟时间(分钟)
	Sign            interface{} // 加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并
	CreateTime      interface{} // 创建时间(时间戳)
	UpdateTime      interface{} // 更新时间(时间戳)
	DeleteTime      interface{} // 删除时间(时间戳)
}
