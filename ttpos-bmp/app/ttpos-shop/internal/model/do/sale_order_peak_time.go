// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderPeakTime is the golang structure of table ttpos_sale_order_peak_time for DAO operations like Where/Data.
type SaleOrderPeakTime struct {
	g.Meta      `orm:"table:ttpos_sale_order_peak_time, do:true"`
	Id          interface{} // 自增ID
	Uuid        interface{} // 唯一ID
	Date        interface{} // 日期（天）
	Hour        interface{} // 小时
	Num         interface{} // 订单数
	Amount      interface{} // 订单金额
	CashierUuid interface{} // 收银员ID
	DeleteTime  interface{} // 删除时间
	CreateTime  interface{} // 创建时间
	UpdateTime  interface{} // 更新时间
}
