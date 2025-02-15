package model

// Tax 税种表,定义税种的相关信息 ttpos_tax
type Tax struct {
	BaseModel
	Name    string  `gorm:"column:name;type:varchar(255);not null;default:'';comment:'名称'"`
	TaxRate float64 `gorm:"column:tax_rate;type:decimal(12,2);not null;default:0.00;comment:'税率'"`
}
