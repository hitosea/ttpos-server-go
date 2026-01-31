package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITransferOrderFileRepo 调拨单附件Repository接口
type ITransferOrderFileRepo interface {
	// Create 创建附件关联
	Create(file *model.TransferOrderFile) error

	// BatchCreate 批量创建附件关联
	BatchCreate(files []model.TransferOrderFile) error

	// GetByTransferOrderUuid 根据调拨单UUID查询附件
	GetByTransferOrderUuid(transferOrderUuid uint64) ([]model.TransferOrderFile, error)

	// GetByUuid 根据UUID查询附件关联
	GetByUuid(uuid uint64) (*model.TransferOrderFile, error)

	// DeleteByUuid 删除附件关联
	DeleteByUuid(uuid uint64) error

	// DeleteByTransferOrderUuid 删除调拨单的所有附件
	DeleteByTransferOrderUuid(transferOrderUuid uint64) error

	// DeleteByFileUuidAndTransferOrderUuid 删除指定文件的关联
	DeleteByFileUuidAndTransferOrderUuid(fileUuid uint64, transferOrderUuid uint64) error

	// CountByTransferOrderUuid 统计调拨单附件数量
	CountByTransferOrderUuid(transferOrderUuid uint64) (int64, error)

	// GetByTransferOrderUuidWithFiles 根据调拨单UUID查询附件（预加载文件信息）
	GetByTransferOrderUuidWithFiles(transferOrderUuid uint64) ([]model.TransferOrderFile, error)
}

// TransferOrderFileRepoImpl 调拨单附件Repository实现
type TransferOrderFileRepoImpl struct {
	db *gorm.DB
}

// NewTransferOrderFileRepo 创建调拨单附件Repository
func NewTransferOrderFileRepo(db *gorm.DB) ITransferOrderFileRepo {
	return &TransferOrderFileRepoImpl{db: db}
}

// Create 创建附件关联
func (r *TransferOrderFileRepoImpl) Create(file *model.TransferOrderFile) error {
	return r.db.Create(file).Error
}

// BatchCreate 批量创建附件关联
func (r *TransferOrderFileRepoImpl) BatchCreate(files []model.TransferOrderFile) error {
	if len(files) == 0 {
		return nil
	}
	return r.db.Create(&files).Error
}

// GetByTransferOrderUuid 根据调拨单UUID查询附件
func (r *TransferOrderFileRepoImpl) GetByTransferOrderUuid(transferOrderUuid uint64) ([]model.TransferOrderFile, error) {
	var files []model.TransferOrderFile
	err := r.db.Where("transfer_order_uuid = ? AND delete_time = 0", transferOrderUuid).
		Order("sort_order ASC, create_time ASC").
		Find(&files).Error
	return files, err
}

// GetByUuid 根据UUID查询附件关联
func (r *TransferOrderFileRepoImpl) GetByUuid(uuid uint64) (*model.TransferOrderFile, error) {
	var file model.TransferOrderFile
	err := r.db.Where("uuid = ? AND delete_time = 0", uuid).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteByUuid 删除附件关联（软删除）
func (r *TransferOrderFileRepoImpl) DeleteByUuid(uuid uint64) error {
	return r.db.Model(&model.TransferOrderFile{}).
		Where("uuid = ?", uuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// DeleteByTransferOrderUuid 删除调拨单的所有附件（软删除）
func (r *TransferOrderFileRepoImpl) DeleteByTransferOrderUuid(transferOrderUuid uint64) error {
	return r.db.Model(&model.TransferOrderFile{}).
		Where("transfer_order_uuid = ? AND delete_time = 0", transferOrderUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// DeleteByFileUuidAndTransferOrderUuid 删除指定文件的关联（软删除）
func (r *TransferOrderFileRepoImpl) DeleteByFileUuidAndTransferOrderUuid(fileUuid uint64, transferOrderUuid uint64) error {
	return r.db.Model(&model.TransferOrderFile{}).
		Where("file_uuid = ? AND transfer_order_uuid = ? AND delete_time = 0", fileUuid, transferOrderUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// CountByTransferOrderUuid 统计调拨单附件数量
func (r *TransferOrderFileRepoImpl) CountByTransferOrderUuid(transferOrderUuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.TransferOrderFile{}).
		Where("transfer_order_uuid = ? AND delete_time = 0", transferOrderUuid).
		Count(&count).Error
	return count, err
}

// GetByTransferOrderUuidWithFiles 根据调拨单UUID查询附件（预加载文件信息）
func (r *TransferOrderFileRepoImpl) GetByTransferOrderUuidWithFiles(transferOrderUuid uint64) ([]model.TransferOrderFile, error) {
	var files []model.TransferOrderFile
	err := r.db.Where("transfer_order_uuid = ? AND delete_time = 0", transferOrderUuid).
		Preload("File").
		Order("sort_order ASC, create_time ASC").
		Find(&files).Error
	return files, err
}
