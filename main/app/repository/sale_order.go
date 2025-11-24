package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

type ISaleOrderRepo interface {
	ISaleOrderQueryRepo
	UpdateSaleOrderCashier(ctx context.Context, saleOrderUuid uint64, cashierUuid uint64, cashierName string) error
	UpdateSaleOrderRecord(obj model.SaleOrder) error
	UpdateSaleOrderPointsExchange(saleOrderUuid uint64, payPoints float64, payPointsAmount float64, pointsExchangeRate float64, autoPointsExchange uint) error // 更新销售订单的积分抵扣信息
	SetCheckoutZeroRuleCancel(saleOrderUuid uint64) error                                                                                                      // 取消结账抹零
	CreateSaleOrderRecord(obj model.SaleOrder) error
	UpdateOrCreateSaleOrderRecord(obj model.SaleOrder) error
	UpdateSaleOrderSoftDeleteByUuid(uuid uint64) error
	DeleteSaleOrder(saleOrderUuid uint64) error
	UpdateSaleOrderErpInvoice(saleOrderUuid uint64, productsInvoiceName string, materialInvoiceName string) error
	UpdateSaleOrderActivity(saleOrderUuid uint64, fullReductionActivityUuid uint64, fullReductionActivityMessage string, activityAmount float64, autoPointsExchange uint) error // 更新销售订单的满减活动信息
}

// ISaleOrderQueryRepo 销售账单查询
type ISaleOrderQueryRepo interface {
	GetSaleOrder(opts ...DBOption) (model.SaleOrder, error)
	GetSaleOrderByUuid(uuid uint64) (*model.SaleOrder, error)
	GetSaleOrderMemberUuid(saleOrderUuid uint64) (uint64, error)
}

type saleOrderRepo struct {
	db *gorm.DB
}

func NewSaleOrderRepo(db *gorm.DB) ISaleOrderRepo {
	return &saleOrderRepo{db: db}
}

func (r *saleOrderRepo) GetSaleOrder(opts ...DBOption) (model.SaleOrder, error) {
	var saleOrder model.SaleOrder
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&saleOrder)
	if result.Error != nil {
		return saleOrder, result.Error
	}

	return saleOrder, nil
}

func (r *saleOrderRepo) GetSaleOrderByUuid(uuid uint64) (*model.SaleOrder, error) {
	order, err := r.GetSaleOrder(CommonRepo.WhereByUuid(uuid))
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &order, nil
}

func (r *saleOrderRepo) GetSaleOrderMemberUuid(saleOrderUuid uint64) (uint64, error) {
	var memberUuid uint64
	err := r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Select("consumer_uuid").Scan(&memberUuid).Error
	return memberUuid, errors.WithMessage(err)
}

func (r *saleOrderRepo) UpdateSaleOrderRecord(obj model.SaleOrder) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.New("销售订单没有主键")
	}
	return r.db.Model(&model.SaleOrder{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(obj).Error
}

func (r *saleOrderRepo) UpdateSaleOrderCashier(ctx context.Context, saleOrderUuid uint64, cashierUuid uint64, cashierName string) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
		"cashier_uuid": cashierUuid,
		"cashier_name": cashierName,
	}).Error
}

func (r *saleOrderRepo) CreateSaleOrderRecord(obj model.SaleOrder) error {
	obj.SetNil()
	return r.db.Model(&model.SaleOrder{}).Create(obj).Error
}

func (r *saleOrderRepo) UpdateOrCreateSaleOrderRecord(obj model.SaleOrder) error {
	// 如果主键id为0则create，否则update
	obj.SetNil()
	if obj.ID == 0 {
		return r.CreateSaleOrderRecord(obj)
	}
	return r.UpdateSaleOrderRecord(obj)
}

func (r *saleOrderRepo) UpdateSaleOrderSoftDeleteByUuid(uuid uint64) error {
	return r.db.Model(&model.SaleOrder{}).Select("delete_time").Where("uuid = ?", uuid).Updates(model.SaleOrder{BaseModel: model.BaseModel{DeleteTime: time.Now().Unix()}}).Error
}

func (r *saleOrderRepo) DeleteSaleOrder(saleOrderUuid uint64) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Update("delete_time", time.Now().Unix()).Error
}

func (r *saleOrderRepo) UpdateSaleOrderPointsExchange(saleOrderUuid uint64, payPoints float64, payPointsAmount float64, pointsExchangeRate float64, autoPointsExchange uint) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
		"pay_points":           payPoints,
		"pay_points_amount":    payPointsAmount,
		"points_exchange_rate": pointsExchangeRate,
		"auto_points_exchange": autoPointsExchange, // 手动改过积分后，订单扣变为手动抵扣
	}).Error
}

func (r *saleOrderRepo) SetCheckoutZeroRuleCancel(saleOrderUuid uint64) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
		"zero_checkout_rule": constant.SaleBillSettingCheckoutZeroingMethodNone,
		"zero_checkout_fee":  0,
	}).Error
}

func (r *saleOrderRepo) UpdateSaleOrderErpInvoice(saleOrderUuid uint64, productsInvoiceName string, materialInvoiceName string) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
		"erp_products_invoice_name": productsInvoiceName,
		"erp_material_invoice_name": materialInvoiceName,
	}).Error
}

func (r *saleOrderRepo) UpdateSaleOrderActivity(saleOrderUuid uint64, fullReductionActivityUuid uint64, fullReductionActivityMessage string, activityAmount float64, autoPointsExchange uint) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
		"full_reduction_activity_uuid":    fullReductionActivityUuid,
		"full_reduction_activity_message": fullReductionActivityMessage,
		"activity_amount":                 activityAmount,
		"auto_points_exchange":            autoPointsExchange,
	}).Error
}
