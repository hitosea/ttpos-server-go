package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITransferOrderItemUnitRepo 调拨单明细单位Repository接口
type ITransferOrderItemUnitRepo interface {
	Create(unit *model.TransferOrderItemUnit) error
	CreateBatch(units []*model.TransferOrderItemUnit) error
	Update(unit *model.TransferOrderItemUnit) error
	Delete(uuid uint64) error
	DeleteByItemUuid(itemUuid uint64) error
	DeleteByTransferOrderUuid(transferOrderUuid uint64) error
	GetByUuid(uuid uint64) (*model.TransferOrderItemUnit, error)
	GetListByItemUuid(itemUuid uint64) ([]model.TransferOrderItemUnit, error)
	GetListByTransferOrderUuid(transferOrderUuid uint64) ([]model.TransferOrderItemUnit, error)
}

// TransferOrderItemUnitRepoImpl 调拨单明细单位Repository实现
type TransferOrderItemUnitRepoImpl struct {
	db *gorm.DB
}

// NewTransferOrderItemUnitRepo 创建调拨单明细单位Repository
func NewTransferOrderItemUnitRepo(db *gorm.DB) ITransferOrderItemUnitRepo {
	return &TransferOrderItemUnitRepoImpl{db: db}
}

func (r *TransferOrderItemUnitRepoImpl) Create(unit *model.TransferOrderItemUnit) error {
	return r.db.Create(unit).Error
}

func (r *TransferOrderItemUnitRepoImpl) CreateBatch(units []*model.TransferOrderItemUnit) error {
	if len(units) == 0 {
		return nil
	}
	return r.db.Create(units).Error
}

func (r *TransferOrderItemUnitRepoImpl) Update(unit *model.TransferOrderItemUnit) error {
	return r.db.Model(&model.TransferOrderItemUnit{}).Where("uuid = ?", unit.Uuid).Updates(unit).Error
}

func (r *TransferOrderItemUnitRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.TransferOrderItemUnit{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

func (r *TransferOrderItemUnitRepoImpl) DeleteByItemUuid(itemUuid uint64) error {
	return r.db.Model(&model.TransferOrderItemUnit{}).Where("item_uuid = ?", itemUuid).Update("delete_time", time.Now().Unix()).Error
}

func (r *TransferOrderItemUnitRepoImpl) DeleteByTransferOrderUuid(transferOrderUuid uint64) error {
	return r.db.Model(&model.TransferOrderItemUnit{}).Where("transfer_order_uuid = ?", transferOrderUuid).Update("delete_time", time.Now().Unix()).Error
}

func (r *TransferOrderItemUnitRepoImpl) GetByUuid(uuid uint64) (*model.TransferOrderItemUnit, error) {
	var unit model.TransferOrderItemUnit
	err := r.db.Model(&model.TransferOrderItemUnit{}).Where("uuid = ?", uuid).Where("delete_time = ?", constant.NotDeleted).First(&unit).Error
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *TransferOrderItemUnitRepoImpl) GetListByItemUuid(itemUuid uint64) ([]model.TransferOrderItemUnit, error) {
	var units []model.TransferOrderItemUnit
	err := r.db.Model(&model.TransferOrderItemUnit{}).Where("item_uuid = ?", itemUuid).Where("delete_time = ?", constant.NotDeleted).Find(&units).Error
	return units, err
}

func (r *TransferOrderItemUnitRepoImpl) GetListByTransferOrderUuid(transferOrderUuid uint64) ([]model.TransferOrderItemUnit, error) {
	var units []model.TransferOrderItemUnit
	err := r.db.Model(&model.TransferOrderItemUnit{}).Where("transfer_order_uuid = ?", transferOrderUuid).Where("delete_time = ?", constant.NotDeleted).Find(&units).Error
	return units, err
}
