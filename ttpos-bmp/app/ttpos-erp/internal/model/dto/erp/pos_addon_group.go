package erp

// PosAddonGroup POS附加项组数据传输对象
// 用于定义POS系统中附加项组的相关信息
type PosAddonGroup struct {
	GroupCode  string           `json:"group_code,omitempty"`  // 组代码，必填
	GroupName  string           `json:"group_name,omitempty"`  // 组名称
	IsRequired bool             `json:"is_required,omitempty"` // 是否必选
	IsActive   bool             `json:"is_active,omitempty"`   // 是否激活
	MinSelect  int              `json:"min_select,omitempty"`  // 最小选择数量，必填
	MaxSelect  int              `json:"max_select,omitempty"`  // 最大选择数量，必填
	SortOrder  int              `json:"sort_order,omitempty"`  // 排序顺序
	AddonList  []PosAddonOption `json:"addon_list,omitempty"`  // 附加项选项列表，必填
	ItemGroup  string           `json:"item_group,omitempty"`  // 物品组
}

// PosAddonOption POS附加项选项数据传输对象
// 用于定义附加项组中的具体选项信息
type PosAddonOption struct {
	// 这里需要根据实际的 Pos Addon Option 结构来定义字段
	// 暂时定义基础字段，可根据实际需求调整
	OptionCode string  `json:"option_code,omitempty"` // 选项代码
	OptionName string  `json:"option_name,omitempty"` // 选项名称
	Price      float64 `json:"price,omitempty"`       // 价格
	IsActive   bool    `json:"is_active,omitempty"`   // 是否激活
}
