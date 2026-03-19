package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IPrinterLogDataRepo interface {
	Create(printerLogData model.PrinterLogData) (model.PrinterLogData, error)
	Delete7DaysAgo() error
}

func NewPrinterLogDataRepo(db *gorm.DB) IPrinterLogDataRepo {
	return NewPrinterLogDataRepoImpl(db)
}

type printerLogDataRepo struct {
	db *gorm.DB
}

func NewPrinterLogDataRepoImpl(db *gorm.DB) IPrinterLogDataRepo {
	return &printerLogDataRepo{db: db}
}

func (r *printerLogDataRepo) Create(printerLogData model.PrinterLogData) (model.PrinterLogData, error) {
	err := r.db.Model(&model.PrinterLogData{}).Create(&printerLogData).Error
	return printerLogData, errors.WithMessage(err)
}

// 删除7天前的数据 - 分批物理删除，每批10000条
func (r *printerLogDataRepo) Delete7DaysAgo() error {
	threshold := time.Now().Add(-7 * 24 * time.Hour).Unix()
	batchSize := 10000
	for {
		result := r.db.Where("create_time < ?", threshold).Limit(batchSize).Delete(&model.PrinterLogData{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			break
		}
	}
	return nil
}
