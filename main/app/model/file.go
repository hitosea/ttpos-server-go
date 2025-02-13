package model

// File 文件表 ttpos_file
type File struct {
	BaseModel
	Storage        string `gorm:"column:storage;not null;default:'';comment:'存储方式'"`
	GroupID        int    `gorm:"column:group_id;not null;default:0;comment:'文件分组id'"`
	FileUrl        string `gorm:"column:file_url;not null;default:'';comment:'存储域名'"`
	SaveName       string `gorm:"column:save_name;not null;default:'';comment:'保存路径'"`
	FileName       string `gorm:"column:file_name;not null;default:'';comment:'文件路径'"`
	FileSize       int    `gorm:"column:file_size;not null;default:0;comment:'文件大小(字节)'"`
	FileType       string `gorm:"column:file_type;not null;default:'';comment:'文件类型'"`
	RealName       string `gorm:"column:real_name;default:'';comment:'文件真实名'"`
	UrlParam       string `gorm:"column:url_param;type:text;default:'';comment:'签名参数'"`
	IndexFileName  string `gorm:"column:index_file_name;default:'';comment:'文件唯一名'"`
	Extension      string `gorm:"column:extension;not null;default:'';comment:'文件扩展名'"`
	IsUser         int    `gorm:"column:is_user;not null;default:0;comment:'是否为c端用户上传'"`
	IsRecycle      int    `gorm:"column:is_recycle;not null;default:0;comment:'是否已回收'"`
	ShopSupplierID int    `gorm:"column:shop_supplier_id;default:0;comment:'供应商id'"`
	IsDelete       int    `gorm:"column:is_delete;not null;default:0;comment:'软删除'"`
	AppID          int    `gorm:"column:app_id;default:0;comment:'应用id'"`
}

// GetUrl 获取地址。file_url + save_name + url_param
func (model *File) GetUrl() string {
	return model.FileUrl + model.SaveName + model.UrlParam
}
