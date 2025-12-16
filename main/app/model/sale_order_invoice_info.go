package model

// SaleOrderInvoiceInfo 销售订单发票信息 `ttpos_sale_order_invoice_info`
type SaleOrderInvoiceInfo struct {
	BaseModel
	// 关联ID字段
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:'销售订单ID';index" json:"sale_order_uuid"`
	InvoiceNumber string `gorm:"column:invoice_number;type:varchar(64);default:'';comment:'发票编号'" json:"invoice_number"`

	// 发票信息字段
	CompanyName      string `gorm:"column:company_name;type:varchar(255);comment:'公司名称'" json:"company_name"`
	CompanyAddr      string `gorm:"column:company_addr;type:varchar(255);comment:'公司地址'" json:"company_addr"`
	CompanyTaxNumber string `gorm:"column:company_tax_number;type:varchar(100);comment:'公司税号'" json:"company_tax_number"`
	CompanyPhone     string `gorm:"column:company_phone;type:varchar(50);comment:'公司电话'" json:"company_phone"`
	PrintNum         int    `gorm:"column:print_num;type:int(11);default:0;comment:'打印次数'" json:"print_num"`
}

// HasContent 检查是否有内容需要打印
func (i *SaleOrderInvoiceInfo) HasContent() bool {
	return i != nil && (i.CompanyName != "" || i.CompanyAddr != "" || i.CompanyTaxNumber != "" || i.CompanyPhone != "")
}

func (i *SaleOrderInvoiceInfo) GetInvoiceNumber() string {
	return i.InvoiceNumber
}
