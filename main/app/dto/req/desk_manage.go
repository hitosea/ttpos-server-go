package req

// DeskRegionAddReq 桌台区域添加请求
type DeskRegionAddReq struct {
	Name string `json:"name" binding:"required"` // 区域名称
	Sort uint   `json:"sort"`                    // 排序
}

// DeskRegionEditReq 桌台区域编辑请求
type DeskRegionEditReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 区域UUID
	Name string `json:"name" binding:"required"` // 区域名称
	Sort uint   `json:"sort"`                    // 排序
}

// DeskRegionDeleteReq 桌台区域删除请求
type DeskRegionDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 区域UUID
}

// DeskTypeAddReq 桌台类型添加请求
type DeskTypeAddReq struct {
	Name     string `json:"name" binding:"required"` // 类型名称
	Sort     uint   `json:"sort"`                    // 排序
	RangeMin uint   `json:"range_min"`               // 最少人数
	RangeMax uint   `json:"range_max"`               // 最多人数
}

// DeskTypeEditReq 桌台类型编辑请求
type DeskTypeEditReq struct {
	Uuid     uint64 `json:"uuid" binding:"required"` // 类型UUID
	Name     string `json:"name" binding:"required"` // 类型名称
	Sort     uint   `json:"sort"`                    // 排序
	RangeMin uint   `json:"range_min"`               // 最少人数
	RangeMax uint   `json:"range_max"`               // 最多人数
}

// DeskTypeDeleteReq 桌台类型删除请求
type DeskTypeDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 类型UUID
}

// DeskAddReq 桌台添加请求
type DeskAddReq struct {
	DeskNo                 string `json:"desk_no" binding:"required"` // 桌位编号
	RegionUuid             uint64 `json:"region_uuid"`                // 区域UUID
	TypeUuid               uint64 `json:"type_uuid"`                  // 类型UUID
	Sort                   uint   `json:"sort"`                       // 排序
	DefaultPeopleNum       uint   `json:"default_people_num"`         // 默认人数
	IsOpenDefaultPeopleNum uint   `json:"is_open_default_people_num"` // 是否开启默认人数 0-否 1-是
}

// DeskEditReq 桌台编辑请求
type DeskEditReq struct {
	Uuid                   uint64 `json:"uuid" binding:"required"`    // 桌台UUID
	DeskNo                 string `json:"desk_no" binding:"required"` // 桌位编号
	RegionUuid             uint64 `json:"region_uuid"`                // 区域UUID
	TypeUuid               uint64 `json:"type_uuid"`                  // 类型UUID
	Sort                   uint   `json:"sort"`                       // 排序
	IsDisable              uint   `json:"is_disable"`                 // 是否禁用 0-否 1-是
	DefaultPeopleNum       uint   `json:"default_people_num"`         // 默认人数
	IsOpenDefaultPeopleNum uint   `json:"is_open_default_people_num"` // 是否开启默认人数 0-否 1-是
}

// DeskDeleteReq 桌台删除请求
type DeskDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 桌台UUID
}

// DeskListReq 桌台列表请求
type DeskManageListReq struct {
	PageNo     int    `form:"page_no,default=1"`    // 页码
	PageSize   int    `form:"page_size,default=20"` // 每页数量
	RegionUuid uint64 `form:"region_uuid"`          // 区域UUID筛选
	TypeUuid   uint64 `form:"type_uuid"`            // 类型UUID筛选
	Keyword    string `form:"keyword"`              // 关键词搜索
}
