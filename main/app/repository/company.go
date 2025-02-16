package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ICompanyRepo 公司
type ICompanyRepo interface {
	GetCompanyInfo(companyUuid uint64, opts ...DBOption) (model.Company, error)
}

func NewCompanyRepo(db *gorm.DB) ICompanyRepo {
	return NewCompanyRepoImpl(db)
}

// NewCompanyRepoImpl 创建新的公司仓库实现
func NewCompanyRepoImpl(db *gorm.DB) *CompanyRepoImpl {
	return &CompanyRepoImpl{db: db}
}

type CompanyRepoImpl struct {
	db *gorm.DB
}

// GetCompanyInfo 获取公司信息
func (r *CompanyRepoImpl) GetCompanyInfo(companyUuid uint64, opts ...DBOption) (model.Company, error) {
	var company model.Company

	db := r.db.Where("uuid = ?", companyUuid)

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&company)
	if result.Error != nil {
		return company, result.Error
	}

	return company, nil
}
