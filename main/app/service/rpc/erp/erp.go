package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/delivery_note"
	"ttpos-bmp/app/ttpos-erp/api/file"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"

	"google.golang.org/grpc/metadata"

	cc "ttpos-server-go/pkg/context"
	pkgCtx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IErpSrv interface {
	GetCompanyList(ctx pkgCtx.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error)
	UpdateTtposCompanyParentUuids(siteCode string) error
	InitShop(ctx pkgCtx.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error)
	GetUomList(ctx context.Context, getUomListReq req.GetUomListReq) (resp.GetUomListResp, error)
	GetAttributeList(ctx context.Context, getAttributeListReq req.GetAttributeListReq) (resp.GetAttributeListResp, error)
	SyncUomAndAttribute(ctx pkgCtx.Context, syncUomAndAttributeReq req.SyncUomAndAttributeReq) error
	GetPosProfileList(ctx context.Context, getPosProfileListReq req.GetPosProfileListReq) (resp.GetPosProfileListResp, error)
	AddLianPayment(ctx pkgCtx.Context, addLianPaymentReq req.ErpnextSiteAddLianLianPaymentReq) error
	OpenPosEntry(ctx context.Context, openEntryReq req.OpenPosEntryReq) (string, error)
	ClosePosEntry(ctx context.Context, closeEntryReq req.ClosePosEntryReq) (string, error)
	SaveUom(ctx context.Context, saveUomReq req.SaveUomReq) error
	DeleteUom(ctx context.Context, deleteUomReq req.DeleteUomReq) error
	SaveAttribute(ctx context.Context, saveAttributeReq req.SaveAttributeReq) error
	SavePosInvoice(ctx pkgCtx.Context, savePosInvoiceReq req.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error)
	ParseSavePosInvoiceError(err error) (*resp.GetPosInvoiceErrorResp, error)
	CancelPosInvoice(ctx pkgCtx.Context, cancelPosInvoiceReq req.CancelPosInvoiceReq) error
	ReturnPosInvoice(ctx pkgCtx.Context, returnPosInvoiceReq req.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error)
	GetPaymentMethodList(ctx pkgCtx.Context, getPaymentReq req.GetPaymentMethodListReq) (*resp.GetPaymentMethodListResp, error)
	AddPaymentMethod(ctx pkgCtx.Context, addPaymentMethodReq req.AddPaymentMethodReq) error
	SaveModeOfPayment(ctx pkgCtx.Context, saveModeOfPaymentReq req.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error)

	// 采购单
	GetMaterialRequestList(ctx pkgCtx.Context, getMaterialRequestListReq *stock.GetMaterialRequestListReq) (*stock.GetMaterialRequestListResp, error)
	SaveMaterialRequest(ctx pkgCtx.Context, companySetting model.CompanySetting, createPurchaseOrderReq *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error)
	CreatePurchaseOrder(ctx pkgCtx.Context, createPurchaseOrderReq *buying.CreatePurchaseOrderReq) (*buying.CreatePurchaseOrderResp, error)
	GetPurchaseOrder(ctx pkgCtx.Context, getPurchaseOrderReq *buying.GetPurchaseOrderReq) (*buying.GetPurchaseOrderResp, error)
	GetPurchaseOrderList(ctx pkgCtx.Context, getPurchaseOrderListReq *buying.GetPurchaseOrderListReq) (*buying.GetPurchaseOrderListResp, error)
	SavePurchaseReceipt(ctx pkgCtx.Context, savePurchaseReceiptReq *buying.SavePurchaseReceiptReq) (*buying.SavePurchaseReceiptResp, error)

	// 盘点单
	SubmitStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, saveStockReconciliationReq *stock.SaveStockReconciliationReq) (*stock.SaveStockReconciliationResp, error)
	ApproveStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, saveStockReconciliationReq *stock.SubmitStockReconciliationReq) (*stock.SubmitStockReconciliationResp, error)
	RejectStockReconciliation(ctx cc.Context, companySetting model.CompanySetting, cancelStockReconciliationReq *stock.CancelStockReconciliationReq) (*stock.CancelStockReconciliationReq, error)
	GetMaterialStockNumByBin(ctx cc.Context, warehouseErpCode string) ([]*stock.ItemStockBin, error)

	// 报损单（库存变动出库）
	SubmitStockEntry(ctx cc.Context, companySetting model.CompanySetting, submitStockEntryReq *stock.SubmitStockEntryReq) (*stock.SubmitStockEntryResp, error)

	// 调拨单
	SaveMaterialTransfer(ctx pkgCtx.Context, companySetting model.CompanySetting, saveMaterialTransferReq *material_transfer.MaterialTransferReq) (*material_transfer.MaterialTransferResp, error)

	// 供应商
	GetSupplierList(ctx pkgCtx.Context) (*buying.GetSupplierListResp, error)                                                          // 获取内部供应商
	ListSuppliers(ctx cc.Context, listSuppliersReq req.GetErpnextSupplierListReq) ([]*buying.SupplierData, error)                     // 获取自己和总部创建可以看到的
	GetSupplierItemList(ctx cc.Context, getSupplierItemListReq req.GetErpnextSupplierItemListReq) ([]*buying.SupplierItemInfo, error) // 获取供应商物品列表
	CreateSupplier(ctx context.Context, createSupplierReq req.CreateSupplierReq) (string, error)
	UpdateSupplier(ctx context.Context, updateSupplierReq req.UpdateSupplierReq) error
	DeleteSupplier(ctx context.Context, deleteSupplierReq req.DeleteSupplierReq) error

	// 物品
	AddMaterial(ctx pkgCtx.Context, params req.MaterialAddErpReq) (*item.ItemInfo, error)                                   // 添加物品
	AddProductBomCard(ctx pkgCtx.Context, params ProductBomCardAddErpReq) (*manufacturing.SaveBomResp, error)               // 添加成本卡
	AddProduct(ctx pkgCtx.Context, params req.ProductAddErpReq) (*item.ItemInfo, error)                                     // 添加商品
	DeleteProduct(ctx pkgCtx.Context, params req.DeleteProductErpReq) error                                                 // 删除商品。删除所有规格和商品模版
	AddProductBom(ctx pkgCtx.Context, params req.ProductBomAddErpReq) (*item.CreateSingleVariantItemResp, error)            // 添加套餐bom
	DeleteProductBom(ctx pkgCtx.Context, params req.DeleteProductBomErpReq) error                                           // 删除套餐bom
	AddPackage(ctx pkgCtx.Context, params req.PackageAddErpReq) (*item.ItemInfo, error)                                     // 添加套餐
	GetMaterialStockNum(ctx pkgCtx.Context, warehouseErpCode string) ([]*item.ItemStock, error)                             // 获取仓库物品库存数量
	GetHeadquarterMaterialList(ctx pkgCtx.Context, params req.GetHeadquarterMaterialListReq) (*item.GetItemListResp, error) // 获取总部物品列表
	GetSubShopMaterialList(ctx pkgCtx.Context) (*item.GetItemListResp, error)                                               // 获取子公司物品列表
	GetMaterialList(ctx pkgCtx.Context, params GetMaterialListReq) (*item.GetItemListResp, error)                           // 获取物品列表
	GetProductBomCardList(ctx pkgCtx.Context) (*manufacturing.GetBomListResp, error)                                        // 获取成本卡列表
	GetProductBomCardDetail(ctx pkgCtx.Context, params req.ErpProductBomCardDetailReq) (*manufacturing.GetBomResp, error)   // 获取成本卡详情

	// 仓库
	CreateWarehouse(ctx context.Context, createWarehouseReq req.CreateErpnextWarehouseReq) (string, error)                        // 创建仓库
	UpdateWarehouse(ctx context.Context, updateWarehouseReq req.UpdateErpnextWarehouseReq) error                                  // 更新仓库
	DeleteWarehouse(ctx context.Context, deleteWarehouseReq req.DeleteErpnextWarehouseReq) error                                  // 删除仓库
	GetWarehouseList(ctx context.Context, getWarehouseListReq req.GetErpnextWarehouseListReq) ([]*warehouse.WarehouseInfo, error) // 获取仓库列表

	// 规格
	SaveFlavor(ctx context.Context, params req.SaveErpFlavorReq) error                                    // 保存规格
	GetFlavorList(ctx context.Context, params req.GetErpFlavorListReq) (resp.GetErpFlavorListResp, error) // 获取规格列表

	// 加料
	GetSauceList(ctx pkgCtx.Context, sourceListReq req.GetErpSauceListReq) ([]*item.ItemInfo, error) // 获取加料列表

	// 商品
	GetProductList(ctx pkgCtx.Context, params GetErpProductListReq) (*item.GetItemListResp, error) // 获取商品列表
	UpdateProduct(ctx pkgCtx.Context, params UpdateProductReq) error                               // 更新商品

	// 送货单
	GetDeliveryNoteList(ctx pkgCtx.Context, getDeliveryNoteListReq *delivery_note.GetDeliveryNoteListReq) (*delivery_note.GetDeliveryNoteListResp, error)
	GetDeliveryNote(ctx pkgCtx.Context, dnName string) (*delivery_note.DeliveryNote, error) // 获取单个送货单详情

	// 文件
	UploadFileUrl(ctx pkgCtx.Context, req *file.UploadFileUrlReq) (*file.UploadFileResp, error) // 通过 URL 上传文件到 ERPNext
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
