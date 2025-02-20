package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IH5OrderRepo 接单
type IH5OrderRepo interface {
	GetQrcodeOrderList(pageNo, pageSize int) ([]model.H5Order, int64, error)
	GetH5OrderInfoByDeskUuid(qrcodeOrderUuid uint64) (model.H5Order, error)
	UpdateQrcodeOrder(qrcodeOrderUuid uint64, qrcodeOrder model.H5Order) error
	CreateQrcodeOrder(qrcodeOrder model.H5Order) (uint64, error)
	DeleteQrcodeOrder(qrcodeOrderUuid uint64) error
	Reject(DeskUuid uint64) error
}

func NewQrcodeOrderRepo(db *gorm.DB) IH5OrderRepo {
	return NewH5OrderRepoImpl(db)
}

// NewH5OrderRepoImpl 创建新的仓库实现
func NewH5OrderRepoImpl(db *gorm.DB) *H5OrderRepoImpl {
	return &H5OrderRepoImpl{db: db}
}

type H5OrderRepoImpl struct {
	db *gorm.DB
}

// GetQrcodeOrderList 获取接单列表，排除逻辑删除的接单
func (r *H5OrderRepoImpl) GetQrcodeOrderList(pageNo, pageSize int) ([]model.H5Order, int64, error) {
	var qrcodeOrders []model.H5Order
	var total int64

	query := r.db.Model(&model.H5Order{}).Where("delete_time = ?", 0)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&qrcodeOrders).Error
	return qrcodeOrders, total, err
}

// GetH5OrderInfoByDeskUuid 获取接单信息
func (r *H5OrderRepoImpl) GetH5OrderInfoByDeskUuid(deskUuid uint64) (model.H5Order, error) {
	var qrcodeOrder model.H5Order
	if err := r.db.Model(&model.H5Order{}).Preload("H5OrderProducts").Where("desk_uuid = ? AND status = ？ AND delete_time = 0", deskUuid, model.H5OrderStatusUnpaid).First(&qrcodeOrder).Error; err != nil {
		return model.H5Order{}, err
	}
	return qrcodeOrder, nil
}

// UpdateQrcodeOrder 更新接单
func (r *H5OrderRepoImpl) UpdateQrcodeOrder(qrcodeOrderUuid uint64, qrcodeOrder model.H5Order) error {
	if err := r.db.Model(&model.H5Order{}).Where("uuid = ?", qrcodeOrderUuid).Updates(qrcodeOrder).Error; err != nil {
		return err
	}
	return nil
}

// CreateQrcodeOrder 创建接单
func (r *H5OrderRepoImpl) CreateQrcodeOrder(qrcodeOrder model.H5Order) (uint64, error) {
	// 创建桌台
	if err := r.db.Create(&qrcodeOrder).Error; err != nil {
		return 0, err
	}
	return qrcodeOrder.Uuid, nil
}

// DeleteQrcodeOrder 软删除接单
func (r *H5OrderRepoImpl) DeleteQrcodeOrder(qrcodeOrderUuid uint64) error {
	return r.db.Model(&model.H5Order{}).Where("uuid = ?", qrcodeOrderUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// Reject 拒绝接单
func (r *H5OrderRepoImpl) Reject(DeskUuid uint64) error {
	return r.db.Model(&model.H5Order{}).
		Where("status = ?", 0).
		Where("desk_uuid = ?", DeskUuid).
		Updates(map[string]interface{}{
			"status":      2,
			"handle_time": uint(time.Now().Unix()),
		}).Error
}
