package request

// ImportMenuRequest 导入 Grab 菜单请求
type ImportMenuRequest struct {
	Platform string      `json:"platform" binding:"required,oneof=grab lineman hamburger"` // 平台名称：grab, lineman, hamburger 等
	MenuData interface{} `json:"menuData" binding:"required"`                              // 平台菜单 JSON 数据
}

// ExportMenuRequest 导出菜单请求
type ExportMenuRequest struct {
	Platform                  string `json:"platform" binding:"required"`     // 平台名称：grab, lineman 等
	CompanyUuid               uint64 `json:"company_uuid" binding:"required"` // 公司 UUID
	CurrencyUnit              string `json:"currency_unit"`                   // 货币单位
	IsSaveSingleFlavorMapping bool   `json:"is_save_single_flavor_mapping"`   // 是否保存单规格映射，默认不保存
}

// ToggleTakeoutStatusRequest 切换外卖状态请求
type ToggleTakeoutStatusRequest struct {
	Platform string `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	Enabled  bool   `json:"enabled"`                     // 是否开启外卖
}

// UpdateBindingStatusRequest 更新绑定状态请求
type UpdateBindingStatusRequest struct {
	Platform string `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
}

// GetImportLogsRequest 获取导入日志列表请求
type GetImportLogsRequest struct {
	Platform   string `form:"platform" json:"platform"`       // 外卖平台筛选
	ImportType *int8  `form:"import_type" json:"import_type"` // 导入类型筛选（1-TTPOS推送到平台 2-平台推送到TTPOS）
	Status     *int8  `form:"status" json:"status"`           // 状态筛选（0-进行中 1-成功 2-失败）
	PageNo     int    `form:"page_no" json:"page_no"`         // 页码
	PageSize   int    `form:"page_size" json:"page_size"`     // 每页数量
}

// PushTakeoutMenuRequest 推送菜单请求
type PushTakeoutMenuRequest struct {
	Platform string `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
}

// UpdateMenuItemRequest 更新菜单项请求
type UpdateMenuItemRequest struct {
	Platform        string `json:"platform" binding:"required,oneof=grab"` // 平台名称，目前仅支持 grab
	ItemId          string `json:"item_id" binding:"required"`             // 商品ID (partner item id)
	Price           *int64 `json:"price,omitempty"`                        // 价格 (单位：分)
	AvailableStatus string `json:"available_status,omitempty"`             // 可用状态: AVAILABLE, UNAVAILABLE, UNAVAILABLEHIDE
	MaxStock        *int64 `json:"max_stock,omitempty"`                    // 库存数量，设为 0 时需同时设置 AvailableStatus 为 UNAVAILABLE
}

// UpdateMenuModifierRequest 更新菜单修饰符请求
type UpdateMenuModifierRequest struct {
	Platform        string `json:"platform" binding:"required,oneof=grab"` // 平台名称，目前仅支持 grab
	ModifierId      string `json:"modifier_id" binding:"required"`         // 修饰符ID (partner modifier id)
	ModifierName    string `json:"modifier_name" binding:"required"`       // 修饰符名称 (用于定位记录)
	Price           *int64 `json:"price,omitempty"`                        // 价格 (单位：分)
	AvailableStatus string `json:"available_status,omitempty"`             // 可用状态: AVAILABLE, UNAVAILABLE
}

// PageReq 分页请求参数
type PageReq struct {
	PageNo   int `form:"page_no,default=1"  json:"page_no,default=1" binding:"omitempty,min=1"`               // 页码
	PageSize int `form:"page_size,default=20" json:"page_size,default=20" binding:"omitempty,min=1,max=1100"` // 每页大小
}
