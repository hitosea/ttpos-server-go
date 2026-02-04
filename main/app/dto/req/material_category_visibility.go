package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// VisibilityRoleConfig 角色配置（用于请求）
type VisibilityRoleConfig struct {
	RoleUuid    uint64 `json:"role_uuid"`
	CompanyUuid uint64 `json:"company_uuid"`
}

// VisibilityListReq 可见性配置列表请求
type VisibilityListReq struct {
}

// VisibilityDetailReq 可见性配置详情请求
type VisibilityDetailReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 配置UUID
}

// VisibilityCreateReq 创建可见性配置请求
type VisibilityCreateReq struct {
	LocaleName    dto.LocaleResponse     `json:"locale_name"`                      // 配置名称（多语言）
	IsAllRoles    int                    `json:"is_all_roles" binding:"oneof=0 1"` // 是否全部门店角色：1-全部，0-部分
	CategoryUuids []uint64               `json:"category_uuids"`                   // 物品类别UUID列表
	RoleConfigs   []VisibilityRoleConfig `json:"role_configs"`                     // 角色配置列表
}

// Validate 验证请求参数
func (r *VisibilityCreateReq) Validate() error {
	if r.LocaleName.IsNull() {
		return errors.New("名称不能为空")
	}
	if !r.LocaleName.CheckLen(100) {
		return errors.New("名称不能超过100字符")
	}
	// 如果不是全部角色，则必须选择至少一个角色
	if r.IsAllRoles == 0 && len(r.RoleConfigs) == 0 {
		return errors.New("至少选择一个门店角色")
	}
	return nil
}

// VisibilityUpdateReq 更新可见性配置请求
type VisibilityUpdateReq struct {
	Uuid          uint64                 `json:"uuid" binding:"required"`          // 配置UUID
	LocaleName    dto.LocaleResponse     `json:"locale_name"`                      // 配置名称（多语言）
	IsAllRoles    int                    `json:"is_all_roles" binding:"oneof=0 1"` // 是否全部门店角色：1-全部，0-部分
	CategoryUuids []uint64               `json:"category_uuids"`                   // 物品类别UUID列表
	RoleConfigs   []VisibilityRoleConfig `json:"role_configs"`                     // 角色配置列表
}

// Validate 验证请求参数
func (r *VisibilityUpdateReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("配置UUID不能为空")
	}
	if r.LocaleName.IsNull() {
		return errors.New("名称不能为空")
	}
	if !r.LocaleName.CheckLen(100) {
		return errors.New("名称不能超过100字符")
	}
	// 如果不是全部角色，则必须选择至少一个角色
	if r.IsAllRoles == 0 && len(r.RoleConfigs) == 0 {
		return errors.New("至少选择一个门店角色")
	}
	return nil
}

// VisibilityDeleteReq 删除可见性配置请求
type VisibilityDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 配置UUID
}
