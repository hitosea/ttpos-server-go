package model

// ProductMap 外卖平台商品与店内商品映射
type ProductMap struct {
	BaseModel
	Source             string `gorm:"column:source;type:varchar(50);default:'';index:idx_source_product_id;comment:'来源平台(grab/foodpanda等)'"`
	SourceProductId    string `gorm:"column:source_product_id;type:varchar(191);default:'';index:idx_source_product_id;comment:'来源平台商品唯一ID'"`
	ProductPackageUuid uint64 `gorm:"column:product_package_uuid;type:bigint(20) unsigned;default:0;comment:'店内商品包UUID'"`
	Status             int    `gorm:"column:status;type:int(11);default:1;comment:'状态 1-有效'"` // 预留
	SyncTime           int64  `gorm:"column:sync_time;type:bigint(20);default:0;comment:'同步时间戳'"`
}

func (ProductMap) TableName() string {
	return GetTableName("product_map")
}
