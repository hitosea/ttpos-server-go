// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsProduct is the golang structure of table ttpos_statistics_product for DAO operations like Where/Data.
type StatisticsProduct struct {
	g.Meta             `orm:"table:ttpos_statistics_product, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // UUID
	SaleBillUuid       interface{} // 销售单UUID
	SaleOrderUuid      interface{} // 销售订单UUID
	DutyNo             interface{} // 当班编号
	DeskUuid           interface{} // 桌台UUID
	ProductPackageUuid interface{} // 商品包uuid
	ProductBomUuid     interface{} // 商品清单uuid
	ProductPrice       interface{} // 商品单价: 未含税
	ProductSalePrice   interface{} // 商品销售价: 规格+加料
	ProductFinalPrice  interface{} // 商品最终价
	ProductNum         interface{} // 商品数量
	TaxRate            interface{} // 税率
	TaxFee             interface{} // 税费
	ServiceFee         interface{} // 服务费
	ServiceTax         interface{} // 服务税
	GiveNum            interface{} // 赠菜数量
	FreeNum            interface{} // 免单数量
	RefundNum          interface{} // 退款数量
	CompleteTime       interface{} // 完成时间
	RefundTime         interface{} // 完成时间
	CreateTime         interface{} // 创建时间
	UpdateTime         interface{} // 更新时间
	DeleteTime         interface{} // 删除时间
}
