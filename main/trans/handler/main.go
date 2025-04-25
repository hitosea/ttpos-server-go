package handler

import (
	"gorm.io/gorm"
	v1 "ttpos-server-go/trans/v1"
)

func Run(sourceDB *gorm.DB, targetDB *gorm.DB) error {
	userService := v1.NewUserGradeService(sourceDB, targetDB)
	err := userService.ConvertUserGrade()
	if err != nil {
		return err
	}

	return nil
}
