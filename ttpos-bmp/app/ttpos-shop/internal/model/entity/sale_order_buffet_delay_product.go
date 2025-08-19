// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderBuffetDelayProduct is the golang structure for table sale_order_buffet_delay_product.
type SaleOrderBuffetDelayProduct struct {
	Id              uint    `json:"id"              orm:"id"                description:"自增ID"`                                                      // 自增ID
	Uuid            uint64  `json:"uuid"            orm:"uuid"              description:"自助餐加钟价格ID"`                                                 // 自助餐加钟价格ID
	SaleOrderUuid   uint64  `json:"saleOrderUuid"   orm:"sale_order_uuid"   description:"销售订单ID"`                                                    // 销售订单ID
	BuffetDelayUuid uint64  `json:"buffetDelayUuid" orm:"buffet_delay_uuid" description:"自助餐加钟价格ID"`                                                 // 自助餐加钟价格ID
	Name            string  `json:"name"            orm:"name"              description:"自助餐加钟商品名称，下单时固定不受后台改变"`                                     // 自助餐加钟商品名称，下单时固定不受后台改变
	Num             int     `json:"num"             orm:"num"               description:"数量"`                                                        // 数量
	Price           float64 `json:"price"           orm:"price"             description:"价格,下单时固定不受后台改变，结账时再检查是否改变"`                                 // 价格,下单时固定不受后台改变，结账时再检查是否改变
	DelayTime       uint    `json:"delayTime"       orm:"delay_time"        description:"加钟时间(分钟)"`                                                  // 加钟时间(分钟)
	Sign            string  `json:"sign"            orm:"sign"              description:"加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并"` // 加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并
	CreateTime      uint    `json:"createTime"      orm:"create_time"       description:"创建时间(时间戳)"`                                                 // 创建时间(时间戳)
	UpdateTime      uint    `json:"updateTime"      orm:"update_time"       description:"更新时间(时间戳)"`                                                 // 更新时间(时间戳)
	DeleteTime      uint    `json:"deleteTime"      orm:"delete_time"       description:"删除时间(时间戳)"`                                                 // 删除时间(时间戳)
}
