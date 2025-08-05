// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// H5Order is the golang structure for table h5_order.
type H5Order struct {
	Id                     uint    `json:"id"                     orm:"id"                        description:"自增ID"`                                                  // 自增ID
	Uuid                   uint64  `json:"uuid"                   orm:"uuid"                      description:"扫码订单ID"`                                                // 扫码订单ID
	DeskUuid               uint64  `json:"deskUuid"               orm:"desk_uuid"                 description:"桌台uuid"`                                                // 桌台uuid
	DeskNo                 string  `json:"deskNo"                 orm:"desk_no"                   description:"桌台编号"`                                                  // 桌台编号
	Status                 int     `json:"status"                 orm:"status"                    description:"状态, 0-未下单 1-未接单 2-已接单 3-已拒单"`                           // 状态, 0-未下单 1-未接单 2-已接单 3-已拒单
	IsAutoAccept           int     `json:"isAutoAccept"           orm:"is_auto_accept"            description:"是否自动接单, 0-否 1-是"`                                       // 是否自动接单, 0-否 1-是
	IsBuffet               int     `json:"isBuffet"               orm:"is_buffet"                 description:"是否是自助餐, 0-非自助餐 1-自助餐"`                                  // 是否是自助餐, 0-非自助餐 1-自助餐
	MemberDiscountRate     float64 `json:"memberDiscountRate"     orm:"member_discount_rate"      description:"会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变"`       // 会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	MemberCardDiscountRate float64 `json:"memberCardDiscountRate" orm:"member_card_discount_rate" description:"会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变"`      // 会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	CustomDiscountRate     float64 `json:"customDiscountRate"     orm:"custom_discount_rate"      description:"自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变"`      // 自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	ProductTotalPrice      float64 `json:"productTotalPrice"      orm:"product_total_price"       description:"商品总价。接单和拒单后从sale_order_product表获取，不再改变"`                // 商品总价。接单和拒单后从sale_order_product表获取，不再改变
	TotalAmount            float64 `json:"totalAmount"            orm:"total_amount"              description:"订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变"` // 订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变
	StaffUuid              uint64  `json:"staffUuid"              orm:"staff_uuid"                description:"接单或拒单员工ID"`                                             // 接单或拒单员工ID
	HandleTime             uint    `json:"handleTime"             orm:"handle_time"               description:"接单或拒单时间(时间戳)"`                                          // 接单或拒单时间(时间戳)
	OrderTime              uint    `json:"orderTime"              orm:"order_time"                description:"下单时间(时间戳)"`                                             // 下单时间(时间戳)
	SaleOrderUuid          uint64  `json:"saleOrderUuid"          orm:"sale_order_uuid"           description:"销售订单uuid"`                                              // 销售订单uuid
	SaleBillUuid           uint64  `json:"saleBillUuid"           orm:"sale_bill_uuid"            description:"销售账单uuid"`                                              // 销售账单uuid
	CreateTime             uint    `json:"createTime"             orm:"create_time"               description:"创建时间(时间戳)，扫码下单时间"`                                      // 创建时间(时间戳)，扫码下单时间
	UpdateTime             uint    `json:"updateTime"             orm:"update_time"               description:"更新时间(时间戳)"`                                             // 更新时间(时间戳)
	DeleteTime             uint    `json:"deleteTime"             orm:"delete_time"               description:"删除时间(时间戳)"`                                             // 删除时间(时间戳)
}
