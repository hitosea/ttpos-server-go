// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPackageAttributeGroup is the golang structure of table ttpos_product_package_attribute_group for DAO operations like Where/Data.
type ProductPackageAttributeGroup struct {
	g.Meta                    `orm:"table:ttpos_product_package_attribute_group, do:true"`
	Id                        interface{} // 自增ID
	Uuid                      interface{} // 商品包属性组ID
	IsMust                    interface{} // 是否必选, 0-否 1-是
	MaxSelection              interface{} // 最大选择数量
	ProductPackageUuid        interface{} // 商品包ID
	ProductAttributeGroupUuid interface{} // 商品属性组ID
	CreateTime                interface{} // 创建时间(时间戳)
	UpdateTime                interface{} // 更新时间(时间戳)
	DeleteTime                interface{} // 删除时间(时间戳)
}
