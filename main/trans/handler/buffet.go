package handler

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	oldModel "ttpos-server-go/trans/v1"
	newModel "ttpos-server-go/trans/v2/model"

	"gorm.io/gorm"
)

func testConvertBuffet() error {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	buffetService := BuffetService{db: db, targetDB: targetDB}

	err = buffetService.ConvertBuffet()
	if err != nil {
		return errors.WithMessage(err, "ConvertBuffet failed")
	}
	return nil
}

type BuffetInterface interface {
	GetBuffetList() ([]oldModel.Buffet, error)
	ConvertBuffet() error
}

func NewBuffetService(db *gorm.DB, targetDB *gorm.DB) BuffetInterface {
	return &BuffetService{db: db, targetDB: targetDB}
}

type BuffetService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *BuffetService) GetBuffetList() ([]oldModel.Buffet, error) {
	var buffets []oldModel.Buffet
	err := s.db.Preload("BuffetProducts").Preload("BuffetCustomers").Preload("BuffetTaxes").Find(&buffets).Error
	if err != nil {
		return nil, err
	}
	return buffets, nil
}

func (s *BuffetService) ConvertBuffet() error {
	buffets, err := s.GetBuffetList()
	if err != nil {
		return err
	}
	if err := s.targetDB.Transaction(func(tx *gorm.DB) error {
		for index, _ := range buffets {
			buffet := &buffets[index]
			fmt.Println(fmt.Sprintf("-----迁移自助餐：%+v", buffet))
			buffetPackage, err := newModel.NewBuffet(buffet)
			if err != nil {
				return err
			}
			repo := repository.NewBuffetRepo(tx)
			if err := repo.CreateBuffet(*buffetPackage); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
