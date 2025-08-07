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
	Id        interface{} // 主键
	Uuid      interface{} // UUID
	SiteName  interface{} // 站点名称
	SiteUrl   interface{} // 站点地址
	Remark    interface{} // 备注
	SiteCode  interface{} // 站点编码， 与ttpos映射
	ApiKey    interface{} //
	ApiSecret interface{} //
}
