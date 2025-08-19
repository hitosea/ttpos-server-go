// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberRechargeOrderAbnormalRecord is the golang structure for table member_recharge_order_abnormal_record.
type MemberRechargeOrderAbnormalRecord struct {
	Id                uint   `json:"id"                orm:"id"                  description:"自增ID"`      // 自增ID
	Uuid              uint64 `json:"uuid"              orm:"uuid"                description:"UUID"`      // UUID
	RechargeOrderUuid uint64 `json:"rechargeOrderUuid" orm:"recharge_order_uuid" description:"充值订单ID"`    // 充值订单ID
	DutyNo            string `json:"dutyNo"            orm:"duty_no"             description:"当班编号"`      // 当班编号
	Action            string `json:"action"            orm:"action"              description:"行为"`        // 行为
	SubAction         string `json:"subAction"         orm:"sub_action"          description:"自定义子行为"`    // 自定义子行为
	Sign              string `json:"sign"              orm:"sign"                description:"操作签名"`      // 操作签名
	Remark            string `json:"remark"            orm:"remark"              description:"备注"`        // 备注
	CashierUuid       uint64 `json:"cashierUuid"       orm:"cashier_uuid"        description:"收银员ID"`     // 收银员ID
	DeleteTime        uint   `json:"deleteTime"        orm:"delete_time"         description:"删除时间(时间戳)"` // 删除时间(时间戳)
	CreateTime        uint   `json:"createTime"        orm:"create_time"         description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime        uint   `json:"updateTime"        orm:"update_time"         description:"更新时间(时间戳)"` // 更新时间(时间戳)
}
