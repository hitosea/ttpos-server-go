// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// H5Order is the golang structure of table ttpos_h5_order for DAO operations like Where/Data.
type H5Order struct {
	g.Meta                 `orm:"table:ttpos_h5_order, do:true"`
	Id                     interface{} // 自增ID
	Uuid                   interface{} // 扫码订单ID
	DeskUuid               interface{} // 桌台uuid
	DeskNo                 interface{} // 桌台编号
	Status                 interface{} // 状态, 0-未下单 1-未接单 2-已接单 3-已拒单
	IsAutoAccept           interface{} // 是否自动接单, 0-否 1-是
	IsBuffet               interface{} // 是否是自助餐, 0-非自助餐 1-自助餐
	MemberDiscountRate     interface{} // 会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	MemberCardDiscountRate interface{} // 会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	CustomDiscountRate     interface{} // 自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	ProductTotalPrice      interface{} // 商品总价。接单和拒单后从sale_order_product表获取，不再改变
	TotalAmount            interface{} // 订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变
	StaffUuid              interface{} // 接单或拒单员工ID
	HandleTime             interface{} // 接单或拒单时间(时间戳)
	OrderTime              interface{} // 下单时间(时间戳)
	SaleOrderUuid          interface{} // 销售订单uuid
	SaleBillUuid           interface{} // 销售账单uuid
	CreateTime             interface{} // 创建时间(时间戳)，扫码下单时间
	UpdateTime             interface{} // 更新时间(时间戳)
	DeleteTime             interface{} // 删除时间(时间戳)
}
