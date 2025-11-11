package erp

import (
	"context"
	"errors"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/model"

	pkgCtx "ttpos-server-go/pkg/context"

	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpMaterialTransferClient() (material_transfer.MaterialTransferServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return material_transfer.NewMaterialTransferServiceClient(conn), conn, nil
}

// SaveMaterialTransfer 保存调拨单请求
func (s *erpSrv) SaveMaterialTransfer(ctx pkgCtx.Context, companySetting model.CompanySetting, saveMaterialTransferReq *material_transfer.MaterialTransferReq) (*material_transfer.MaterialTransferResp, error) {
	client, conn, err := NewErpMaterialTransferClient()
	if err != nil {
		return &material_transfer.MaterialTransferResp{}, err
	}
	defer conn.Close()

	result, err := client.MaterialTransfer(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), saveMaterialTransferReq)
	if err != nil {
		return &material_transfer.MaterialTransferResp{}, err
	}
	if result.Code != "0" {
		logger.Logger.Error("SaveTransferOrder-SaveTransferOrder", zap.String("result", result.String()))
		return &material_transfer.MaterialTransferResp{}, errors.New("调用erp接口失败-10001-" + result.GetMessage())
	}
	if result.Data != nil {
		var resp material_transfer.MaterialTransferResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SaveTransferOrder-UnmarshalTo", zap.Any("err", err))
			return &material_transfer.MaterialTransferResp{}, err
		}
		return &resp, nil
	}
	return &material_transfer.MaterialTransferResp{}, nil
}
