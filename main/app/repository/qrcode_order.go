package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IQrcodeOrderRepo 接单
type IQrcodeOrderRepo interface {
	GetQrcodeOrderList(pageNo, pageSize int) ([]model.QrcodeOrder, int64, error)
	GetQrcodeOrderInfo(qrcodeOrderUuid uint64) (model.QrcodeOrder, error)
	UpdateQrcodeOrder(qrcodeOrderUuid uint64, qrcodeOrder model.QrcodeOrder) error
	CreateQrcodeOrder(qrcodeOrder model.QrcodeOrder) (uint64, error)
	DeleteQrcodeOrder(qrcodeOrderUuid uint64) error
	Reject(DeskUuid uint64) error
}

func NewQrcodeOrderRepo(db *gorm.DB) IQrcodeOrderRepo {
	return NewQrcodeOrderRepoImpl(db)
}

// NewProductFlavorRepoImpl 创建新的仓库实现
func NewQrcodeOrderRepoImpl(db *gorm.DB) *QrcodeOrderRepoImpl {
	return &QrcodeOrderRepoImpl{db: db}
}

type QrcodeOrderRepoImpl struct {
	db *gorm.DB
}

// GetQrcodeOrderList 获取接单列表，排除逻辑删除的接单
func (r *QrcodeOrderRepoImpl) GetQrcodeOrderList(pageNo, pageSize int) ([]model.QrcodeOrder, int64, error) {
	var qrcodeOrders []model.QrcodeOrder
	var total int64

	query := r.db.Model(&model.QrcodeOrder{}).Where("delete_time = ?", 0)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&qrcodeOrders).Error
	return qrcodeOrders, total, err
}

// GetQrcodeOrderInfo 获取接单信息
func (r *QrcodeOrderRepoImpl) GetQrcodeOrderInfo(qrcodeOrderUuid uint64) (model.QrcodeOrder, error) {
	var qrcodeOrder model.QrcodeOrder
	if err := r.db.Model(&model.QrcodeOrder{}).Where("uuid = ?", qrcodeOrderUuid).First(&qrcodeOrder).Error; err != nil {
		return model.QrcodeOrder{}, err
	}
	return qrcodeOrder, nil
}

// UpdateQrcodeOrder 更新接单
func (r *QrcodeOrderRepoImpl) UpdateQrcodeOrder(qrcodeOrderUuid uint64, qrcodeOrder model.QrcodeOrder) error {
	if err := r.db.Model(&model.QrcodeOrder{}).Where("uuid = ?", qrcodeOrderUuid).Updates(qrcodeOrder).Error; err != nil {
		return err
	}
	return nil
}

// CreateQrcodeOrder 创建接单
func (r *QrcodeOrderRepoImpl) CreateQrcodeOrder(qrcodeOrder model.QrcodeOrder) (uint64, error) {
	// 创建桌台
	if err := r.db.Create(&qrcodeOrder).Error; err != nil {
		return 0, err
	}
	return qrcodeOrder.Uuid, nil
}

// DeleteQrcodeOrder 软删除接单
func (r *QrcodeOrderRepoImpl) DeleteQrcodeOrder(qrcodeOrderUuid uint64) error {
	return r.db.Model(&model.QrcodeOrder{}).Where("uuid = ?", qrcodeOrderUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// Reject 拒绝接单
func (r *QrcodeOrderRepoImpl) Reject(DeskUuid uint64) error {
	return r.db.Model(&model.QrcodeOrder{}).
		Where("status = ?", 0).
		Where("desk_uuid = ?", DeskUuid).
		Updates(map[string]interface{}{
			"status":      2,
			"handle_time": uint(time.Now().Unix()),
		}).Error
}
