package erp

import (
	"context"
	"errors"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpWarehouseClient() (warehouse.WarehouseServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return warehouse.NewWarehouseServiceClient(conn), conn, nil
}

// CreateWarehouse 创建仓库
func (s *erpSrv) CreateWarehouse(ctx context.Context, createWarehouseReq req.CreateErpnextWarehouseReq) (string, error) {
	client, conn, err := NewErpWarehouseClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	result, err := client.CreateWarehouse(WithSiteCode(context.Background(), createWarehouseReq.SiteCode), &warehouse.WarehouseInfo{
		WarehouseName: createWarehouseReq.WarehouseName,
		AliasName:     createWarehouseReq.WarehouseName,
		CompanyAbbr:   createWarehouseReq.CompanyAbbr,
		Branch:        createWarehouseReq.Branch,
		Disabled:      createWarehouseReq.Disabled,
		WarehouseType: createWarehouseReq.WarehouseType,
	})
	if err != nil {
		return "", err
	}
	if result.Code != "0" {
		return "", errors.New(result.Message)
	}
	if result.Data != nil {
		var resp warehouse.WarehouseInfo
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("CreateWarehouse-UnmarshalTo", zap.Any("err", err))
			return "", err
		}
		return resp.Name, nil
	}
	return "", nil
}

// UpdateWarehouse 更新仓库
func (s *erpSrv) UpdateWarehouse(ctx context.Context, updateWarehouseReq req.UpdateErpnextWarehouseReq) error {
	client, conn, err := NewErpWarehouseClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := client.UpdateWarehouse(WithSiteCode(context.Background(), updateWarehouseReq.SiteCode), &warehouse.UpdateWarehouseReq{
		Warehouse: &warehouse.WarehouseInfo{
			WarehouseName: updateWarehouseReq.WarehouseName,
			AliasName:     updateWarehouseReq.WarehouseName,
			CompanyAbbr:   updateWarehouseReq.CompanyAbbr,
			Branch:        updateWarehouseReq.Branch,
			Disabled:      updateWarehouseReq.Disabled,
			Name:          updateWarehouseReq.Name,
		},
	})
	if err != nil {
		return err
	}

	if result.Code != "0" {
		return errors.New(result.Message)
	}
	if result.Data != nil {
		var resp warehouse.UpdateWarehouseResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("UpdateWarehouse-UnmarshalTo", zap.Any("err", err))
			return err
		}
		return nil
	}

	return nil
}

// DeleteWarehouse 删除仓库
func (s *erpSrv) DeleteWarehouse(ctx context.Context, deleteWarehouseReq req.DeleteErpnextWarehouseReq) error {
	client, conn, err := NewErpWarehouseClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := client.DeleteWarehouse(WithSiteCode(context.Background(), deleteWarehouseReq.SiteCode), &warehouse.DeleteWarehouseReq{
		Name: deleteWarehouseReq.Name,
	})
	if err != nil {
		return err
	}
	if result.Code != "0" {
		return errors.New(result.Message)
	}
	if result.Data != nil {
		var resp warehouse.DeleteWarehouseResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("DeleteWarehouse-UnmarshalTo", zap.Any("err", err))
			return err
		}
		return nil
	}
	return nil
}

func (s *erpSrv) GetWarehouseList(ctx context.Context, getWarehouseListReq req.GetErpnextWarehouseListReq) ([]*warehouse.WarehouseInfo, error) {
	client, conn, err := NewErpWarehouseClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	result, err := client.GetWarehouseList(WithSiteCode(context.Background(), getWarehouseListReq.SiteCode), &warehouse.GetWarehouseListReq{
		CompanyAbbr: getWarehouseListReq.CompanyAbbr,
		Branch:      getWarehouseListReq.Branch,
	})
	if err != nil {
		return nil, err
	}
	if result.Code != "0" {
		return nil, errors.New(result.Message)
	}
	if result.Data != nil {
		var listResp warehouse.GetWarehouseListResp
		if err := result.Data.UnmarshalTo(&listResp); err != nil {
			logger.Logger.Error("GetWarehouseList-UnmarshalTo", zap.Any("err", err))
			return nil, err
		}
		return listResp.WarehouseList, nil
	}
	return nil, nil
}
