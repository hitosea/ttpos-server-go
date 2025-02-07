package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ICompanyRepo interface {
	Create(company model.Company) error
	Update(company model.Company) error
	Delete(id uint) error
}

func NewCompanyRepoImpl(dbm *database.DBManager) *CompanyRepoImpl {
	return &CompanyRepoImpl{dbm: dbm}
}

type CompanyRepoImpl struct {
	dbm *database.DBManager
}

func (r *CompanyRepoImpl) Create(company model.Company) error {
	return r.dbm.GetDB(constant.DefaultDB).Create(&company).Error
}

func (r *CompanyRepoImpl) Update(company model.Company) error {
	return r.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("id = ?", company.ID).Updates(company).Error
}

func (r *CompanyRepoImpl) Delete(id uint) error {
	return r.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("id = ?", id).Update("is_delete", 1).Error
}
