package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// ICompanyRepo 公司
type ICompanyRepo interface {
	WhereName(name string) DBOption
	WhereNotUuid(uuid uint64) DBOption

	GetCompany(opts ...DBOption) (model.Company, error)                               // 获取公司
	GetListByHeadquarterUuid(headquarterUuid uint64) ([]model.Company, error)         // 获取公司列表
	GetNoDeleteListByHeadquarterUuid(headquarterUuid uint64) ([]model.Company, error) // 获取公司列表
	GetCompanyInfo(ctx context.Context, opts ...DBOption) (model.Company, error)
	GetCompanyInfoByUuid(uuid uint64) (*model.Company, error)
	CreateCompany(obj model.Company) error
	UpdateCompany(uuid uint64, vars map[string]any) error
}

func NewCompanyRepo(db *gorm.DB) ICompanyRepo {
	return NewCompanyRepoImpl(db)
}

// NewCompanyRepoImpl 创建新的公司仓库实现
func NewCompanyRepoImpl(db *gorm.DB) ICompanyRepo {
	return &companyRepo{db: db}
}

type companyRepo struct {
	db *gorm.DB
}

func (r *companyRepo) WhereName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

func (r *companyRepo) WhereNotUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid <> ?", uuid)
	}
}

// CreateCompany 创建公司
func (r *companyRepo) CreateCompany(obj model.Company) error {
	company := obj
	companySetting := obj.CompanySetting

	company.SetNil()
	result := r.db.Create(&company)
	if result.Error != nil {
		return result.Error
	}

	result = r.db.Create(&companySetting)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 获取总部下的所有门店
func (r *companyRepo) GetListByHeadquarterUuid(headquarterUuid uint64) ([]model.Company, error) {
	var companies []model.Company
	db := r.db.Model(&model.Company{})
	db = db.Joins("LEFT JOIN ttpos_company_setting ON ttpos_company.uuid = ttpos_company_setting.company_uuid")
	db = db.Where("ttpos_company_setting.headquarter_uuid = ?", headquarterUuid)
	result := db.Find(&companies)
	if result.Error != nil {
		return companies, result.Error
	}
	return companies, nil
}

// 获取总部下的所有门店（不包含已删除的门店） 包括总部
func (r *companyRepo) GetNoDeleteListByHeadquarterUuid(headquarterUuid uint64) ([]model.Company, error) {
	var companies []model.Company
	db := r.db.Model(&model.Company{})
	db = db.Joins("LEFT JOIN ttpos_company_setting ON ttpos_company.uuid = ttpos_company_setting.company_uuid")
	db = db.Where("ttpos_company_setting.headquarter_uuid = ? OR ttpos_company.uuid = ?", headquarterUuid, headquarterUuid)
	db = db.Where("ttpos_company.delete_time = ?", 0)
	result := db.Find(&companies)
	if result.Error != nil {
		return companies, result.Error
	}
	return companies, nil
}

// GetCompanyInfo 获取公司信息
func (r *companyRepo) GetCompanyInfo(ctx context.Context, opts ...DBOption) (model.Company, error) {
	companyUuid := ctx.GetCompanyUuid()
	var company model.Company

	db := r.db.Model(&model.Company{}).Where("uuid = ?", companyUuid)

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&company)
	if result.Error != nil {
		return company, result.Error
	}

	return company, nil
}

// GetCompany 获取公司
func (r *companyRepo) GetCompany(opts ...DBOption) (model.Company, error) {
	var company model.Company

	db := r.db.Model(&model.Company{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&company)
	if result.Error != nil {
		return company, result.Error
	}

	return company, nil
}

func (r *companyRepo) GetCompanyInfoByUuid(uuid uint64) (*model.Company, error) {
	companyInfo, err := r.GetCompany(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "CompanySetting",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &companyInfo, nil
}

func (r *companyRepo) UpdateCompany(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Company{}).Where("uuid = ?", uuid).Updates(vars).Error
}
