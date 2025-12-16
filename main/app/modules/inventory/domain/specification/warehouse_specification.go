package specification

import (
	"strings"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
)

// WarehouseQuerySpec 仓库查询规格
type WarehouseQuerySpec struct {
	keyword         string                       // 关键字（名称或编码）
	warehouseType   *valueobject.WarehouseType   // 仓库类型
	status          *valueobject.WarehouseStatus // 状态
	isHeadquarter   *bool                        // 是否总部仓库
	headquarterUuid *uint64                      // 总部UUID
}

// NewWarehouseQuerySpec 创建查询规格
func NewWarehouseQuerySpec() *WarehouseQuerySpec {
	return &WarehouseQuerySpec{}
}

// WithKeyword 设置关键字
func (s *WarehouseQuerySpec) WithKeyword(keyword string) *WarehouseQuerySpec {
	s.keyword = keyword
	return s
}

// WithType 设置仓库类型
func (s *WarehouseQuerySpec) WithType(warehouseType valueobject.WarehouseType) *WarehouseQuerySpec {
	s.warehouseType = &warehouseType
	return s
}

// WithStatus 设置状态
func (s *WarehouseQuerySpec) WithStatus(status valueobject.WarehouseStatus) *WarehouseQuerySpec {
	s.status = &status
	return s
}

// WithIsHeadquarter 设置是否总部仓库
func (s *WarehouseQuerySpec) WithIsHeadquarter(isHeadquarter bool) *WarehouseQuerySpec {
	s.isHeadquarter = &isHeadquarter
	return s
}

// WithHeadquarterUuid 设置总部UUID
func (s *WarehouseQuerySpec) WithHeadquarterUuid(headquarterUuid uint64) *WarehouseQuerySpec {
	s.headquarterUuid = &headquarterUuid
	return s
}

// IsSatisfiedBy 判断仓库是否满足规格
func (s *WarehouseQuerySpec) IsSatisfiedBy(warehouse *entity.Warehouse) bool {
	// 关键字匹配
	if s.keyword != "" {
		nameMatch := strings.Contains(
			strings.ToLower(warehouse.Name().ZH()),
			strings.ToLower(s.keyword),
		)
		codeMatch := strings.Contains(
			strings.ToLower(warehouse.Code().Value()),
			strings.ToLower(s.keyword),
		)
		if !nameMatch && !codeMatch {
			return false
		}
	}

	// 类型匹配
	if s.warehouseType != nil && warehouse.WarehouseType() != *s.warehouseType {
		return false
	}

	// 状态匹配
	if s.status != nil && warehouse.Status() != *s.status {
		return false
	}

	// 总部仓库匹配
	if s.isHeadquarter != nil {
		if *s.isHeadquarter && !warehouse.IsHeadquarter() {
			return false
		}
		if !*s.isHeadquarter && warehouse.IsHeadquarter() {
			return false
		}
	}

	// 总部UUID匹配
	if s.headquarterUuid != nil && warehouse.HeadquarterUuid() != *s.headquarterUuid {
		return false
	}

	return true
}

// Keyword 获取关键字
func (s *WarehouseQuerySpec) Keyword() string {
	return s.keyword
}

// WarehouseType 获取仓库类型
func (s *WarehouseQuerySpec) WarehouseType() *valueobject.WarehouseType {
	return s.warehouseType
}

// Status 获取状态
func (s *WarehouseQuerySpec) Status() *valueobject.WarehouseStatus {
	return s.status
}

// IsHeadquarter 获取是否总部仓库
func (s *WarehouseQuerySpec) IsHeadquarter() *bool {
	return s.isHeadquarter
}

// HeadquarterUuid 获取总部UUID
func (s *WarehouseQuerySpec) HeadquarterUuid() *uint64 {
	return s.headquarterUuid
}


