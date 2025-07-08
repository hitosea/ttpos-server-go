// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderOperationRecord is the golang structure of table ttpos_sale_order_operation_record for DAO operations like Where/Data.
type SaleOrderOperationRecord struct {
	g.Meta        `orm:"table:ttpos_sale_order_operation_record, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // 桌台账单记录ID
	Source        interface{} // 操作来源 cashier-收银端 assistant-点餐助手 shop-商家后台 h5-扫码点餐
	Action        interface{} // 操作行为
	Data          interface{} // 数据
	Remark        interface{} // 备注
	SaleBillUuid  interface{} // 销售账单ID
	SaleOrderUuid interface{} // 销售订单ID
	H5OrderUuid   interface{} // h5订单Uuid
	OperatorUuid  interface{} // 操作员ID
	CreateTime    interface{} // 创建时间(时间戳)
	UpdateTime    interface{} // 更新时间(时间戳)
	DeleteTime    interface{} // 删除时间(时间戳)
}
