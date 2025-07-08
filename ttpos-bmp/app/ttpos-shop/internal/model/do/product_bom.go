// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductBom is the golang structure of table ttpos_product_bom for DAO operations like Where/Data.
type ProductBom struct {
	g.Meta             `orm:"table:ttpos_product_bom, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 商品BOM ID
	PurchasePrice      interface{} // 采购单价
	Price              interface{} // 价格
	Name               interface{} // 商品名称或小料名称(不用于业务显示)
	ProductFlavorUuid  interface{} // 商品规格ID(仅商品使用)
	ProductSauceUuid   interface{} // 商品小料ID(仅小料使用)
	ProductPackageUuid interface{} // 商品包ID
	StockNum           interface{} // 库存数量
	BarcodeValue       interface{} // 条形码值
	IsDefaultSelect    interface{} // 是否默认选择, 0-否 1-是
	Status             interface{} // 状态, 0-下架 1-上架. 同步商品包的状态
	IsSoldOut          interface{} // 是否沽清, 0-否 1-是
	ActualSaleNum      interface{} // 实际销量。每次卖出时,实际销量增加
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
