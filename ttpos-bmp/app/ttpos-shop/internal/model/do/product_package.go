// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPackage is the golang structure of table ttpos_product_package for DAO operations like Where/Data.
type ProductPackage struct {
	g.Meta                `orm:"table:ttpos_product_package, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 商品包ID
	Name                  interface{} // 商品包名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	ImageName             interface{} // 图片名称
	ImageFileUuid         interface{} // 图片ID
	DeductStockType       interface{} // 库存计算方法, 0-付款减库存 1-下单减库存
	UnitUuid              interface{} // 单位UUID
	DineTaxUuid           interface{} // 堂食税UUID
	CategoryUuid          interface{} // 类别UUID
	SpecialCategoryUuid   interface{} // 特殊类别UUID
	TakeoutTaxUuid        interface{} // 外卖税UUID
	PrinterTagUuid        interface{} // 打印机标签UUID
	SupplierUuid          interface{} // 供应商UUID
	Status                interface{} // 状态,0-下架 1-上架
	IsShowCashier         interface{} // 是否在收银设备显示, 0-否 1-是
	IsShowTablet          interface{} // 是否在平板设备显示, 0-否 1-是
	IsShowKitchen         interface{} // 是否在厨房设备显示, 0-否 1-是
	IsShowAssistant       interface{} // 是否在助手设备显示, 0-否 1-是
	IsShowH5              interface{} // 是否在H5设备显示, 0-否 1-是
	Sort                  interface{} // 排序
	LimitNum              interface{} // 限购数量
	SauceRequired         interface{} // 是否必选小料, 0-否 1-是
	SauceMaxSelection     interface{} // 小料最大选择数量
	Describe              interface{} // 卖点描述
	OpenDiscount          interface{} // 是否开启会员折扣, 0-否 1-是
	ActualSaleNum         interface{} // 实际销量。每次卖出时,实际销量增加
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
