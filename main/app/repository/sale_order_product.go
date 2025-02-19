package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderProductRepo interface {
	CreateSaleOrderProduct(model model.SaleOrderProduct) (uint64, error)
}

type saleOrderProductRepo struct {
	db *gorm.DB
}

func NewSaleOrderProductRepo(db *gorm.DB) ISaleOrderProductRepo {
	return &saleOrderProductRepo{db: db}
}
func (r *saleOrderProductRepo) CreateSaleOrderProduct(model model.SaleOrderProduct) (uint64, error) {
	db := r.db
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return model.Uuid, nil
}
