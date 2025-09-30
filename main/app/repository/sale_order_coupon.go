package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderCouponRepo interface {
	CreateSaleOrderCoupon(saleOrderCoupon model.SaleOrderCoupon) error            // 创建销售订单优惠券
	UpdateSaleOrderCoupon(saleOrderCoupon model.SaleOrderCoupon) error            // 更新销售订单优惠券 更换优惠券或删除优惠券
	UpdateSaleOrderMemberDiscountCancel(saleOrderUuid uint64) error               // 当订单取消会员时，删除销售订单中已经选中的会员优惠券
	UpdateSaleOrderCouponCancelAll(saleOrderUuid uint64) error                    // 删除销售订单中所有优惠券
	UpdateSaleOrderCouponAmount(saleOrderUuid uint64, couponAmount float64) error // 更新销售订单优惠券抵扣金额
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

// 软删除已经选中的会员优惠券
func (r *saleOrderCouponRepo) UpdateSaleOrderMemberDiscountCancel(saleOrderUuid uint64) error {
	if err := r.db.Model(&model.SaleOrderCoupon{}).Where("sale_order_uuid = ? AND coupon_requirement = ?", saleOrderUuid, constant.CouponRequirementMember).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return err
	}
	return nil
}

// 软删除已经选中的会员优惠券
func (r *saleOrderCouponRepo) UpdateSaleOrderCouponCancelAll(saleOrderUuid uint64) error {
	if err := r.db.Model(&model.SaleOrderCoupon{}).Where("sale_order_uuid = ?", saleOrderUuid).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return err
	}
	return nil
}

// 更新销售订单优惠券抵扣金额
func (r *saleOrderCouponRepo) UpdateSaleOrderCouponAmount(saleOrderUuid uint64, couponAmount float64) error {
	if err := r.db.Model(&model.SaleOrderCoupon{}).Where("sale_order_uuid = ?", saleOrderUuid).Update("coupon_amount", couponAmount).Error; err != nil {
		return err
	}
	return nil
}
