package resp

// CompanyStaffResp 账号-门店关联响应
type CompanyStaffResp struct {
	CompanyUuid uint64   `json:"company_uuid"` // 门店UUID
	CompanyName string   `json:"company_name"` // 门店名称
	StoreCode   string   `json:"store_code"`   // 店铺编号
	CreateTime  int64    `json:"-"`            // 创建时间（内部排序用，不输出到JSON）
	Roles       []string `json:"roles"`        // 角色列表
	IsSuper     int      `json:"is_super"`     // 是否超级管理员
}

// SearchStaffCompanyInfo 搜索员工的门店信息
type SearchStaffCompanyInfo struct {
	CompanyUuid uint64     `json:"company_uuid"` // 门店UUID
	CompanyName string     `json:"company_name"` // 门店名称
	Roles       []RoleItem `json:"roles"`        // 角色列表
	IsSuper     int        `json:"is_super"`     // 是否超级管理员
}

type RoleItem struct {
	Uuid uint64 `json:"uuid"` // 角色UUID
	Name string `json:"name"` // 角色名称
}

// SearchStaffResp 搜索员工响应
type SearchStaffResp struct {
	Uuid        uint64                    `json:"uuid"`         // 员工UUID，大于0表示找到员工
	Email       string                    `json:"email"`        // 邮箱
	Phone       string                    `json:"phone"`        // 手机号
	RealName    string                    `json:"real_name"`    // 姓名
	CompanyList []*SearchStaffCompanyInfo `json:"company_list"` // 在当前门店可见范围内的门店列表
}

// QueryStaffByContactItem 查询员工响应项
type QueryStaffByContactItem struct {
	Uuid     uint64 `json:"uuid"`      // 员工UUID
	RealName string `json:"real_name"` // 姓名
	Email    string `json:"email"`     // 邮箱
	Phone    string `json:"phone"`     // 手机号
}

// QueryStaffByContactResp 查询员工响应
type QueryStaffByContactResp struct {
	List []*QueryStaffByContactItem `json:"list"` // 员工列表
}
