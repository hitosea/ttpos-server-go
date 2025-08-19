// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetProduct is the golang structure of table ttpos_buffet_product for DAO operations like Where/Data.
type BuffetProduct struct {
	g.Meta             `orm:"table:ttpos_buffet_product, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 自助餐商品ID
	BuffetPackageUuid  interface{} // 自助餐套餐ID
	ProductPackageUuid interface{} // 商品包ID
	IsShowCashier      interface{} // 是否在收银台显示, 0-否 1-是
	IsShowTablet       interface{} // 是否在平板显示, 0-否 1-是
	IsShowKitchen      interface{} // 是否在厨房显示, 0-否 1-是
	IsShowAssistant    interface{} // 是否在助手显示, 0-否 1-是
	Limit              interface{} // 限购数量
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
