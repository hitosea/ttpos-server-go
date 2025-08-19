// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// LossReportForm is the golang structure for table loss_report_form.
type LossReportForm struct {
	Id             uint   `json:"id"             orm:"id"               description:"自增ID"`                                        // 自增ID
	Uuid           uint64 `json:"uuid"           orm:"uuid"             description:"报损单ID"`                                       // 报损单ID
	FormNo         string `json:"formNo"         orm:"form_no"          description:"编号"`                                          // 编号
	Scene          int    `json:"scene"          orm:"scene"            description:"报损类型,0-loss损耗 1-lost丢失"`                      // 报损类型,0-loss损耗 1-lost丢失
	Num            int    `json:"num"            orm:"num"              description:"数量"`                                          // 数量
	Remark         string `json:"remark"         orm:"remark"           description:"备注"`                                          // 备注
	ProductBomUuid uint64 `json:"productBomUuid" orm:"product_bom_uuid" description:"商品清单bom uuid"`                                // 商品清单bom uuid
	MaterialUuid   uint64 `json:"materialUuid"   orm:"material_uuid"    description:"物料ID"`                                        // 物料ID
	ApplicantUuid  uint64 `json:"applicantUuid"  orm:"applicant_uuid"   description:"申请人ID"`                                       // 申请人ID
	RejectReason   string `json:"rejectReason"   orm:"reject_reason"    description:"拒绝原因"`                                        // 拒绝原因
	Status         int    `json:"status"         orm:"status"           description:"状态,0-pending待审核 1-approved已通过 2-rejected已驳回"` // 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
	OperatorUuid   uint64 `json:"operatorUuid"   orm:"operator_uuid"    description:"操作员ID"`                                       // 操作员ID
	ApprovedTime   uint   `json:"approvedTime"   orm:"approved_time"    description:"通过时间(时间戳)"`                                   // 通过时间(时间戳)
	RevokeTime     uint   `json:"revokeTime"     orm:"revoke_time"      description:"撤销时间(时间戳)"`                                   // 撤销时间(时间戳)
	CreateTime     uint   `json:"createTime"     orm:"create_time"      description:"创建时间(时间戳)"`                                   // 创建时间(时间戳)
	UpdateTime     uint   `json:"updateTime"     orm:"update_time"      description:"更新时间(时间戳)"`                                   // 更新时间(时间戳)
	DeleteTime     uint   `json:"deleteTime"     orm:"delete_time"      description:"删除时间(时间戳)"`                                   // 删除时间(时间戳)
}
