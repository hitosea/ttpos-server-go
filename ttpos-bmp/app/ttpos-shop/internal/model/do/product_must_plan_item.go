// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductMustPlanItem is the golang structure of table ttpos_product_must_plan_item for DAO operations like Where/Data.
type ProductMustPlanItem struct {
	g.Meta              `orm:"table:ttpos_product_must_plan_item, do:true"`
	Id                  interface{} // 自增ID
	Uuid                interface{} // 商品必选商品计划商品明细ID
	ProductMustPlanUuid interface{} // 商品必选商品计划ID
	ProductPackageUuid  interface{} // 商品包ID
	CreateTime          interface{} // 创建时间(时间戳)
	UpdateTime          interface{} // 更新时间(时间戳)
	DeleteTime          interface{} // 删除时间(时间戳)
}
