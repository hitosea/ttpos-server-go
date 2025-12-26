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
	Id           any //
	ShopUuid     any // 商店UUID
	AdminUuid    any // 商店管理员UUID
	CashierEmail any // 收银员邮箱
	ApiKey       any //
	ApiSecret    any //
	CompanyAbbr  any // 公司缩写
	Branch       any // 分支
	SiteCode     any // 站点编码
}
