package handler

import (
	v1 "ttpos-server-go/trans/v1"

	"gorm.io/gorm"
)

func Run(sourceDB *gorm.DB, targetDB *gorm.DB) error {
	userService := v1.NewUserGradeService(sourceDB, targetDB)
	err := userService.ConvertUserGrade()
	if err != nil {
		return err
	}

	tableAreaService := v1.NewTableAreaService(sourceDB, targetDB)
	err = tableAreaService.ConvertTableArea()
	if err != nil {
		return err
	}

	appService := v1.NewAppService(sourceDB, targetDB)
	err = appService.ConvertApp()
	if err != nil {
		return err
	}

	return nil
}
