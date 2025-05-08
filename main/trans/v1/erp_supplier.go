package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type ErpSupplier struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;comment:供应商ID"`
	Name           string `gorm:"type:varchar(255);default:'';comment:供应商名称"`
	Address        string `gorm:"type:varchar(2000);default:'';comment:供应商地址"`
	ContactPerson  string `gorm:"type:varchar(255);default:'';comment:联系人"`
	ContactPhone   string `gorm:"type:varchar(100);default:'';comment:联系电话"`
	Position       string `gorm:"type:varchar(255);default:'';comment:职位"`
	PurchaserID    int    `gorm:"default:0;comment:采购负责人id"`
	ShopSupplierID int    `gorm:"default:0;comment:门店id"`
	AppID          int    `gorm:"default:0;comment:应用id"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime     int64  `gorm:"comment:删除时间"`
}

type ErpSupplierRepository interface {
	GetErpSupplierList() ([]*ErpSupplier, error)
	ConvertErpSupplier() error
}

func NewErpSupplierService(db *gorm.DB, targetDB *gorm.DB) ErpSupplierRepository {
	return &ErpSupplierService{db: db, targetDB: targetDB}
}

type ErpSupplierService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ErpSupplierService) GetErpSupplierList() ([]*ErpSupplier, error) {
	var erpSuppliers []*ErpSupplier
	if err := s.db.Find(&erpSuppliers).Error; err != nil {
		return nil, err
	}
	return erpSuppliers, nil
}

func (s *ErpSupplierService) ConvertErpSupplier() error {
	var erpSuppliers []*ErpSupplier
	if err := s.db.Find(&erpSuppliers).Error; err != nil {
		return errors.WithMessage(err)
	}
	for _, erpSupplier := range erpSuppliers {
		fmt.Println(fmt.Sprintf("erpSupplier: %+v", erpSupplier))
		supplier := model.Supplier{
			BaseModel: model.BaseModel{
				Uuid:       uint64(erpSupplier.ID),
				CreateTime: erpSupplier.CreateTime,
				UpdateTime: erpSupplier.UpdateTime,
			},
			Name:         erpSupplier.Name,
			Address:      erpSupplier.Address,
			ContactName:  erpSupplier.ContactPerson,
			ContactPhone: erpSupplier.ContactPhone,
			Position:     erpSupplier.Position,
			StaffUuid:    uint64(erpSupplier.PurchaserID),
		}
		_, err := base.NewSupplierRepo(s.targetDB).CreateSupplier(supplier)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
