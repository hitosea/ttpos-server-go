package model

// 仓库表 `ttpos_warehouse`
type Warehouse struct {
	BaseModel
	Name                  string `gorm:"column:name;type:text;comment:名称;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;not null;default:0;comment:'多语言名称UUID'" json:"multi_language_name_uuid"`
	Type                  string `gorm:"column:type;type:varchar(255);comment:仓库类型;NOT NULL" json:"type"`
	Code                  string `gorm:"column:code;type:varchar(255);comment:仓库编码;NOT NULL" json:"code"`
	Status                int    `gorm:"column:status;type:int(11);default:0;comment:仓库状态;NOT NULL" json:"status"`
	Contact               string `gorm:"column:contact;type:varchar(255);comment:联系人;NOT NULL" json:"contact"`
	Phone                 string `gorm:"column:phone;type:varchar(255);comment:联系电话;NOT NULL" json:"phone"`
	Address               string `gorm:"column:address;type:varchar(255);comment:地址;NOT NULL" json:"address"`
	IsDefault             int    `gorm:"column:is_default;type:int(11);default:0;comment:是否默认：0-否；1-是;NOT NULL" json:"is_default"`
	ErpCode               string `gorm:"column:erp_code;type:varchar(255);comment:关联erpnext;NOT NULL" json:"erp_code"`
	HeadquarterUuid       uint64 `gorm:"column:headquarter_uuid;not null;default:0;comment:'总部Uuid'" json:"headquarter_uuid"`

	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	Items             []*WarehouseItem   `gorm:"foreignKey:warehouse_uuid;references:uuid"`
}

func (w Warehouse) IsTransit() bool {
	return w.Type == "transit"
}

func (w Warehouse) IsHeadquarter() bool {
	return w.HeadquarterUuid > 0
}

// 是否禁用
func (w Warehouse) IsDisabled() bool {
	return w.Status == 0
}
