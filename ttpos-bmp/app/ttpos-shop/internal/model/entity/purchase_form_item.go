// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PurchaseFormItem is the golang structure for table purchase_form_item.
type PurchaseFormItem struct {
	Id               uint    `json:"id"               orm:"id"                 description:"自增ID"`           // 自增ID
	Uuid             uint64  `json:"uuid"             orm:"uuid"               description:"采购单明细ID"`        // 采购单明细ID
	PurchaseFormUuid uint64  `json:"purchaseFormUuid" orm:"purchase_form_uuid" description:"采购单ID"`          // 采购单ID
	MaterialType     int     `json:"materialType"     orm:"material_type"      description:"物料类型,0-商品 1-原料"` // 物料类型,0-商品 1-原料
	MaterialUuid     uint64  `json:"materialUuid"     orm:"material_uuid"      description:"物料ID"`           // 物料ID
	EstimateNum      int     `json:"estimateNum"      orm:"estimate_num"       description:"预计数量"`           // 预计数量
	EstimatePrice    float64 `json:"estimatePrice"    orm:"estimate_price"     description:"预计单价"`           // 预计单价
	EstimateAmount   float64 `json:"estimateAmount"   orm:"estimate_amount"    description:"预计金额"`           // 预计金额
	Num              int     `json:"num"              orm:"num"                description:"数量"`             // 数量
	Price            float64 `json:"price"            orm:"price"              description:"单价"`             // 单价
	Amount           float64 `json:"amount"           orm:"amount"             description:"金额"`             // 金额
	CreateTime       uint    `json:"createTime"       orm:"create_time"        description:"创建时间(时间戳)"`      // 创建时间(时间戳)
	UpdateTime       uint    `json:"updateTime"       orm:"update_time"        description:"更新时间(时间戳)"`      // 更新时间(时间戳)
	DeleteTime       uint    `json:"deleteTime"       orm:"delete_time"        description:"删除时间(时间戳)"`      // 删除时间(时间戳)
}
