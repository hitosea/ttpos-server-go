package model

// Access 权限表 ttpos_access
type Access struct {
	BaseModel
	Name             string `gorm:"column:name;type:varchar(255);comment:权限名称;NOT NULL" json:"name"`
	Path             string `gorm:"column:path;type:varchar(255);comment:路由地址" json:"path"`
	ApiPath          string `gorm:"column:api_path;type:varchar(255);comment:后端路由地址" json:"api_path"`
	ParentUuid       uint64 `gorm:"column:parent_uuid;type:bigint(20) unsigned;default:0;comment:父级ID;NOT NULL" json:"parent_uuid"`
	Sort             int    `gorm:"column:sort;type:int(11);default:0;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`
	Icon             string `gorm:"column:icon;type:varchar(128);comment:菜单图标" json:"icon"`
	RedirectName     string `gorm:"column:redirect_name;type:varchar(128);comment:重定向名称" json:"redirect_name"`
	IsRoute          int    `gorm:"column:is_route;type:tinyint(1);default:0;comment:是否是路由 0=不是1=是;NOT NULL" json:"is_route"`
	IsMenu           int    `gorm:"column:is_menu;type:tinyint(1);default:0;comment:是否是菜单 0不是 1是;NOT NULL" json:"is_menu"`
	IsShow           int    `gorm:"column:is_show;type:tinyint(1);default:0;comment:是否显示1=显示0=不显示;NOT NULL" json:"is_show"`
	PlusCategoryUuid uint64 `gorm:"column:plus_category_uuid;type:bigint(20) unsigned;default:0;comment:插件分类ID" json:"plus_category_uuid"`
	Remark           string `gorm:"column:remark;type:varchar(255);comment:描述" json:"remark"`
	IsSupplier       int    `gorm:"column:is_supplier;type:tinyint(1);default:0;comment:是否门店菜单0否1是;NOT NULL" json:"is_supplier"`
}

// Role 角色表 ttpos_role
type Role struct {
	BaseModel
	Name string `gorm:"column:name;type:varchar(255);comment:角色名称;NOT NULL" json:"name"`
	Sort int    `gorm:"column:sort;type:int(11);default:0;comment:排序(数字越小越靠前);NOT NULL" json:"sort"`

	Accesses []RoleAccess `gorm:"foreignKey:RoleUuid;references:Uuid" json:"accesses"`
}

// RoleAccess 角色权限关系表 ttpos_role_access
type RoleAccess struct {
	BaseModel
	RoleUuid   uint64 `gorm:"column:role_uuid;type:bigint(20) unsigned;default:0;comment:角色ID;NOT NULL" json:"role_uuid"`
	AccessUuid uint64 `gorm:"column:access_uuid;type:bigint(20) unsigned;default:0;comment:权限ID;NOT NULL" json:"access_uuid"`
}
