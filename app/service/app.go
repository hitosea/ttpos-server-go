package service

import (
	"errors"
	"time"

	"ttpos-server-go/app/model"
)

type AppService struct {
}

func NewAppService() *AppService {
	return &AppService{}
}

func (s *AppService) GetLicense(app model.Company) error {
	if app.ExpireTime > 0 && app.ExpireTime < int(time.Now().Unix()) {
		return errors.New("店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if app.Status != 0 {
		return errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	return nil
}
