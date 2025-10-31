package resp

import "ttpos-server-go/app/dto"

// ProductLabelListResp 商品标签列表响应
type ProductLabelListResp struct {
	List []ProductLabelDetail `json:"list"`
}

// ProductLabelDetail 商品标签详情
type ProductLabelDetail struct {
	Uuid                uint64                    `json:"uuid"`                  // 标签UUID
	Name                string                    `json:"name"`                  // 标签名称
	Style               string                    `json:"style"`                 // 标签样式
	IsShowCashier       uint                      `json:"is_show_cashier"`       // 是否在收银机显示, 0-否 1-是
	IsShowTablet        uint                      `json:"is_show_tablet"`        // 是否在平板显示, 0-否 1-是
	IsShowAssistant     uint                      `json:"is_show_assistant"`     // 是否在助手显示, 0-否 1-是
	IsShowH5            uint                      `json:"is_show_h5"`            // 是否在H5显示, 0-否 1-是
	IsShowDelivery      uint                      `json:"is_show_delivery"`      // 是否在外送显示, 0-否 1-是
	IsShowMenu          uint                      `json:"is_show_menu"`          // 是否在电子菜单显示, 0-否 1-是
	ProductPackageCount int                       `json:"product_package_count"` // 关联商品数量
	ProductPackages     []ProductLabelPackageItem `json:"product_packages"`      // 关联商品列表
	CreateTime          int64                     `json:"create_time"`           // 创建时间
}

// ProductLabelPackageItem 标签关联的商品项
type ProductLabelPackageItem struct {
	Uuid       uint64             `json:"uuid"`        // 商品UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品名称
}

// ProductLabelInfo 商品标签信息（用于商品列表接口）
type ProductLabelInfo struct {
	Uuid  uint64 `json:"uuid"`  // 标签UUID
	Name  string `json:"name"`  // 标签名称
	Style string `json:"style"` // 标签样式
}
