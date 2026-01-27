package service

import (
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/pkg/context"
)

// IErpIntegrationService ERP集成服务接口（领域层定义，基础设施层实现）
type IErpIntegrationService interface {
	// CreateWarehouse 在ERP中创建仓库
	CreateWarehouse(ctx context.Context, warehouse *entity.Warehouse, companySetting ErpCompanySetting) (erpCode string, err error)

	// UpdateWarehouse 在ERP中更新仓库
	UpdateWarehouse(ctx context.Context, warehouse *entity.Warehouse, companySetting ErpCompanySetting) error

	// DeleteWarehouse 在ERP中删除仓库（禁用）
	DeleteWarehouse(ctx context.Context, erpCode string, companySetting ErpCompanySetting) error

	// FetchWarehouses 从ERP获取仓库列表
	FetchWarehouses(ctx context.Context, companySetting ErpCompanySetting) ([]*ErpWarehouseDTO, error)
}

// ErpCompanySetting ERP公司设置
type ErpCompanySetting struct {
	SiteCode    string
	CompanyAbbr string
	BranchName  string
}

// ErpWarehouseDTO ERP仓库数据传输对象
type ErpWarehouseDTO struct {
	Name          string // ERP仓库编码
	AliasName     string // 别名
	WarehouseType string // 仓库类型
	Disabled      bool   // 是否禁用
}
