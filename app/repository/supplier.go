package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
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
