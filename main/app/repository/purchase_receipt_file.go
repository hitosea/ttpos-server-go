package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseReceiptFileRepo 收货单附件Repository接口
type IPurchaseReceiptFileRepo interface {
	// Create 创建附件关联
	Create(file *model.PurchaseReceiptFile) error

	// BatchCreate 批量创建附件关联
	BatchCreate(files []model.PurchaseReceiptFile) error

	// GetByReceiptOrderUuid 根据收货单UUID查询附件
	GetByReceiptOrderUuid(receiptOrderUuid uint64) ([]model.PurchaseReceiptFile, error)

	// GetByUuid 根据UUID查询附件关联
	GetByUuid(uuid uint64) (*model.PurchaseReceiptFile, error)

	// DeleteByUuid 删除附件关联
	DeleteByUuid(uuid uint64) error

	// DeleteByReceiptOrderUuid 删除收货单的所有附件
	DeleteByReceiptOrderUuid(receiptOrderUuid uint64) error

	// DeleteByFileUuidAndReceiptOrderUuid 删除指定文件的关联
	DeleteByFileUuidAndReceiptOrderUuid(fileUuid uint64, receiptOrderUuid uint64) error

	// CountByReceiptOrderUuid 统计收货单附件数量
	CountByReceiptOrderUuid(receiptOrderUuid uint64) (int64, error)

	// GetByReceiptOrderUuidWithFiles 根据收货单UUID查询附件（预加载文件信息）
	GetByReceiptOrderUuidWithFiles(receiptOrderUuid uint64) ([]model.PurchaseReceiptFile, error)
}

// PurchaseReceiptFileRepoImpl 收货单附件Repository实现
type PurchaseReceiptFileRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseReceiptFileRepo 创建收货单附件Repository
func NewPurchaseReceiptFileRepo(db *gorm.DB) IPurchaseReceiptFileRepo {
	return &PurchaseReceiptFileRepoImpl{db: db}
}

// Create 创建附件关联
func (r *PurchaseReceiptFileRepoImpl) Create(file *model.PurchaseReceiptFile) error {
	return r.db.Create(file).Error
}

// BatchCreate 批量创建附件关联
func (r *PurchaseReceiptFileRepoImpl) BatchCreate(files []model.PurchaseReceiptFile) error {
	if len(files) == 0 {
		return nil
	}
	return r.db.Create(&files).Error
}

// GetByReceiptOrderUuid 根据收货单UUID查询附件
func (r *PurchaseReceiptFileRepoImpl) GetByReceiptOrderUuid(receiptOrderUuid uint64) ([]model.PurchaseReceiptFile, error) {
	var files []model.PurchaseReceiptFile
	err := r.db.Where("receipt_order_uuid = ? AND delete_time = 0", receiptOrderUuid).
		Order("sort_order ASC, create_time ASC").
		Find(&files).Error
	return files, err
}

// GetByUuid 根据UUID查询附件关联
func (r *PurchaseReceiptFileRepoImpl) GetByUuid(uuid uint64) (*model.PurchaseReceiptFile, error) {
	var file model.PurchaseReceiptFile
	err := r.db.Where("uuid = ? AND delete_time = 0", uuid).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteByUuid 删除附件关联（软删除）
func (r *PurchaseReceiptFileRepoImpl) DeleteByUuid(uuid uint64) error {
	return r.db.Model(&model.PurchaseReceiptFile{}).
		Where("uuid = ?", uuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// DeleteByReceiptOrderUuid 删除收货单的所有附件（软删除）
func (r *PurchaseReceiptFileRepoImpl) DeleteByReceiptOrderUuid(receiptOrderUuid uint64) error {
	return r.db.Model(&model.PurchaseReceiptFile{}).
		Where("receipt_order_uuid = ? AND delete_time = 0", receiptOrderUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// DeleteByFileUuidAndReceiptOrderUuid 删除指定文件的关联（软删除）
func (r *PurchaseReceiptFileRepoImpl) DeleteByFileUuidAndReceiptOrderUuid(fileUuid uint64, receiptOrderUuid uint64) error {
	return r.db.Model(&model.PurchaseReceiptFile{}).
		Where("file_uuid = ? AND receipt_order_uuid = ? AND delete_time = 0", fileUuid, receiptOrderUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// CountByReceiptOrderUuid 统计收货单附件数量
func (r *PurchaseReceiptFileRepoImpl) CountByReceiptOrderUuid(receiptOrderUuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.PurchaseReceiptFile{}).
		Where("receipt_order_uuid = ? AND delete_time = 0", receiptOrderUuid).
		Count(&count).Error
	return count, err
}

// GetByReceiptOrderUuidWithFiles 根据收货单UUID查询附件（预加载文件信息）
func (r *PurchaseReceiptFileRepoImpl) GetByReceiptOrderUuidWithFiles(receiptOrderUuid uint64) ([]model.PurchaseReceiptFile, error) {
	var files []model.PurchaseReceiptFile
	err := r.db.Where("receipt_order_uuid = ? AND delete_time = 0", receiptOrderUuid).
		Preload("File").
		Order("sort_order ASC, create_time ASC").
		Find(&files).Error
	return files, err
}
