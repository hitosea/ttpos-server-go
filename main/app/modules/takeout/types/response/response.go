package response

// BindingLinkResponse 绑定链接响应
type BindingLinkResponse struct {
	BindingLink string `json:"binding_link"` // 绑定链接 URL
}

// BindingStatusResponse 绑定状态响应
type BindingStatusResponse struct {
	IsBound bool `json:"is_bound"` // 是否已绑定
}

// GrabMenuResponse Grab 菜单响应
type GrabMenuResponse struct {
	Platform string      `json:"platform"` // 平台名称
	Menu     interface{} `json:"menu"`     // Grab 菜单数据
}

// TakeoutStatusResponse 外卖状态响应
type TakeoutStatusResponse struct {
	Platform  string `json:"platform"`   // 外卖平台
	Enabled   bool   `json:"enabled"`    // 是否开启
	IsBound   bool   `json:"is_bound"`   // 是否已绑定
	Skip      bool   `json:"skip"`       // 是否跳过绑定
	UpdatedAt int64  `json:"updated_at"` // 更新时间
}

// TakeoutStatusListResponse 外卖状态列表响应
type TakeoutStatusListResponse struct {
	List []*TakeoutStatusResponse `json:"list"` // 平台状态列表
}

// ImportProgressResponse 导入进度响应
type ImportProgressResponse struct {
	UUID            uint64 `json:"uuid"`             // 日志UUID
	Platform        string `json:"platform"`         // 外卖平台
	ImportType      int8   `json:"import_type"`      // 导入类型
	ImportDirection string `json:"import_direction"` // 导入方向描述
	Status          int8   `json:"status"`           // 导入状态(-1-未开启 0-进行中 1-成功 2-失败)
	Progress        int    `json:"progress"`         // 进度百分比(0-100)
	CurrentStep     string `json:"current_step"`     // 当前步骤描述
	SuccessCount    int    `json:"success_count"`    // 成功数量
	FailureCount    int    `json:"failure_count"`    // 失败数量
	TotalCount      int    `json:"total_count"`      // 总数量
	ErrorMessage    string `json:"error_message"`    // 错误信息
	StartTime       int64  `json:"start_time"`       // 开始时间
	EndTime         int64  `json:"end_time"`         // 结束时间
	Duration        int64  `json:"duration"`         // 耗时(秒)
	EstimatedTime   int64  `json:"estimated_time"`   // 预估剩余时间(秒)
}

// ImportLogResponse 导入日志响应
type ImportLogResponse struct {
	UUID            uint64 `json:"uuid"`             // 日志UUID
	Platform        string `json:"platform"`         // 外卖平台
	ImportType      int8   `json:"import_type"`      // 导入类型
	ImportDirection string `json:"import_direction"` // 导入方向描述
	Status          int8   `json:"status"`           // 导入状态(0-进行中 1-成功 2-失败)
	Progress        int    `json:"progress"`         // 进度百分比(0-100)
	SuccessCount    int    `json:"success_count"`    // 成功数量
	FailureCount    int    `json:"failure_count"`    // 失败数量
	TotalCount      int    `json:"total_count"`      // 总数量
	ErrorMessage    string `json:"error_message"`    // 错误信息
	StartTime       int64  `json:"start_time"`       // 开始时间
	EndTime         int64  `json:"end_time"`         // 结束时间
	Duration        int64  `json:"duration"`         // 耗时(秒)
	CreateTime      int64  `json:"create_time"`      // 创建时间
	CanReimport     bool   `json:"can_reimport"`     // 是否可以重新导入(true-可以  false-不可以)
}

// ImportLogListResponse 导入日志列表响应
type ImportLogListResponse struct {
	List     []ImportLogResponse `json:"list"`      // 日志列表
	PageInfo PageResponse        `json:"page_info"` // 分页信息
}

// PageResponse 分页响应
type PageResponse struct {
	PageNo   int   `json:"page_no"`   // 当前页码
	PageSize int   `json:"page_size"` // 每页大小
	Total    int64 `json:"total"`     // 总数
}

// UpdateMenuItemResponse 更新菜单项响应
type UpdateMenuItemResponse struct {
	MerchantId      string `json:"merchant_id"`      // 商户ID
	ItemId          string `json:"item_id"`          // 商品ID
	Price           int64  `json:"price"`            // 价格
	AvailableStatus string `json:"available_status"` // 可用状态
	MaxStock        int64  `json:"max_stock"`        // 库存数量
}

// UpdateMenuModifierResponse 更新菜单修饰符响应
type UpdateMenuModifierResponse struct {
	MerchantId      string `json:"merchant_id"`      // 商户ID
	ModifierId      string `json:"modifier_id"`      // 修饰符ID
	ModifierName    string `json:"modifier_name"`    // 修饰符名称
	Price           int64  `json:"price"`            // 价格
	AvailableStatus string `json:"available_status"` // 可用状态
}

// MenuSyncResult 菜单同步结果
type MenuSyncResult struct {
	TotalItems       int                  `json:"total_items"`       // 总商品数
	TotalModifiers   int                  `json:"total_modifiers"`   // 总修饰符数
	ChangedItems     int                  `json:"changed_items"`     // 变更商品数
	ChangedModifiers int                  `json:"changed_modifiers"` // 变更修饰符数
	SuccessItems     int                  `json:"success_items"`     // 成功更新商品数
	SuccessModifiers int                  `json:"success_modifiers"` // 成功更新修饰符数
	FailedItems      int                  `json:"failed_items"`      // 失败商品数
	FailedModifiers  int                  `json:"failed_modifiers"`  // 失败修饰符数
	ItemChanges      []MenuItemChange     `json:"item_changes"`      // 商品变更详情
	ModifierChanges  []MenuModifierChange `json:"modifier_changes"`  // 修饰符变更详情
	Errors           []string             `json:"errors,omitempty"`  // 错误信息列表
}

// MenuItemChange 菜单项变更详情
type MenuItemChange struct {
	ItemID       string `json:"item_id"`                 // 商品ID
	ItemName     string `json:"item_name"`               // 商品名称
	ChangeType   string `json:"change_type"`             // 变更类型: price, status, stock
	OldPrice     *int64 `json:"old_price,omitempty"`     // 旧价格
	NewPrice     *int64 `json:"new_price,omitempty"`     // 新价格
	OldStatus    string `json:"old_status,omitempty"`    // 旧状态
	NewStatus    string `json:"new_status,omitempty"`    // 新状态
	OldStock     *int64 `json:"old_stock,omitempty"`     // 旧库存
	NewStock     *int64 `json:"new_stock,omitempty"`     // 新库存
	Success      bool   `json:"success"`                 // 是否成功
	ErrorMessage string `json:"error_message,omitempty"` // 错误信息
}

// MenuModifierChange 修饰符变更详情
type MenuModifierChange struct {
	ModifierID   string `json:"modifier_id"`             // 修饰符ID
	ModifierName string `json:"modifier_name"`           // 修饰符名称
	ChangeType   string `json:"change_type"`             // 变更类型: price, status
	OldPrice     *int64 `json:"old_price,omitempty"`     // 旧价格
	NewPrice     *int64 `json:"new_price,omitempty"`     // 新价格
	OldStatus    string `json:"old_status,omitempty"`    // 旧状态
	NewStatus    string `json:"new_status,omitempty"`    // 新状态
	Success      bool   `json:"success"`                 // 是否成功
	ErrorMessage string `json:"error_message,omitempty"` // 错误信息
}
