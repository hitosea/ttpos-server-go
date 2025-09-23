package erp

import (
	"context"
	"errors"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/cloud"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpStockClient() (stock.StockServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return stock.NewStockServiceClient(conn), conn, nil
}

// SaveMaterialRequestReq 保存物品申请单请求
func (s *erpSrv) SaveMaterialRequest(ctx cc.Context, createPurchaseOrderReq *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &stock.SaveMaterialRequestResp{}, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting
	createPurchaseOrderReq.CompanyAbbr = companySetting.ErpnextCompanyAbbr
	createPurchaseOrderReq.Branch = companySetting.ErpnextBranchName

	result, err := client.SaveMaterialRequest(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), createPurchaseOrderReq)
	if err != nil {
		return &stock.SaveMaterialRequestResp{}, err
	}
	if result.Code != "0" {
		logger.Logger.Error("SaveMaterialRequest-SaveMaterialRequest", zap.Any("err", err))
		return &stock.SaveMaterialRequestResp{}, errors.New("调用erp接口失败 - 001")
	}
	if result.Data != nil {
		var resp stock.SaveMaterialRequestResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SaveMaterialRequest-UnmarshalTo", zap.Any("err", err))
			return &stock.SaveMaterialRequestResp{}, err
		}
		return &resp, nil
	}
	return &stock.SaveMaterialRequestResp{}, nil
}

// GetMaterialRequestList 获取物品申请单列表
func (s *erpSrv) GetMaterialRequestList(ctx cc.Context, getMaterialRequestListReq *stock.GetMaterialRequestListReq) (*stock.GetMaterialRequestListResp, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &stock.GetMaterialRequestListResp{}, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting
	getMaterialRequestListReq.CompanyAbbr = companySetting.ErpnextCompanyAbbr
	getMaterialRequestListReq.Branch = companySetting.ErpnextBranchName

	result, err := client.GetMaterialRequestList(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), getMaterialRequestListReq)
	if err != nil {
		return &stock.GetMaterialRequestListResp{}, err
	}
	if result.Code != "0" {
		return &stock.GetMaterialRequestListResp{}, errors.New(result.Message)
	}
	if result.Data != nil {
		var resp stock.GetMaterialRequestListResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("GetMaterialRequestList-UnmarshalTo", zap.Any("err", err))
			return &stock.GetMaterialRequestListResp{}, err
		}
		return &resp, nil
	}
	return &stock.GetMaterialRequestListResp{}, nil
}
