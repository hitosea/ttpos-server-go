// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LossReportForm is the golang structure of table ttpos_loss_report_form for DAO operations like Where/Data.
type LossReportForm struct {
	g.Meta         `orm:"table:ttpos_loss_report_form, do:true"`
	Id             interface{} // 自增ID
	Uuid           interface{} // 报损单ID
	FormNo         interface{} // 编号
	Scene          interface{} // 报损类型,0-loss损耗 1-lost丢失
	Num            interface{} // 数量
	Remark         interface{} // 备注
	ProductBomUuid interface{} // 商品清单bom uuid
	MaterialUuid   interface{} // 物料ID
	ApplicantUuid  interface{} // 申请人ID
	RejectReason   interface{} // 拒绝原因
	Status         interface{} // 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
	OperatorUuid   interface{} // 操作员ID
	ApprovedTime   interface{} // 通过时间(时间戳)
	RevokeTime     interface{} // 撤销时间(时间戳)
	CreateTime     interface{} // 创建时间(时间戳)
	UpdateTime     interface{} // 更新时间(时间戳)
	DeleteTime     interface{} // 删除时间(时间戳)
}
