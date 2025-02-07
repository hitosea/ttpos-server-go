package repository

import "gorm.io/gorm"

type Where func(*gorm.DB) *gorm.DB
type With func(*gorm.DB) *gorm.DB
type DBOption func(*gorm.DB) *gorm.DB

// ICommonRepo 公共仓库接口
type ICommonRepo interface {
	WhereByID(id uint) DBOption
	WhereLikeByName(name string) DBOption
	OrderByID(id uint, order string) DBOption
}

// commonRepo 公共仓库实现
type commonRepo struct{}

// NewCommonRepo 创建新的公共仓库
func NewCommonRepo() ICommonRepo {
	return NewCommonRepoImpl()
}

// NewCommonRepoImpl 创建新的公共仓库实现
func NewCommonRepoImpl() ICommonRepo {
	return &commonRepo{}
}

// WhereByID 根据ID查询
func (r *commonRepo) WhereByID(id uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

// WhereLikeByName 根据名称查询
func (r *commonRepo) WhereLikeByName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

// OrderByID 根据ID排序
func (r *commonRepo) OrderByID(id uint, order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("id " + order)
	}
}
