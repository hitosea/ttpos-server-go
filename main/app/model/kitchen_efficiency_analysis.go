package model

// KitchenEfficiencyAnalysis 后厨效率分析表 `ttpos_kitchen_efficiency_analysis`
type KitchenEfficiencyAnalysis struct {
	BaseModel
	ProductPackageUuid uint64  `gorm:"column:product_package_uuid;type:bigint(20) unsigned;not null" json:"product_package_uuid"` // 商品包UUID
	Min                float64 `gorm:"column:min;type:int(10) unsigned;not null" json:"min"`                                      // 最短出品时长
	Max                float64 `gorm:"column:max;type:int(10) unsigned;not null" json:"max"`                                      // 最长出品时长
	Avg                float64 `gorm:"column:avg;type:int(10) unsigned;not null" json:"avg"`                                      // 平均出品时长
	Total              float64 `gorm:"column:total;type:int(10) unsigned;not null" json:"total"`                                  // 总出品时长
	Count              float64 `gorm:"column:count;type:int(10) unsigned;not null" json:"count"`                                  // 出品次数
	Date               int64   `gorm:"column:date;type:int(10) unsigned;not null" json:"date"`                                    // 统计日期(unix时间戳)
	DateString         string  `gorm:"column:date_string;type:varchar(255);not null" json:"date_string"`                          // 统计日期,某天零点的时间戳.一个商品一天只有唯一的一条记录
	Timezone           string  `gorm:"column:timezone;type:varchar(255);not null" json:"timezone"`                                // 时区
}

// TableName 获取表名
func (m *KitchenEfficiencyAnalysis) TableName() string {
	return "ttpos_kitchen_efficiency_analysis"
}
