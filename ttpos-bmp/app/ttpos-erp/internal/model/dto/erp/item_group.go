package erp

// ItemGroupInfo 物品分组信息数据传输对象
// 用于定义物品分组的相关信息
type ItemGroupInfo struct {
	ItemGroupName   string `json:"item_group_name"`             // 分组名称，必填
	ParentItemGroup string `json:"parent_item_group,omitempty"` // 父级分组，可选
	IsGroup         bool   `json:"is_group,omitempty"`          // 是否为分组，可选
	Branch          string `json:"custom_branch,omitempty"`     // 分支名称，可选
	Company         string `json:"custom_company,omitempty"`    //公司名称, 可选
	AliasName       string `json:"alias_name,omitempty"`        //别名,可选
}

// ItemGroupListReq 获取物品分组列表请求数据传输对象
type ItemGroupListReq struct {
	Branch          string `json:"custom_branch,omitempty"`     // 分支名称，可选
	ItemGroupName   string `json:"item_group_name,omitempty"`   // 分组名称模糊查询，可选
	ParentItemGroup string `json:"parent_item_group,omitempty"` // 父级分组，可选
	Company         string `json:"custom_company,omitempty"`    // 公司名称，可选
}
