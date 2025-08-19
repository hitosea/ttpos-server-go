// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductPrinterProductItem is the golang structure for table product_printer_product_item.
type ProductPrinterProductItem struct {
	Id                 uint   `json:"id"                 orm:"id"                   description:"自增ID"`        // 自增ID
	Uuid               uint64 `json:"uuid"               orm:"uuid"                 description:"商品打印机商品关联ID"` // 商品打印机商品关联ID
	ProductPrinterUuid uint64 `json:"productPrinterUuid" orm:"product_printer_uuid" description:"商品打印机ID"`     // 商品打印机ID
	ProductPackageUuid uint64 `json:"productPackageUuid" orm:"product_package_uuid" description:"商品包ID"`       // 商品包ID
	CreateTime         uint   `json:"createTime"         orm:"create_time"          description:"创建时间(时间戳)"`   // 创建时间(时间戳)
	UpdateTime         uint   `json:"updateTime"         orm:"update_time"          description:"更新时间(时间戳)"`   // 更新时间(时间戳)
	DeleteTime         uint   `json:"deleteTime"         orm:"delete_time"          description:"删除时间(时间戳)"`   // 删除时间(时间戳)
}
