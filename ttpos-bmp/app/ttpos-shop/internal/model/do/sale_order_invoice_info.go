// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderInvoiceInfo is the golang structure of table ttpos_sale_order_invoice_info for DAO operations like Where/Data.
type SaleOrderInvoiceInfo struct {
	g.Meta           `orm:"table:ttpos_sale_order_invoice_info, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 唯一ID
	SaleOrderUuid    interface{} // 销售订单ID
	CompanyName      interface{} // 公司名称
	CompanyAddr      interface{} // 公司地址
	CompanyTaxNumber interface{} // 公司税号
	CompanyPhone     interface{} // 公司电话
	PrintNum         interface{} // 打印次数
	CreateTime       interface{} // 创建时间
	UpdateTime       interface{} // 更新时间
	DeleteTime       interface{} // 删除时间
}
