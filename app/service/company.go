package service

import (
	"errors"
	"time"

	"jjjshop-server-go/app/model"
)

type CompanyService struct {
}

func NewCompanyService() *CompanyService {
	return &CompanyService{}
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
