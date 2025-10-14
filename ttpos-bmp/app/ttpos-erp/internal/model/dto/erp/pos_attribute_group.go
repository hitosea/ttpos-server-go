package erp

// PosAttributeGroup POS属性组数据传输对象
// 用于定义POS系统中商品属性组的相关信息
type PosAttributeGroup struct {
	GroupCode     string               `json:"group_code,omitempty"`     // 组代码，必填
	GroupName     string               `json:"group_name,omitempty"`     // 组名称
	IsRequired    bool                 `json:"is_required,omitempty"`    // 是否必选
	IsActive      bool                 `json:"is_active,omitempty"`      // 是否激活
	MinSelect     int                  `json:"min_select,omitempty"`     // 最小选择数量，必填
	MaxSelect     int                  `json:"max_select,omitempty"`     // 最大选择数量，必填
	SortOrder     int                  `json:"sort_order,omitempty"`     // 排序顺序
	AttributeList []PosAttributeOption `json:"attribute_list,omitempty"` // 属性选项列表，必填
	ItemGroup     string               `json:"item_group,omitempty"`     // 物品组
}

// PosAttributeOption POS属性选项数据传输对象
// 用于定义属性组中的具体选项信息
type PosAttributeOption struct {
	Item          string `json:"item,omitempty"`           // 物品，必填
	AttributeName string `json:"attribute_name,omitempty"` // 属性名称
	IsDefault     bool   `json:"is_default,omitempty"`     // 是否默认
	Quantity      string `json:"quantity,omitempty"`       // 数量
	SortOrder     int    `json:"sort_order,omitempty"`     // 排序顺序
}
