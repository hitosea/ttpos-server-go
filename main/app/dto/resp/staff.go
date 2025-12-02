package resp

import "ttpos-server-go/app/dto"

type Staff struct {
	Uuid              uint64      `json:"uuid"`                // 管理员UUID
	Username          string      `json:"username"`            // 用户名
	Phone             string      `json:"phone"`               // 手机号
	RealName          string      `json:"real_name"`           // 真实姓名
	Roles             []StaffRole `json:"roles"`               // 角色
	IsDisable         int         `json:"is_disable"`          // 是否禁用
	IsSuper           int         `json:"is_super"`            // 是否超级管理员
	HasDataPermission bool        `json:"has_data_permission"` // 是否有数据管理权限
	CreateTime        int64       `json:"create_time"`         // 创建时间
}

type StaffRole struct {
	Uuid uint64 `json:"uuid"` // 角色UUID
	Name string `json:"name"` // 角色名称
}

// StaffListPaginationResp 管理员列表
type StaffListPaginationResp struct {
	List []Staff          `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

type RoleListResp struct {
	List []Role           `json:"list"` // 角色列表
	Meta dto.PageResponse `json:"meta"`
}

type Role struct {
	Uuid       uint64 `json:"uuid"`        // 角色UUID
	Name       string `json:"name"`        // 角色名称
	CreateTime int64  `json:"create_time"` // 创建时间
}

type RoleDetailResp struct {
	Uuid              uint64   `json:"uuid"`                // 角色UUID
	AccessUuids       []uint64 `json:"access_uuids"`        // 权限ID列表
	Name              string   `json:"name"`                // 角色名称
	StaffCount        int      `json:"staff_count"`         // 关联员工数量
	StaffUuids        []uint64 `json:"staff_uuids"`         // 关联员工UUID列表（编辑时返回）
	SelectedLeafCount int      `json:"selected_leaf_count"` // 已选择叶子节点权限数量
	TotalLeafCount    int      `json:"total_leaf_count"`    // 公司管理APP、收银机、点餐助手三个权限组的叶子节点数量之和
}
