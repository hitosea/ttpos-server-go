// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PurchaseFormLog is the golang structure of table ttpos_purchase_form_log for DAO operations like Where/Data.
type PurchaseFormLog struct {
	g.Meta           `orm:"table:ttpos_purchase_form_log, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 采购单日志UUID
	PurchaseFormUuid interface{} // 采购单uuid
	OperatorUuid     interface{} // 操作人uuid
	Username         interface{} // 操作人员
	Status           interface{} // 操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
	Operation        interface{} // 操作动作
	Remark           interface{} // 备注
	CreateTime       interface{} // 创建时间(时间戳)
	UpdateTime       interface{} // 更新时间(时间戳)
	DeleteTime       interface{} // 删除时间(时间戳)
}
