// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PurchaseFormLog is the golang structure for table purchase_form_log.
type PurchaseFormLog struct {
	Id               uint   `json:"id"               orm:"id"                 description:"自增ID"`                               // 自增ID
	Uuid             uint64 `json:"uuid"             orm:"uuid"               description:"采购单日志UUID"`                          // 采购单日志UUID
	PurchaseFormUuid int64  `json:"purchaseFormUuid" orm:"purchase_form_uuid" description:"采购单uuid"`                            // 采购单uuid
	OperatorUuid     int64  `json:"operatorUuid"     orm:"operator_uuid"      description:"操作人uuid"`                            // 操作人uuid
	Username         string `json:"username"         orm:"username"           description:"操作人员"`                               // 操作人员
	Status           int    `json:"status"           orm:"status"             description:"操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库"` // 操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
	Operation        string `json:"operation"        orm:"operation"          description:"操作动作"`                               // 操作动作
	Remark           string `json:"remark"           orm:"remark"             description:"备注"`                                 // 备注
	CreateTime       uint   `json:"createTime"       orm:"create_time"        description:"创建时间(时间戳)"`                          // 创建时间(时间戳)
	UpdateTime       uint   `json:"updateTime"       orm:"update_time"        description:"更新时间(时间戳)"`                          // 更新时间(时间戳)
	DeleteTime       uint   `json:"deleteTime"       orm:"delete_time"        description:"删除时间(时间戳)"`                          // 删除时间(时间戳)
}
