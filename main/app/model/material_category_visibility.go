package model

import "encoding/json"

// MaterialCategoryVisibility 物品分类可见性配置表 `ttpos_material_category_visibility`
type MaterialCategoryVisibility struct {
	BaseModel
	Name            string `gorm:"type:text;column:name;comment:'配置名称（多语言JSON）'"`
	IsAllRoles      int    `gorm:"default:0;column:is_all_roles;comment:'是否全部门店角色：1-全部，0-部分'"`
	CategoryUuids   string `gorm:"type:text;column:category_uuids;comment:'物品类别UUID列表，JSON数组格式'"`
	RoleConfigs     string `gorm:"type:text;column:role_configs;comment:'角色配置，JSON数组格式'"`
	HeadquarterUuid uint64 `gorm:"default:0;column:headquarter_uuid;comment:'总部UUID，0=当前店创建，>0=来自总部'"`
}

// TableName 指定表名
func (MaterialCategoryVisibility) TableName() string {
	return GetTableName("material_category_visibility")
}

// RoleConfig 角色配置项
type RoleConfig struct {
	RoleUuid    uint64 `json:"role_uuid"`
	CompanyUuid uint64 `json:"company_uuid"`
}

// GetCategoryUuidList 获取物品类别UUID列表
func (m *MaterialCategoryVisibility) GetCategoryUuidList() []uint64 {
	if m.CategoryUuids == "" {
		return make([]uint64, 0)
	}
	var uuids []uint64
	if err := json.Unmarshal([]byte(m.CategoryUuids), &uuids); err != nil {
		return make([]uint64, 0)
	}
	return uuids
}

// SetCategoryUuidList 设置物品类别UUID列表
func (m *MaterialCategoryVisibility) SetCategoryUuidList(uuids []uint64) {
	if len(uuids) == 0 {
		m.CategoryUuids = "[]"
		return
	}
	data, err := json.Marshal(uuids)
	if err != nil {
		m.CategoryUuids = "[]"
		return
	}
	m.CategoryUuids = string(data)
}

// GetRoleConfigList 获取角色配置列表
func (m *MaterialCategoryVisibility) GetRoleConfigList() []RoleConfig {
	if m.RoleConfigs == "" {
		return make([]RoleConfig, 0)
	}
	var configs []RoleConfig
	if err := json.Unmarshal([]byte(m.RoleConfigs), &configs); err != nil {
		return make([]RoleConfig, 0)
	}
	return configs
}

// SetRoleConfigList 设置角色配置列表
func (m *MaterialCategoryVisibility) SetRoleConfigList(configs []RoleConfig) {
	if len(configs) == 0 {
		m.RoleConfigs = "[]"
		return
	}
	data, err := json.Marshal(configs)
	if err != nil {
		m.RoleConfigs = "[]"
		return
	}
	m.RoleConfigs = string(data)
}

// IsHeadquarter 判断是否来自总部
func (m *MaterialCategoryVisibility) IsHeadquarter() bool {
	return m.HeadquarterUuid != 0
}

// ContainsRole 判断配置是否包含指定角色
func (m *MaterialCategoryVisibility) ContainsRole(roleUuid, companyUuid uint64) bool {
	// 如果是全部角色，则包含所有角色
	if m.IsAllRoles == 1 {
		return true
	}
	// 检查部分角色配置
	for _, config := range m.GetRoleConfigList() {
		if config.RoleUuid == roleUuid && config.CompanyUuid == companyUuid {
			return true
		}
	}
	return false
}

// GetRoleCount 获取角色数量（用于列表显示）
func (m *MaterialCategoryVisibility) GetRoleCount() int {
	if m.IsAllRoles == 1 {
		return -1 // -1 表示全部
	}
	return len(m.GetRoleConfigList())
}

// GetCategoryCount 获取类别数量
func (m *MaterialCategoryVisibility) GetCategoryCount() int {
	return len(m.GetCategoryUuidList())
}
