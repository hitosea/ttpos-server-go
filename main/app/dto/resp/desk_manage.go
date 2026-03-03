package resp

// DeskRegionListResp 桌台区域列表响应
type DeskRegionListResp struct {
	List []DeskRegionItem `json:"list"` // 区域列表
}

// DeskRegionItem 桌台区域项
type DeskRegionItem struct {
	Uuid      uint64 `json:"uuid"`       // 区域UUID
	Name      string `json:"name"`       // 区域名称
	Sort      uint   `json:"sort"`       // 排序
	DeskCount int64  `json:"desk_count"` // 桌台数量
}

// DeskTypeListResp 桌台类型列表响应
type DeskTypeListResp struct {
	List []DeskTypeItem `json:"list"` // 类型列表
}

// DeskTypeItem 桌台类型项
type DeskTypeItem struct {
	Uuid     uint64 `json:"uuid"`      // 类型UUID
	Name     string `json:"name"`      // 类型名称
	Sort     uint   `json:"sort"`      // 排序
	RangeMin uint   `json:"range_min"` // 最少人数
	RangeMax uint   `json:"range_max"` // 最多人数
}

// DeskManageListResp 桌台管理列表响应
type DeskManageListResp struct {
	List  []DeskManageItem `json:"list"`  // 桌台列表
	Total int64            `json:"total"` // 总数
}

// DeskManageItem 桌台管理项
type DeskManageItem struct {
	Uuid                   uint64 `json:"uuid"`                       // 桌台UUID
	DeskNo                 string `json:"desk_no"`                    // 桌位编号
	RegionUuid             uint64 `json:"region_uuid"`                // 区域UUID
	RegionName             string `json:"region_name"`                // 区域名称
	TypeUuid               uint64 `json:"type_uuid"`                  // 类型UUID
	TypeName               string `json:"type_name"`                  // 类型名称
	Sort                   uint   `json:"sort"`                       // 排序
	Status                 uint   `json:"status"`                     // 状态 0-未开台 1-已开台
	IsDisable              uint   `json:"is_disable"`                 // 是否禁用 0-否 1-是
	DefaultPeopleNum       uint   `json:"default_people_num"`         // 默认人数
	IsOpenDefaultPeopleNum uint   `json:"is_open_default_people_num"` // 是否开启默认人数
}
