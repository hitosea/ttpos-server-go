package model

// File 文件库记录表 ttpos_file
type File struct {
	BaseModel
	Storage       string `gorm:"column:storage;type:varchar(20);comment:存储方式;NOT NULL" json:"storage"`
	GroupUuid     uint64 `gorm:"column:group_uuid;type:bigint(20) unsigned;default:0;comment:文件分组UUID;NOT NULL" json:"group_uuid"`
	FileUrl       string `gorm:"column:file_url;type:varchar(255);comment:存储域名;NOT NULL" json:"file_url"`
	SaveName      string `gorm:"column:save_name;type:varchar(255);comment:保存路径;NOT NULL" json:"save_name"`
	FileName      string `gorm:"column:file_name;type:varchar(255);comment:文件路径;NOT NULL" json:"file_name"`
	FileSize      int    `gorm:"column:file_size;type:int(11);default:0;comment:文件大小(字节);NOT NULL" json:"file_size"`
	FileType      string `gorm:"column:file_type;type:varchar(20);comment:文件类型;NOT NULL" json:"file_type"`
	RealName      string `gorm:"column:real_name;type:varchar(255);comment:文件真实名" json:"real_name"`
	UrlParam      string `gorm:"column:url_param;type:text;comment:签名参数" json:"url_param"`
	IndexFileName string `gorm:"column:index_file_name;type:varchar(500);comment:文件唯一名" json:"index_file_name"`
	Extension     string `gorm:"column:extension;type:varchar(20);comment:文件扩展名;NOT NULL" json:"extension"`
	IsUser        int    `gorm:"column:is_user;type:int(11);default:0;comment:是否为c端用户上传;NOT NULL" json:"is_user"`
	IsRecycle     int    `gorm:"column:is_recycle;type:tinyint(3);default:0;comment:是否已回收;NOT NULL" json:"is_recycle"`

	FileGroup *FileGroup `gorm:"foreignKey:GroupUuid;references:Uuid"` // 关联组
}

// GetUrl 获取地址。file_url + save_name + url_param
func (model *File) GetUrl() string {
	return model.FileUrl + model.SaveName + model.UrlParam
}

// FileGroup 文件库分组记录表 ttpos_file_group
type FileGroup struct {
	BaseModel
	GroupType string `gorm:"column:group_type;type:varchar(10);comment:文件类型;NOT NULL" json:"group_type"`
	GroupName string `gorm:"column:group_name;type:varchar(30);comment:分类名称;NOT NULL" json:"group_name"`
	Sort      uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:分类排序(数字越小越靠前);NOT NULL" json:"sort"`
}
