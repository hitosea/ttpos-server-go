package v1

// 加料 `jjjfood_feed`
type Feed struct {
	FeedID         int     `gorm:"primaryKey;autoIncrement;comment:加料id"`
	FeedName       string  `gorm:"type:varchar(2000);default:'';comment:加料名称"`
	Price          float64 `gorm:"type:decimal(12,2);default:0.00;comment:加料价格"`
	ShopSupplierID int     `gorm:"default:0;comment:门店id"`
	Sort           int     `gorm:"default:0;comment:排序"`
	AppID          int     `gorm:"default:0;comment:应用id"`
	CreateTime     int     `gorm:"not null;default:0;comment:创建时间"`
	UpdateTime     int     `gorm:"not null;default:0;comment:更新时间"`

	ProductFeedMaterials []ProductFeedMaterial `gorm:"foreignKey:FeedID;references:FeedID"`
}

// 产品加料材料关联表 `jjjfood_product_feed_material`
type ProductFeedMaterial struct {
	ID             uint    `gorm:"primaryKey;autoIncrement;comment:ID"`
	FeedID         int     `gorm:"default:0;comment:加料id"`
	ProductFeedID  int     `gorm:"default:0;comment:产品加料id"`
	MaterialID     int     `gorm:"default:0;comment:材料id"`
	MaterialNum    float64 `gorm:"type:decimal(12,4);default:0.0000;comment:使用库存数量"`
	ShopSupplierID int     `gorm:"default:0;comment:门店id"`
	AppID          int     `gorm:"default:0;comment:应用id"`
	CreateTime     int     `gorm:"not null;default:0;comment:创建时间"`
}
