// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ShopCashier is the golang structure of table erp_shop_cashier for DAO operations like Where/Data.
type ShopCashier struct {
	g.Meta       `orm:"table:erp_shop_cashier, do:true"`
	Id           interface{} //
	ShopUuid     interface{} // 商店UUID
	AdminUuid    interface{} // 商店管理员UUID
	CashierEmail interface{} // 收银员邮箱
	ApiKey       interface{} //
	ApiSecret    interface{} //
	CompanyAbbr  interface{} // 公司缩写
	Branch       interface{} // 分支
}
