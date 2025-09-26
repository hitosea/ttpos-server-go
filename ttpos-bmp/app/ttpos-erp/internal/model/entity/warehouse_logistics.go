// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WarehouseLogistics is the golang structure for table warehouse_logistics.
type WarehouseLogistics struct {
	Id            int64  `json:"id"            orm:"id"             description:"ID"`                          // ID
	SiteCode      string `json:"siteCode"      orm:"site_code"      description:"站点编码。 关联 erp_site.site_code"` // 站点编码。 关联 erp_site.site_code
	ShopUuid      string `json:"shopUuid"      orm:"shop_uuid"      description:"ttpos商铺ID"`                   // ttpos商铺ID
	WarehouseCode string `json:"warehouseCode" orm:"warehouse_code" description:"仓库编码. erpnext warehouse"`     // 仓库编码. erpnext warehouse
	LogisticsId   int64  `json:"logisticsId"   orm:"logistics_id"   description:"物流ID"`                        // 物流ID
}
