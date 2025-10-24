package erp

const (
	ApiMethodMakeMappedDoc = "frappe.model.mapper.make_mapped_doc"
	//ApiSaveCancel 取消已提交记录
	ApiSaveCancel = "frappe.desk.form.save.cancel"
	//ApiIsDocumentAmend 记录是否已修订
	ApiIsDocumentAmend = "frappe.client.is_document_amended"
	//ApiMethodCreateVariantItem 创建单规格Item
	ApiMethodCreateVariantItem = "erpnext.controllers.item_variant.create_variant"
)

// 文档类型
const (
	DocTypePosProfile        = "POS Profile"
	DocTypePosInvoice        = "POS Invoice"
	DocTypePosOpeningEntry   = "POS Opening Entry"
	DocTypePosClosingEntry   = "POS Closing Entry"
	DocTypeModeOfPayment     = "Mode of Payment"
	DocTypeBranch            = "Branch"
	DocTypeSupplier          = "Supplier"
	DocTypeItem              = "Item"
	DocTypePurchaseOrder     = "Purchase Order"
	DocTypeSaleOrder         = "Sales Order"
	DocTypePurchaseReceipt   = "Purchase Receipt"
	DocTypeBom               = "BOM"
	DocTypeMaterialRequest   = "Material Request"
	DocTypeWarehouse         = "Warehouse"
	DocTypePosPermissionRule = "Pos Permission Rule"
	DocTypeItemAttribute     = "Item Attribute"
	//DocTypeStockProjectedQty 预估库存查询
	DocTypeStockProjectedQty = "Stock Projected Qty"
	DocTypeContact           = "Contact"
	DocTypeAddress           = "Address"
	DocTypeItemGroup         = "Item Group"
	DocTypeCustomer          = "Customer"
	DocTypeDeliveryNote      = "Delivery Note"
	DocTypePosPriceList      = "Pos Price List"

	//DocTypeStockLedger 库存台账查询
	DocTypeStockLedger = "Stock Ledger"
	//DocTypeUom 商品单位
	DocTypeUom = "UOM"

	// DocTypeStockReconciliation 库存盘点单据类型
	DocTypeStockReconciliation = "Stock Reconciliation"
	// DocTypeStockReconciliationItem 库存盘点明细类型
	DocTypeStockReconciliationItem = "Stock Reconciliation Item"
)

const (
	//0 - 已保存, 1 - 已提交, 2 - 已取消

	DocstatusDraft     = "0"
	DocstatusSubmitted = "1"
	DocstatusCancelled = "2"
)

const (
	// HeadquartersSupplier 总部供应商，连锁模式默认
	HeadquartersSupplier = "Headquarters - Supplier"
)

const (
	//ColumnCustomPermissionRule 自定义权限规则
	ColumnCustomPermissionRule = "custom_permission_rule"
)

// 物品组
const (
	ItemGroupPosAttribute = "Pos Attribute"
	ItemGroupPosAddon     = "Pos Addon"
)
