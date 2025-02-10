package model

// Access 权限表
type Access struct {
	ID               uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid             uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:权限ID;NOT NULL" json:"uuid"`
	Name             string `gorm:"column:name;type:varchar(255);comment:权限名称;NOT NULL" json:"name"`
	Path             string `gorm:"column:path;type:varchar(255);comment:路由地址" json:"path"`
	ApiPath          string `gorm:"column:api_path;type:varchar(255);comment:后端路由地址" json:"api_path"`
	ParentUuid       uint64 `gorm:"column:parent_uuid;type:bigint(20) unsigned;default:0;comment:父级ID;NOT NULL" json:"parent_uuid"`
	Sort             int    `gorm:"column:sort;type:int(11);default:100;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	Icon             string `gorm:"column:icon;type:varchar(128);comment:菜单图标" json:"icon"`
	RedirectName     string `gorm:"column:redirect_name;type:varchar(128);comment:重定向名称" json:"redirect_name"`
	IsRoute          int    `gorm:"column:is_route;type:tinyint(1);default:0;comment:是否是路由 0=不是1=是;NOT NULL" json:"is_route"`
	IsMenu           int    `gorm:"column:is_menu;type:tinyint(1);default:0;comment:是否是菜单 0不是 1是;NOT NULL" json:"is_menu"`
	IsShow           int    `gorm:"column:is_show;type:tinyint(1);default:1;comment:是否显示1=显示0=不显示;NOT NULL" json:"is_show"`
	PlusCategoryUuid uint64 `gorm:"column:plus_category_uuid;type:bigint(20) unsigned;default:0;comment:插件分类ID" json:"plus_category_uuid"`
	Remark           string `gorm:"column:remark;type:varchar(255);comment:描述" json:"remark"`
	IsSupplier       int    `gorm:"column:is_supplier;type:tinyint(1);default:0;comment:是否门店菜单0否1是;NOT NULL" json:"is_supplier"`
	CreateTime       int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime       int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime       int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`
}

// Staff 员工表
type Staff struct {
	ID               uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid             uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"uuid"`
	CompanyUuid      uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	Username         string `gorm:"column:username;type:varchar(255);comment:用户名;NOT NULL" json:"username"`
	Password         string `gorm:"column:password;type:varchar(255);comment:登录密码;NOT NULL" json:"password"`
	Phone            string `gorm:"column:phone;type:varchar(20);comment:手机号" json:"phone"`
	PasswordChange   int    `gorm:"column:password_change;type:int(11);default:0;comment:修改密码次数" json:"password_change"`
	RealName         string `gorm:"column:real_name;type:varchar(255);comment:姓名;NOT NULL" json:"real_name"`
	IsSuper          int    `gorm:"column:is_super;type:tinyint(3);default:0;comment:是否为超级管理员0不是,1是;NOT NULL" json:"is_super"`
	UserType         int    `gorm:"column:user_type;type:tinyint(1);default:0;comment:账号类型0总台1门店;NOT NULL" json:"user_type"`
	IsDisable        int    `gorm:"column:is_disable;type:tinyint(3);default:0;comment:是否禁用1禁用，0未禁用;NOT NULL" json:"is_disable"`
	BindKey          string `gorm:"column:bind_key;type:varchar(255);comment:绑定的设备key" json:"bind_key"`
	CashierOnline    int    `gorm:"column:cashier_online;type:tinyint(1);default:0;comment:收银员当班 0-不在线 1-在线;NOT NULL" json:"cashier_online"`
	CashierLoginTime int    `gorm:"column:cashier_login_time;type:int(11);default:0;comment:收银员当班登录时间;NOT NULL" json:"cashier_login_time"`
	DutyNo           string `gorm:"column:duty_no;type:varchar(64);comment:当班编号" json:"duty_no"`
	CreateTime       int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime       int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime       int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`

	Company *Company `gorm:"foreignKey:company_uuid;references:Uuid"`
}

// Role 角色表
type Role struct {
	ID         uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid       uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:角色ID;NOT NULL" json:"uuid"`
	Name       string `gorm:"column:name;type:varchar(255);comment:角色名称;NOT NULL" json:"name"`
	Sort       int    `gorm:"column:sort;type:int(11);default:100;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	CreateTime int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`
}

// StaffRole 员工角色
type StaffRole struct {
	ID         uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid       uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:员工角色关系ID;NOT NULL" json:"uuid"`
	StaffUuid  int64  `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:超管用户ID;NOT NULL" json:"staff_uuid"`
	RoleUuid   int64  `gorm:"column:role_uuid;type:bigint(20);default:0;comment:角色ID;NOT NULL" json:"role_uuid"`
	CreateTime int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`
}

// RoleAccess 角色权限
type RoleAccess struct {
	ID         uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid       uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:角色权限关系ID;NOT NULL" json:"uuid"`
	RoleUuid   uint64 `gorm:"column:role_uuid;type:bigint(20) unsigned;default:0;comment:角色ID;NOT NULL" json:"role_uuid"`
	AccessUuid uint64 `gorm:"column:access_uuid;type:bigint(20) unsigned;default:0;comment:权限ID;NOT NULL" json:"access_uuid"`
	CreateTime int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`
}
