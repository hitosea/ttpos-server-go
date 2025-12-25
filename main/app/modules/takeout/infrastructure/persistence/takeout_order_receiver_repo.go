package persistence

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

// ITakeoutOrderReceiverRepo 外卖订单收货人信息仓储接口
type ITakeoutOrderReceiverRepo interface {
	// Create 创建收货人信息
	Create(receiver *model.TakeoutOrderReceiver) error
	// Update 更新收货人信息
	Update(receiver *model.TakeoutOrderReceiver) error
	// GetByOrderUuid 根据订单UUID获取收货人信息
	GetByOrderUuid(orderUuid uint64, options ...DBOption) (*model.TakeoutOrderReceiver, error)
	// Delete 删除收货人信息
	Delete(uuid uint64) error
}

// TakeoutOrderReceiverRepoImpl 外卖订单收货人信息仓储实现
type TakeoutOrderReceiverRepoImpl struct {
	db *gorm.DB
}

// NewTakeoutOrderReceiverRepo 创建外卖订单收货人信息仓储
func NewTakeoutOrderReceiverRepo(db *gorm.DB) ITakeoutOrderReceiverRepo {
	return &TakeoutOrderReceiverRepoImpl{db: db}
}

// Create 创建收货人信息
func (r *TakeoutOrderReceiverRepoImpl) Create(receiver *model.TakeoutOrderReceiver) error {
	if receiver == nil {
		return errors.New("收货人信息不能为空")
	}
	return r.db.Create(receiver).Error
}

// Update 更新收货人信息
func (r *TakeoutOrderReceiverRepoImpl) Update(receiver *model.TakeoutOrderReceiver) error {
	if receiver == nil {
		return errors.New("收货人信息不能为空")
	}
	return r.db.Save(receiver).Error
}

// GetByOrderUuid 根据订单UUID获取收货人信息
func (r *TakeoutOrderReceiverRepoImpl) GetByOrderUuid(orderUuid uint64, options ...DBOption) (*model.TakeoutOrderReceiver, error) {
	var receiver model.TakeoutOrderReceiver
	db := r.db.Model(&model.TakeoutOrderReceiver{}).Where("delete_time = ?", 0)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("takeout_order_uuid = ?", orderUuid).First(&receiver).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 返回 nil 表示未找到，而不是错误
		}
		return nil, errors.WithMessage(err, "查询收货人信息失败")
	}

	return &receiver, nil
}

// Delete 删除收货人信息（软删除）
func (r *TakeoutOrderReceiverRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.TakeoutOrderReceiver{}).
		Where("uuid = ?", uuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}
