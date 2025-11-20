package resp

// DeskMapAreaItem 区域列表项
//
// 任务: story-admin-desktop-table-map Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapAreaItem struct {
	AreaUuid     uint64 `json:"area_uuid"`     // 区域UUID
	AreaName     string `json:"area_name"`     // 区域名称
	DeskCount    uint   `json:"desk_count"`    // 桌台数量
	LayoutStatus string `json:"layout_status"` // 布局状态: set-已配置, unset-未配置
}

// DeskMapAreaListResp 区域列表响应
//
// 任务: story-admin-desktop-table-map Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapAreaListResp struct {
	List []DeskMapAreaItem `json:"list"` // 区域列表
}

// DeskMapTableItem 桌台基础信息项
//
// 任务: story-admin-desktop-table-map Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapTableItem struct {
	DeskUuid uint64 `json:"desk_uuid"` // 桌台UUID
	DeskName string `json:"desk_name"` // 桌台名称
	RangeMin uint   `json:"range_min"` // 最少人数
	RangeMax uint   `json:"range_max"` // 最多人数
	Selected bool   `json:"selected"`  // 是否已选中（在布局中）
}

// DeskMapLayoutResp 布局详情响应
//
// 任务: story-admin-desktop-table-map Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type DeskMapLayoutResp struct {
	Area   DeskMapAreaInfo    `json:"area"`   // 区域信息
	Desks  []DeskMapTableItem `json:"desks"`  // 桌台列表
	Layout DeskMapLayoutData  `json:"layout"` // 布局数据
}

// DeskMapAreaInfo 区域信息
type DeskMapAreaInfo struct {
	AreaUuid uint64 `json:"area_uuid"` // 区域UUID
	AreaName string `json:"area_name"` // 区域名称
}

// DeskMapLayoutData 布局数据（JSON 解析后的结构）
type DeskMapLayoutData struct {
	Desks []DeskMapLayoutTable `json:"desks"` // 桌台布局列表
}

// DeskMapLayoutTable 单个桌台的布局信息
type DeskMapLayoutTable struct {
	DeskUuid uint64  `json:"desk_uuid"` // 桌台UUID
	Shape    string  `json:"shape"`     // 形状: circle-圆形, rect-矩形
	RangeMin uint    `json:"range_min"` // 最少人数
	RangeMax uint    `json:"range_max"` // 最多人数
	X        float64 `json:"x"`         // X坐标
	Y        float64 `json:"y"`         // Y坐标
	Width    float64 `json:"width"`     // 宽度
	Height   float64 `json:"height"`    // 高度
	Rotation float64 `json:"rotation"`  // 旋转角度
}
