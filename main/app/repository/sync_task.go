package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ISyncTaskRepo 同步任务Repository接口
type ISyncTaskRepo interface {
	// 基础操作
	Create(task *model.SyncTask) error
	Update(uuid uint64, data map[string]any) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.SyncTask, error)
	GetList(opts ...DBOption) ([]model.SyncTask, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.SyncTask, int64, error)

	// 条件查询选项
	WhereStatus(status uint8) DBOption
	OrderByCreateTime(desc bool) DBOption
	PreloadItems() DBOption
}

// ISyncTaskItemRepo 同步任务明细Repository接口
type ISyncTaskItemRepo interface {
	// 基础操作
	Create(item *model.SyncTaskItem) error
	BatchCreate(items []model.SyncTaskItem) error
	Update(uuid uint64, data map[string]any) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.SyncTaskItem, error)
	GetList(opts ...DBOption) ([]model.SyncTaskItem, error)

	// 条件查询选项
	WhereSyncTaskUuid(syncTaskUuid uint64) DBOption
	WhereTaskType(taskType string) DBOption
	WhereStatus(status uint8) DBOption
	OrderByCreateTime(desc bool) DBOption
}

// SyncTaskRepoImpl 同步任务Repository实现
type SyncTaskRepoImpl struct {
	db *gorm.DB
}

// NewSyncTaskRepo 创建同步任务Repository
func NewSyncTaskRepo(db *gorm.DB) ISyncTaskRepo {
	return &SyncTaskRepoImpl{db: db}
}

// Create 创建同步任务
func (r *SyncTaskRepoImpl) Create(task *model.SyncTask) error {
	return r.db.Create(task).Error
}

// Update 更新同步任务
func (r *SyncTaskRepoImpl) Update(uuid uint64, data map[string]any) error {
	data["update_time"] = time.Now().Unix()
	return r.db.Model(&model.SyncTask{}).Where("uuid = ?", uuid).Updates(data).Error
}

// GetByUuid 根据UUID查询同步任务
func (r *SyncTaskRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.SyncTask, error) {
	var task model.SyncTask
	db := r.db.Model(&model.SyncTask{}).Where("uuid = ?", uuid)
	db = r.applyOptions(db, opts...)
	err := db.First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetList 获取同步任务列表
func (r *SyncTaskRepoImpl) GetList(opts ...DBOption) ([]model.SyncTask, error) {
	var tasks []model.SyncTask
	db := r.db.Model(&model.SyncTask{})
	db = r.applyOptions(db, opts...)
	err := db.Find(&tasks).Error
	return tasks, err
}

// GetListWithPagination 分页获取同步任务列表
func (r *SyncTaskRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.SyncTask, int64, error) {
	var tasks []model.SyncTask
	var total int64

	// 构建基础查询
	db := r.db.Model(&model.SyncTask{})
	db = r.applyOptions(db, opts...)

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// applyOptions 应用查询选项
func (r *SyncTaskRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

// 查询选项实现

// WhereStatus 状态条件
func (r *SyncTaskRepoImpl) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *SyncTaskRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// PreloadItems 预加载明细
func (r *SyncTaskRepoImpl) PreloadItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items")
	}
}

// SyncTaskItemRepoImpl 同步任务明细Repository实现
type SyncTaskItemRepoImpl struct {
	db *gorm.DB
}

// NewSyncTaskItemRepo 创建同步任务明细Repository
func NewSyncTaskItemRepo(db *gorm.DB) ISyncTaskItemRepo {
	return &SyncTaskItemRepoImpl{db: db}
}

// Create 创建同步任务明细
func (r *SyncTaskItemRepoImpl) Create(item *model.SyncTaskItem) error {
	return r.db.Create(item).Error
}

// BatchCreate 批量创建同步任务明细
func (r *SyncTaskItemRepoImpl) BatchCreate(items []model.SyncTaskItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// Update 更新同步任务明细
func (r *SyncTaskItemRepoImpl) Update(uuid uint64, data map[string]any) error {
	data["update_time"] = time.Now().Unix()
	return r.db.Model(&model.SyncTaskItem{}).Where("uuid = ?", uuid).Updates(data).Error
}

// GetByUuid 根据UUID查询同步任务明细
func (r *SyncTaskItemRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.SyncTaskItem, error) {
	var item model.SyncTaskItem
	db := r.db.Model(&model.SyncTaskItem{}).Where("uuid = ?", uuid)
	db = r.applyOptions(db, opts...)
	err := db.First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetList 获取同步任务明细列表
func (r *SyncTaskItemRepoImpl) GetList(opts ...DBOption) ([]model.SyncTaskItem, error) {
	var items []model.SyncTaskItem
	db := r.db.Model(&model.SyncTaskItem{})
	db = r.applyOptions(db, opts...)
	err := db.Find(&items).Error
	return items, err
}

// applyOptions 应用查询选项
func (r *SyncTaskItemRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

// 查询选项实现

// WhereSyncTaskUuid 同步任务UUID条件
func (r *SyncTaskItemRepoImpl) WhereSyncTaskUuid(syncTaskUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sync_task_uuid = ?", syncTaskUuid)
	}
}

// WhereTaskType 任务类型条件
func (r *SyncTaskItemRepoImpl) WhereTaskType(taskType string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("task_type = ?", taskType)
	}
}

// WhereStatus 状态条件
func (r *SyncTaskItemRepoImpl) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *SyncTaskItemRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}
