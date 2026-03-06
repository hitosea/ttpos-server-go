package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IMultiLanguageNameRepo 定义多语言名称仓库接口
type IMultiLanguageNameRepo interface {
	IMultiLanguageNameQueryRepo
	CreateMultiLanguageName(multiLanguageName model.MultiLanguageName) (uint64, error)          // 创建多语言名称
	CreateMultiLanguageNameDUPLICATE(multiLanguageName model.MultiLanguageName) (uint64, error) // 创建或更新多语言名称（ON DUPLICATE KEY UPDATE）
	CreateMultiLanguageNameList(multiLanguageNames []model.MultiLanguageName) error             // 创建多语言名称列表
	UpdateMultiLanguageName(id uint64, multiLanguageName model.MultiLanguageName) error         // 更新多语言名称
	UpdateMultiLanguageNameData(data map[string]any, opts ...DBOption) error                    // 更新多语言名称数据
	DeleteMultiLanguageName(id uint64) error                                                    // 删除多语言名称
	DestroyMultiLanguageName(opts ...DBOption) error                                            // 销毁多语言名称
	NotOverwriteMultiLanguageName(id uint64) error                                              // 不要覆盖多语言名称
	FindInBatches(batchSize int, fn func(batch []model.MultiLanguageName, batchRepo IMultiLanguageNameRepo) error, opts ...DBOption) error // 分批查询并处理
}

type IMultiLanguageNameQueryRepo interface {
	GetMultiLanguageName(opts ...DBOption) (*model.MultiLanguageName, error)
	GetMultiLanguageNameByUuid(uuid uint64) (model.MultiLanguageName, error)                                     // 获取多语言名称
	GetMultiLanguageNameListByUuids(uuids []uint64) ([]*model.MultiLanguageName, error)                          // 根据UUID列表批量获取多语言名称
	CountNameExistsByTable(tableName string, lang string, text string, excludeUuid uint64) (int64, error) // 按表名和语言查询多语言名称是否重复
}

// NewMultiLanguageNameRepo 创建新的多语言名称仓库
func NewMultiLanguageNameRepo(db *gorm.DB) IMultiLanguageNameRepo {
	return NewMultiLanguageNameRepositoryImpl(db)
}

// MultiLanguageNameRepoImpl 多语言名称仓库实现
type MultiLanguageNameRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// NewMultiLanguageNameRepositoryImpl 创建新的多语言名称仓库实现
func NewMultiLanguageNameRepositoryImpl(db *gorm.DB) IMultiLanguageNameRepo {
	return &MultiLanguageNameRepoImpl{db: db}
}

// GetMultiLanguageName 获取多语言名称
func (r *MultiLanguageNameRepoImpl) GetMultiLanguageName(opts ...DBOption) (*model.MultiLanguageName, error) {
	var multiLanguageName model.MultiLanguageName
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&multiLanguageName).Error
	return &multiLanguageName, errors.WithMessage(err)
}

// GetMultiLanguageNameByUuid 获取多语言名称
func (r *MultiLanguageNameRepoImpl) GetMultiLanguageNameByUuid(uuid uint64) (model.MultiLanguageName, error) {
	multiLanguageName, err := r.GetMultiLanguageName(
		CommonRepo.WhereByUuid(uuid),
	)
	if err != nil {
		return model.MultiLanguageName{}, errors.WithMessage(err)
	}
	return *multiLanguageName, nil
}

// GetMultiLanguageNameListByUuids 根据UUID列表批量获取多语言名称
func (r *MultiLanguageNameRepoImpl) GetMultiLanguageNameListByUuids(uuids []uint64) ([]*model.MultiLanguageName, error) {
	if len(uuids) == 0 {
		return []*model.MultiLanguageName{}, nil
	}
	var multiLanguageNames []model.MultiLanguageName
	err := r.db.Model(&model.MultiLanguageName{}).
		Where("uuid IN ?", uuids).
		Find(&multiLanguageNames).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 转换为指针切片
	result := make([]*model.MultiLanguageName, len(multiLanguageNames))
	for i := range multiLanguageNames {
		result[i] = &multiLanguageNames[i]
	}
	return result, nil
}

// CreateMultiLanguageName 创建多语言名称
func (r *MultiLanguageNameRepoImpl) CreateMultiLanguageName(multiLanguageName model.MultiLanguageName) (uint64, error) {
	err := r.db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error // 将多语言名称插入数据库
	return multiLanguageName.Uuid, errors.WithMessage(err)
}

// CreateMultiLanguageNameDUPLICATE 创建或更新多语言名称（ON DUPLICATE KEY UPDATE）
// 如果uuid已存在，则更新现有记录；如果不存在，则创建新记录
func (r *MultiLanguageNameRepoImpl) CreateMultiLanguageNameDUPLICATE(multiLanguageName model.MultiLanguageName) (uint64, error) {
	// 方法1：使用GORM的OnConflict（推荐）
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uuid"}}, // 指定冲突的列
		DoUpdates: clause.AssignmentColumns([]string{
			"en_name", "zh_name", "zh_tw_name", "th_name", "my_name",
			"ja_name", "ko_name", "tr_name", "sv_name", "update_time",
		}), // 指定要更新的字段
	}).Create(&multiLanguageName).Error

	if err != nil {
		return 0, errors.WithMessage(err, "创建或更新多语言名称失败")
	}

	return multiLanguageName.Uuid, nil
}

// CreateMultiLanguageNameList 创建多语言名称列表
func (r *MultiLanguageNameRepoImpl) CreateMultiLanguageNameList(multiLanguageNames []model.MultiLanguageName) error {
	if err := r.db.Create(&multiLanguageNames).Error; err != nil {
		return errors.WithMessage(err, "创建多语言名称列表失败")
	}
	return nil
}

// UpdateMultiLanguageName 更新多语言名称
func (r *MultiLanguageNameRepoImpl) UpdateMultiLanguageName(uuid uint64, multiLanguageName model.MultiLanguageName) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"zh_name":    multiLanguageName.ZhName,
		"th_name":    multiLanguageName.ThName,
		"en_name":    multiLanguageName.EnName,
		"zh_tw_name": multiLanguageName.ZhTwName,
		"ja_name":    multiLanguageName.JaName,
		"ko_name":    multiLanguageName.KoName,
		"my_name":    multiLanguageName.MyName,
		"tr_name":    multiLanguageName.TrName,
		"sv_name":    multiLanguageName.SvName,
	}).Error // 更新数据库中的多语言名称
}

// UpdateMultiLanguageNameData 更新多语言名称数据
func (r *MultiLanguageNameRepoImpl) UpdateMultiLanguageNameData(data map[string]any, opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.MultiLanguageName{}).Updates(data).Error
	if err != nil {
		return errors.WithMessage(err, "更新多语言名称数据失败")
	}
	return nil
}

// DeleteMultiLanguageName 删除多语言名称
func (r *MultiLanguageNameRepoImpl) DeleteMultiLanguageName(id uint64) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", id).Update("delete_time", uint(time.Now().Unix())).Error // 逻辑删除多语言名称
}

// DestroyMultiLanguageName 销毁多语言名称
func (r *MultiLanguageNameRepoImpl) DestroyMultiLanguageName(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	return db.Delete(&model.MultiLanguageName{}).Error
}

// NotOverwriteMultiLanguageName 不要覆盖多语言名称
func (r *MultiLanguageNameRepoImpl) NotOverwriteMultiLanguageName(id uint64) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", id).Update("not_overwrite", 1).Error
}

// CountNameExistsByTable 按表名和语言查询多语言名称是否重复
func (r *MultiLanguageNameRepoImpl) CountNameExistsByTable(tableName string, lang string, text string, excludeUuid uint64) (int64, error) {
	langColumnMap := map[string]string{
		"zh":   "zh_name",
		"zhtw": "zh_tw_name",
		"en":   "en_name",
		"ja":   "ja_name",
		"ko":   "ko_name",
		"my":   "my_name",
		"sv":   "sv_name",
		"th":   "th_name",
		"tr":   "tr_name",
	}
	langColumn, ok := langColumnMap[lang]
	if !ok {
		return 0, errors.New("不支持的语言: " + lang)
	}

	mlnTable := r.db.Table("multi_language_name").Name()
	bizTable := r.db.Table(tableName).Name()
	var count int64
	query := r.db.Table(bizTable).
		Joins(fmt.Sprintf("JOIN %s ON %s.multi_language_name_uuid = %s.uuid", mlnTable, bizTable, mlnTable)).
		Where(fmt.Sprintf("%s.%s = ?", mlnTable, langColumn), text).
		Where(fmt.Sprintf("%s.delete_time = 0", mlnTable))
	if excludeUuid != 0 {
		query = query.Where(fmt.Sprintf("%s.uuid != ?", bizTable), excludeUuid)
	}
	err := query.Count(&count).Error
	return count, err
}

// FindInBatches 分批查询并处理多语言名称
func (r *MultiLanguageNameRepoImpl) FindInBatches(batchSize int, fn func(batch []model.MultiLanguageName, batchRepo IMultiLanguageNameRepo) error, opts ...DBOption) error {
	var results []model.MultiLanguageName
	query := r.db.Model(&model.MultiLanguageName{}).Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	return query.FindInBatches(&results, batchSize, func(tx *gorm.DB, batch int) error {
		batchRepo := NewMultiLanguageNameRepo(tx)
		return fn(results, batchRepo)
	}).Error
}
