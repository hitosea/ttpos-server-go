package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IWarehouseFormRepo interface {
	GetWarehouseForm(opts ...DBOption) (*model.WarehouseForm, error)                                    // 获取入库单
	GetWarehouseOutFormItem(opts ...DBOption) ([]*model.WarehouseOutFormItem, error)                    // 获取出库单明细
	GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid uint64) ([]*model.WarehouseOutFormItem, error) // 获取该销售订单的所有出库单记录
	CreateWarehouseOutFormRecord(obj model.WarehouseOutForm) error                                      // 创建出库单记录
	CreateWarehouseOutFormItemRecord(obj model.WarehouseOutFormItem) error                              // 创建出库单记录
	CreateWarehouseOutFormItemRecords(list []*model.WarehouseOutFormItem) error                         // 批量创建出库单明细记录
	UpdateWarehouseOutFormItemRecordsStatus(saleOrderUuid uint64) error                                 // 更新该销售订单的所有出库单记录为已出库
}

type warehouseFormRepoImpl struct {
	db *gorm.DB
}

func NewWarehouseFormRepo(db *gorm.DB) IWarehouseFormRepo {
	return &warehouseFormRepoImpl{db: db}
}

func (r *warehouseFormRepoImpl) GetWarehouseForm(opts ...DBOption) (*model.WarehouseForm, error) {
	var warehouseForm model.WarehouseForm
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&warehouseForm)
	if result.Error != nil {
		return nil, result.Error
	}

	return &warehouseForm, nil
}

func (r *warehouseFormRepoImpl) GetWarehouseOutFormItem(opts ...DBOption) ([]*model.WarehouseOutFormItem, error) {
	var warehouseOutFormItems []*model.WarehouseOutFormItem
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Model(&model.WarehouseOutFormItem{}).Find(&warehouseOutFormItems)
	if result.Error != nil {
		return nil, result.Error
	}

	return warehouseOutFormItems, nil
}

func (r *warehouseFormRepoImpl) GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid uint64) ([]*model.WarehouseOutFormItem, error) {
	warehouseOutFormItems, err := r.GetWarehouseOutFormItem(
		CommonRepo.WhereBySaleOrderUuid(saleOrderUuid),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return warehouseOutFormItems, nil
}

func (r *warehouseFormRepoImpl) CreateWarehouseOutFormRecord(obj model.WarehouseOutForm) error {
	obj.SetNil()
	return r.db.Model(&model.WarehouseOutForm{}).Create(&obj).Error
}

func (r *warehouseFormRepoImpl) CreateWarehouseOutFormItemRecord(obj model.WarehouseOutFormItem) error {
	obj.SetNil()
	return r.db.Model(&model.WarehouseOutFormItem{}).Create(obj).Error
}

func (r *warehouseFormRepoImpl) CreateWarehouseOutFormItemRecords(list []*model.WarehouseOutFormItem) error {
	items := make([]model.WarehouseOutFormItem, 0)
	for _, item := range list {
		items = append(items, *item)
	}
	for index, _ := range items {
		items[index].SetNil()
	}
	return r.db.Model(&model.WarehouseOutFormItem{}).Create(items).Error
}

func (r *warehouseFormRepoImpl) UpdateWarehouseOutFormItemRecordsStatus(saleOrderUuid uint64) error {
	return r.db.Model(&model.WarehouseOutFormItem{}).Where("sale_order_uuid = ? AND delete_time = ?", saleOrderUuid, constant.NotDeleted).Update("status", constant.WarehouseOutFormStatusSuccess).Error
}
