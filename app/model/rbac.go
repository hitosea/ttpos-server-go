package model

// Access 权限表
type Access struct {
	ID             uint   `gorm:"column:id;type:int(11) unsigned;primary_key;comment:主键id" json:"id"`
	Name           string `gorm:"column:name;type:varchar(255);comment:权限名称;NOT NULL" json:"name"`
	Path           string `gorm:"column:path;type:varchar(255);comment:路由地址" json:"path"`
	ApiPath        string `gorm:"column:api_path;type:varchar(255);comment:后端路由地址" json:"api_path"`
	ParentId       uint   `gorm:"column:parent_id;type:int(11) unsigned;default:0;comment:父级id;NOT NULL" json:"parent_id"`
	Sort           uint   `gorm:"column:sort;type:tinyint(3) unsigned;default:100;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	Icon           string `gorm:"column:icon;type:varchar(128);comment:菜单图标" json:"icon"`
	RedirectName   string `gorm:"column:redirect_name;type:varchar(128);comment:重定向名称" json:"redirect_name"`
	IsRoute        int    `gorm:"column:is_route;type:tinyint(1);default:0;comment:是否是路由 0=不是1=是;NOT NULL" json:"is_route"`
	IsMenu         uint   `gorm:"column:is_menu;type:tinyint(1) unsigned;default:0;comment:是否是菜单 0不是 1是;NOT NULL" json:"is_menu"`
	IsShow         uint   `gorm:"column:is_show;type:tinyint(1) unsigned;default:1;comment:是否显示1=显示0=不显示;NOT NULL" json:"is_show"`
	PlusCategoryId int    `gorm:"column:plus_category_id;type:int(11);default:0;comment:插件分类id" json:"plus_category_id"`
	Remark         string `gorm:"column:remark;type:varchar(255);comment:描述" json:"remark"`
	IsSupplier     int    `gorm:"column:is_supplier;type:tinyint(1);default:0;comment:是否门店菜单0否1是;NOT NULL" json:"is_supplier"`
	CreateTime     uint   `gorm:"autoCreateTime;column:create_time;type:int(11) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     uint   `gorm:"autoUpdateTime;column:update_time;type:int(11) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime     uint   `gorm:"column:delete_time;type:int(11) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}

// Staff 员工表
type Staff struct {
	ID               uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	CompanyId        uint   `gorm:"column:company_id;type:int(11) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_id"`
	Username         string `gorm:"column:username;type:varchar(255);comment:用户名;NOT NULL" json:"username"`
	Password         string `gorm:"column:password;type:varchar(255);comment:登录密码;NOT NULL" json:"password"`
	Phone            string `gorm:"column:phone;type:varchar(20);comment:手机号" json:"phone"`
	PasswordChange   int    `gorm:"column:password_change;type:int(11);default:0;comment:修改密码次数" json:"password_change"`
	RealName         string `gorm:"column:real_name;type:varchar(255);comment:姓名;NOT NULL" json:"real_name"`
	IsSuper          uint   `gorm:"column:is_super;type:tinyint(3) unsigned;default:0;comment:是否为超级管理员0不是,1是;NOT NULL" json:"is_super"`
	UserType         int    `gorm:"column:user_type;type:tinyint(1);default:0;comment:账号类型0总台1门店;NOT NULL" json:"user_type"`
	IsDisable        uint   `gorm:"column:is_disable;type:tinyint(3) unsigned;default:0;comment:是否禁用1禁用，0未禁用;NOT NULL" json:"is_disable"`
	BindKey          string `gorm:"column:bind_key;type:varchar(255);comment:绑定的设备key" json:"bind_key"`
	CashierOnline    int    `gorm:"column:cashier_online;type:tinyint(4);default:0;comment:收银员当班 0-不在线 1-在线;NOT NULL" json:"cashier_online"`
	CashierLoginTime int    `gorm:"column:cashier_login_time;type:int(11);default:0;comment:收银员当班登录时间;NOT NULL" json:"cashier_login_time"`
	DutyNo           string `gorm:"column:duty_no;type:varchar(64);comment:当班编号" json:"duty_no"`
	IsDelete         uint   `gorm:"column:is_delete;type:tinyint(3) unsigned;default:0;comment:0=显示1=伪删除;NOT NULL" json:"is_delete"`
	CreateTime       uint   `gorm:"autoCreateTime;column:create_time;type:int(11) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime       uint   `gorm:"autoUpdateTime;column:update_time;type:int(11) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime       uint   `gorm:"column:delete_time;type:int(11) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`

	Company *Company `gorm:"foreignKey:company_id"`
}

// Role 角色表
type Role struct {
	ID         uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:角色id" json:"id"`
	Name       string `gorm:"column:name;type:varchar(2000);comment:角色名称;NOT NULL" json:"name"`
	Sort       uint   `gorm:"column:sort;type:int(10) unsigned;default:100;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	CreateTime uint   `gorm:"autoCreateTime;column:create_time;type:int(11) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime uint   `gorm:"autoUpdateTime;column:update_time;type:int(11) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime uint   `gorm:"column:delete_time;type:int(11) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}

// StaffRole 员工角色
type StaffRole struct {
	ID         uint `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	StaffId    uint `gorm:"column:staff_id;type:int(11) unsigned;default:0;comment:超管用户id;NOT NULL" json:"staff_id"`
	RoleId     uint `gorm:"column:role_id;type:int(11) unsigned;default:0;comment:角色id;NOT NULL" json:"role_id"`
	CreateTime uint `gorm:"autoCreateTime;column:create_time;type:int(11) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime uint `gorm:"autoUpdateTime;column:update_time;type:int(11) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime uint `gorm:"column:delete_time;type:int(11) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}

// RoleAccess 角色权限
type RoleAccess struct {
	ID         uint `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	RoleId     uint `gorm:"column:role_id;type:int(11) unsigned;default:0;comment:角色id;NOT NULL" json:"role_id"`
	AccessId   uint `gorm:"column:access_id;type:int(11) unsigned;default:0;comment:权限id;NOT NULL" json:"access_id"`
	CreateTime uint `gorm:"autoCreateTime;column:create_time;type:int(11) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime uint `gorm:"autoUpdateTime;column:update_time;type:int(11) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime uint `gorm:"column:delete_time;type:int(11) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}
