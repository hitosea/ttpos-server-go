package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderCouponRepo interface {
	CreateSaleOrderCoupon(saleOrderCoupon model.SaleOrderCoupon) error // 创建销售订单优惠券
	UpdateSaleOrderCoupon(saleOrderCoupon model.SaleOrderCoupon) error // 更新销售订单优惠券 更换优惠券或删除优惠券
}

func NewSaleOrderCouponRepo(db *gorm.DB) ISaleOrderCouponRepo {
	return NewSaleOrderCouponRepoImpl(db)
}

type saleOrderCouponRepo struct {
	db *gorm.DB
}

func NewSaleOrderCouponRepoImpl(db *gorm.DB) ISaleOrderCouponRepo {
	return &saleOrderCouponRepo{db: db}
}

func (r *saleOrderCouponRepo) CreateSaleOrderCoupon(saleOrderCoupon model.SaleOrderCoupon) error {
	saleOrderCoupon.SetNil()
	if err := r.db.Create(&saleOrderCoupon).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderCouponRepo) UpdateSaleOrderCoupon(saleOrderCoupon model.SaleOrderCoupon) error {
	saleOrderCoupon.SetNil()
	if err := r.db.Save(&saleOrderCoupon).Error; err != nil {
		return err
	}
	return nil
}
