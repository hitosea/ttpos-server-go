// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MaterialAttribute is the golang structure for table material_attribute.
type MaterialAttribute struct {
	Id                           uint   `json:"id"                           orm:"id"                              description:"自增ID"`      // 自增ID
	Uuid                         uint64 `json:"uuid"                         orm:"uuid"                            description:"原料属性ID"`    // 原料属性ID
	MaterialUuid                 uint64 `json:"materialUuid"                 orm:"material_uuid"                   description:"原料ID"`      // 原料ID
	HistoricalPurchaseQuantity   int    `json:"historicalPurchaseQuantity"   orm:"historical_purchase_quantity"    description:"历史采购数量"`    // 历史采购数量
	HistoricalLossReportQuantity int    `json:"historicalLossReportQuantity" orm:"historical_loss_report_quantity" description:"历史报损数量"`    // 历史报损数量
	HistoricalSaleQuantity       int    `json:"historicalSaleQuantity"       orm:"historical_sale_quantity"        description:"历史销售数量"`    // 历史销售数量
	CreateTime                   uint   `json:"createTime"                   orm:"create_time"                     description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime                   uint   `json:"updateTime"                   orm:"update_time"                     description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime                   uint   `json:"deleteTime"                   orm:"delete_time"                     description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
