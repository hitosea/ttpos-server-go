package resp

import "ttpos-server-go/app/dto"

// CompanyAreaItem 区域列表项
type CompanyAreaItem struct {
	Uuid       uint64             `json:"uuid"`        // 区域UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 多语言名称
	Sort       int                `json:"sort"`        // 排序
	ShopCount  int                `json:"shop_count"`  // 关联门店数量
}

// CompanyAreaListResp 区域列表响应（含摘要）
type CompanyAreaListResp struct {
	TotalShopCount      int               `json:"total_shop_count"`      // 总门店数
	UnassignedShopCount int               `json:"unassigned_shop_count"` // 未分配区域的门店数
	List                []CompanyAreaItem  `json:"list"`
}

// CompanyAreaInfoResp 区域详情响应
type CompanyAreaInfoResp struct {
	Uuid       uint64             `json:"uuid"`        // 区域UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 多语言名称
	Sort       int                `json:"sort"`        // 排序
	ShopList   []CompanyAreaShop  `json:"shop_list"`   // 关联门店列表
}

// CompanyAreaShop 区域关联的门店信息
type CompanyAreaShop struct {
	Uuid   uint64 `json:"uuid"`   // 门店UUID
	Name   string `json:"name"`   // 门店名称
	Status int    `json:"status"` // 状态
}

// CompanyAreaUnassignedResp 未分配门店列表响应
type CompanyAreaUnassignedResp struct {
	List []CompanyAreaShop `json:"list"`
}

// CompanyAreaShopItem 区域管理用门店列表项
type CompanyAreaShopItem struct {
	Uuid     uint64 `json:"uuid"`      // 门店UUID
	Name     string `json:"name"`      // 门店名称
	AreaUuid uint64 `json:"area_uuid"` // 所属区域UUID: 0-未分配
	AreaName string `json:"area_name"` // 所属区域名称
}

// CompanyAreaShopListResp 区域管理用门店列表响应
type CompanyAreaShopListResp struct {
	List []CompanyAreaShopItem `json:"list"`
}
