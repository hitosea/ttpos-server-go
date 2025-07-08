// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderAbnormalRecord is the golang structure of table ttpos_sale_order_abnormal_record for DAO operations like Where/Data.
type SaleOrderAbnormalRecord struct {
	g.Meta        `orm:"table:ttpos_sale_order_abnormal_record, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // UUID
	SaleBillUuid  interface{} // 销售账单ID
	SaleOrderUuid interface{} // 销售订单ID
	DutyNo        interface{} // 当班编号
	Action        interface{} // 行为
	SubAction     interface{} // 自定义子行为
	Sign          interface{} // 操作签名
	Remark        interface{} // 备注
	CashierUuid   interface{} // 收银员ID
	DeleteTime    interface{} // 删除时间(时间戳)
	CreateTime    interface{} // 创建时间(时间戳)
	UpdateTime    interface{} // 更新时间(时间戳)
}
