package repository

import (
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

type ITakeoutRepo interface {
	GetEnabledTakeouts() ([]takeoutModel.Takeout, error) // 获取已启用的外卖平台
	BatchMarkErpStockDeducted(orderUuids []uint64) error // 批量标记外卖订单ERP库存已扣减
}

func NewTakeoutRepo(db *gorm.DB) ITakeoutRepo {
	return &takeoutRepo{db: db}
}

type takeoutRepo struct {
	db *gorm.DB
}

// GetEnabledTakeouts 获取已启用的外卖平台
func (r *takeoutRepo) GetEnabledTakeouts() ([]takeoutModel.Takeout, error) {
	var takeouts []takeoutModel.Takeout
	err := r.db.Model(&takeoutModel.Takeout{}).Where("enabled = 1").Find(&takeouts).Error
	return takeouts, err
}

// BatchMarkErpStockDeducted 批量标记外卖订单ERP库存已扣减
func (r *takeoutRepo) BatchMarkErpStockDeducted(orderUuids []uint64) error {
	if len(orderUuids) == 0 {
		return nil
	}
	return r.db.Model(&takeoutModel.TakeoutOrder{}).
		Where("uuid IN ?", orderUuids).
		Update("erp_stock_deducted", 1).Error
}
