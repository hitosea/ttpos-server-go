package erp

import (
	"context"
	"errors"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"

	"ttpos-server-go/app/model"
	cc "ttpos-server-go/pkg/context"
	pkgCtx "ttpos-server-go/pkg/context"

	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpBuyingClient() (buying.BuyingServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return buying.NewBuyingServiceClient(conn), conn, nil
}

// GetSupplierList 获取供应商列表
func (s *erpSrv) GetSupplierList(ctx pkgCtx.Context) (*buying.GetSupplierListResp, error) {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return &buying.GetSupplierListResp{}, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting
	result, err := client.GetSupplierList(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), &buying.GetSupplierListReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
	})
	if err != nil {
		return &buying.GetSupplierListResp{}, err
	}
	if result.Code != "0" {
		logger.Logger.Error("GetSupplierList-GetSupplierList", zap.Any("err", err))
		return &buying.GetSupplierListResp{}, errors.New("调用erp接口失败 - 20001")
	}
	if result.Data != nil {
		var resp buying.GetSupplierListResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("GetSupplierList-UnmarshalTo", zap.Any("err", err))
			return &buying.GetSupplierListResp{}, err
		}
		return &resp, nil
	}
	return &buying.GetSupplierListResp{}, nil
}

// ListSuppliers 获取内部供应商列表
func (s *erpSrv) ListSuppliers(ctx cc.Context, companySetting model.CompanySetting, listSuppliersReq *buying.ListSuppliersReq) (*buying.GetSupplierListResp, error) {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return &buying.GetSupplierListResp{}, err
	}
	defer conn.Close()
	result, err := client.ListSuppliers(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), &buying.ListSuppliersReq{
		CompanyAbbr: listSuppliersReq.CompanyAbbr,
		PageNo:      listSuppliersReq.PageNo,
		PageSize:    listSuppliersReq.PageSize,
	})
	if err != nil {
		return &buying.GetSupplierListResp{}, err
	}
	if result.Code != "0" {
		logger.Logger.Error("ListSuppliers-ListSuppliers", zap.Any("err", err))
		return &buying.GetSupplierListResp{}, errors.New("调用erp接口失败 - 20002")
	}
	if result.Data != nil {
		var resp buying.GetSupplierListResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("ListSuppliers-UnmarshalTo", zap.Any("err", err))
			return &buying.GetSupplierListResp{}, err
		}
		return &resp, nil
	}
	return &buying.GetSupplierListResp{}, nil
}

// SavePurchaseReceipt 保存收货单
func (s *erpSrv) SavePurchaseReceipt(ctx pkgCtx.Context, savePurchaseReceiptReq *buying.SavePurchaseReceiptReq) (*buying.SavePurchaseReceiptResp, error) {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return &buying.SavePurchaseReceiptResp{}, err
	}
	defer conn.Close()

	companySetting := ctx.GetCompany().CompanySetting

	result, err := client.SavePurchaseReceipt(WithSiteCode(context.Background(), companySetting.ErpnextSiteCode), savePurchaseReceiptReq)
	if err != nil {
		return &buying.SavePurchaseReceiptResp{}, err
	}
	if result.Code != "0" {
		logger.Logger.Error("SavePurchaseReceipt-SavePurchaseReceipt", zap.Any("err", err))
		return &buying.SavePurchaseReceiptResp{}, errors.New("调用erp接口失败 - 20003")
	}
	if result.Data != nil {
		var resp buying.SavePurchaseReceiptResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("SavePurchaseReceipt-UnmarshalTo", zap.Any("err", err))
			return &buying.SavePurchaseReceiptResp{}, err
		}
		return &resp, nil
	}
	return &buying.SavePurchaseReceiptResp{}, nil
}

// CreateSupplier 创建供应商
func (s *erpSrv) CreateSupplier(ctx context.Context, createSupplierReq req.CreateSupplierReq) (string, error) {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	result, err := client.CreateSupplier(WithSiteCode(ctx, createSupplierReq.SiteCode), &buying.CreateSupplierReq{
		Supplier: &buying.SupplierData{
			SupplierName: createSupplierReq.SupplierName, // 供应商名称
			CompanyAbbr:  createSupplierReq.CompanyAbbr,  // 总部在移动管理端创建供应商时不需要传
			Branch:       createSupplierReq.Branch,       // 总部在移动管理端创建供应商时不需要传
			Disabled:     createSupplierReq.Disabled,     // 是否禁用
		},
	})
	if err != nil {
		return "", err
	}
	if result.Code != "0" {
		logger.Logger.Error("CreateSupplier-CreateSupplier", zap.Any("err", err))
		return "", errors.New(result.GetMessage())
	}
	if result.Data != nil {
		var resp buying.CreateSupplierResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("CreateSupplier-UnmarshalTo", zap.Any("err", err))
			return "", err
		}
		return resp.Supplier.Name, nil
	}
	return "", nil
}

// UpdateSupplier 更新供应商
func (s *erpSrv) UpdateSupplier(ctx context.Context, updateSupplierReq req.UpdateSupplierReq) error {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := client.UpdateSupplier(WithSiteCode(ctx, updateSupplierReq.SiteCode), &buying.UpdateSupplierReq{
		Supplier: &buying.SupplierData{
			Name:         updateSupplierReq.Name,
			SupplierName: updateSupplierReq.SupplierName,
			CompanyAbbr:  updateSupplierReq.CompanyAbbr,
			Branch:       updateSupplierReq.Branch,
			Disabled:     updateSupplierReq.Disabled,
		},
	})
	if err != nil {
		return err
	}
	if result.Code != "0" {
		logger.Logger.Error("UpdateSupplier-UpdateSupplier", zap.Any("err", err))
		return errors.New(result.GetMessage())
	}
	if result.Data != nil {
		var resp buying.UpdateSupplierResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("UpdateSupplier-UnmarshalTo", zap.Any("err", err))
			return err
		}
		return nil
	}
	return nil
}

// DeleteSupplier 删除供应商
func (s *erpSrv) DeleteSupplier(ctx context.Context, deleteSupplierReq req.DeleteSupplierReq) error {
	client, conn, err := NewErpBuyingClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := client.DeleteSupplier(WithSiteCode(ctx, deleteSupplierReq.SiteCode), &buying.DeleteSupplierReq{
		Name: deleteSupplierReq.Name,
	})
	if err != nil {
		return err
	}
	if result.Code != "0" {
		logger.Logger.Error("DeleteSupplier-DeleteSupplier", zap.Any("err", err))
		return errors.New(result.GetMessage())
	}
	if result.Data != nil {
		var resp buying.DeleteSupplierResp
		if err := result.Data.UnmarshalTo(&resp); err != nil {
			logger.Logger.Error("DeleteSupplier-UnmarshalTo", zap.Any("err", err))
			return err
		}
		return nil
	}
	return nil
}
