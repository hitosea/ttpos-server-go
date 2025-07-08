// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CashBox is the golang structure for table cash_box.
type CashBox struct {
	Id              uint    `json:"id"              orm:"id"               description:"自增ID"`                                                // 自增ID
	Uuid            uint64  `json:"uuid"            orm:"uuid"             description:"钱箱ID"`                                                // 钱箱ID
	Name            string  `json:"name"            orm:"name"             description:"名称"`                                                  // 名称
	Balance         float64 `json:"balance"         orm:"balance"          description:"钱箱余额"`                                                // 钱箱余额
	FrozenBalance   float64 `json:"frozenBalance"   orm:"frozen_balance"   description:"冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额"` // 冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额
	PreviousBalance float64 `json:"previousBalance" orm:"previous_balance" description:"上一班遗留备用金"`                                            // 上一班遗留备用金
	CashWithdrawal  float64 `json:"cashWithdrawal"  orm:"cash_withdrawal"  description:"中途取出金额"`                                              // 中途取出金额
	CashDeposit     float64 `json:"cashDeposit"     orm:"cash_deposit"     description:"中途存入金额"`                                              // 中途存入金额
	CreateTime      uint    `json:"createTime"      orm:"create_time"      description:"创建时间(时间戳)"`                                           // 创建时间(时间戳)
	UpdateTime      uint    `json:"updateTime"      orm:"update_time"      description:"更新时间(时间戳)"`                                           // 更新时间(时间戳)
	DeleteTime      uint    `json:"deleteTime"      orm:"delete_time"      description:"删除时间(时间戳)"`                                           // 删除时间(时间戳)
}
