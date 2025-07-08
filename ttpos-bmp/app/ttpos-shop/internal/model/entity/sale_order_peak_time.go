// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderPeakTime is the golang structure for table sale_order_peak_time.
type SaleOrderPeakTime struct {
	Id          uint    `json:"id"          orm:"id"           description:"自增ID"`  // 自增ID
	Uuid        uint64  `json:"uuid"        orm:"uuid"         description:"唯一ID"`  // 唯一ID
	Date        int     `json:"date"        orm:"date"         description:"日期（天）"` // 日期（天）
	Hour        int     `json:"hour"        orm:"hour"         description:"小时"`    // 小时
	Num         int     `json:"num"         orm:"num"          description:"订单数"`   // 订单数
	Amount      float64 `json:"amount"      orm:"amount"       description:"订单金额"`  // 订单金额
	CashierUuid int64   `json:"cashierUuid" orm:"cashier_uuid" description:"收银员ID"` // 收银员ID
	DeleteTime  int     `json:"deleteTime"  orm:"delete_time"  description:"删除时间"`  // 删除时间
	CreateTime  int     `json:"createTime"  orm:"create_time"  description:"创建时间"`  // 创建时间
	UpdateTime  int     `json:"updateTime"  orm:"update_time"  description:"更新时间"`  // 更新时间
}
