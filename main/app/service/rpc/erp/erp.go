package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

	"google.golang.org/grpc/metadata"

	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IErpSrv interface {
	GetCompanyList(ctx context.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error)
	InitShop(ctx cc.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error)
	GetUomList(ctx context.Context, getUomListReq req.GetUomListReq) (resp.GetUomListResp, error)
	GetAttributeList(ctx context.Context, getAttributeListReq req.GetAttributeListReq) (resp.GetAttributeListResp, error)
	SyncUomAndAttribute(ctx cc.Context, syncUomAndAttributeReq req.SyncUomAndAttributeReq) error
	GetPosProfileList(ctx context.Context, getPosProfileListReq req.GetPosProfileListReq) (resp.GetPosProfileListResp, error)
	AddLianPayment(ctx cc.Context, addLianPaymentReq req.ErpnextSiteAddLianLianPaymentReq) error

	SaveUom(ctx context.Context, saveUomReq req.SaveUomReq) error
	SaveAttribute(ctx context.Context, saveAttributeReq req.SaveAttributeReq) error

	// 采购单
	CreatePurchaseOrder(ctx cc.Context, createPurchaseOrderReq *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error)
	GetMaterialRequestList(ctx cc.Context, getMaterialRequestListReq *stock.GetMaterialRequestListReq) (*stock.GetMaterialRequestListResp, error)
	SavePurchaseReceipt(ctx cc.Context, savePurchaseReceiptReq *buying.SavePurchaseReceiptReq) (*buying.SavePurchaseReceiptResp, error)

	// 供应商
	GetSupplierList(ctx cc.Context) (*buying.GetSupplierListResp, error)

	// 物品
	AddMaterial(ctx cc.Context, params req.MaterialAddErpReq) (*item.ItemInfo, error)                     // 添加物品
	AddProductBomCard(ctx cc.Context, params ProductBomCardAddErpReq) (*manufacturing.SaveBomResp, error) // 添加成本卡
	AddProduct(ctx cc.Context, params req.ProductAddErpReq) (*item.ItemInfo, error)                       // 添加商品
}
type erpSrv struct {
	dbm *database.DBManager
}

func NewIErpSrv(dbm *database.DBManager) IErpSrv {
	return &erpSrv{
		dbm: dbm,
	}
}

func WithSiteCode(ctx context.Context, siteCode string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md.Set("erp_site_code", siteCode)
	return metadata.NewOutgoingContext(ctx, md)
}
