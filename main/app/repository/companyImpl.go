package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type CompanyRepositoryInterface interface {
	CreateCompanyInRepo(company req.ParamCreateCompany) error
	UpdateCompanyInRepo(company req.ParamUpdateCompany) error
	DeleteCompanyInRepo(id uint) error
}

func NewCompanyRepository(dbm *database.DBManager) CompanyRepositoryInterface {
	return NewCompanyRepositoryImpl(dbm)
}

type CompanyRepositoryImpl struct {
	dbm                *database.DBManager
	companySettingRepo *CompanySettingRepository
	companyRepo        *CompanyRepository
}

func NewCompanyRepositoryImpl(dbm *database.DBManager) *CompanyRepositoryImpl {
	return &CompanyRepositoryImpl{
		dbm:                dbm,
		companySettingRepo: NewCompanySettingRepository(dbm),
		companyRepo:        NewCompanyRepo(dbm),
	}
}

func (r *CompanyRepositoryImpl) CreateCompanyInRepo(param req.ParamCreateCompany) error {
	// 开始事务
	db := r.dbm.GetDB(constant.DefaultDB)
	tx := db.Begin()

	// 1. 创建公司
	company := model.Company{
		Name: param.Name,
		Logo: param.Logo,
	}
	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 2. 创建公司设置
	companySetting := model.CompanySetting{
		CompanyId: company.ID,
		// 其他设置
	}
	if err := tx.Create(&companySetting).Error; err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 提交事务
	return tx.Commit().Error
}
func (r *CompanyRepositoryImpl) UpdateCompanyInRepo(param req.ParamUpdateCompany) error {
	db := r.dbm.GetDB(constant.DefaultDB)
	tx := db.Begin()

	// 1. 直接更新公司
	if err := r.companyRepo.Update(model.Company{
		ID:   uint(param.ID),
		Name: param.Name,
		Logo: param.Logo,
	}); err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 2. 直接更新公司设置
	if err := r.companySettingRepo.Update(model.CompanySetting{
		CompanyId:      uint(param.ID),
		Address:        param.Address,
		LinkPhone:      param.LinkPhone,
		CashLimit:      param.CashLimit,
		KitchenLimit:   param.KitchenLimit,
		TabletLimit:    param.TabletLimit,
		AssistantLimit: param.AssistantLimit,
	}); err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

func (r *CompanyRepositoryImpl) DeleteCompanyInRepo(id uint) error {
	db := r.dbm.GetDB(constant.DefaultDB)
	tx := db.Begin()

	// 1. 删除公司
	if err := r.companyRepo.Delete(id); err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 2. 删除公司设置
	if err := r.companySettingRepo.Delete(id); err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 提交事务
	return tx.Commit().Error
}
