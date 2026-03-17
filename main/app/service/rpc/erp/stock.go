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
	if result.Data == nil {
		return &stock.SaveMaterialRequestResp{}, nil
	}
	var resp stock.SaveMaterialRequestResp
	if err := result.Data.UnmarshalTo(&resp); err != nil {
		logger.Logger.Error("SaveMaterialRequest-UnmarshalTo", zap.Any("err", err))
		return &stock.SaveMaterialRequestResp{}, err
	}
	return &resp, nil
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
	var resp stock.SaveStockReconciliationResp
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &resp, err
	}
	defer conn.Close()

	saveStockReconciliationReq.CompanyAbbr = companySetting.ErpnextCompanyAbbr
	saveStockReconciliationReq.Branch = companySetting.ErpnextBranchName

	result, err := client.SaveStockReconciliation(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), saveStockReconciliationReq)
	if err != nil {
		return &resp, err
	}
	if result.Code != "0" {
		return &resp, errors.New(result.Message)
	}
	if result.Data != nil {
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SaveStockReconciliation-UnmarshalTo", zap.Any("err", err))
			return &resp, err
		}
		return &resp, nil
	}
	return &resp, nil
}

// 审核盘点单，对应erp的提交盘点单
func (s *erpSrv) ApproveStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, saveStockReconciliationReq *stock.SubmitStockReconciliationReq) (*stock.SubmitStockReconciliationResp, error) {
	var resp stock.SubmitStockReconciliationResp
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &resp, err
	}
	defer conn.Close()

	result, err := client.SubmitStockReconciliation(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), saveStockReconciliationReq)
	if err != nil {
		return &resp, err
	}
	if result.Code != "0" {
		return &resp, errors.New(result.Message)
	}
	if result.Data != nil {
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SubmitStockReconciliation-UnmarshalTo", zap.Any("err", err))
			return &resp, err
		}
		return &resp, nil
	}
	return &resp, nil
}

// 驳回盘点单，对应erp取消盘点单
func (s *erpSrv) RejectStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, cancelStockReconciliationReq *stock.CancelStockReconciliationReq) (*stock.CancelStockReconciliationReq, error) {
	var resp stock.CancelStockReconciliationReq
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &resp, err
	}
	defer conn.Close()

	result, err := client.CancelStockReconciliation(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), cancelStockReconciliationReq)
	if err != nil {
		return &resp, err
	}
	if result.Code != "0" {
		return &resp, errors.New(result.Message)
	}
	if result.Data != nil {
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("CancelStockReconciliation-UnmarshalTo", zap.Any("err", err))
			return &resp, err
		}
		return &resp, nil
	}
	return &resp, nil
}

// 获取物品在指定仓库的 Bin 记录
func (s *erpSrv) GetMaterialStockNumByBin(ctx cc.Context, warehouseErpCode string) ([]*stock.ItemStockBin, error) {
	client, conn, err := NewErpStockClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result, err := client.GetBin(WithSiteCode(ctx.GetContext(), ctx.GetCompany().CompanySetting.ErpnextSiteCode), &stock.GetBinReq{
		Warehouse: warehouseErpCode,
	})
	if err != nil {
		return nil, err
	}
	if result.Code != "0" {
		return nil, errors.New(result.Message)
	}
	if result.Data != nil {
		var resp stock.GetBinResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("GetBin-UnmarshalTo", zap.Any("err", err))
			return nil, err
		}
		return resp.Items, nil
	}
	return nil, nil
}

// SubmitStockEntry 提交库存变动（物料出库），用于报损单审核通过时调用
func (s *erpSrv) SubmitStockEntry(ctx cc.Context, companySetting model.CompanySetting, submitStockEntryReq *stock.SubmitStockEntryReq) (*stock.SubmitStockEntryResp, error) {
	var resp stock.SubmitStockEntryResp
	client, conn, err := NewErpStockClient()
	if err != nil {
		return &resp, err
	}
	defer conn.Close()

	submitStockEntryReq.CompanyAbbr = companySetting.ErpnextCompanyAbbr
	submitStockEntryReq.Branch = companySetting.ErpnextBranchName

	result, err := client.SubmitStockEntry(WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), submitStockEntryReq)
	if err != nil {
		return &resp, err
	}
	if result.Code != "0" {
		return &resp, errors.New(result.Message)
	}
	if result.Data != nil {
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SubmitStockEntry-UnmarshalTo", zap.Any("err", err))
			return &resp, err
		}
		return &resp, nil
	}
	return &resp, nil
}
