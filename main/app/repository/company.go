package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type CompanyRepoInterface interface {
	Create(company model.Company) error
	Update(company model.Company) error
	Delete(id uint) error
}

func NewCompanyRepo(dbm *database.DBManager) *CompanyRepository {
	return &CompanyRepository{dbm: dbm}
}

type CompanyRepository struct {
	dbm *database.DBManager
}

func (r *CompanyRepository) Create(company model.Company) error {
	return r.dbm.GetDB(constant.DefaultDB).Create(&company).Error
}

func (r *CompanyRepository) Update(company model.Company) error {
	return r.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("id = ?", company.ID).Updates(company).Error
}

func (r *CompanyRepository) Delete(id uint) error {
	return r.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("id = ?", id).Update("is_delete", 1).Error
}
