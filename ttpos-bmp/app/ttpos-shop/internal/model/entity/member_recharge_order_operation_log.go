// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberRechargeOrderOperationLog is the golang structure for table member_recharge_order_operation_log.
type MemberRechargeOrderOperationLog struct {
	Id                uint   `json:"id"                orm:"id"                  description:"自增ID"`         // 自增ID
	Uuid              uint64 `json:"uuid"              orm:"uuid"                description:"会员充值订单操作日志ID"` // 会员充值订单操作日志ID
	OperatorName      string `json:"operatorName"      orm:"operator_name"       description:"操作员姓名"`        // 操作员姓名
	OperatorEmail     string `json:"operatorEmail"     orm:"operator_email"      description:"操作员电子邮件"`      // 操作员电子邮件
	Client            string `json:"client"            orm:"client"              description:"客户端信息"`        // 客户端信息
	Message           string `json:"message"           orm:"message"             description:"消息内容"`         // 消息内容
	Action            string `json:"action"            orm:"action"              description:"操作"`           // 操作
	Data              string `json:"data"              orm:"data"                description:"数据"`           // 数据
	RechargeOrderUuid uint64 `json:"rechargeOrderUuid" orm:"recharge_order_uuid" description:"充值订单ID"`       // 充值订单ID
	CreateTime        uint   `json:"createTime"        orm:"create_time"         description:"创建时间(时间戳)"`    // 创建时间(时间戳)
	UpdateTime        uint   `json:"updateTime"        orm:"update_time"         description:"更新时间(时间戳)"`    // 更新时间(时间戳)
	DeleteTime        uint   `json:"deleteTime"        orm:"delete_time"         description:"删除时间(时间戳)"`    // 删除时间(时间戳)
}
