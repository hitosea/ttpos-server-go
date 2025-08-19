// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderInvoiceInfo is the golang structure for table sale_order_invoice_info.
type SaleOrderInvoiceInfo struct {
	Id               uint   `json:"id"               orm:"id"                 description:"自增ID"`   // 自增ID
	Uuid             uint64 `json:"uuid"             orm:"uuid"               description:"唯一ID"`   // 唯一ID
	SaleOrderUuid    uint64 `json:"saleOrderUuid"    orm:"sale_order_uuid"    description:"销售订单ID"` // 销售订单ID
	CompanyName      string `json:"companyName"      orm:"company_name"       description:"公司名称"`   // 公司名称
	CompanyAddr      string `json:"companyAddr"      orm:"company_addr"       description:"公司地址"`   // 公司地址
	CompanyTaxNumber string `json:"companyTaxNumber" orm:"company_tax_number" description:"公司税号"`   // 公司税号
	CompanyPhone     string `json:"companyPhone"     orm:"company_phone"      description:"公司电话"`   // 公司电话
	PrintNum         uint   `json:"printNum"         orm:"print_num"          description:"打印次数"`   // 打印次数
	CreateTime       uint64 `json:"createTime"       orm:"create_time"        description:"创建时间"`   // 创建时间
	UpdateTime       uint64 `json:"updateTime"       orm:"update_time"        description:"更新时间"`   // 更新时间
	DeleteTime       uint64 `json:"deleteTime"       orm:"delete_time"        description:"删除时间"`   // 删除时间
}
