// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// OrderStatusLog is the golang structure for table order_status_log.
type OrderStatusLog struct {
	Id           int64  `json:"id"           orm:"id"            description:"主键"`                         // 主键
	Uuid         string `json:"uuid"         orm:"uuid"          description:"唯一ID"`                       // 唯一ID
	OrderUuid    string `json:"orderUuid"    orm:"order_uuid"    description:"关联订单UUID"`                   // 关联订单UUID
	ProviderName string `json:"providerName" orm:"provider_name" description:"渠道: grab"`                   // 渠道: grab
	StatusBefore string `json:"statusBefore" orm:"status_before" description:"变更前状态"`                      // 变更前状态
	StatusAfter  string `json:"statusAfter"  orm:"status_after"  description:"变更后状态"`                      // 变更后状态
	ChangeSource string `json:"changeSource" orm:"change_source" description:"变更来源: WEBHOOK, API, SYSTEM"` // 变更来源: WEBHOOK, API, SYSTEM
	DriverEta    int    `json:"driverEta"    orm:"driver_eta"    description:"骑手预计到达时间(分钟)"`               // 骑手预计到达时间(分钟)
	Remark       string `json:"remark"       orm:"remark"        description:"备注或原因"`                      // 备注或原因
	RawData      string `json:"rawData"      orm:"raw_data"      description:"原始JSON数据"`                   // 原始JSON数据
	CreatedAt    int    `json:"createdAt"    orm:"created_at"    description:"创建时间"`                       // 创建时间
}
