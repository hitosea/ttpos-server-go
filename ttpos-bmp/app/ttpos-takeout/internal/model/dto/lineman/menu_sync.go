// Package lineman 定义 Lineman API 数据传输对象
package lineman

// ============================================================================
// 菜单同步 API 数据结构
// ============================================================================

// MenuSyncRequest Lineman 菜单同步请求
type MenuSyncRequest struct {
	MenuGroups []*MenuGroup `json:"menuGroups"`
}

// MenuSyncResponse Lineman 菜单同步响应
type MenuSyncResponse struct {
	BaseResponse             // 嵌入公共响应字段
	MenuSyncRequestId string `json:"menuSyncRequestId,omitempty"` // 请求ID
}

// ============================================================================
// 菜单数据结构
// ============================================================================

// MenuGroup 菜单分类
type MenuGroup struct {
	ID             string      `json:"id"`             // 分类ID（value_object.PrefixCategory{id}）
	Name           NameTrans   `json:"name"`           // 分类名称（多语言）
	UseSellingTime bool        `json:"useSellingTime"` // 是否使用时段销售（固定false）
	MenuItems      []*MenuItem `json:"menuItems"`      // 商品列表
}

// MenuItem 菜单商品
type MenuItem struct {
	ID                        string                `json:"id"`                        // 商品ID（TTPOS-ITEM-{id}）
	Name                      NameTrans             `json:"name"`                      // 商品名称
	Description               DescTrans             `json:"description"`               // 商品描述
	Price                     float64               `json:"price"`                     // 价格（分）
	PhotoUrl                  string                `json:"photoUrl,omitempty"`        // 图片URL
	MenuStatus                string                `json:"menuStatus"`                // 状态：AVAILABLE, SOLD_OUT_TODAY, SUSPENDED
	SalesChannelsAvailability *ChannelsAvailability `json:"salesChannelsAvailability"` // 渠道可用性
	Properties                []*Property           `json:"properties,omitempty"`      // 规格/属性
}

// Property 属性组（规格/加料/属性）
type Property struct {
	ID     string       `json:"id"`            // 属性组ID（TTPOS-PROP-{id}）
	Name   NameTrans    `json:"name"`          // 属性组名称
	Min    int          `json:"min"`           // 最小选择数
	Max    int          `json:"max,omitempty"` // 最大选择数
	Type   string       `json:"type"`          // 类型：1=单选 2=复选
	Values []*PropValue `json:"values"`        // 属性值列表
}

// PropValue 属性值
type PropValue struct {
	ID     string    `json:"id"`     // 属性值ID（TTPOS-PROPVAL-{id}）
	Name   NameTrans `json:"name"`   // 属性值名称
	Price  float64   `json:"price"`  // 价格（分）
	Status int       `json:"status"` // 状态：1=可用 2=售罄 3=暂停
}

// ============================================================================
// 通用数据结构
// ============================================================================

// NameTrans 名称翻译
type NameTrans struct {
	Thai    string `json:"thai"`    // 泰语
	English string `json:"english"` // 英语
}

// DescTrans 描述翻译
type DescTrans struct {
	Thai    string `json:"thai"`
	English string `json:"english"`
}

// ChannelsAvailability 渠道可用性
type ChannelsAvailability struct {
	Delivery bool `json:"delivery"` // 配送（固定true）
	Pickup   bool `json:"pickup"`   // 自提（固定true）
}
