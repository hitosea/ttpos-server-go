package handler

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	oldModel "ttpos-server-go/trans/v1"
	newModel "ttpos-server-go/trans/v2/model"

	"gorm.io/gorm"
)

func testConvertProductSauce() error {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	productSauceService := ProductSauceService{db: db, targetDB: targetDB}

	err = productSauceService.ConvertProductSauce()
	if err != nil {
		return errors.WithMessage(err, "ConvertProduct failed")
	}
	return nil
}

type ProductSauceInterface interface {
	GetProductSauceList() ([]oldModel.Feed, error)
	ConvertProductSauce() error
}

func NewProductSauceService(db *gorm.DB, targetDB *gorm.DB) ProductSauceInterface {
	return &ProductSauceService{db: db, targetDB: targetDB}
}

type ProductSauceService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductSauceService) GetProductSauceList() ([]oldModel.Feed, error) {
	var sauces []oldModel.Feed
	err := s.db.Preload("ProductFeedMaterials").Order("create_time asc").Find(&sauces).Error
	if err != nil {
		return nil, err
	}
	return sauces, nil
}

func (s *ProductSauceService) ConvertProductSauce() error {
	sauces, err := s.GetProductSauceList()
	if err != nil {
		return err
	}
	if err := s.targetDB.Transaction(func(tx *gorm.DB) error {
		sort := 1
		for index, _ := range sauces {
			feed := &sauces[index]
			fmt.Println(fmt.Sprintf("-----迁移小料：%+v", feed))
			productSauce, err := newModel.NewProductSauce(feed)
			if err != nil {
				return err
			}
			productSauce.Sort = sort
			repo := repository.NewProductSauceRepo(tx)
			if err := repo.CreateProductSauce(productSauce); err != nil {
				return err
			}
			sort++
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
