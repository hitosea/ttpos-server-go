package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDataManageRepo 数据管理仓库
type IDataManageRepo interface {
	WhereByType(typ int) DBOption // 根据数据类型查询

	Count(opts ...DBOption) (int64, error)         // 统计数据管理数据
	List(opts ...DBOption) []model.DataManage      // 查询数据管理数据
	Creates(dataManages []*model.DataManage) error // 批量创建数据管理数据
	Delete(opts ...DBOption) error                 // 删除数据管理数据
}

// NewDataManageRepo 创建新的数据管理仓库
func NewDataManageRepo(db *gorm.DB) IDataManageRepo {
	return NewDataManageRepoImpl(db)
}

// NewDataManageRepoImpl 创建新的数据管理仓库实现
func NewDataManageRepoImpl(db *gorm.DB) IDataManageRepo {
	return &dataManageRepo{db: db, model: &model.DataManage{}}
}

// dataManageRepo 数据管理仓库实现
type dataManageRepo struct {
	db    *gorm.DB          // 数据库连接
	model *model.DataManage // 数据管理模型
}

// WhereByType 根据数据类型查询
func (r *dataManageRepo) WhereByType(typ int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", typ)
	}
}

// Count 统计数据管理数据
func (r *dataManageRepo) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(r.model)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// List 查询数据管理数据
func (r *dataManageRepo) List(opts ...DBOption) []model.DataManage {
	var dataManages []model.DataManage
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	db.Find(&dataManages)
	return dataManages
}

// Delete 删除数据管理数据
func (r *dataManageRepo) Delete(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(r.model).Error
}

// Creates 批量创建数据管理数据
func (r *dataManageRepo) Creates(dataManages []*model.DataManage) error {
	return r.db.Create(dataManages).Error
}
