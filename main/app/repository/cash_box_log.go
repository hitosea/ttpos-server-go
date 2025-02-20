package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type ICashBoxLogRepo interface {
	Create(cashBox model.CashBoxLog) (model.CashBoxLog, error)
}

func NewCashBoxLogRepo(db *gorm.DB) ICashBoxLogRepo {
	return NewCashBoxLogRepoImpl(db)
}

type cashBoxLogRepo struct {
	db *gorm.DB
}

func NewCashBoxLogRepoImpl(db *gorm.DB) ICashBoxLogRepo {
	return &cashBoxLogRepo{db: db}
}

func (r *cashBoxLogRepo) Create(cashBoxLog model.CashBoxLog) (model.CashBoxLog, error) {
	err := r.db.Create(&cashBoxLog).Error
	return cashBoxLog, err
}
