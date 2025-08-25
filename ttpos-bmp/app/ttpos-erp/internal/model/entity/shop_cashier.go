// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ShopCashier is the golang structure for table shop_cashier.
type ShopCashier struct {
	Id           int64  `json:"id"           orm:"id"            description:""`          //
	ShopUuid     string `json:"shopUuid"     orm:"shop_uuid"     description:"商店UUID"`    // 商店UUID
	AdminUuid    string `json:"adminUuid"    orm:"admin_uuid"    description:"商店管理员UUID"` // 商店管理员UUID
	CashierEmail string `json:"cashierEmail" orm:"cashier_email" description:"收银员邮箱"`     // 收银员邮箱
	ApiKey       string `json:"apiKey"       orm:"api_key"       description:""`          //
	ApiSecret    string `json:"apiSecret"    orm:"api_secret"    description:""`          //
	CompanyAbbr  string `json:"companyAbbr"  orm:"company_abbr"  description:"公司缩写"`      // 公司缩写
	Branch       string `json:"branch"       orm:"branch"        description:"分支"`        // 分支
}
