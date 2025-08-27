package repository

import (
	"math/rand"
	"strconv"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type IWarehouseFormRepo interface {
	IWarehouseFormQueryRepo
	CreateWarehouseOutFormRecord(obj model.WarehouseOutForm) error              // 创建出库单记录
	CreateWarehouseFormRecord(obj model.WarehouseForm) error                    // 创建入库单记录
	CreateWarehouseOutFormItemRecord(obj model.WarehouseOutFormItem) error      // 创建出库单记录
	CreateWarehouseOutFormItemRecords(list []*model.WarehouseOutFormItem) error // 批量创建出库单明细记录
	CreateWarehouseFormItemRecords(list []*model.WarehouseFormItem) error       // 批量创建入库单明细记录
	UpdateWarehouseOutFormItemRecordsStatus(saleOrderUuid uint64) error         // 更新该销售订单的所有出库单记录为已出库
	UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid uint64) error     // 更新该销售账单的所有出库单记录为已扣减库存
	UpdateWarehouseFormItemRecordsAddStock(saleBillUuid uint64) error           // 更新该销售账单的所有入库单记录为已加库存
	UpdateWarehouseOutFormRecord(obj model.WarehouseOutForm) error              // 更新出库单记录
	UpdateWarehouseOutFormItemRecord(obj model.WarehouseOutFormItem) error      // 更新出库单明细记录
	GenerateWarehouseFormNo(timezone string) string                             // 生成入库单编号
	GenerateWarehouseOutFormNo(timezone string) string                          // 生成出库单编号
}

type IWarehouseFormQueryRepo interface {
	GetWarehouseForm(opts ...DBOption) (*model.WarehouseForm, error)                                    // 获取入库单
	GetWarehouseOutForms(opts ...DBOption) ([]*model.WarehouseOutForm, error)                           // 获取出库单
	GetWarehouseOutFormsBySaleBillUuid(saleBillUuid uint64) ([]*model.WarehouseOutForm, error)          // 获取该销售账单的所有出库单记录
	GetWarehouseOutFormItem(opts ...DBOption) ([]*model.WarehouseOutFormItem, error)                    // 获取出库单明细
	GetWarehouseFormItem(opts ...DBOption) ([]*model.WarehouseFormItem, error)                          // 获取入库单明细
	GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid uint64) ([]*model.WarehouseOutFormItem, error) // 获取该销售订单的所有出库单记录
	GetWarehouseOutFormItemBySaleBillUuid(saleBillUuid uint64) ([]*model.WarehouseOutFormItem, error)   // 获取该销售账单的所有出库单记录
	GetWarehouseOutFormItemNotProcessed(saleBillUuid uint64) ([]*model.WarehouseOutFormItem, error)     // 获取该销售账单的所有未减库存的出库单记录
	GetWarehouseFormItemNotProcessed(saleBillUuid uint64) ([]*model.WarehouseFormItem, error)           // 获取该销售账单的所有未加库存的入库单记录
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
		return nil, errors.WithMessage(result.Error)
	}

	return warehouseOutFormItems, nil
}

func (r *warehouseFormRepoImpl) GetWarehouseOutForms(opts ...DBOption) ([]*model.WarehouseOutForm, error) {
	var warehouseOutForms []*model.WarehouseOutForm
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Model(&model.WarehouseOutForm{}).Find(&warehouseOutForms)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}
	return warehouseOutForms, nil
}

func (r *warehouseFormRepoImpl) GetWarehouseOutFormsBySaleBillUuid(saleBillUuid uint64) ([]*model.WarehouseOutForm, error) {
	warehouseOutForms, err := r.GetWarehouseOutForms(
		CommonRepo.WhereByAssociatedOrderUuid(saleBillUuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "WarehouseOutFormItems",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return warehouseOutForms, nil
}

func (r *warehouseFormRepoImpl) GetWarehouseFormItem(opts ...DBOption) ([]*model.WarehouseFormItem, error) {
	var warehouseFormItems []*model.WarehouseFormItem
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Model(&model.WarehouseFormItem{}).Find(&warehouseFormItems)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return warehouseFormItems, nil
}

func (r *warehouseFormRepoImpl) GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid uint64) ([]*model.WarehouseOutFormItem, error) {
	warehouseOutFormItems, err := r.GetWarehouseOutFormItem(
		CommonRepo.WhereBySaleOrderUuid(saleOrderUuid), // 销售订单uuid
		CommonRepo.WhereByNotRevoked(),                 // 未撤销的出库记录
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return warehouseOutFormItems, nil
}

// GetWarehouseOutFormItemBySaleBillUuid 获取销售账单的所有出库单记录
func (r *warehouseFormRepoImpl) GetWarehouseOutFormItemBySaleBillUuid(saleBillUuid uint64) ([]*model.WarehouseOutFormItem, error) {
	warehouseOutFormItems, err := r.GetWarehouseOutFormItem(
		CommonRepo.WhereBySaleBillUuid(saleBillUuid),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return warehouseOutFormItems, nil
}

// GetWarehouseOutFormItemNotProcessed 获取销售账单的所有未减库存的出库单记录
func (r *warehouseFormRepoImpl) GetWarehouseOutFormItemNotProcessed(saleBillUuid uint64) ([]*model.WarehouseOutFormItem, error) {
	warehouseOutFormItems, err := r.GetWarehouseOutFormItem(
		CommonRepo.WhereBySaleBillUuid(saleBillUuid),
		CommonRepo.WhereByReduceStock(constant.WarehouseOutFormItemReduceStockNotProcessed),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductBom.ProductPackage",
			},
			WithPreload{
				Query: "Material",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return warehouseOutFormItems, nil
}

// GetWarehouseFormItemNotProcessed 获取销售账单的所有未加库存的入库单记录
func (r *warehouseFormRepoImpl) GetWarehouseFormItemNotProcessed(saleBillUuid uint64) ([]*model.WarehouseFormItem, error) {
	warehouseFormItems, err := r.GetWarehouseFormItem(
		CommonRepo.WhereBySaleBillUuid(saleBillUuid),
		CommonRepo.WhereByAddStock(constant.WarehouseOutFormItemReduceStockNotProcessed),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductBom",
			},
			WithPreload{
				Query: "Material",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return warehouseFormItems, nil
}

func (r *warehouseFormRepoImpl) CreateWarehouseOutFormRecord(obj model.WarehouseOutForm) error {
	obj.SetNil()
	return r.db.Model(&model.WarehouseOutForm{}).Create(&obj).Error
}

func (r *warehouseFormRepoImpl) CreateWarehouseFormRecord(obj model.WarehouseForm) error {
	obj.SetNil()
	return r.db.Model(&model.WarehouseForm{}).Create(&obj).Error
}

func (r *warehouseFormRepoImpl) CreateWarehouseOutFormItemRecord(obj model.WarehouseOutFormItem) error {
	obj.SetNil()
	return r.db.Model(&model.WarehouseOutFormItem{}).Create(&obj).Error
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

func (r *warehouseFormRepoImpl) CreateWarehouseFormItemRecords(list []*model.WarehouseFormItem) error {
	items := make([]model.WarehouseFormItem, 0)
	for _, item := range list {
		items = append(items, *item)
	}
	for index, _ := range items {
		items[index].SetNil()
	}
	return r.db.Model(&model.WarehouseFormItem{}).Create(items).Error
}

// UpdateWarehouseOutFormItemRecordsStatus 更新该销售订单的所有出库单记录为已出库
func (r *warehouseFormRepoImpl) UpdateWarehouseOutFormItemRecordsStatus(saleOrderUuid uint64) error {
	return r.db.Model(&model.WarehouseOutFormItem{}).Where("sale_order_uuid = ? AND delete_time = ?", saleOrderUuid, constant.NotDeleted).Update("status", constant.WarehouseOutFormItemStatusSuccess).Error
}

// UpdateWarehouseOutFormItemRecordsReduceStock 更新该销售账单的所有出库单记录为已扣减库存
func (r *warehouseFormRepoImpl) UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid uint64) error {
	return r.db.Model(&model.WarehouseOutFormItem{}).Where("sale_bill_uuid = ? AND delete_time = ?", saleBillUuid, constant.NotDeleted).Update("reduce_stock", constant.WarehouseOutFormItemReduceStockSuccess).Error
}

// UpdateWarehouseFormItemRecordsAddStock 更新该销售账单的所有入库单记录为已加库存
func (r *warehouseFormRepoImpl) UpdateWarehouseFormItemRecordsAddStock(saleBillUuid uint64) error {
	return r.db.Model(&model.WarehouseFormItem{}).Where("sale_bill_uuid = ? AND delete_time = ?", saleBillUuid, constant.NotDeleted).Update("add_stock", constant.WarehouseOutFormItemReduceStockSuccess).Error
}

// UpdateWarehouseOutFormRecord 更新出库单记录
func (r *warehouseFormRepoImpl) UpdateWarehouseOutFormRecord(obj model.WarehouseOutForm) error {
	obj.SetNil()
	if obj.Uuid == 0 {
		return errors.WithMessage(errors.New("uuid is required"))
	}
	err := r.db.Model(&model.WarehouseOutForm{}).Where("uuid = ?", obj.Uuid).Updates(&obj).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateWarehouseOutFormItemRecord 更新出库单明细记录
func (r *warehouseFormRepoImpl) UpdateWarehouseOutFormItemRecord(obj model.WarehouseOutFormItem) error {
	obj.SetNil()
	if obj.Uuid == 0 {
		return errors.WithMessage(errors.New("uuid is required"))
	}
	err := r.db.Model(&model.WarehouseOutFormItem{}).Where("uuid = ?", obj.Uuid).Updates(&obj).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// GenerateFormNo 生成入库单编号。入库编号：18位纯数字（前2位WR，2-10位是年月日，中间位是0000，后4位随机生成）
func (r *warehouseFormRepoImpl) GenerateWarehouseFormNo(timezone string) string {
	date := utils.SetTimezone(timezone).Now().Format("20060102")
	rand := rand.Intn(9000) + 1000
	code := "WR" + date + "0000" + strconv.Itoa(rand)
	return code
}

// GenerateWarehouseOutFormNo 生成出库单编号。出库编号：18位纯数字（前2位OO，2-10位是年月日，中间位是0000，后4位随机生成）
func (r *warehouseFormRepoImpl) GenerateWarehouseOutFormNo(timezone string) string {
	date := utils.SetTimezone(timezone).Now().Format("20060102")
	rand := rand.Intn(9000) + 1000
	code := "OO" + date + "0000" + strconv.Itoa(rand)
	return code
}
