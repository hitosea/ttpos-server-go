package model

// GrabProductMap Grab 商品与店内商品映射
type GrabProductMap struct {
	BaseModel
	CompanyUuid        uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;index:idx_company_grab_product;comment:'公司UUID'"`
	GrabProductId      string `gorm:"column:grab_product_id;type:varchar(191);default:'';index:idx_company_grab_product;comment:'Grab 商品唯一ID'"`
	ProductPackageUuid uint64 `gorm:"column:product_package_uuid;type:bigint(20) unsigned;default:0;comment:'店内商品包UUID'"`
	Status             int    `gorm:"column:status;type:int(11);default:1;comment:'状态 1-有效'"` // 预留
	SyncTime           int64  `gorm:"column:sync_time;type:bigint(20);default:0;comment:'同步时间戳'"`
}

func (GrabProductMap) TableName() string {
	return GetTableName("grab_product_map")
}
