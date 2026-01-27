package entity

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
)

// Warehouse 仓库聚合根
type Warehouse struct {
	uuid            uint64
	name            valueobject.MultiLanguageName
	warehouseType   valueobject.WarehouseType
	code            valueobject.WarehouseCode
	status          valueobject.WarehouseStatus
	contactInfo     valueobject.ContactInfo
	isDefault       bool
	erpCode         string
	headquarterUuid uint64
	createTime      int64
	updateTime      int64
	hasItems        bool // 是否有关联的物品
}

// NewWarehouse 创建新仓库
func NewWarehouse(
	name valueobject.MultiLanguageName,
	warehouseType valueobject.WarehouseType,
	code valueobject.WarehouseCode,
	status valueobject.WarehouseStatus,
	contactInfo valueobject.ContactInfo,
) *Warehouse {
	now := time.Now().Unix()
	return &Warehouse{
		name:          name,
		warehouseType: warehouseType,
		code:          code,
		status:        status,
		contactInfo:   contactInfo,
		isDefault:     false,
		createTime:    now,
		updateTime:    now,
	}
}

// ReconstructWarehouse 重建仓库（从数据库加载）
func ReconstructWarehouse(
	uuid uint64,
	name valueobject.MultiLanguageName,
	warehouseType valueobject.WarehouseType,
	code valueobject.WarehouseCode,
	status valueobject.WarehouseStatus,
	contactInfo valueobject.ContactInfo,
	isDefault bool,
	erpCode string,
	headquarterUuid uint64,
	createTime int64,
	updateTime int64,
	hasItems bool,
) *Warehouse {
	return &Warehouse{
		uuid:            uuid,
		name:            name,
		warehouseType:   warehouseType,
		code:            code,
		status:          status,
		contactInfo:     contactInfo,
		isDefault:       isDefault,
		erpCode:         erpCode,
		headquarterUuid: headquarterUuid,
		createTime:      createTime,
		updateTime:      updateTime,
		hasItems:        hasItems,
	}
}

// Getters

// Uuid 获取UUID
func (w *Warehouse) Uuid() uint64 {
	return w.uuid
}

// SetUuid 设置UUID（仅用于持久化后）
func (w *Warehouse) SetUuid(uuid uint64) {
	w.uuid = uuid
}

// Name 获取名称
func (w *Warehouse) Name() valueobject.MultiLanguageName {
	return w.name
}

// WarehouseType 获取仓库类型
func (w *Warehouse) WarehouseType() valueobject.WarehouseType {
	return w.warehouseType
}

// Code 获取编码
func (w *Warehouse) Code() valueobject.WarehouseCode {
	return w.code
}

// Status 获取状态
func (w *Warehouse) Status() valueobject.WarehouseStatus {
	return w.status
}

// ContactInfo 获取联系信息
func (w *Warehouse) ContactInfo() valueobject.ContactInfo {
	return w.contactInfo
}

// IsDefault 是否默认仓库
func (w *Warehouse) IsDefault() bool {
	return w.isDefault
}

// ErpCode 获取ERP编码
func (w *Warehouse) ErpCode() string {
	return w.erpCode
}

// HeadquarterUuid 获取总部UUID
func (w *Warehouse) HeadquarterUuid() uint64 {
	return w.headquarterUuid
}

// CreateTime 获取创建时间
func (w *Warehouse) CreateTime() int64 {
	return w.createTime
}

// UpdateTime 获取更新时间
func (w *Warehouse) UpdateTime() int64 {
	return w.updateTime
}

// HasItems 是否有关联的物品
func (w *Warehouse) HasItems() bool {
	return w.hasItems
}

// SetHasItems 设置是否有关联的物品
func (w *Warehouse) SetHasItems(hasItems bool) {
	w.hasItems = hasItems
}

// 业务方法

// Activate 启用仓库
func (w *Warehouse) Activate() error {
	if w.status.IsActive() {
		return errors.New("仓库已经是启用状态")
	}
	w.status = valueobject.StatusActive
	w.updateTime = time.Now().Unix()
	return nil
}

// Deactivate 禁用仓库
func (w *Warehouse) Deactivate() error {
	if w.status.IsDisabled() {
		return errors.New("仓库已经是禁用状态")
	}

	// 业务规则：默认仓库不能禁用
	if w.isDefault {
		return errors.New("默认仓库不可禁用")
	}

	// 业务规则：在途仓库不能禁用
	if w.warehouseType.IsTransit() {
		return errors.New("在途仓库不可禁用")
	}

	w.status = valueobject.StatusDisabled
	w.updateTime = time.Now().Unix()
	return nil
}

// SetAsDefault 设置为默认仓库
func (w *Warehouse) SetAsDefault() error {
	if w.isDefault {
		return errors.New("仓库已经是默认仓库")
	}

	// 业务规则：在途仓库不能设置为默认仓库
	if w.warehouseType.IsTransit() {
		return errors.New("在途仓库不可设置为默认仓库")
	}

	w.isDefault = true
	w.updateTime = time.Now().Unix()
	return nil
}

// UnsetDefault 取消默认仓库
func (w *Warehouse) UnsetDefault() {
	w.isDefault = false
	w.updateTime = time.Now().Unix()
}

// UpdateContactInfo 更新联系信息
func (w *Warehouse) UpdateContactInfo(info valueobject.ContactInfo) error {
	if w.contactInfo.Equals(info) {
		return nil // 没有变化
	}
	w.contactInfo = info
	w.updateTime = time.Now().Unix()
	return nil
}

// UpdateName 更新名称
func (w *Warehouse) UpdateName(name valueobject.MultiLanguageName) {
	w.name = name
	w.updateTime = time.Now().Unix()
}

// UpdateType 更新仓库类型
func (w *Warehouse) UpdateType(warehouseType valueobject.WarehouseType) error {
	// 业务规则：在途仓库不能修改类型
	if w.warehouseType.IsTransit() && !warehouseType.IsTransit() {
		return errors.New("在途仓库不可修改类型")
	}

	w.warehouseType = warehouseType
	w.updateTime = time.Now().Unix()
	return nil
}

// UpdateStatus 更新状态
func (w *Warehouse) UpdateStatus(status valueobject.WarehouseStatus) error {
	if status.IsDisabled() {
		return w.Deactivate()
	}
	return w.Activate()
}

// SetErpCode 设置ERP编码
func (w *Warehouse) SetErpCode(erpCode string) {
	w.erpCode = erpCode
	w.updateTime = time.Now().Unix()
}

// SetHeadquarterUuid 设置总部UUID
func (w *Warehouse) SetHeadquarterUuid(headquarterUuid uint64) {
	w.headquarterUuid = headquarterUuid
}

// CanBeDeleted 是否可以删除
func (w *Warehouse) CanBeDeleted() error {
	// 业务规则：默认仓库不能删除
	if w.isDefault {
		return errors.New("默认仓库不可删除")
	}

	// 业务规则：在途仓库不能删除
	if w.warehouseType.IsTransit() {
		return errors.New("在途仓库不可删除")
	}

	// 业务规则：有关联物品的仓库不能删除
	if w.hasItems {
		return errors.New("仓库存在关联的物品，不可删除")
	}

	return nil
}

// CanBeEdited 是否可以编辑
func (w *Warehouse) CanBeEdited(currentCompanyUuid uint64) error {
	// 业务规则：总部仓库在子店不可编辑
	if w.headquarterUuid > 0 && w.headquarterUuid != currentCompanyUuid {
		return errors.New("仓库不可编辑")
	}
	return nil
}

// IsHeadquarter 是否为总部仓库
func (w *Warehouse) IsHeadquarter() bool {
	return w.headquarterUuid > 0
}

// IsTransit 是否为在途仓库
func (w *Warehouse) IsTransit() bool {
	return w.warehouseType.String() == constant.WarehouseTypeTransit
}

// IsNormal 是否为普通仓库
func (w *Warehouse) IsNormal() bool {
	return w.warehouseType.String() == constant.WarehouseTypeNormal
}
