package resp

import "ttpos-server-go/app/dto"

// VisibilityListResp 可见性配置列表响应
type VisibilityListResp struct {
	List []VisibilityListItem `json:"list"`
}

// VisibilityListItem 可见性配置列表项
type VisibilityListItem struct {
	Uuid          uint64             `json:"uuid"`           // 配置UUID
	LocaleName    dto.LocaleResponse `json:"locale_name"`    // 配置名称（多语言）
	IsAllRoles    int                `json:"is_all_roles"`   // 是否全部门店角色
	RoleCount     int                `json:"role_count"`     // 角色数量（-1表示全部）
	CategoryCount int                `json:"category_count"` // 类别数量
}

// VisibilityRoleConfig 角色配置（用于响应）
type VisibilityRoleConfig struct {
	RoleUuid    uint64 `json:"role_uuid"`
	CompanyUuid uint64 `json:"company_uuid"`
}

// VisibilityDetailResp 可见性配置详情响应
type VisibilityDetailResp struct {
	Uuid          uint64                 `json:"uuid"`           // 配置UUID
	LocaleName    dto.LocaleResponse     `json:"locale_name"`    // 配置名称（多语言）
	IsAllRoles    int                    `json:"is_all_roles"`   // 是否全部门店角色
	CategoryUuids []uint64               `json:"category_uuids"` // 物品类别UUID列表
	RoleConfigs   []VisibilityRoleConfig `json:"role_configs"`   // 角色配置列表
}

// SubShopRolesResp 子店角色列表响应
type SubShopRolesResp struct {
	Shops []ShopRoles `json:"shops"`
}

// ShopRoles 门店角色
type ShopRoles struct {
	CompanyUuid uint64     `json:"company_uuid"` // 门店UUID
	CompanyCode string     `json:"company_code"` // 门店编码
	CompanyName string     `json:"company_name"` // 门店名称
	Roles       []RoleItem `json:"roles"`        // 角色列表（复用 saas_staff.go 中的 RoleItem）
}
