package service

import (
	"ttpos-server-go/app/repository"
)

type LoginLogService struct {
	shopLoginLogRepo *repository.LoginLogRepository
}

func NewLoginLogService(
	shopLoginLogRepo *repository.LoginLogRepository,
) *LoginLogService {
	return &LoginLogService{
		shopLoginLogRepo: shopLoginLogRepo,
	}
}

func (s *LoginLogService) Save() error {
	return s.shopLoginLogRepo.Save(1, "", "", "")
}
