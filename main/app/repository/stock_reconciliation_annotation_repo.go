package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IStockReconciliationAnnotationRepo 盘点单批注仓库接口
type IStockReconciliationAnnotationRepo interface {
	Create(annotation *model.StockReconciliationAnnotation) error
	GetListByStockReconciliationUuid(stockReconciliationUuid uint64) ([]*model.StockReconciliationAnnotation, error)
}

// NewStockReconciliationAnnotationRepo 创建新的盘点单批注仓库
func NewStockReconciliationAnnotationRepo(db *gorm.DB) IStockReconciliationAnnotationRepo {
	return &StockReconciliationAnnotationRepoImpl{db: db}
}

// StockReconciliationAnnotationRepoImpl 盘点单批注仓库实现
type StockReconciliationAnnotationRepoImpl struct {
	db *gorm.DB
}

// Create 创建批注
func (r *StockReconciliationAnnotationRepoImpl) Create(annotation *model.StockReconciliationAnnotation) error {
	if err := r.db.Create(annotation).Error; err != nil {
		return errors.WithMessage(err, "创建盘点单批注失败")
	}
	return nil
}

// GetListByStockReconciliationUuid 根据盘点单UUID获取批注列表（按创建时间倒序）
func (r *StockReconciliationAnnotationRepoImpl) GetListByStockReconciliationUuid(stockReconciliationUuid uint64) ([]*model.StockReconciliationAnnotation, error) {
	var list []*model.StockReconciliationAnnotation
	err := r.db.Scopes(NotDeleted).
		Where("stock_reconciliation_uuid = ?", stockReconciliationUuid).
		Order("create_time DESC").
		Find(&list).Error
	if err != nil {
		return nil, errors.WithMessage(err, "获取盘点单批注列表失败")
	}
	return list, nil
}
