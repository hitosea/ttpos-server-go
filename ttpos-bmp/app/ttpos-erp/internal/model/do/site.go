// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Site is the golang structure of table erp_site for DAO operations like Where/Data.
type Site struct {
	g.Meta    `orm:"table:erp_site, do:true"`
	Id        any // 主键
	Uuid      any // UUID
	SiteName  any // 站点名称
	SiteUrl   any // 站点地址
	Remark    any // 备注
	SiteCode  any // 站点编码， 与ttpos映射
	ApiKey    any //
	ApiSecret any //
}
