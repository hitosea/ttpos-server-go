package resp

import "ttpos-server-go/app/dto"

type Staff struct {
	Uuid       uint64      `json:"uuid"`        // 自助餐UUID
	Username   string      `json:"username"`    // 用户名
	RealName   string      `json:"real_name"`   // 真实姓名
	Roles      []StaffRole `json:"roles"`       // 角色
	IsDisable  int         `json:"is_disable"`  // 是否禁用
	IsSuper    int         `json:"is_super"`    // 是否超级管理员
	CreateTime int64       `json:"create_time"` // 创建时间
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
