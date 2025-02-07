package service

import (
	"ttpos-server-go/app/repository"
)

type LoginLogService struct {
	shopLoginLogRepo *repository.LoginLogRepo
}

func NewLoginLogService(
	shopLoginLogRepo *repository.LoginLogRepo,
) *LoginLogService {
	return &LoginLogService{
		shopLoginLogRepo: shopLoginLogRepo,
	}
}

func (s *LoginLogService) Save() error {
	return s.shopLoginLogRepo.Save(1, "", "", "")
}
