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

// 删除7天前的数据 - 从数据库上删除
func (r *printerLogDataRepo) Delete7DaysAgo() error {
	return r.db.Model(&model.PrinterLogData{}).Where("create_time < ?", time.Now().Add(time.Duration(-7*24)*time.Hour).Unix()).Delete(&model.PrinterLogData{}).Error
}
