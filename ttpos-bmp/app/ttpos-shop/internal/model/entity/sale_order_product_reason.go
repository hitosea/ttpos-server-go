// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderProductReason is the golang structure for table sale_order_product_reason.
type SaleOrderProductReason struct {
	Id                    uint   `json:"id"                    orm:"id"                       description:"自增ID"`                                                                             // 自增ID
	Uuid                  uint64 `json:"uuid"                  orm:"uuid"                     description:"自增UUID"`                                                                           // 自增UUID
	SaleOrderUuid         uint64 `json:"saleOrderUuid"         orm:"sale_order_uuid"          description:"销售订单ID"`                                                                           // 销售订单ID
	SaleOrderProductUuid  uint64 `json:"saleOrderProductUuid"  orm:"sale_order_product_uuid"  description:"销售订单商品ID，如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0"` // 销售订单商品ID，如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0
	ReturnFoodReasonUuid  uint64 `json:"returnFoodReasonUuid"  orm:"return_food_reason_uuid"  description:"退菜原因ID"`                                                                           // 退菜原因ID
	FreeReasonUuid        uint64 `json:"freeReasonUuid"        orm:"free_reason_uuid"         description:"免单原因ID"`                                                                           // 免单原因ID
	GiftReasonUuid        uint64 `json:"giftReasonUuid"        orm:"gift_reason_uuid"         description:"赠菜原因ID"`                                                                           // 赠菜原因ID
	MultiLanguageNameUuid uint64 `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"原因-多语言名称ID"`                                                                       // 原因-多语言名称ID
	CreateTime            uint   `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`                                                                        // 创建时间(时间戳)
	UpdateTime            uint   `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`                                                                        // 更新时间(时间戳)
	DeleteTime            uint   `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`                                                                        // 删除时间(时间戳)
}
