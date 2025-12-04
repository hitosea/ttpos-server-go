package entity

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
)

// WarehouseItem 库存物品聚合根
type WarehouseItem struct {
	uuid          uint64
	warehouseUuid uint64            // 所属仓库UUID
	materialUuid  uint64            // 物品/商品UUID
	materialCode  string            // 物品编码
	stock         valueobject.Stock // 库存数量
	reservedStock valueobject.Stock // 预留库存数量
	valuation     float64           // 估值单价
	createTime    int64
	updateTime    int64
}

// NewWarehouseItem 创建新的库存物品
func NewWarehouseItem(
	warehouseUuid uint64,
	materialUuid uint64,
	materialCode string,
	valuation float64,
) *WarehouseItem {
	now := time.Now().Unix()
	return &WarehouseItem{
		warehouseUuid: warehouseUuid,
		materialUuid:  materialUuid,
		materialCode:  materialCode,
		stock:         valueobject.NewStock(0),
		reservedStock: valueobject.NewStock(0),
		valuation:     valuation,
		createTime:    now,
		updateTime:    now,
	}
}

// ReconstructWarehouseItem 重建库存物品（从数据库加载）
func ReconstructWarehouseItem(
	uuid uint64,
	warehouseUuid uint64,
	materialUuid uint64,
	materialCode string,
	stock float64,
	reservedStock float64,
	valuation float64,
	createTime int64,
	updateTime int64,
) *WarehouseItem {
	return &WarehouseItem{
		uuid:          uuid,
		warehouseUuid: warehouseUuid,
		materialUuid:  materialUuid,
		materialCode:  materialCode,
		stock:         valueobject.NewStock(stock),
		reservedStock: valueobject.NewStock(reservedStock),
		valuation:     valuation,
		createTime:    createTime,
		updateTime:    updateTime,
	}
}

// Getters

// Uuid 获取UUID
func (w *WarehouseItem) Uuid() uint64 {
	return w.uuid
}

// SetUuid 设置UUID（仅用于持久化后）
func (w *WarehouseItem) SetUuid(uuid uint64) {
	w.uuid = uuid
}

// WarehouseUuid 获取仓库UUID
func (w *WarehouseItem) WarehouseUuid() uint64 {
	return w.warehouseUuid
}

// MaterialUuid 获取物品UUID
func (w *WarehouseItem) MaterialUuid() uint64 {
	return w.materialUuid
}

// MaterialCode 获取物品编码
func (w *WarehouseItem) MaterialCode() string {
	return w.materialCode
}

// Stock 获取库存数量
func (w *WarehouseItem) Stock() valueobject.Stock {
	return w.stock
}

// ReservedStock 获取预留库存数量
func (w *WarehouseItem) ReservedStock() valueobject.Stock {
	return w.reservedStock
}

// Valuation 获取估值单价
func (w *WarehouseItem) Valuation() float64 {
	return w.valuation
}

// CreateTime 获取创建时间
func (w *WarehouseItem) CreateTime() int64 {
	return w.createTime
}

// UpdateTime 获取更新时间
func (w *WarehouseItem) UpdateTime() int64 {
	return w.updateTime
}

// 业务方法

// AvailableStock 获取可用库存（库存 - 预留库存）
func (w *WarehouseItem) AvailableStock() valueobject.Stock {
	return valueobject.NewStock(w.stock.Value() - w.reservedStock.Value())
}

// AddStock 增加库存
func (w *WarehouseItem) AddStock(quantity float64) error {
	if quantity < 0 {
		return errors.New("增加的库存数量不能为负数")
	}

	w.stock = valueobject.NewStock(w.stock.Value() + quantity)
	w.updateTime = time.Now().Unix()
	return nil
}

// ReduceStock 减少库存
func (w *WarehouseItem) ReduceStock(quantity float64) error {
	if quantity < 0 {
		return errors.New("减少的库存数量不能为负数")
	}

	// 业务规则：库存不能为负
	if w.stock.Value()-quantity < 0 {
		return errors.New("库存不足")
	}

	w.stock = valueobject.NewStock(w.stock.Value() - quantity)
	w.updateTime = time.Now().Unix()
	return nil
}

// ReserveStock 预留库存
func (w *WarehouseItem) ReserveStock(quantity float64) error {
	if quantity < 0 {
		return errors.New("预留的库存数量不能为负数")
	}

	// 业务规则：预留库存不能超过可用库存
	if w.AvailableStock().Value() < quantity {
		return errors.New("可用库存不足，无法预留")
	}

	w.reservedStock = valueobject.NewStock(w.reservedStock.Value() + quantity)
	w.updateTime = time.Now().Unix()
	return nil
}

// ReleaseReservedStock 释放预留库存
func (w *WarehouseItem) ReleaseReservedStock(quantity float64) error {
	if quantity < 0 {
		return errors.New("释放的库存数量不能为负数")
	}

	// 业务规则：释放的预留库存不能超过已预留的库存
	if w.reservedStock.Value() < quantity {
		return errors.New("释放的预留库存不能超过已预留的库存")
	}

	w.reservedStock = valueobject.NewStock(w.reservedStock.Value() - quantity)
	w.updateTime = time.Now().Unix()
	return nil
}

// ConsumeReservedStock 消耗预留库存（同时减少库存和预留库存）
func (w *WarehouseItem) ConsumeReservedStock(quantity float64) error {
	if quantity < 0 {
		return errors.New("消耗的库存数量不能为负数")
	}

	// 业务规则：消耗的库存不能超过已预留的库存
	if w.reservedStock.Value() < quantity {
		return errors.New("消耗的库存不能超过已预留的库存")
	}

	// 业务规则：库存不能为负
	if w.stock.Value() < quantity {
		return errors.New("库存不足")
	}

	w.stock = valueobject.NewStock(w.stock.Value() - quantity)
	w.reservedStock = valueobject.NewStock(w.reservedStock.Value() - quantity)
	w.updateTime = time.Now().Unix()
	return nil
}

// UpdateValuation 更新估值单价
func (w *WarehouseItem) UpdateValuation(valuation float64) error {
	if valuation < 0 {
		return errors.New("估值单价不能为负数")
	}

	w.valuation = valuation
	w.updateTime = time.Now().Unix()
	return nil
}

// SetStock 设置库存（用于同步场景）
func (w *WarehouseItem) SetStock(stock float64) {
	w.stock = valueobject.NewStock(stock)
	w.updateTime = time.Now().Unix()
}

// SetReservedStock 设置预留库存（用于同步场景）
func (w *WarehouseItem) SetReservedStock(reservedStock float64) {
	w.reservedStock = valueobject.NewStock(reservedStock)
	w.updateTime = time.Now().Unix()
}

// HasStock 是否有库存
func (w *WarehouseItem) HasStock() bool {
	return w.stock.Value() > 0
}

// HasAvailableStock 是否有可用库存
func (w *WarehouseItem) HasAvailableStock() bool {
	return w.AvailableStock().Value() > 0
}

// TotalValue 获取库存总价值
func (w *WarehouseItem) TotalValue() float64 {
	return w.stock.Value() * w.valuation
}
