package model

// Access 权限表
type Access struct {
	ID             uint   `gorm:"primary_key;comment:主键id"`
	Name           string `gorm:"default:'';comment:权限名称;"`
	Path           string `gorm:"default:'';comment:路由地址"`
	ApiPath        string `gorm:"default:'';comment:后端路由地址"`
	ParentId       uint   `gorm:"default:0;comment:父级id"`
	Sort           uint   `gorm:"default:100;comment:排序(数字越小越靠前)"`
	Icon           string `gorm:"default:'';comment:菜单图标"`
	RedirectName   string `gorm:"default:'';comment:重定向名称"`
	IsRoute        uint   `gorm:"default:0;comment:是否是路由 0=不是1=是"`
	IsMenu         uint   `gorm:"default:0;comment:是否是菜单 0不是 1是"`
	IsShow         uint   `gorm:"default:1;comment:是否显示1=显示0=不显示"`
	PlusCategoryId uint   `gorm:"default:0;comment:插件分类id"`
	Remark         string `gorm:"default:'';comment:描述"`
	IsSupplier     uint   `gorm:"default:0;comment:是否门店菜单0否1是"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime     int64  `gorm:"default:0;comment:删除时间"`
}

// Staff 员工表
type Staff struct {
	ID               uint   `gorm:"primary_key;AUTO_INCREMENT;comment:主键id"`
	Uuid             uint   `gorm:"default:0;comment:主键id"`
	CompanyUuid      uint   `gorm:"default:0;comment:集团ID"`
	Username         string `gorm:"default:'';comment:用户名"`
	Password         string `gorm:"default:'';comment:登录密码"`
	Phone            string `gorm:"default:'';comment:手机号"`
	PasswordChange   uint   `gorm:"default:0;comment:修改密码次数"`
	RealName         string `gorm:"default:'';comment:姓名"`
	IsSuper          uint   `gorm:"default:0;comment:是否为超级管理员0不是,1是"`
	UserType         uint   `gorm:"default:0;comment:账号类型0总台1门店"`
	IsDisable        uint   `gorm:"default:0;comment:是否禁用1禁用，0未禁用"`
	BindKey          string `gorm:"default:'';comment:绑定的设备key"`
	CashierOnline    uint   `gorm:"default:0;comment:收银员当班 0-不在线 1-在线"`
	CashierLoginTime uint   `gorm:"default:0;comment:收银员当班登录时间"`
	DutyNo           string `gorm:"default:'';comment:当班编号"`
	CreateTime       int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime       int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime       int64  `gorm:"default:0;comment:删除时间"`

	Company *Company `gorm:"foreignKey:company_uuid"`
}

// Role 角色表
type Role struct {
	ID         uint   `gorm:"primary_key;AUTO_INCREMENT;comment:角色id"`
	Name       string `gorm:"comment:角色名称"`
	Sort       uint   `gorm:"default:100;comment:排序(数字越小越靠前)"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime int64  `gorm:"default:0;comment:删除时间"`
}

// StaffRole 员工角色
type StaffRole struct {
	ID         uint  `gorm:"primary_key;AUTO_INCREMENT;comment:主键id"`
	StaffId    uint  `gorm:"default:0;comment:超管用户id"`
	RoleId     uint  `gorm:"default:0;comment:角色id"`
	CreateTime int64 `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime int64 `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime int64 `gorm:"default:0;comment:删除时间"`
}

// RoleAccess 角色权限
type RoleAccess struct {
	ID         uint  `gorm:"primary_key;AUTO_INCREMENT;comment:主键id"`
	RoleId     uint  `gorm:"default:0;comment:角色id"`
	AccessId   uint  `gorm:"default:0;comment:权限id"`
	CreateTime int64 `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime int64 `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime int64 `gorm:"default:0;comment:删除时间"`
}
