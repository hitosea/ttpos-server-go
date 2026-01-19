package repository

import (
	"strings"
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IExportRecordRepo 导出记录Repository接口
type IExportRecordRepo interface {
	// 基础操作
	Create(record *model.ExportRecord) error
	Update(uuid uint64, data map[string]any) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.ExportRecord, error)
	GetUnfinishedExportRecord(exportType uint8) (*model.ExportRecord, error)                   // 查询3小时内未完成的导出记录
	GetByDateAndType(exportType uint8, startTime, endTime int64) ([]model.ExportRecord, error) // 查询指定日期范围内指定类型的导出记录
	GetList(opts ...DBOption) ([]model.ExportRecord, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.ExportRecord, int64, error)
	GetByUuids(uuids []uint64, opts ...DBOption) ([]model.ExportRecord, error)
	BatchDelete(uuids []uint64) error

	// 条件查询选项
	WhereExportType(exportType *uint8) DBOption
	WhereStatus(status uint8) DBOption
	WhereNotDeleted() DBOption
	OrderByCreateTime(desc bool) DBOption
	PreloadFile() DBOption
}

// ExportRecordRepoImpl 导出记录Repository实现
type ExportRecordRepoImpl struct {
	db *gorm.DB
}

// NewExportRecordRepo 创建导出记录Repository
func NewExportRecordRepo(db *gorm.DB) IExportRecordRepo {
	return &ExportRecordRepoImpl{db: db}
}

// Create 创建导出记录
func (r *ExportRecordRepoImpl) Create(record *model.ExportRecord) error {
	return r.db.Create(record).Error
}

// Update 更新导出记录
func (r *ExportRecordRepoImpl) Update(uuid uint64, data map[string]any) error {
	data["update_time"] = time.Now().Unix()
	return r.db.Model(&model.ExportRecord{}).Where("uuid = ?", uuid).Updates(data).Error
}

// GetByUuid 根据UUID查询导出记录
func (r *ExportRecordRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.ExportRecord, error) {
	var record model.ExportRecord
	db := r.db.Model(&model.ExportRecord{}).Where("uuid = ?", uuid)
	db = r.applyOptions(db, opts...)
	err := db.First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetList 获取导出记录列表
func (r *ExportRecordRepoImpl) GetList(opts ...DBOption) ([]model.ExportRecord, error) {
	var records []model.ExportRecord
	db := r.db.Model(&model.ExportRecord{})
	db = r.applyOptions(db, opts...)
	err := db.Find(&records).Error
	return records, err
}

// GetByUuids 根据UUID列表批量查询导出记录
func (r *ExportRecordRepoImpl) GetByUuids(uuids []uint64, opts ...DBOption) ([]model.ExportRecord, error) {
	if len(uuids) == 0 {
		return []model.ExportRecord{}, nil
	}

	var records []model.ExportRecord
	db := r.db.Model(&model.ExportRecord{}).Where("uuid IN ?", uuids)
	db = r.applyOptions(db, opts...)
	err := db.Find(&records).Error
	return records, err
}

// GetListWithPagination 分页获取导出记录列表
func (r *ExportRecordRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.ExportRecord, int64, error) {
	var records []model.ExportRecord
	var total int64

	// 构建基础查询
	db := r.db.Model(&model.ExportRecord{})
	db = r.applyOptions(db, opts...)

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// BatchDelete 批量删除导出记录（软删除）
func (r *ExportRecordRepoImpl) BatchDelete(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	return r.db.Model(&model.ExportRecord{}).
		Where("uuid IN ?", uuids).
		Update("delete_time", time.Now().Unix()).Error
}

// applyOptions 应用查询选项
func (r *ExportRecordRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

// 查询选项实现

// WhereExportType 导出类型条件
func (r *ExportRecordRepoImpl) WhereExportType(exportType *uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if exportType != nil {
			return db.Where("export_type = ?", *exportType)
		}
		return db
	}
}

// WhereStatus 状态条件
func (r *ExportRecordRepoImpl) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereNotDeleted 未删除条件
func (r *ExportRecordRepoImpl) WhereNotDeleted() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ?", 0)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *ExportRecordRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// PreloadFile 预加载文件信息
func (r *ExportRecordRepoImpl) PreloadFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("File")
	}
}

// 查询1小时内未完成的导出记录.
// 判定没有哪个导出会超过1个小时未完成,如果1小时后还未完成说明可能系统已经不在处理,则不再限时用户因为还有导出记录而不能新建导出
func (r *ExportRecordRepoImpl) GetUnfinishedExportRecord(exportType uint8) (*model.ExportRecord, error) {
	var record model.ExportRecord
	beforeHours := time.Now().Add(-1 * time.Hour).Unix()
	err := r.db.Model(&model.ExportRecord{}).
		Where("export_type = ?", exportType).
		Where("status = ?", model.ExportStatusPending).
		Where("create_time > ?", beforeHours).
		First(&record).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, nil
		}
		return nil, err
	}
	return &record, err
}

// GetByDateAndType 查询指定日期范围内指定类型的导出记录
// 注意：数据库连接已包含商户隔离，无需额外的 company_uuid 过滤
// 只查询导出成功的记录（status = 1），用于计算文件名序号
func (r *ExportRecordRepoImpl) GetByDateAndType(
	exportType uint8,
	startTime, endTime int64,
) ([]model.ExportRecord, error) {
	var records []model.ExportRecord
	err := r.db.Model(&model.ExportRecord{}).
		Where("export_type = ?", exportType).
		Where("create_time >= ?", startTime).
		Where("create_time <= ?", endTime).
		Where("status = ?", model.ExportStatusSuccess).
		Where("delete_time = ?", 0).
		Find(&records).Error
	// GORM 的 Find 方法在查询为空时会返回空切片 []model.ExportRecord{}，而不是 nil
	return records, err
}
