// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PurchaseFormItem is the golang structure of table ttpos_purchase_form_item for DAO operations like Where/Data.
type PurchaseFormItem struct {
	g.Meta           `orm:"table:ttpos_purchase_form_item, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 采购单明细ID
	PurchaseFormUuid interface{} // 采购单ID
	MaterialType     interface{} // 物料类型,0-商品 1-原料
	MaterialUuid     interface{} // 物料ID
	EstimateNum      interface{} // 预计数量
	EstimatePrice    interface{} // 预计单价
	EstimateAmount   interface{} // 预计金额
	Num              interface{} // 数量
	Price            interface{} // 单价
	Amount           interface{} // 金额
	CreateTime       interface{} // 创建时间(时间戳)
	UpdateTime       interface{} // 更新时间(时间戳)
	DeleteTime       interface{} // 删除时间(时间戳)
}
