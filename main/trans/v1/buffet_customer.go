package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type BuffetCustomer struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;comment:'自助餐顾客关联表ID'"`
	BuffetID       uint64  `gorm:"default:0;comment:'自助餐ID'"`
	CustomerTypeID uint64  `gorm:"default:0;comment:'顾客类型ID'"`
	Price          float64 `gorm:"type:decimal(12,2);default:0.00;comment:'价格'"`
	AppID          int     `gorm:"default:0;comment:'应用ID'"`
	ShopSupplierID int     `gorm:"default:0;comment:'门店id'"`
	CreateTime     int64   `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64   `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type BuffetCustomerInterface interface {
	GetBuffetCustomerList() ([]BuffetCustomer, error)
	ConvertBuffetCustomer() error
}

type BuffetCustomerRepository struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (r *BuffetCustomerRepository) GetBuffetCustomerList() ([]BuffetCustomer, error) {
	var buffetCustomers []BuffetCustomer
	err := r.db.Find(&buffetCustomers).Error
	if err != nil {
		return nil, err
	}
	return buffetCustomers, nil
}

func (r *BuffetCustomerRepository) ConvertBuffetCustomer() error {
	buffetCustomers, err := r.GetBuffetCustomerList()
	if err != nil {
		return err
	}

	for _, buffetCustomer := range buffetCustomers {
		fmt.Println(fmt.Sprintf("-------迁移buffet_customer: %+v", buffetCustomer))

		// 创建自助餐顾客类型
		price := model.BuffetCustomerTypePrice{
			BaseModel: model.BaseModel{
				Uuid:       uint64(buffetCustomer.ID),
				CreateTime: buffetCustomer.CreateTime,
				UpdateTime: buffetCustomer.UpdateTime,
			},
			BuffetPackageUuid: buffetCustomer.BuffetID,
			CustomerTypeUuid:  buffetCustomer.CustomerTypeID,
			Price:             buffetCustomer.Price,
		}
		_, err := base.NewBuffetCustomerTypePriceRepo(r.targetDB).CreateBuffetCustomerTypePrice(price)
		if err != nil {
			return err
		}
	}
	return nil
}
