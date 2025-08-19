// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPackageAttribute is the golang structure of table ttpos_product_package_attribute for DAO operations like Where/Data.
type ProductPackageAttribute struct {
	g.Meta                           `orm:"table:ttpos_product_package_attribute, do:true"`
	Id                               interface{} // 自增ID
	Uuid                             interface{} // 商品包属性ID
	ProductPackageAttributeGroupUuid interface{} // 商品包属性组ID
	AttributeUuid                    interface{} // 商品属性ID
	IsDefaultSelected                interface{} // 是否默认选中, 0-否 1-是
	CreateTime                       interface{} // 创建时间(时间戳)
	UpdateTime                       interface{} // 更新时间(时间戳)
	DeleteTime                       interface{} // 删除时间(时间戳)
}
