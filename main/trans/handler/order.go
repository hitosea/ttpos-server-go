package handler

import (
	"ttpos-server-go/app/errors"

	"gorm.io/gorm"
)

func testConvertOrder() error {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	orderService := OrderService{db: db, targetDB: targetDB}

	err = orderService.ConvertOrder()
	if err != nil {
		return errors.WithMessage(err, "ConvertOrder failed")
	}
	return nil
}

type OrderInterface interface {
	// GetOrderList() ([]oldModel.Product, error)
	ConvertOrder() error
}

type OrderService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

// func (s *OrderService) GetOrderList() ([]oldModel.SaleOrder, error) {
// 	var orders []oldModel.SaleOrder
// 	err := s.db.Find(&orders).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return orders, nil
// }

func (s *OrderService) ConvertOrder() error {
	return nil
}
