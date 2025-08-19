// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Material is the golang structure of table ttpos_material for DAO operations like Where/Data.
type Material struct {
	g.Meta                `orm:"table:ttpos_material, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 原料ID
	Name                  interface{} // 原料名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	CategoryUuid          interface{} // 类别ID
	SupplierUuid          interface{} // 供应商ID
	ImageUuid             interface{} // 图片ID
	ImageName             interface{} // 图片名称
	UnitUuid              interface{} // 单位ID
	Price                 interface{} // 采购单价
	StockNum              interface{} // 库存数量
	BarcodeValue          interface{} // 条形码值
	Status                interface{} // 状态, 1-上架 0-下架
	ActualSaleNum         interface{} // 实际销量。每次卖出时,实际销量增加
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
