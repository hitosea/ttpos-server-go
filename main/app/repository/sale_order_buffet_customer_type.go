package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderBuffetCustomerTypeRepo interface {
	UpdateOrCreateSaleOrderBuffetCustomerTypeRecord(saleOrderBuffetCustomerType model.SaleOrderBuffetCustomerType) error
	CreateSaleOrderBuffetCustomerTypeRecord(saleOrderBuffetCustomerType model.SaleOrderBuffetCustomerType) error
	UpdateSaleOrderBuffetCustomerTypeRecord(saleOrderBuffetCustomerType model.SaleOrderBuffetCustomerType) error
	DeleteSaleOrderBuffetCustomerTypeBySaleBillUuid(saleBillUuid uint64) error                      // 根据销售账单uuid删除销售订单自助餐顾客。delete_time赋值为当前时间
	GetSaleOrderBuffetCustomerTypes(opts ...DBOption) ([]*model.SaleOrderBuffetCustomerType, error) // 根据销售订单uuid获取销售订单自助餐顾客
}

func NewSaleOrderBuffetCustomerTypeRepo(db *gorm.DB) ISaleOrderBuffetCustomerTypeRepo {
	return NewSaleOrderBuffetCustomerTypeRepoImpl(db)
}

type saleOrderBuffetCustomerTypeRepo struct {
	db *gorm.DB
}

func NewSaleOrderBuffetCustomerTypeRepoImpl(db *gorm.DB) ISaleOrderBuffetCustomerTypeRepo {
	return &saleOrderBuffetCustomerTypeRepo{db: db}
}

func (r *saleOrderBuffetCustomerTypeRepo) CreateSaleOrderBuffetCustomerTypeRecord(obj model.SaleOrderBuffetCustomerType) error {
	obj.SetNil()
	return r.db.Model(&model.SaleOrderBuffetCustomerType{}).Create(&obj).Error
}

func (r *saleOrderBuffetCustomerTypeRepo) UpdateSaleOrderBuffetCustomerTypeRecord(obj model.SaleOrderBuffetCustomerType) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.WithMessage(errors.New("销售订单自助餐顾客类型UUID或ID不能为0"))
	}
	return r.db.Model(&model.SaleOrderBuffetCustomerType{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
}

func (r *saleOrderBuffetCustomerTypeRepo) UpdateOrCreateSaleOrderBuffetCustomerTypeRecord(obj model.SaleOrderBuffetCustomerType) error {
	// 如果主键id为0则create，否则update
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return r.CreateSaleOrderBuffetCustomerTypeRecord(obj)
	}
	return r.UpdateSaleOrderBuffetCustomerTypeRecord(obj)
}

func (r *saleOrderBuffetCustomerTypeRepo) DeleteSaleOrderBuffetCustomerTypeBySaleBillUuid(saleBillUuid uint64) error {
	now := time.Now().Unix()
	return r.db.Model(&model.SaleOrderBuffetCustomerType{}).Where("sale_bill_uuid = ?", saleBillUuid).Update("delete_time", now).Error
}

func (r *saleOrderBuffetCustomerTypeRepo) GetSaleOrderBuffetCustomerTypes(opts ...DBOption) ([]*model.SaleOrderBuffetCustomerType, error) {
	var buffetCustomerTypes []*model.SaleOrderBuffetCustomerType
	db := r.db.Model(&model.SaleOrderBuffetCustomerType{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&buffetCustomerTypes).Error
	return buffetCustomerTypes, err
}
