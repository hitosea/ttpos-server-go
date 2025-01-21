package model

import "jjjshop-server-go/config"

// User 商家用户表
type User struct {
	ShopUserId       uint   `gorm:"column:shop_user_id;type:int(11);primary_key;AUTO_INCREMENT;comment:主键id" json:"shop_user_id"`
	UserName         string `gorm:"column:user_name;type:varchar(255);comment:用户名;NOT NULL" json:"user_name"`
	Password         string `gorm:"column:password;type:varchar(255);comment:登录密码;NOT NULL" json:"password"`
	Phone            string `gorm:"column:phone;type:varchar(20);comment:手机号" json:"phone"`
	PasswordChange   int    `gorm:"column:password_change;type:int(11);default:0;comment:修改密码次数" json:"password_change"`
	RealName         string `gorm:"column:real_name;type:varchar(255);comment:姓名;NOT NULL" json:"real_name"`
	IsSuper          uint   `gorm:"column:is_super;type:tinyint(3);default:0;comment:是否为超级管理员0不是,1是;NOT NULL" json:"is_super"`
	ShopSupplierId   int    `gorm:"column:shop_supplier_id;type:int(11);default:0;comment:总店id;NOT NULL" json:"shop_supplier_id"`
	IsDelete         uint   `gorm:"column:is_delete;type:tinyint(3);default:0;comment:0=显示1=伪删除;NOT NULL" json:"is_delete"`
	UserType         int    `gorm:"column:user_type;type:tinyint(1);default:0;comment:账号类型0总台1门店;NOT NULL" json:"user_type"`
	IsStatus         uint   `gorm:"column:is_status;type:tinyint(3);default:0;comment:是否禁用1禁用，0未禁用;NOT NULL" json:"is_status"` // ToDo 改为is_disabled
	AppId            uint   `gorm:"column:app_id;type:int(11);default:0;comment:程序id;NOT NULL" json:"app_id"`
	BindKey          string `gorm:"column:bind_key;type:varchar(255);comment:绑定的设备key" json:"bind_key"`
	CashierOnline    int    `gorm:"column:cashier_online;type:tinyint(4);default:0;comment:收银员当班 0-不在线 1-在线;NOT NULL" json:"cashier_online"`
	CashierLoginTime int    `gorm:"column:cashier_login_time;type:int(11);default:0;comment:收银员当班登录时间;NOT NULL" json:"cashier_login_time"`
	DutyNo           string `gorm:"column:duty_no;type:varchar(64);comment:当班编号" json:"duty_no"`
	CreateTime       uint   `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime       int    `gorm:"column:update_time;type:int(11);comment:更新时间;NOT NULL" json:"update_time"`

	Supplier *Supplier `gorm:"foreignKey:shop_supplier_id;references:shop_supplier_id"`
	App      *App      `gorm:"foreignKey:app_id;references:app_id"`
}

func (User) TableName() string {
	return config.Database.TablePrefix + "shop_user"
}

// ShopRole 商家用户角色表
type ShopRole struct {
	RoleId     uint   `gorm:"column:role_id;type:int(11);primary_key;AUTO_INCREMENT;comment:角色id" json:"role_id"`
	RoleName   string `gorm:"column:role_name;type:varchar(2000);comment:角色名称;NOT NULL" json:"role_name"`
	Sort       uint   `gorm:"column:sort;type:int(10);default:100;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	AppId      uint   `gorm:"column:app_id;type:int(11);default:0;comment:小程序id;NOT NULL" json:"app_id"`
	CreateTime uint   `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime uint   `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`
}

// ShopAccess 商家权限表
type ShopAccess struct {
	AccessId       uint   `gorm:"column:access_id;type:int(11);primary_key;comment:主键id" json:"access_id"`
	Name           string `gorm:"column:name;type:varchar(255);comment:权限名称;NOT NULL" json:"name"`
	Path           string `gorm:"column:path;type:varchar(255);comment:路由地址" json:"path"`
	ApiPath        string `gorm:"column:api_path;type:varchar(255);comment:后端路由地址" json:"api_path"`
	ParentId       uint   `gorm:"column:parent_id;type:int(11);default:0;comment:父级id;NOT NULL" json:"parent_id"`
	Sort           uint   `gorm:"column:sort;type:tinyint(3);default:100;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	Icon           string `gorm:"column:icon;type:varchar(128);comment:菜单图标" json:"icon"`
	RedirectName   string `gorm:"column:redirect_name;type:varchar(128);comment:重定向名称" json:"redirect_name"`
	IsRoute        int    `gorm:"column:is_route;type:tinyint(1);default:0;comment:是否是路由 0=不是1=是;NOT NULL" json:"is_route"`
	IsMenu         uint   `gorm:"column:is_menu;type:tinyint(1);default:0;comment:是否是菜单 0不是 1是;NOT NULL" json:"is_menu"`
	Alias          string `gorm:"column:alias;type:varchar(128);comment:别名(废弃)" json:"alias"`
	IsShow         uint   `gorm:"column:is_show;type:tinyint(1);default:1;comment:是否显示1=显示0=不显示;NOT NULL" json:"is_show"`
	PlusCategoryId int    `gorm:"column:plus_category_id;type:int(11);default:0;comment:插件分类id" json:"plus_category_id"`
	Remark         string `gorm:"column:remark;type:varchar(255);comment:描述" json:"remark"`
	IsSupplier     int    `gorm:"column:is_supplier;type:tinyint(1);default:0;comment:是否门店菜单0否1是;NOT NULL" json:"is_supplier"`
	AppId          uint   `gorm:"column:app_id;type:int(10);default:0;comment:app_id" json:"app_id"`
	CreateTime     uint   `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     uint   `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`
}

// UserRole 商家用户角色
type UserRole struct {
	Id         uint `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	ShopUserId uint `gorm:"column:shop_user_id;type:int(11);default:0;comment:超管用户id;NOT NULL" json:"shop_user_id"`
	RoleId     uint `gorm:"column:role_id;type:int(11);default:0;comment:角色id;NOT NULL" json:"role_id"`
	AppId      uint `gorm:"column:app_id;type:int(11);default:0;comment:小程序id;NOT NULL" json:"app_id"`
	CreateTime uint `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime uint `gorm:"column:update_time;type:int(10);default:0;comment:更新时间" json:"update_time"`
}

func (UserRole) TableName() string {
	return config.Database.TablePrefix + "shop_user_role"
}

// ShopRoleAccess 商家角色权限
type ShopRoleAccess struct {
	Id         uint `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	RoleId     uint `gorm:"column:role_id;type:int(11);default:0;comment:角色id;NOT NULL" json:"role_id"`
	AccessId   uint `gorm:"column:access_id;type:int(11);default:0;comment:权限id;NOT NULL" json:"access_id"`
	AppId      uint `gorm:"column:app_id;type:int(11);default:0;comment:小程序id;NOT NULL" json:"app_id"`
	CreateTime uint `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
}
