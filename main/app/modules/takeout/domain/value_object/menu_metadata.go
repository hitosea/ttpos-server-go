package value_object

import (
	"github.com/grab/grabfood-api-sdk-go"
)

// SingleFlavorInfo 单规格信息
// 用于记录被跳过推送的单规格商品的规格信息
type SingleFlavorInfo struct {
	ProductPackageUuid uint64 `json:"product_package_uuid"` // 商品UUID
	FlavorBomUuid      uint64 `json:"flavor_bom_uuid"`      // 规格BOM UUID
	FlavorName         string `json:"flavor_name"`          // 规格名称
	FlavorPrice        int64  `json:"flavor_price"`         // 规格价格（分）
}

// TtposMenuWithMetadata 带元数据的TTPOS菜单
// 包装 Grab 菜单格式，额外存储单规格映射等元数据
type TtposMenuWithMetadata struct {
	// 原始 Grab 格式菜单数据
	Menu *grabfood.GetMenuNewResponse `json:"menu"`

	// 单规格映射表: key = 商品ID (如 "TTPOS-ITEM-123456"), value = 单规格信息
	// 任务 #39752: 记录被跳过推送的单规格信息，订单回传时用于自动补全
	SingleFlavorMapping map[string]*SingleFlavorInfo `json:"single_flavor_mapping,omitempty"`
}

// NewTtposMenuWithMetadata 创建带元数据的菜单
func NewTtposMenuWithMetadata(menu *grabfood.GetMenuNewResponse) *TtposMenuWithMetadata {
	return &TtposMenuWithMetadata{
		Menu:                menu,
		SingleFlavorMapping: make(map[string]*SingleFlavorInfo),
	}
}

// AddSingleFlavorMapping 添加单规格映射
func (m *TtposMenuWithMetadata) AddSingleFlavorMapping(itemId string, info *SingleFlavorInfo) {
	if m.SingleFlavorMapping == nil {
		m.SingleFlavorMapping = make(map[string]*SingleFlavorInfo)
	}
	m.SingleFlavorMapping[itemId] = info
}

// GetSingleFlavorInfo 获取单规格信息
func (m *TtposMenuWithMetadata) GetSingleFlavorInfo(itemId string) *SingleFlavorInfo {
	if m.SingleFlavorMapping == nil {
		return nil
	}
	return m.SingleFlavorMapping[itemId]
}
