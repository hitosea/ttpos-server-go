package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IStocktakeSnapshotRepo 盘点快照仓库接口
type IStocktakeSnapshotRepo interface {
	// BatchCreate 批量创建盘点快照
	BatchCreate(snapshots []model.StocktakeSnapshot, batchSize int) error
}

// NewStocktakeSnapshotRepo 创建新的盘点快照仓库
func NewStocktakeSnapshotRepo(db *gorm.DB) IStocktakeSnapshotRepo {
	return &stocktakeSnapshotRepoImpl{db: db}
}

type stocktakeSnapshotRepoImpl struct {
	db *gorm.DB
}

// BatchCreate 批量创建盘点快照
func (r *stocktakeSnapshotRepoImpl) BatchCreate(snapshots []model.StocktakeSnapshot, batchSize int) error {
	if len(snapshots) == 0 {
		return nil
	}
	return r.db.CreateInBatches(&snapshots, batchSize).Error
}
