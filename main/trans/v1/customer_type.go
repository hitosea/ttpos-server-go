package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type CustomerType struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement;comment:ID"`
	Name           string `gorm:"default:'';comment:名称"`
	Status         uint   `gorm:"default:1;comment:状态"`
	AppID          uint   `gorm:"default:0;comment:应用ID"`
	ShopSupplierID uint   `gorm:"default:0;comment:门店id"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type CustomerTypeRepository interface {
	GetCustomerTypeList() ([]*CustomerType, error)
	ConvertCustomerType() error
}

func NewCustomerTypeService(db *gorm.DB, targetDB *gorm.DB) CustomerTypeRepository {
	return &CustomerTypeService{
		db:       db,
		targetDB: targetDB,
	}
}

type CustomerTypeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *CustomerTypeService) GetCustomerTypeList() ([]*CustomerType, error) {
	var customerTypes []*CustomerType
	err := s.db.Find(&customerTypes).Error
	if err != nil {
		return nil, err
	}
	return customerTypes, nil
}

func (s *CustomerTypeService) ConvertCustomerType() error {
	customerTypes, err := s.GetCustomerTypeList()
	if err != nil {
		return err
	}
	for _, customerType := range customerTypes {

		fmt.Println(fmt.Sprintf("customerType: %+v", customerType))

		isDelete := 0
		if customerType.Status == 0 {
			isDelete = 1
		}
		customerType := model.BuffetCustomerType{
			BaseModel: model.BaseModel{
				Uuid:       uint64(customerType.ID),
				CreateTime: customerType.CreateTime,
				UpdateTime: customerType.UpdateTime,
				DeleteTime: int64(isDelete),
			},
			Name: customerType.Name,
		}
		_, err = base.NewBuffetCustomerTypeRepo(s.targetDB).CreateBuffetCustomerType(customerType)
		if err != nil {
			return err
		}
	}
	return nil
}
