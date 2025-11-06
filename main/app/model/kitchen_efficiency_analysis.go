package model

// KitchenEfficiencyAnalysis 后厨效率分析表 `ttpos_kitchen_efficiency_analysis`
type KitchenEfficiencyAnalysis struct {
	ID                 uint64 `gorm:"primaryKey;column:id;type:int(10) unsigned;not null;autoIncrement" json:"id"` // 主键ID
	UUID               uint64 `gorm:"uniqueIndex:idx_uuid;column:uuid;type:bigint(20) unsigned;not null" json:"uuid"`
	CompanyUUID        uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;not null" json:"company_uuid"`                 // 公司UUID
	ProductPackageUUID uint64 `gorm:"column:product_package_uuid;type:bigint(20) unsigned;not null" json:"product_package_uuid"` // 商品包UUID
	Min                uint   `gorm:"column:min;type:int(10) unsigned;not null" json:"min"`                                      // 最短出品时长
	Max                uint   `gorm:"column:max;type:int(10) unsigned;not null" json:"max"`                                      // 最长出品时长
	Avg                uint   `gorm:"column:avg;type:int(10) unsigned;not null" json:"avg"`                                      // 平均出品时长
	Total              uint   `gorm:"column:total;type:int(10) unsigned;not null" json:"total"`                                  // 总出品时长
	Count              uint   `gorm:"column:count;type:int(10) unsigned;not null" json:"count"`                                  // 出品次数
	Date               uint   `gorm:"column:date;type:int(10) unsigned;not null" json:"date"`                                    // 统计日期(unix时间戳)
	CreateTime         uint   `gorm:"column:create_time;type:int(10) unsigned;not null" json:"create_time"`                      // 创建时间
	UpdateTime         uint   `gorm:"column:update_time;type:int(10) unsigned;not null" json:"update_time"`                      // 更新时间
	DeleteTime         uint   `gorm:"column:delete_time;type:int(10) unsigned;not null" json:"delete_time"`                      // 删除时间
}

// TableName 获取表名
func (m *KitchenEfficiencyAnalysis) TableName() string {
	return "ttpos_kitchen_efficiency_analysis"
}
