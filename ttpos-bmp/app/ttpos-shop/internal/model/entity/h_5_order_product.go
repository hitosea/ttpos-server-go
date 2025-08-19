// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// H5OrderProduct is the golang structure for table h5_order_product.
type H5OrderProduct struct {
	Id                   uint    `json:"id"                   orm:"id"                      description:"自增ID"`                                        // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                    description:"扫码订单商品uuid"`                                  // 扫码订单商品uuid
	Name                 string  `json:"name"                 orm:"name"                    description:"商品名称.接单和拒单后从sale_order_product表获取，不再改变"`      // 商品名称.接单和拒单后从sale_order_product表获取，不再改变
	Price                float64 `json:"price"                orm:"price"                   description:"最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变"` // 最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变
	SalePrice            float64 `json:"salePrice"            orm:"sale_price"              description:"销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变"`  // 销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变
	Num                  int     `json:"num"                  orm:"num"                     description:"最终商品数量.接单和拒单后从sale_order_product表获取，不再改变"`    // 最终商品数量.接单和拒单后从sale_order_product表获取，不再改变
	AttributeText        string  `json:"attributeText"        orm:"attribute_text"          description:"商品属性文本。接单和拒单后从sale_order_product表获取，不再改变"`    // 商品属性文本。接单和拒单后从sale_order_product表获取，不再改变
	Remark               string  `json:"remark"               orm:"remark"                  description:"备注。接单和拒单后从sale_order_product表获取，不再改变"`        // 备注。接单和拒单后从sale_order_product表获取，不再改变
	SaleOrderProductUuid uint64  `json:"saleOrderProductUuid" orm:"sale_order_product_uuid" description:"销售订单商品uuid"`                                  // 销售订单商品uuid
	H5OrderUuid          uint64  `json:"h5OrderUuid"          orm:"h5_order_uuid"           description:"扫码订单uuid"`                                    // 扫码订单uuid
	SaleBillUuid         uint64  `json:"saleBillUuid"         orm:"sale_bill_uuid"          description:"销售账单uuid"`                                    // 销售账单uuid
	CreateTime           uint    `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`                                   // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`                                   // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`                                   // 删除时间(时间戳)
}
