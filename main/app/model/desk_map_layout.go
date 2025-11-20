package model

// DeskMapLayout 桌台地图布局表，存储各区域的桌台地图布局配置 ttpos_desk_map_layout
//
// 任务: story-admin-desktop-table-map Phase 2.2
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapLayout struct {
	BaseModel
	AreaUuid   uint64 `gorm:"column:area_uuid;type:bigint(20) unsigned;default:0;comment:'区域UUID';NOT NULL;uniqueIndex:uk_area_uuid" json:"area_uuid"`
	LayoutJson string `gorm:"column:layout_json;type:text;comment:'画布布局JSON（含桌台坐标、尺寸、样式等）';NOT NULL" json:"layout_json"`
}

// TableName 指定表名
func (DeskMapLayout) TableName() string {
	return GetTableName("desk_map_layout")
}
