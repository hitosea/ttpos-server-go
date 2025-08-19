// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CashBox is the golang structure of table ttpos_cash_box for DAO operations like Where/Data.
type CashBox struct {
	g.Meta          `orm:"table:ttpos_cash_box, do:true"`
	Id              interface{} // 自增ID
	Uuid            interface{} // 钱箱ID
	Name            interface{} // 名称
	Balance         interface{} // 钱箱余额
	FrozenBalance   interface{} // 冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额
	PreviousBalance interface{} // 上一班遗留备用金
	CashWithdrawal  interface{} // 中途取出金额
	CashDeposit     interface{} // 中途存入金额
	CreateTime      interface{} // 创建时间(时间戳)
	UpdateTime      interface{} // 更新时间(时间戳)
	DeleteTime      interface{} // 删除时间(时间戳)
}
