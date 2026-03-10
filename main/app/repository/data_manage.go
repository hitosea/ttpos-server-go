package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDataManageRepo 数据管理仓库
type IDataManageRepo interface {
	WhereByType(typ int) DBOption             // 根据数据类型查询
	WhereInDataUuids(uuids []uint64) DBOption // 根据数据UUID列表查询

	Count(opts ...DBOption) (int64, error)                    // 统计数据管理数据
	List(opts ...DBOption) []model.DataManage                 // 查询数据管理数据
	Creates(dataManages []*model.DataManage) error            // 批量创建数据管理数据
	Delete(opts ...DBOption) error                            // 删除数据管理数据
	GetDataUuids(opts ...DBOption) []uint64                   // 获取数据管理数据UUID列表
	DeleteByDataUuid(typ int, dataUuid uint64) error          // 根据数据UUID和类型删除单条记录
	BatchDeleteByDataUuids(typ int, dataUuids []uint64) error // 根据数据UUID列表和类型批量删除
	ExistsByDataUuid(typ int, dataUuid uint64) bool           // 判断数据UUID是否存在
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
	return r.db.CreateInBatches(dataManages, 100).Error
}

// WhereInDataUuids 根据数据UUID列表查询
func (r *dataManageRepo) WhereInDataUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(uuids) == 0 {
			return db.Where("1 = 0")
		}
		return db.Where("data_uuid IN ?", uuids)
	}
}

// GetDataUuids 获取数据管理数据UUID列表
func (r *dataManageRepo) GetDataUuids(opts ...DBOption) []uint64 {
	var dataUuids []uint64
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	db.Model(r.model).Select("data_uuid").Pluck("data_uuid", &dataUuids)
	return dataUuids
}

// DeleteByDataUuid 根据数据UUID和类型删除单条记录
func (r *dataManageRepo) DeleteByDataUuid(typ int, dataUuid uint64) error {
	return r.db.Where("type = ? AND data_uuid = ?", typ, dataUuid).Delete(r.model).Error
}

// ExistsByDataUuid 判断数据UUID是否存在
func (r *dataManageRepo) ExistsByDataUuid(typ int, dataUuid uint64) bool {
	var count int64
	r.db.Model(r.model).Where("type = ? AND data_uuid = ?", typ, dataUuid).Count(&count)
	return count > 0
}

// BatchDeleteByDataUuids 根据数据UUID列表和类型批量删除
func (r *dataManageRepo) BatchDeleteByDataUuids(typ int, dataUuids []uint64) error {
	if len(dataUuids) == 0 {
		return nil
	}
	const batchSize = 1000
	for i := 0; i < len(dataUuids); i += batchSize {
		end := i + batchSize
		if end > len(dataUuids) {
			end = len(dataUuids)
		}
		if err := r.db.Where("type = ? AND data_uuid IN ?", typ, dataUuids[i:end]).Delete(r.model).Error; err != nil {
			return err
		}
	}
	return nil
}
