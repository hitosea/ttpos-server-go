// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderBuffetCustomerType is the golang structure of table ttpos_sale_order_buffet_customer_type for DAO operations like Where/Data.
type SaleOrderBuffetCustomerType struct {
	g.Meta                      `orm:"table:ttpos_sale_order_buffet_customer_type, do:true"`
	Id                          interface{} // 自增ID
	Uuid                        interface{} // 销售订单顾客类型ID
	Name                        interface{} // 顾客类型名称
	Num                         interface{} // 人数
	SalePrice                   interface{} // 原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变
	Price                       interface{} // 最终单价（折后价），只进行自定义打折，不进行会员打折
	CustomDiscountRate          interface{} // 自定义折扣率, 值为0-1之间(0-100%)
	CustomDiscountFee           interface{} // 自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率
	TaxRate                     interface{} // 税率,值为0-1之间.加购时记录税率,结账时再重新核算
	ServiceTaxFee               interface{} // 服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率
	TaxFee                      interface{} // 自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率
	ServiceFee                  interface{} // 服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例
	TotalPrice                  interface{} // 应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费
	SaleOrderUuid               interface{} // 销售订单ID
	BuffetPackageUuid           interface{} // 自助餐套餐ID
	BuffetCustomerTypePriceUuid interface{} // 自助餐客户类型价格ID
	CreateTime                  interface{} // 创建时间(时间戳)
	UpdateTime                  interface{} // 更新时间(时间戳)
	DeleteTime                  interface{} // 删除时间(时间戳)
}
