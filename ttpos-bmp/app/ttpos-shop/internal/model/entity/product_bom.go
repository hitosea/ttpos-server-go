// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductBom is the golang structure for table product_bom.
type ProductBom struct {
	Id                 uint    `json:"id"                 orm:"id"                   description:"自增ID"`                    // 自增ID
	Uuid               uint64  `json:"uuid"               orm:"uuid"                 description:"商品BOM ID"`                // 商品BOM ID
	PurchasePrice      float64 `json:"purchasePrice"      orm:"purchase_price"       description:"采购单价"`                    // 采购单价
	Price              float64 `json:"price"              orm:"price"                description:"价格"`                      // 价格
	Name               string  `json:"name"               orm:"name"                 description:"商品名称或小料名称(不用于业务显示)"`      // 商品名称或小料名称(不用于业务显示)
	ProductFlavorUuid  uint64  `json:"productFlavorUuid"  orm:"product_flavor_uuid"  description:"商品规格ID(仅商品使用)"`           // 商品规格ID(仅商品使用)
	ProductSauceUuid   uint64  `json:"productSauceUuid"   orm:"product_sauce_uuid"   description:"商品小料ID(仅小料使用)"`           // 商品小料ID(仅小料使用)
	ProductPackageUuid uint64  `json:"productPackageUuid" orm:"product_package_uuid" description:"商品包ID"`                   // 商品包ID
	StockNum           float64 `json:"stockNum"           orm:"stock_num"            description:"库存数量"`                    // 库存数量
	BarcodeValue       string  `json:"barcodeValue"       orm:"barcode_value"        description:"条形码值"`                    // 条形码值
	IsDefaultSelect    int     `json:"isDefaultSelect"    orm:"is_default_select"    description:"是否默认选择, 0-否 1-是"`         // 是否默认选择, 0-否 1-是
	Status             int     `json:"status"             orm:"status"               description:"状态, 0-下架 1-上架. 同步商品包的状态"` // 状态, 0-下架 1-上架. 同步商品包的状态
	IsSoldOut          int     `json:"isSoldOut"          orm:"is_sold_out"          description:"是否沽清, 0-否 1-是"`           // 是否沽清, 0-否 1-是
	ActualSaleNum      float64 `json:"actualSaleNum"      orm:"actual_sale_num"      description:"实际销量。每次卖出时,实际销量增加"`       // 实际销量。每次卖出时,实际销量增加
	CreateTime         uint    `json:"createTime"         orm:"create_time"          description:"创建时间(时间戳)"`               // 创建时间(时间戳)
	UpdateTime         uint    `json:"updateTime"         orm:"update_time"          description:"更新时间(时间戳)"`               // 更新时间(时间戳)
	DeleteTime         uint    `json:"deleteTime"         orm:"delete_time"          description:"删除时间(时间戳)"`               // 删除时间(时间戳)
}
