// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PurchaseForm is the golang structure for table purchase_form.
type PurchaseForm struct {
	Id            uint    `json:"id"            orm:"id"             description:"自增ID"`                             // 自增ID
	Uuid          uint64  `json:"uuid"          orm:"uuid"           description:"采购单ID"`                            // 采购单ID
	FormNo        string  `json:"formNo"        orm:"form_no"        description:"编号"`                               // 编号
	Name          string  `json:"name"          orm:"name"           description:"采购单名称"`                            // 采购单名称
	ApplicantUuid uint64  `json:"applicantUuid" orm:"applicant_uuid" description:"申请人ID"`                            // 申请人ID
	Remark        string  `json:"remark"        orm:"remark"         description:"备注"`                               // 备注
	Num           float64 `json:"num"           orm:"num"            description:"总数量"`                              // 总数量
	Amount        float64 `json:"amount"        orm:"amount"         description:"总金额"`                              // 总金额
	Status        int     `json:"status"        orm:"status"         description:"状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库"` // 状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
	ArrivalTime   int     `json:"arrivalTime"   orm:"arrival_time"   description:"到达时间(时间戳)"`                        // 到达时间(时间戳)
	CreateTime    uint    `json:"createTime"    orm:"create_time"    description:"创建时间(时间戳)"`                        // 创建时间(时间戳)
	UpdateTime    uint    `json:"updateTime"    orm:"update_time"    description:"更新时间(时间戳)"`                        // 更新时间(时间戳)
	DeleteTime    uint    `json:"deleteTime"    orm:"delete_time"    description:"删除时间(时间戳)"`                        // 删除时间(时间戳)
}
