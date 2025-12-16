package valueobject

import "ttpos-server-go/app/constant"

// warehouseTypeValue 内部类型值
type warehouseTypeValue int

const (
	typeNormal  warehouseTypeValue = 0
	typeTransit warehouseTypeValue = 1
)

// WarehouseType 仓库类型值对象（不可变）
type WarehouseType struct {
	value warehouseTypeValue
}

// 预定义的类型常量
var (
	// TypeNormal 普通仓库
	TypeNormal = WarehouseType{value: typeNormal}
	// TypeTransit 在途仓库
	TypeTransit = WarehouseType{value: typeTransit}
)

// NewWarehouseType 从字符串创建仓库类型
func NewWarehouseType(typeStr string) WarehouseType {
	switch typeStr {
	case constant.WarehouseTypeNormal:
		return TypeNormal
	case constant.WarehouseTypeTransit:
		return TypeTransit
	default:
		return TypeNormal
	}
}

// String 转换为字符串
func (t WarehouseType) String() string {
	switch t.value {
	case typeNormal:
		return constant.WarehouseTypeNormal
	case typeTransit:
		return constant.WarehouseTypeTransit
	default:
		return constant.WarehouseTypeNormal
	}
}

// IsNormal 是否为普通仓库
func (t WarehouseType) IsNormal() bool {
	return t.value == typeNormal
}

// IsTransit 是否为在途仓库
func (t WarehouseType) IsTransit() bool {
	return t.value == typeTransit
}

// Equals 比较两个类型是否相等
func (t WarehouseType) Equals(other WarehouseType) bool {
	return t.value == other.value
}
