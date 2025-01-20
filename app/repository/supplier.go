package repository

import (
	"jjjshop-server-go/app/constant"
	"jjjshop-server-go/app/model"
	"jjjshop-server-go/pkg/database"
)

type SupplierRepository struct {
	dbm *database.DBManager
}

func NewSupplierRepository(dbm *database.DBManager) *SupplierRepository {
	return &SupplierRepository{dbm: dbm}
}

func (r *SupplierRepository) GetById(id uint) model.Supplier {
	var supplier model.Supplier
	r.dbm.GetDB(constant.DefaultDB).Model(&model.Supplier{}).First(&supplier, id)
	return supplier
}
