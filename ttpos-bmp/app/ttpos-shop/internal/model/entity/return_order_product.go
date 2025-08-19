// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ReturnOrderProduct is the golang structure for table return_order_product.
type ReturnOrderProduct struct {
	Id                   uint    `json:"id"                   orm:"id"                      description:"自增ID"`                                                                                             // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                    description:"退货单商品唯一标识符"`                                                                                       // 退货单商品唯一标识符
	SaleOrderUuid        uint64  `json:"saleOrderUuid"        orm:"sale_order_uuid"         description:"销售订单ID"`                                                                                           // 销售订单ID
	SaleOrderProductUuid uint64  `json:"saleOrderProductUuid" orm:"sale_order_product_uuid" description:"销售订单商品表ID"`                                                                                        // 销售订单商品表ID
	ReturnOrderUuid      uint64  `json:"returnOrderUuid"      orm:"return_order_uuid"       description:"退货单ID"`                                                                                            // 退货单ID
	ProductType          int     `json:"productType"          orm:"product_type"            description:"商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct"` // 商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct
	ProductPackageUuid   uint64  `json:"productPackageUuid"   orm:"product_package_uuid"    description:"商品包ID"`                                                                                            // 商品包ID
	ProductName          string  `json:"productName"          orm:"product_name"            description:"商品名称"`                                                                                             // 商品名称
	ProductPrice         float64 `json:"productPrice"         orm:"product_price"           description:"商品单价"`                                                                                             // 商品单价
	TaxRate              float64 `json:"taxRate"              orm:"tax_rate"                description:"税率,根据结账时税率计算"`                                                                                     // 税率,根据结账时税率计算
	Num                  int     `json:"num"                  orm:"num"                     description:"商品数量,退货的商品数量"`                                                                                     // 商品数量,退货的商品数量
	ProductDiscount      float64 `json:"productDiscount"      orm:"product_discount"        description:"商品折扣"`                                                                                             // 商品折扣
	ProductTotalAmount   float64 `json:"productTotalAmount"   orm:"product_total_amount"    description:"商品总金额（退款总金额）"`                                                                                     // 商品总金额（退款总金额）
	CreateTime           uint    `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`                                                                                        // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`                                                                                        // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`                                                                                        // 删除时间(时间戳)
}
