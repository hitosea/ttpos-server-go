// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Material is the golang structure for table material.
type Material struct {
	Id                    uint    `json:"id"                    orm:"id"                       description:"自增ID"`              // 自增ID
	Uuid                  uint64  `json:"uuid"                  orm:"uuid"                     description:"原料ID"`              // 原料ID
	Name                  string  `json:"name"                  orm:"name"                     description:"原料名称"`              // 原料名称
	MultiLanguageNameUuid uint64  `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"多语言名称ID"`           // 多语言名称ID
	CategoryUuid          uint64  `json:"categoryUuid"          orm:"category_uuid"            description:"类别ID"`              // 类别ID
	SupplierUuid          uint64  `json:"supplierUuid"          orm:"supplier_uuid"            description:"供应商ID"`             // 供应商ID
	ImageUuid             uint64  `json:"imageUuid"             orm:"image_uuid"               description:"图片ID"`              // 图片ID
	ImageName             string  `json:"imageName"             orm:"image_name"               description:"图片名称"`              // 图片名称
	UnitUuid              uint64  `json:"unitUuid"              orm:"unit_uuid"                description:"单位ID"`              // 单位ID
	Price                 float64 `json:"price"                 orm:"price"                    description:"采购单价"`              // 采购单价
	StockNum              float64 `json:"stockNum"              orm:"stock_num"                description:"库存数量"`              // 库存数量
	BarcodeValue          string  `json:"barcodeValue"          orm:"barcode_value"            description:"条形码值"`              // 条形码值
	Status                int     `json:"status"                orm:"status"                   description:"状态, 1-上架 0-下架"`     // 状态, 1-上架 0-下架
	ActualSaleNum         float64 `json:"actualSaleNum"         orm:"actual_sale_num"          description:"实际销量。每次卖出时,实际销量增加"` // 实际销量。每次卖出时,实际销量增加
	CreateTime            uint    `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`         // 创建时间(时间戳)
	UpdateTime            uint    `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`         // 更新时间(时间戳)
	DeleteTime            uint    `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`         // 删除时间(时间戳)
}
