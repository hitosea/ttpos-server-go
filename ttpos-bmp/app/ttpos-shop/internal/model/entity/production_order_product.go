// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductionOrderProduct is the golang structure for table production_order_product.
type ProductionOrderProduct struct {
	Id                    uint   `json:"id"                    orm:"id"                      description:"自增ID"`                        // 自增ID
	Uuid                  uint64 `json:"uuid"                  orm:"uuid"                    description:"生产订单商品ID"`                    // 生产订单商品ID
	Name                  string `json:"name"                  orm:"name"                    description:"名称"`                          // 名称
	Num                   int    `json:"num"                   orm:"num"                     description:"商品数量"`                        // 商品数量
	FlavorName            string `json:"flavorName"            orm:"flavor_name"             description:"规格名称,不随后台改变"`                 // 规格名称,不随后台改变
	ProductAttributeNames string `json:"productAttributeNames" orm:"product_attribute_names" description:"商品属性名称,多个属性名用逗号分隔,不随后台改变"`    // 商品属性名称,多个属性名用逗号分隔,不随后台改变
	ProductSaucesNames    string `json:"productSaucesNames"    orm:"product_sauces_names"    description:"商品加料名称,多个加料名用逗号分隔,不随后台改变"`    // 商品加料名称,多个加料名用逗号分隔,不随后台改变
	Status                int    `json:"status"                orm:"status"                  description:"状态, 0-待制作 1-制作中 2-已完成 3-已退菜"` // 状态, 0-待制作 1-制作中 2-已完成 3-已退菜
	Remark                string `json:"remark"                orm:"remark"                  description:"商品备注"`                        // 商品备注
	HasMaterial           int    `json:"hasMaterial"           orm:"has_material"            description:"是否无原料, 0-无原料,商品没有关联原料 1-有原料"` // 是否无原料, 0-无原料,商品没有关联原料 1-有原料
	SaleBillUuid          uint64 `json:"saleBillUuid"          orm:"sale_bill_uuid"          description:"销售账单ID"`                      // 销售账单ID
	ProductPackageUuid    uint64 `json:"productPackageUuid"    orm:"product_package_uuid"    description:"商品包ID"`                       // 商品包ID
	SaleOrderProductUuid  uint64 `json:"saleOrderProductUuid"  orm:"sale_order_product_uuid" description:"销售订单商品ID"`                    // 销售订单商品ID
	ProductionOrderUuid   uint64 `json:"productionOrderUuid"   orm:"production_order_uuid"   description:"生产订单ID"`                      // 生产订单ID
	FirstCategoryUuid     uint64 `json:"firstCategoryUuid"     orm:"first_category_uuid"     description:"一级分类ID"`                      // 一级分类ID
	FinishedTime          uint   `json:"finishedTime"          orm:"finished_time"           description:"完成时间(时间戳)"`                   // 完成时间(时间戳)
	CreateTime            uint   `json:"createTime"            orm:"create_time"             description:"创建时间(时间戳),送厨时间"`              // 创建时间(时间戳),送厨时间
	UpdateTime            uint   `json:"updateTime"            orm:"update_time"             description:"更新时间(时间戳)"`                   // 更新时间(时间戳)
	DeleteTime            uint   `json:"deleteTime"            orm:"delete_time"             description:"删除时间(时间戳)"`                   // 删除时间(时间戳)
}
