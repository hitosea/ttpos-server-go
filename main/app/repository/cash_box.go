package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type ICashBoxRepo interface {
	Get() model.CashBox
	Create(cashBox model.CashBox) (model.CashBox, error)

	Update(uuid uint64, vars map[string]any) error
}

func NewCashBoxRepo(db *gorm.DB) ICashBoxRepo {
	return NewCashBoxRepoImpl(db)
}

type CashBoxRepo struct {
	db *gorm.DB
}

func NewCashBoxRepoImpl(db *gorm.DB) *CashBoxRepo {
	return &CashBoxRepo{db: db}
}

func (r *CashBoxRepo) Get() model.CashBox {
	var cashBox model.CashBox
	r.db.First(&cashBox)
	return cashBox
}

func (r *CashBoxRepo) Create(cashBox model.CashBox) (model.CashBox, error) {
	err := r.db.Create(&cashBox).Error
	return cashBox, err
}

func (r *CashBoxRepo) Update(uuid uint64, vars map[string]any) error {
	err := r.db.Model(&model.CashBox{}).Where("uuid = ?", uuid).Updates(vars).Error
	return err
}
