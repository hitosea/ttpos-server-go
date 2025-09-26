package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

	"google.golang.org/grpc/metadata"

	cc "ttpos-server-go/pkg/context"
	pkgCtx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IErpSrv interface {
	GetCompanyList(ctx pkgCtx.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error)
	InitShop(ctx pkgCtx.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error)
	GetUomList(ctx context.Context, getUomListReq req.GetUomListReq) (resp.GetUomListResp, error)
	GetAttributeList(ctx context.Context, getAttributeListReq req.GetAttributeListReq) (resp.GetAttributeListResp, error)
	SyncUomAndAttribute(ctx pkgCtx.Context, syncUomAndAttributeReq req.SyncUomAndAttributeReq) error
	GetPosProfileList(ctx context.Context, getPosProfileListReq req.GetPosProfileListReq) (resp.GetPosProfileListResp, error)
	AddLianPayment(ctx pkgCtx.Context, addLianPaymentReq req.ErpnextSiteAddLianLianPaymentReq) error
	OpenPosEntry(ctx context.Context, openEntryReq req.OpenPosEntryReq) (string, error)
	ClosePosEntry(ctx context.Context, closeEntryReq req.ClosePosEntryReq) (string, error)
	SaveUom(ctx context.Context, saveUomReq req.SaveUomReq) error
	SaveAttribute(ctx context.Context, saveAttributeReq req.SaveAttributeReq) error
	SavePosInvoice(ctx pkgCtx.Context, savePosInvoiceReq req.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error)
	ParseSavePosInvoiceError(err error) (*resp.GetPosInvoiceErrorResp, error)
	CancelPosInvoice(ctx pkgCtx.Context, cancelPosInvoiceReq req.CancelPosInvoiceReq) error
	ReturnPosInvoice(ctx pkgCtx.Context, returnPosInvoiceReq req.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error)
	GetPaymentMethodList(ctx pkgCtx.Context, getPaymentReq req.GetPaymentMethodListReq) (*resp.GetPaymentMethodListResp, error)
	AddPaymentMethod(ctx pkgCtx.Context, addPaymentMethodReq req.AddPaymentMethodReq) error

	// 采购单
	SaveMaterialRequest(ctx pkgCtx.Context, createPurchaseOrderReq *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error)
	GetMaterialRequestList(ctx pkgCtx.Context, getMaterialRequestListReq *stock.GetMaterialRequestListReq) (*stock.GetMaterialRequestListResp, error)
	CreatePurchaseOrder(ctx pkgCtx.Context, createPurchaseOrderReq *buying.CreatePurchaseOrderReq) (*buying.CreatePurchaseOrderResp, error)
	SavePurchaseReceipt(ctx pkgCtx.Context, savePurchaseReceiptReq *buying.SavePurchaseReceiptReq) (*buying.SavePurchaseReceiptResp, error)

	// 供应商
	GetSupplierList(ctx pkgCtx.Context) (*buying.GetSupplierListResp, error)                                      // 获取内部供应商
	ListSuppliers(ctx cc.Context, listSuppliersReq req.GetErpnextSupplierListReq) ([]*buying.SupplierData, error) // 获取自己和总部创建可以看到的
	CreateSupplier(ctx context.Context, createSupplierReq req.CreateSupplierReq) (string, error)
	UpdateSupplier(ctx context.Context, updateSupplierReq req.UpdateSupplierReq) error
	DeleteSupplier(ctx context.Context, deleteSupplierReq req.DeleteSupplierReq) error

	// 物品
	AddMaterial(ctx pkgCtx.Context, params req.MaterialAddErpReq) (*item.ItemInfo, error)                                   // 添加物品
	AddProductBomCard(ctx pkgCtx.Context, params ProductBomCardAddErpReq) (*manufacturing.SaveBomResp, error)               // 添加成本卡
	AddProduct(ctx pkgCtx.Context, params req.ProductAddErpReq) (*item.ItemInfo, error)                                     // 添加商品
	DeleteProduct(ctx pkgCtx.Context, params req.DeleteProductErpReq) error                                                 // 删除商品。删除所有规格和商品模版
	AddPackage(ctx pkgCtx.Context, params req.PackageAddErpReq) (*item.ItemInfo, error)                                     // 添加套餐
	GetMaterialStockNum(ctx pkgCtx.Context, warehouseErpCode string) ([]*item.ItemStock, error)                             // 获取仓库物品库存数量
	GetHeadquarterMaterialList(ctx pkgCtx.Context, params req.GetHeadquarterMaterialListReq) (*item.GetItemListResp, error) // 获取总部物品列表
	GetProductBomCardList(ctx pkgCtx.Context) (*manufacturing.GetBomListResp, error)                                        // 获取成本卡列表
	GetProductBomCardDetail(ctx pkgCtx.Context, params req.ErpProductBomCardDetailReq) (*manufacturing.GetBomResp, error)   // 获取成本卡详情

	// 仓库
	CreateWarehouse(ctx context.Context, createWarehouseReq req.CreateErpnextWarehouseReq) (string, error)                        // 创建仓库
	UpdateWarehouse(ctx context.Context, updateWarehouseReq req.UpdateErpnextWarehouseReq) error                                  // 更新仓库
	DeleteWarehouse(ctx context.Context, deleteWarehouseReq req.DeleteErpnextWarehouseReq) error                                  // 删除仓库
	GetWarehouseList(ctx context.Context, getWarehouseListReq req.GetErpnextWarehouseListReq) ([]*warehouse.WarehouseInfo, error) // 获取仓库列表

	// 规格
	GetFlavorList(ctx context.Context, params req.GetErpFlavorListReq) (resp.GetErpFlavorListResp, error) // 获取规格列表
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
