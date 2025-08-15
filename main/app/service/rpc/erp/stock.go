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
func (s *erpSrv) CreatePurchaseOrder(ctx cc.Context, createPurchaseOrderReq *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error) {
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
		return &stock.SaveMaterialRequestResp{}, errors.New(result.Message)
	}
	if result.Data != nil {

		var resp stock.SaveMaterialRequestResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("CreatePurchaseOrder-UnmarshalTo", zap.Any("err", err))
			return &stock.SaveMaterialRequestResp{}, err
		}
		return &resp, nil
	}
	return &stock.SaveMaterialRequestResp{}, nil
}

// func (s *erpSrv) GetMaterialRequestList(ctx context.Context, getMaterialRequestListReq req.GetMaterialRequestListReq) (resp.GetMaterialRequestListResp, error) {
// 	client, conn, err := NewErpStockClient()
// 	if err != nil {
// 		return resp.GetMaterialRequestListResp{}, err
// 	}
// 	defer conn.Close()
// 	req := &stock.GetMaterialRequestListReq{
// 		Branch:      getMaterialRequestListReq.Branch,
// 		CompanyAbbr: getMaterialRequestListReq.CompanyAbbr,
// 	}
// 	result, err := client.GetMaterialRequestList(WithSiteCode(ctx, getMaterialRequestListReq.SiteCode), req)
// 	if err != nil {
// 		return resp.GetMaterialRequestListResp{}, err
// 	}
// 	response := &stock.GetMaterialRequestListResp{}
// 	if err := result.Data.UnmarshalTo(response); err != nil {
// 		return resp.GetMaterialRequestListResp{}, err
// 	}
// 	return resp.GetMaterialRequestListResp{
// 		MaterialRequestList: response.MaterialRequestList,
// 	}, nil
// }
