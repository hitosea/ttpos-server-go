package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderProductRepo interface {
	CreateSaleOrderProduct(model model.SaleOrderProduct) (uint64, error)
	UpdateSaleOrderProduct(model *model.SaleOrderProduct) error
	UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error
	GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error)
}

type saleOrderProductRepo struct {
	db *gorm.DB
}

func NewSaleOrderProductRepo(db *gorm.DB) ISaleOrderProductRepo {
	return &saleOrderProductRepo{db: db}
}

func (r *saleOrderProductRepo) CreateSaleOrderProduct(model model.SaleOrderProduct) (uint64, error) {
	db := r.db
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.Uuid, nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProduct(model *model.SaleOrderProduct) error {
	db := r.db
	if err := db.Model(&model).Updates(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error {
	db := r.db
	for _, m := range models {
		if err := db.Model(&m).Updates(m).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *saleOrderProductRepo) GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error) {
	db := r.db
	var model model.SaleOrderProduct
	if err := db.Where("uuid = ?", uuid).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}
