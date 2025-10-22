package erp

import (
	"context"
	"errors"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/model"
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
func (s *erpSrv) SaveMaterialRequest(ctx cc.Context, companySetting model.CompanySetting, createPurchaseOrderReq *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &stock.SaveMaterialRequestResp{}, err
	}
	defer conn.Close()

	createPurchaseOrderReq.CompanyAbbr = companySetting.ErpnextCompanyAbbr
	createPurchaseOrderReq.Branch = companySetting.ErpnextBranchName

	result, err := client.SaveMaterialRequest(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), createPurchaseOrderReq)
	if err != nil {
		return &stock.SaveMaterialRequestResp{}, err
	}
	if result.Code != "0" {
		logger.Logger.Error("SaveMaterialRequest-SaveMaterialRequest", zap.Any("err", err))
		return &stock.SaveMaterialRequestResp{}, errors.New("调用erp接口失败-1001-" + result.GetMessage())
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

// 提交盘点单，对应erp的保存盘点单
func (s *erpSrv) SubmitStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, saveStockReconciliationReq *stock.SaveStockReconciliationReq) (*stock.SaveStockReconciliationResp, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &stock.SaveStockReconciliationResp{}, err
	}
	defer conn.Close()

	saveStockReconciliationReq.CompanyAbbr = companySetting.ErpnextCompanyAbbr
	saveStockReconciliationReq.Branch = companySetting.ErpnextBranchName

	result, err := client.SaveStockReconciliation(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), saveStockReconciliationReq)
	if err != nil {
		return &stock.SaveStockReconciliationResp{}, err
	}
	if result.Code != "0" {
		return &stock.SaveStockReconciliationResp{}, errors.New(result.Message)
	}
	if result.Data != nil {
		var resp stock.SaveStockReconciliationResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SaveStockReconciliation-UnmarshalTo", zap.Any("err", err))
			return &stock.SaveStockReconciliationResp{}, err
		}
		return &resp, nil
	}
	return &stock.SaveStockReconciliationResp{}, nil
}

// 审核盘点单，对应erp的提交盘点单
func (s *erpSrv) ApproveStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, saveStockReconciliationReq *stock.SubmitStockReconciliationReq) (*stock.SubmitStockReconciliationReq, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &stock.SubmitStockReconciliationReq{}, err
	}
	defer conn.Close()

	result, err := client.SubmitStockReconciliation(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), saveStockReconciliationReq)
	if err != nil {
		return &stock.SubmitStockReconciliationReq{}, err
	}
	if result.Code != "0" {
		return &stock.SubmitStockReconciliationReq{}, errors.New(result.Message)
	}
	if result.Data != nil {
		var resp stock.SubmitStockReconciliationReq
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SubmitStockReconciliation-UnmarshalTo", zap.Any("err", err))
			return &stock.SubmitStockReconciliationReq{}, err
		}
		return &resp, nil
	}
	return &stock.SubmitStockReconciliationReq{}, nil
}

// 驳回盘点单，对应erp取消盘点单
func (s *erpSrv) RejectStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, cancelStockReconciliationReq *stock.CancelStockReconciliationReq) (*stock.CancelStockReconciliationReq, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &stock.CancelStockReconciliationReq{}, err
	}
	defer conn.Close()

	result, err := client.CancelStockReconciliation(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), cancelStockReconciliationReq)
	if err != nil {
		return &stock.CancelStockReconciliationReq{}, err
	}
	if result.Code != "0" {
		return &stock.CancelStockReconciliationReq{}, errors.New(result.Message)
	}
	if result.Data != nil {
		var resp stock.CancelStockReconciliationReq
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("CancelStockReconciliation-UnmarshalTo", zap.Any("err", err))
			return &stock.CancelStockReconciliationReq{}, err
		}
		return &resp, nil
	}
	return &stock.CancelStockReconciliationReq{}, nil
}
