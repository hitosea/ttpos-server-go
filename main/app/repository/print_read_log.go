package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IPrinterReadLogRepo interface {
	Update(deviceId string, vars map[string]any) error
}

func NewPrinterReadLogRepo(db *gorm.DB) IPrinterReadLogRepo {
	return NewPrinterReadLogRepoImpl(db)
}

type printerReadLogRepo struct {
	db *gorm.DB
}

func NewPrinterReadLogRepoImpl(db *gorm.DB) IPrinterReadLogRepo {
	return &printerReadLogRepo{db: db}
}

func (r *printerReadLogRepo) Update(deviceId string, vars map[string]any) error {
	return r.db.Model(&model.PrinterReadLog{}).Where("device_id = ?", deviceId).Updates(vars).Error
}
