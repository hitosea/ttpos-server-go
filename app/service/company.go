package service

import (
	"errors"
	"jjjshop-server-go/app/dto/req" // 引入请求参数的包
	"jjjshop-server-go/app/model"
	"jjjshop-server-go/app/repository"
	"time"
)

// 定义公司服务接口
type CompanyServiceInterface interface {
	CreateCompany(param req.ParamCreateCompany) error // 创建公司方法
	UpdateCompany(param req.ParamUpdateCompany) error // 更新公司方法
	DeleteCompany(id uint) error                      // 删除公司方法
}

// 创建公司服务的实例
func NewCompanyService(companyRepo repository.CompanyRepositoryInterface) CompanyServiceInterface {
	return NewCompanyServiceImpl(companyRepo) // 返回新的公司服务实例
}

// 公司服务结构体
type CompanyService struct {
	companyRepo repository.CompanyRepositoryInterface
}

// 创建公司服务实例的函数
func NewCompanyServiceImpl(companyRepo repository.CompanyRepositoryInterface) *CompanyService {
	return &CompanyService{companyRepo: companyRepo} // 返回新的公司服务实例
}

// 创建公司方法的实现
func (s *CompanyService) CreateCompany(param req.ParamCreateCompany) error {
	return s.companyRepo.CreateCompanyInRepo(param)
}

// 更新公司方法的实现
func (s *CompanyService) UpdateCompany(param req.ParamUpdateCompany) error {
	return s.companyRepo.UpdateCompanyInRepo(param)
}

// 删除公司方法的实现
func (s *CompanyService) DeleteCompany(id uint) error {
	return s.companyRepo.DeleteCompanyInRepo(id)
}

func (s *CompanyService) GetLicense(company model.Company) error {
	if company.ExpireTime > 0 && company.ExpireTime < int(time.Now().Unix()) {
		return errors.New("店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if company.Status != 0 {
		return errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	return nil
}
