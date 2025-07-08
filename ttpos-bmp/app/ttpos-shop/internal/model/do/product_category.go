// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductCategory is the golang structure of table ttpos_product_category for DAO operations like Where/Data.
type ProductCategory struct {
	g.Meta                `orm:"table:ttpos_product_category, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 商品类别ID
	Name                  interface{} // 名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	Status                interface{} // 状态, 1-开启 0-关闭
	ParentUuid            interface{} // 父级ID
	IsSpecial             interface{} // 特殊分类, 1-是 0-否
	CategoryKey           interface{} // 关键字
	Sort                  interface{} // 排序
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
