// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RelatedMaterial is the golang structure of table ttpos_related_material for DAO operations like Where/Data.
type RelatedMaterial struct {
	g.Meta       `orm:"table:ttpos_related_material, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 关联材料ID
	RelatedUuid  interface{} // 物料清单BOM的ID
	MaterialUuid interface{} // 原料ID
	Num          interface{} // 材料用量,可小数
	CreateTime   interface{} // 创建时间(时间戳)
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}
