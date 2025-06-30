package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderBuffetDelayProductRepo interface {
	UpdateOrCreateSaleOrderBuffetDelayProductRecord(saleOrderBuffetDelay model.SaleOrderBuffetDelayProduct) error
	CreateSaleOrderBuffetDelayProductRecord(saleOrderBuffetDelay model.SaleOrderBuffetDelayProduct) error
	UpdateSaleOrderBuffetDelayProductRecord(saleOrderBuffetDelay model.SaleOrderBuffetDelayProduct) error
	GetSaleOrderBuffetDelayProducts(opts ...DBOption) ([]*model.SaleOrderBuffetDelayProduct, error) // 根据销售订单uuid获取销售订单自助餐延迟商品
}

func NewSaleOrderBuffetDelayProductRepo(db *gorm.DB) ISaleOrderBuffetDelayProductRepo {
	return NewSaleOrderBuffetDelayProductRepoImpl(db)
}

type saleOrderBuffetDelayProductRepo struct {
	db *gorm.DB
}

func NewSaleOrderBuffetDelayProductRepoImpl(db *gorm.DB) ISaleOrderBuffetDelayProductRepo {
	return &saleOrderBuffetDelayProductRepo{db: db}
}

func (r *saleOrderBuffetDelayProductRepo) CreateSaleOrderBuffetDelayProductRecord(obj model.SaleOrderBuffetDelayProduct) error {
	obj.SetNil()
	return r.db.Model(&model.SaleOrderBuffetDelayProduct{}).Create(&obj).Error
}

func (r *saleOrderBuffetDelayProductRepo) UpdateSaleOrderBuffetDelayProductRecord(obj model.SaleOrderBuffetDelayProduct) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.WithMessage(errors.New("销售订单自助餐延迟商品UUID或ID不能为0"))
	}
	return r.db.Model(&model.SaleOrderBuffetDelayProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
}

func (r *saleOrderBuffetDelayProductRepo) UpdateOrCreateSaleOrderBuffetDelayProductRecord(obj model.SaleOrderBuffetDelayProduct) error {
	// 如果主键id为0则create，否则update
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return r.CreateSaleOrderBuffetDelayProductRecord(obj)
	}
	return r.UpdateSaleOrderBuffetDelayProductRecord(obj)
}

func (r *saleOrderBuffetDelayProductRepo) GetSaleOrderBuffetDelayProducts(opts ...DBOption) ([]*model.SaleOrderBuffetDelayProduct, error) {
	var buffetDelayProducts []*model.SaleOrderBuffetDelayProduct
	db := r.db.Model(&model.SaleOrderBuffetDelayProduct{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&buffetDelayProducts).Error
	return buffetDelayProducts, err
}
