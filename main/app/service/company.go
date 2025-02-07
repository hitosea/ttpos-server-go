package service

import (
	"errors"
	"time"
	"ttpos-server-go/app/dto/req" // 引入请求参数的包
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
)

// ICompanySrv 定义公司服务接口
type ICompanySrv interface {
	CreateCompany(param req.ParamCreateCompany) error // 创建公司方法
	UpdateCompany(param req.ParamUpdateCompany) error // 更新公司方法
	DeleteCompany(id uint) error                      // 删除公司方法
}

// NewCompanySrv 创建公司服务的实例
func NewCompanySrv(companyRepo repository.CompanyRepositoryInterface) ICompanySrv {
	return NewCompanySrvImpl(companyRepo) // 返回新的公司服务实例
}

// CompanySrv 公司服务结构体
type CompanySrv struct {
	companyRepo repository.CompanyRepositoryInterface
}

// NewCompanySrvImpl 创建公司服务实例的函数
func NewCompanySrvImpl(companyRepo repository.CompanyRepositoryInterface) *CompanySrv {
	return &CompanySrv{companyRepo: companyRepo} // 返回新的公司服务实例
}

// CreateCompany 创建公司方法的实现
func (s *CompanySrv) CreateCompany(param req.ParamCreateCompany) error {
	return s.companyRepo.CreateCompanyInRepo(param)
}

// UpdateCompany 更新公司方法的实现
func (s *CompanySrv) UpdateCompany(param req.ParamUpdateCompany) error {
	return s.companyRepo.UpdateCompanyInRepo(param)
}

// DeleteCompany 删除公司方法的实现
func (s *CompanySrv) DeleteCompany(id uint) error {
	return s.companyRepo.DeleteCompanyInRepo(id)
}

func (s *CompanySrv) GetLicense(company model.Company) error {
	if company.ExpireTime > 0 && company.ExpireTime < int(time.Now().Unix()) {
		return errors.New("店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if company.Status != 0 {
		return errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	return nil
}
