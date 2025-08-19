// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PurchaseForm is the golang structure of table ttpos_purchase_form for DAO operations like Where/Data.
type PurchaseForm struct {
	g.Meta        `orm:"table:ttpos_purchase_form, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // 采购单ID
	FormNo        interface{} // 编号
	Name          interface{} // 采购单名称
	ApplicantUuid interface{} // 申请人ID
	Remark        interface{} // 备注
	Num           interface{} // 总数量
	Amount        interface{} // 总金额
	Status        interface{} // 状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
	ArrivalTime   interface{} // 到达时间(时间戳)
	CreateTime    interface{} // 创建时间(时间戳)
	UpdateTime    interface{} // 更新时间(时间戳)
	DeleteTime    interface{} // 删除时间(时间戳)
}
