// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StatisticsProduct is the golang structure for table statistics_product.
type StatisticsProduct struct {
	Id                 uint    `json:"id"                 orm:"id"                   description:"自增ID"`         // 自增ID
	Uuid               uint64  `json:"uuid"               orm:"uuid"                 description:"UUID"`         // UUID
	SaleBillUuid       uint64  `json:"saleBillUuid"       orm:"sale_bill_uuid"       description:"销售单UUID"`      // 销售单UUID
	SaleOrderUuid      uint64  `json:"saleOrderUuid"      orm:"sale_order_uuid"      description:"销售订单UUID"`     // 销售订单UUID
	DutyNo             string  `json:"dutyNo"             orm:"duty_no"              description:"当班编号"`         // 当班编号
	DeskUuid           uint64  `json:"deskUuid"           orm:"desk_uuid"            description:"桌台UUID"`       // 桌台UUID
	ProductPackageUuid uint64  `json:"productPackageUuid" orm:"product_package_uuid" description:"商品包uuid"`      // 商品包uuid
	ProductBomUuid     uint64  `json:"productBomUuid"     orm:"product_bom_uuid"     description:"商品清单uuid"`     // 商品清单uuid
	ProductPrice       float64 `json:"productPrice"       orm:"product_price"        description:"商品单价: 未含税"`    // 商品单价: 未含税
	ProductSalePrice   float64 `json:"productSalePrice"   orm:"product_sale_price"   description:"商品销售价: 规格+加料"` // 商品销售价: 规格+加料
	ProductFinalPrice  float64 `json:"productFinalPrice"  orm:"product_final_price"  description:"商品最终价"`        // 商品最终价
	ProductNum         int     `json:"productNum"         orm:"product_num"          description:"商品数量"`         // 商品数量
	TaxRate            float64 `json:"taxRate"            orm:"tax_rate"             description:"税率"`           // 税率
	TaxFee             float64 `json:"taxFee"             orm:"tax_fee"              description:"税费"`           // 税费
	ServiceFee         float64 `json:"serviceFee"         orm:"service_fee"          description:"服务费"`          // 服务费
	ServiceTax         float64 `json:"serviceTax"         orm:"service_tax"          description:"服务税"`          // 服务税
	GiveNum            int     `json:"giveNum"            orm:"give_num"             description:"赠菜数量"`         // 赠菜数量
	FreeNum            int     `json:"freeNum"            orm:"free_num"             description:"免单数量"`         // 免单数量
	RefundNum          int     `json:"refundNum"          orm:"refund_num"           description:"退款数量"`         // 退款数量
	CompleteTime       uint    `json:"completeTime"       orm:"complete_time"        description:"完成时间"`         // 完成时间
	RefundTime         uint    `json:"refundTime"         orm:"refund_time"          description:"完成时间"`         // 完成时间
	CreateTime         uint    `json:"createTime"         orm:"create_time"          description:"创建时间"`         // 创建时间
	UpdateTime         uint    `json:"updateTime"         orm:"update_time"          description:"更新时间"`         // 更新时间
	DeleteTime         uint    `json:"deleteTime"         orm:"delete_time"          description:"删除时间"`         // 删除时间
}
