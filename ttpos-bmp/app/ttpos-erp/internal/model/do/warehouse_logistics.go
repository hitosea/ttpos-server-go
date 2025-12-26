// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseLogistics is the golang structure of table erp_warehouse_logistics for DAO operations like Where/Data.
type WarehouseLogistics struct {
	g.Meta        `orm:"table:erp_warehouse_logistics, do:true"`
	Id            any // ID
	SiteCode      any // 站点编码。 关联 erp_site.site_code
	ShopUuid      any // ttpos商铺ID
	WarehouseCode any // 仓库编码. erpnext warehouse
	LogisticsId   any // 物流ID
}
