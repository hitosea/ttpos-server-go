package integration

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/service"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// ErpWarehouseService ERP仓库集成服务实现
type ErpWarehouseService struct {
	dbm *database.DBManager
}

// NewErpWarehouseService 创建ERP仓库集成服务
func NewErpWarehouseService(dbm *database.DBManager) service.IErpIntegrationService {
	return &ErpWarehouseService{
		dbm: dbm,
	}
}

// CreateWarehouse 在ERP中创建仓库
func (s *ErpWarehouseService) CreateWarehouse(
	ctx context.Context,
	warehouse *entity.Warehouse,
	companySetting service.ErpCompanySetting,
) (string, error) {
	erpSrv := erp.NewIErpSrv(s.dbm)

	// 获取仓库名称（英文）
	warehouseName := warehouse.Name().EN()
	if warehouseName == "" {
		warehouseName = warehouse.Name().ZH()
	}

	// 调用ERP接口（需要标准 context）
	erpCode, err := erpSrv.CreateWarehouse(ctx.GetContext(), req.CreateErpnextWarehouseReq{
		SiteCode:      companySetting.SiteCode,
		WarehouseName: warehouseName,
		AliasName:     warehouseName,
		CompanyAbbr:   companySetting.CompanyAbbr,
		Branch:        companySetting.BranchName,
		Disabled:      warehouse.Status().IsDisabled(),
		WarehouseType: s.convertWarehouseType(warehouse.WarehouseType().String()),
	})

	if err != nil {
		return "", errors.WithMessage(err, "创建ERP仓库失败")
	}

	return erpCode, nil
}

// UpdateWarehouse 在ERP中更新仓库
func (s *ErpWarehouseService) UpdateWarehouse(
	ctx context.Context,
	warehouse *entity.Warehouse,
	companySetting service.ErpCompanySetting,
) error {
	if warehouse.ErpCode() == "" {
		return nil // 没有ERP编码，无需更新
	}

	erpSrv := erp.NewIErpSrv(s.dbm)

	// 调用ERP接口（需要标准 context）
	err := erpSrv.UpdateWarehouse(ctx.GetContext(), req.UpdateErpnextWarehouseReq{
		CreateErpnextWarehouseReq: req.CreateErpnextWarehouseReq{
			SiteCode:      companySetting.SiteCode,
			CompanyAbbr:   companySetting.CompanyAbbr,
			Branch:        companySetting.BranchName,
			Disabled:      warehouse.Status().IsDisabled(),
			WarehouseType: s.convertWarehouseType(warehouse.WarehouseType().String()),
		},
		Name: warehouse.ErpCode(),
	})

	if err != nil {
		return errors.WithMessage(err, "更新ERP仓库失败")
	}

	return nil
}

// DeleteWarehouse 在ERP中删除仓库
func (s *ErpWarehouseService) DeleteWarehouse(
	ctx context.Context,
	erpCode string,
	companySetting service.ErpCompanySetting,
) error {
	if erpCode == "" {
		return nil // 没有ERP编码，无需删除
	}

	erpSrv := erp.NewIErpSrv(s.dbm)

	// 调用ERP接口（实际上是禁用，需要标准 context）
	err := erpSrv.DeleteWarehouse(ctx.GetContext(), req.DeleteErpnextWarehouseReq{
		SiteCode: companySetting.SiteCode,
		Name:     erpCode,
	})

	if err != nil {
		return errors.WithMessage(err, "删除ERP仓库失败")
	}

	return nil
}

// FetchWarehouses 从ERP获取仓库列表
func (s *ErpWarehouseService) FetchWarehouses(
	ctx context.Context,
	companySetting service.ErpCompanySetting,
) ([]*service.ErpWarehouseDTO, error) {
	erpSrv := erp.NewIErpSrv(s.dbm)

	// 调用ERP接口（需要标准 context）
	warehouseList, err := erpSrv.GetWarehouseList(ctx.GetContext(), req.GetErpnextWarehouseListReq{
		SiteCode:    companySetting.SiteCode,
		CompanyAbbr: companySetting.CompanyAbbr,
		Branch:      companySetting.BranchName,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "获取ERP仓库列表失败")
	}

	// 转换为DTO
	dtos := make([]*service.ErpWarehouseDTO, 0, len(warehouseList))
	for _, w := range warehouseList {
		dtos = append(dtos, &service.ErpWarehouseDTO{
			Name:          w.Name,
			AliasName:     w.AliasName,
			WarehouseType: w.WarehouseType,
			Disabled:      w.Disabled,
		})
	}

	return dtos, nil
}

// convertWarehouseType 转换仓库类型为ERP格式
func (s *ErpWarehouseService) convertWarehouseType(warehouseType string) string {
	switch warehouseType {
	case "normal":
		return "Normal"
	case "transit":
		return "Transit"
	default:
		return "Normal"
	}
}
