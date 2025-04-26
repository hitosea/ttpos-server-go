package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// ICompanyRepo 公司
type ICompanyRepo interface {
	GetCompany(opts ...DBOption) (model.Company, error) // 获取公司
	GetCompanyInfo(ctx context.Context, opts ...DBOption) (model.Company, error)
	GetCompanyInfoByUuid(uuid uint64) (*model.Company, error)
	CreateCompany(obj model.Company) error
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
