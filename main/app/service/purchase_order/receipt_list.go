package purchase_order

import (
	"fmt"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/delivery_note"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetReceiptList 获取采购单的收货清单
func (s *purchaseOrderSrv) GetReceiptList(ctx context.Context, purchaseOrder *model.PurchaseOrder) ([]resp.ReceiptListItem, error) {
	db := ctx.GetDB()
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 获取该采购单的所有收货单（只返回待收货和已收货状态，不返回已取消）
	receiptOrders, err := receiptOrderRepo.GetList(
		receiptOrderRepo.WherePurchaseOrderUuid(purchaseOrder.Uuid),
		receiptOrderRepo.WhereStatusIn([]int{constant.ReceiptOrderStatusPending, constant.ReceiptOrderStatusReceived}),
		receiptOrderRepo.WithItems(),
	)
	if err != nil {
		return nil, err
	}

	result := make([]resp.ReceiptListItem, 0)

	// 旧采购单（无 ErpSaleOrderNo）：返回单个收货清单项，包含所有收货单
	if purchaseOrder.ErpSaleOrderNo == "" {
		result = append(result, s.getLegacyReceiptList(ctx, purchaseOrder, receiptOrders))
		return result, nil
	}

	// 新采购单（有 ErpSaleOrderNo）：按 DN 和供应商分类

	// 1. 处理DN类型（集采品牌方发货）
	dnList, err := s.getDNReceiptList(ctx, purchaseOrder, receiptOrders)
	if err != nil {
		logger.Logger.Warn("获取DN清单失败", zap.Error(err), zap.Uint64("purchase_order_uuid", purchaseOrder.Uuid))
	} else {
		result = append(result, dnList...)
	}

	// 2. 处理供应商直接发货类型
	supplierList := s.getSupplierDirectReceiptList(ctx, db, purchaseOrder, receiptOrders)
	result = append(result, supplierList...)

	return result, nil
}

// getDNReceiptList 获取DN类型的收货清单
func (s *purchaseOrderSrv) getDNReceiptList(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
	receiptOrders []model.PurchaseReceiptOrder,
) ([]resp.ReceiptListItem, error) {
	companySetting := ctx.GetCompanySetting()
	// 调用ERP获取DN列表
	erpSrv := erp.NewIErpSrv(s.dbm)
	dnListResp, err := erpSrv.GetDeliveryNoteList(ctx, &delivery_note.GetDeliveryNoteListReq{
		CompanyAbbr:  companySetting.ErpnextHeadquarterAbbr,
		SoNo:         purchaseOrder.ErpSaleOrderNo,
		IncludeItems: true,
	})
	if err != nil {
		return nil, err
	}

	if dnListResp == nil || len(dnListResp.DeliveryNoteList) == 0 {
		return nil, nil
	}

	result := make([]resp.ReceiptListItem, 0, len(dnListResp.DeliveryNoteList))

	// 过滤重复的 DN（按 Name 去重）
	seenDN := make(map[string]bool)
	for _, dn := range dnListResp.DeliveryNoteList {
		if seenDN[dn.Name] {
			continue
		}
		seenDN[dn.Name] = true
		// 找到关联此DN的收货单
		relatedReceipts := s.getReceiptsForDN(dn.Name, receiptOrders)

		// 计算是否完成收货（按item_code:uom分组判断）
		isCompleted := true
		for _, dnItem := range dn.Items {
			receivedQty := s.calculateReceivedQtyForDNItem(dnItem.ItemCode, dnItem.Uom, relatedReceipts)
			if receivedQty < dnItem.Qty {
				isCompleted = false
				break
			}
		}

		// 构建收货单显示列表
		receiptOrderInfos := make([]resp.ReceiptListOrderInfo, 0, len(relatedReceipts))
		for _, receipt := range relatedReceipts {
			receiptOrderInfos = append(receiptOrderInfos, s.buildReceiptOrderInfo(ctx, receipt))
		}

		result = append(result, resp.ReceiptListItem{
			IsDeliveryNote:      true,
			DeliveryNoteNo:      dn.Name,
			IsCompleted:         isCompleted,
			SupplierName:        purchaseOrder.SupplierName,
			SupplierErpCode:     purchaseOrder.SupplierErpCode,
			ErpPurchaseOrderNo:  purchaseOrder.ErpOrderNo,
			LocaleWarehouseName: *language.JsonToLocaleResponse(purchaseOrder.WarehouseName),
			ReceiptOrders:       receiptOrderInfos,
		})
	}

	return result, nil
}

// getSupplierDirectReceiptList 获取供应商直接发货类型的收货清单
func (s *purchaseOrderSrv) getSupplierDirectReceiptList(
	ctx context.Context,
	db *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
	receiptOrders []model.PurchaseReceiptOrder,
) []resp.ReceiptListItem {
	// 收集所有 delivered_by_supplier=1 的物品，按 supplier_erp_code 分组
	supplierItemsMap := make(map[string][]model.PurchaseOrderItem)

	for _, item := range purchaseOrder.Items {
		if item.Material != nil && item.Material.DeliveredBySupplier == 1 {
			supplierCode := item.Material.SupplierErpCode
			if supplierCode == "" {
				continue
			}
			supplierItemsMap[supplierCode] = append(supplierItemsMap[supplierCode], item)
		}
	}

	if len(supplierItemsMap) == 0 {
		return nil
	}

	result := make([]resp.ReceiptListItem, 0, len(supplierItemsMap))
	supplierRepo := repository.NewSupplierRepo(db)

	for supplierCode, items := range supplierItemsMap {
		// 获取供应商信息
		supplier, _ := supplierRepo.GetByErpCode(supplierCode)
		supplierName := ""
		if supplier != nil {
			supplierName = supplier.Name
		}

		// 找到关联此供应商的收货单（非DN类型）
		relatedReceipts := s.getReceiptsForSupplier(supplierCode, receiptOrders)

		// 构建收货单显示列表
		receiptOrderInfos := make([]resp.ReceiptListOrderInfo, 0, len(relatedReceipts))
		for _, receipt := range relatedReceipts {
			receiptOrderInfos = append(receiptOrderInfos, s.buildReceiptOrderInfo(ctx, receipt))
		}

		// 检查是否完成收货
		isCompleted := s.checkSupplierReceiptCompleted(items)

		result = append(result, resp.ReceiptListItem{
			IsDeliveryNote:      false,
			DeliveryNoteNo:      "",
			IsCompleted:         isCompleted,
			SupplierName:        supplierName,
			SupplierErpCode:     supplierCode,
			ErpPurchaseOrderNo:  purchaseOrder.ErpOrderNo,
			LocaleWarehouseName: *language.JsonToLocaleResponse(purchaseOrder.WarehouseName),
			ReceiptOrders:       receiptOrderInfos,
		})
	}

	return result
}

// getLegacyReceiptList 获取旧采购单（无 ErpSaleOrderNo）的收货清单
// 返回单个收货清单项，包含所有收货单，是否完成根据总收货进度判断
func (s *purchaseOrderSrv) getLegacyReceiptList(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
	receiptOrders []model.PurchaseReceiptOrder,
) resp.ReceiptListItem {
	// 构建所有收货单的显示列表
	receiptOrderInfos := make([]resp.ReceiptListOrderInfo, 0, len(receiptOrders))
	for _, receipt := range receiptOrders {
		receiptOrderInfos = append(receiptOrderInfos, s.buildReceiptOrderInfo(ctx, receipt))
	}

	// 根据采购单的收货进度判断是否完成（所有物品的所有单位都已收货完成）
	isCompleted := s.checkItemsReceiptCompleted(purchaseOrder.Items)

	return resp.ReceiptListItem{
		IsLegacy:            true,
		IsDeliveryNote:      false,
		DeliveryNoteNo:      "",
		IsCompleted:         isCompleted,
		SupplierName:        purchaseOrder.SupplierName,
		SupplierErpCode:     purchaseOrder.SupplierErpCode,
		ErpPurchaseOrderNo:  purchaseOrder.ErpOrderNo,
		LocaleWarehouseName: *language.JsonToLocaleResponse(purchaseOrder.WarehouseName),
		ReceiptOrders:       receiptOrderInfos,
	}
}

// getReceiptsForDN 获取关联指定DN的收货单
func (s *purchaseOrderSrv) getReceiptsForDN(dnNo string, receiptOrders []model.PurchaseReceiptOrder) []model.PurchaseReceiptOrder {
	result := make([]model.PurchaseReceiptOrder, 0)
	for _, receipt := range receiptOrders {
		if receipt.DeliveryNoteNo == dnNo {
			result = append(result, receipt)
		}
	}
	return result
}

// getReceiptsForSupplier 获取关联指定供应商的收货单（非DN类型）
func (s *purchaseOrderSrv) getReceiptsForSupplier(supplierCode string, receiptOrders []model.PurchaseReceiptOrder) []model.PurchaseReceiptOrder {
	result := make([]model.PurchaseReceiptOrder, 0)
	for _, receipt := range receiptOrders {
		if receipt.SupplierErpCode == supplierCode && receipt.IsFromDeliveryNote == 0 {
			result = append(result, receipt)
		}
	}
	return result
}

// calculateReceivedQtyForDNItem 计算DN物品的已收货数量
func (s *purchaseOrderSrv) calculateReceivedQtyForDNItem(itemCode string, uom string, receipts []model.PurchaseReceiptOrder) float64 {
	totalReceived := 0.0
	for _, receipt := range receipts {
		// 只计算已确认收货的
		if receipt.Status != constant.ReceiptOrderStatusReceived {
			continue
		}
		for _, item := range receipt.Items {
			if item.MaterialCode != itemCode {
				continue
			}
			// 按单位匹配计算已收货数量
			if len(item.Units) > 0 {
				for _, unit := range item.Units {
					if unit.ErpnextUom == uom {
						totalReceived += unit.Num
					}
				}
			} else if item.ErpnextUom == uom {
				totalReceived += item.Num
			}
		}
	}
	return totalReceived
}

// checkSupplierReceiptCompleted 检查供应商物品是否全部收货完成
func (s *purchaseOrderSrv) checkSupplierReceiptCompleted(items []model.PurchaseOrderItem) bool {
	return s.checkItemsReceiptCompleted(items)
}

// checkItemsReceiptCompleted 检查物品列表是否全部收货完成（基于 ttpos_purchase_order_item_unit 判断）
func (s *purchaseOrderSrv) checkItemsReceiptCompleted(items []model.PurchaseOrderItem) bool {
	for _, item := range items {
		// 只使用 ttpos_purchase_order_item_unit 的数据判断
		for _, unit := range item.Units {
			if unit.ArrivalNum < unit.Num {
				return false
			}
		}
	}
	return true
}

// buildReceiptOrderInfo 构建收货单显示信息
func (s *purchaseOrderSrv) buildReceiptOrderInfo(ctx context.Context, receipt model.PurchaseReceiptOrder) resp.ReceiptListOrderInfo {
	displayNo := receipt.ErpOrderNo
	isConfirmed := receipt.Status == constant.ReceiptOrderStatusReceived

	if !isConfirmed {
		// 草稿状态显示 "草稿（修改时间）"，使用商家时区
		updateTime := time.Unix(int64(receipt.UpdateTime), 0)
		if tz := ctx.GetCompanySetting().Timezone; tz != "" {
			if loc, err := time.LoadLocation(tz); err == nil {
				updateTime = updateTime.In(loc)
			}
		}
		draftText := i18n.Translate(ctx.GetLanguage(), "草稿")
		displayNo = fmt.Sprintf("%s（%s）", draftText, updateTime.Format("2006-01-02 15:04"))
	}

	return resp.ReceiptListOrderInfo{
		Uuid:        receipt.Uuid,
		DisplayNo:   displayNo,
		Status:      receipt.Status,
		ErpOrderNo:  receipt.ErpOrderNo,
		CreateTime:  int64(receipt.CreateTime),
		IsConfirmed: isConfirmed,
	}
}

// GetReceiptPendingItems 获取待收货物品列表 (v2.16.0+)
func (s *purchaseOrderSrv) GetReceiptPendingItems(ctx context.Context, pendingReq req.ReceiptPendingItemsReq) (resp.ReceiptPendingItemsResp, error) {
	db := ctx.GetDB()
	result := resp.ReceiptPendingItemsResp{
		Items: make([]resp.ReceiptPendingItemInfo, 0),
	}

	// 获取采购单及其物品（WithItems 已包含 Material 和 Units 关联）
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(
		pendingReq.PurchaseOrderUuid,
		purchaseOrderRepo.WithItems(),
	)
	if err != nil {
		return result, err
	}
	if purchaseOrder == nil {
		return result, errors.New("采购单不存在")
	}

	// 根据请求参数决定获取DN还是供应商类型的待收货物品
	if pendingReq.DeliveryNoteNo != "" {
		return s.getDNPendingItems(ctx, purchaseOrder, pendingReq.DeliveryNoteNo)
	} else if pendingReq.SupplierErpCode != "" {
		return s.getSupplierPendingItems(ctx, db, purchaseOrder, pendingReq.SupplierErpCode)
	}

	return result, errors.New("请指定DN单号或供应商编码")
}

// getDNPendingItems 获取DN类型的待收货物品
// 注意：数量使用DN的Uom单位（与采购单物品的ErpnextUom匹配）
func (s *purchaseOrderSrv) getDNPendingItems(
	ctx context.Context,
	purchaseOrder *model.PurchaseOrder,
	dnNo string,
) (resp.ReceiptPendingItemsResp, error) {
	result := resp.ReceiptPendingItemsResp{
		IsDeliveryNote:      true,
		DeliveryNoteNo:      dnNo,
		SupplierName:        purchaseOrder.SupplierName,
		SupplierErpCode:     purchaseOrder.SupplierErpCode,
		LocaleWarehouseName: *language.JsonToLocaleResponse(purchaseOrder.WarehouseName),
		OrderNo:             purchaseOrder.OrderNo,
		Items:               make([]resp.ReceiptPendingItemInfo, 0),
	}

	db := ctx.GetDB()

	// 调用ERP获取DN详情（包含物品）
	erpSrv := erp.NewIErpSrv(s.dbm)
	targetDN, err := erpSrv.GetDeliveryNote(ctx, dnNo)
	if err != nil {
		return result, err
	}

	// 将DN物品按item_code分组，收集每个物品的所有单位信息
	// key: item_code, value: []dnUnitInfo (同一物品可能有多个单位)
	dnItemUnitsMap := make(map[string][]dnUnitInfo)
	for _, dnItem := range targetDN.Items {
		dnItemUnitsMap[dnItem.ItemCode] = append(dnItemUnitsMap[dnItem.ItemCode], dnUnitInfo{
			Qty: dnItem.Qty,
			Uom: dnItem.Uom,
		})
	}

	// 获取该采购单关联此DN的所有已确认收货单
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)
	receiptOrders, err := receiptOrderRepo.GetList(
		receiptOrderRepo.WherePurchaseOrderUuid(purchaseOrder.Uuid),
		receiptOrderRepo.WhereDeliveryNoteNo(dnNo),
		receiptOrderRepo.WhereStatusIn([]int{constant.ReceiptOrderStatusReceived}),
		receiptOrderRepo.WithItems(),
	)
	if err != nil {
		return result, err
	}

	// 计算每个物品每个单位的已收货数量
	// key: "item_code:uom", value: 已收货数量
	receivedQtyMap := make(map[string]float64)
	for _, receipt := range receiptOrders {
		if receipt.Status != constant.ReceiptOrderStatusReceived {
			continue
		}
		for _, item := range receipt.Items {
			if len(item.Units) > 0 {
				for _, unit := range item.Units {
					key := item.MaterialCode + ":" + unit.ErpnextUom
					receivedQtyMap[key] += unit.Num
				}
			} else {
				key := item.MaterialCode + ":" + item.ErpnextUom
				receivedQtyMap[key] += item.Num
			}
		}
	}

	// 构建物品编码到采购单物品的映射
	itemCodeMap := make(map[string]*model.PurchaseOrderItem)
	for i := range purchaseOrder.Items {
		item := &purchaseOrder.Items[i]
		itemCodeMap[item.MaterialCode] = item
	}

	// 遍历DN物品（按item_code分组后），构建待收货物品信息
	for itemCode, dnUnits := range dnItemUnitsMap {
		purchaseItem, exists := itemCodeMap[itemCode]
		if !exists {
			continue
		}

		// // 检查该物品是否有剩余可收货数量
		// hasRemaining := false
		// for _, dnUnit := range dnUnits {
		// 	key := itemCode + ":" + dnUnit.Uom
		// 	receivedQty := receivedQtyMap[key]
		// 	if dnUnit.Qty-receivedQty > 0 {
		// 		hasRemaining = true
		// 		break
		// 	}
		// }
		// if !hasRemaining {
		// 	continue
		// }

		// 构建待收货物品信息
		pendingItem := s.buildDNPendingItemInfo(purchaseItem, dnUnits, receivedQtyMap)
		result.Items = append(result.Items, pendingItem)
	}

	return result, nil
}

// dnUnitInfo DN单位信息（用于buildDNPendingItemInfo）
type dnUnitInfo struct {
	Qty float64
	Uom string
}

// buildDNPendingItemInfo 构建DN类型的待收货物品信息
// Units 基于DN的单位，支持同一物品多个单位
func (s *purchaseOrderSrv) buildDNPendingItemInfo(
	item *model.PurchaseOrderItem,
	dnUnits []dnUnitInfo,
	receivedQtyMap map[string]float64, // key: "item_code:uom"
) resp.ReceiptPendingItemInfo {
	// 计算总采购数量、总到货数量、总剩余数量
	var totalPurchaseNum, totalArrivalNum, totalPendingNum float64
	for _, dnUnit := range dnUnits {
		key := item.MaterialCode + ":" + dnUnit.Uom
		receivedQty := receivedQtyMap[key]
		pendingQty := dnUnit.Qty - receivedQty
		if pendingQty < 0 {
			pendingQty = 0
		}
		totalPurchaseNum += dnUnit.Qty
		totalArrivalNum += receivedQty
		totalPendingNum += pendingQty
	}

	pendingItem := resp.ReceiptPendingItemInfo{
		PurchaseOrderItemUuid: item.Uuid,
		MaterialUuid:          item.MaterialUuid,
		MaterialCode:          item.MaterialCode,
		LocaleName:            *language.JsonToLocaleResponse(item.MaterialName),
		PurchaseNum:           totalPurchaseNum,
		ArrivalNum:            totalArrivalNum,
		Num:                   totalPendingNum,
		UnitUuid:              item.UnitUuid,
		LocaleUnitName:        *language.JsonToLocaleResponse(item.UnitName),
		UnitConversionRate:    item.UnitConversionRate,
		BaseUnitUuid:          item.BaseUnitUuid,
		LocaleBaseUnitName:    *language.JsonToLocaleResponse(item.BaseUnitName),
		UnitList:              make([]resp.PurchaseOrderItemMaterialUnit, 0),
		Units:                 make([]resp.ReceiptPendingItemUnit, 0),
	}

	// 构建可选单位列表（从 Material 关联获取）
	if item.Material != nil && len(item.Material.NotBaseUnitList) > 0 {
		for _, unit := range item.Material.NotBaseUnitList {
			pendingItem.UnitList = append(pendingItem.UnitList, resp.PurchaseOrderItemMaterialUnit{
				Uuid:           unit.Uuid,
				ConversionRate: unit.ConversionRate,
				LocaleName:     *language.JsonToLocaleResponse(unit.Name),
			})
		}
	}

	// 构建DN每个单位的待收货信息
	for _, dnUnit := range dnUnits {
		key := item.MaterialCode + ":" + dnUnit.Uom
		receivedQty := receivedQtyMap[key]
		pendingQty := dnUnit.Qty - receivedQty
		if pendingQty < 0 {
			pendingQty = 0
		}

		// 查找与DN单位匹配的采购单物品单位信息
		var unitUuid uint64
		var unitName string
		var conversionRate float64 = 1.0

		if item.Material != nil {
			for _, unit := range item.Material.NotBaseUnitList {
				if unit.Unit != nil && unit.Unit.ErpnextUom == dnUnit.Uom {
					unitUuid = unit.Uuid
					unitName = unit.Name
					conversionRate = unit.ConversionRate
					break
				}
			}
		}

		pendingItem.Units = append(pendingItem.Units, resp.ReceiptPendingItemUnit{
			UnitUuid:       unitUuid,
			LocaleUnitName: *language.JsonToLocaleResponse(unitName),
			ConversionRate: conversionRate,
			PurchaseNum:    dnUnit.Qty,
			ArrivalNum:     receivedQty,
			Num:            dnUnit.Qty, // Num与PurchaseNum一致
		})
	}

	return pendingItem
}

// getSupplierPendingItems 获取供应商直接发货类型的待收货物品
func (s *purchaseOrderSrv) getSupplierPendingItems(
	_ context.Context,
	db *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
	supplierErpCode string,
) (resp.ReceiptPendingItemsResp, error) {
	result := resp.ReceiptPendingItemsResp{
		IsDeliveryNote:      false,
		DeliveryNoteNo:      "",
		SupplierErpCode:     supplierErpCode,
		LocaleWarehouseName: *language.JsonToLocaleResponse(purchaseOrder.WarehouseName),
		OrderNo:             purchaseOrder.OrderNo,
		Items:               make([]resp.ReceiptPendingItemInfo, 0),
	}

	// 获取供应商信息
	supplierRepo := repository.NewSupplierRepo(db)
	supplier, _ := supplierRepo.GetByErpCode(supplierErpCode)
	if supplier != nil {
		result.SupplierName = supplier.Name
	}

	// 筛选属于该供应商的物品（delivered_by_supplier=1 且 supplier_erp_code 匹配）
	for i := range purchaseOrder.Items {
		item := &purchaseOrder.Items[i]
		if item.Material == nil {
			continue
		}
		if item.Material.DeliveredBySupplier != 1 {
			continue
		}
		if item.Material.SupplierErpCode != supplierErpCode {
			continue
		}

		// 判断是否有待收货数量（只基于 ttpos_purchase_order_item_unit）
		hasPendingQty := false
		for _, unit := range item.Units {
			if unit.ArrivalNum < unit.Num {
				hasPendingQty = true
				break
			}
		}
		if !hasPendingQty {
			continue
		}

		// 构建待收货物品信息
		pendingItem := s.buildPendingItemInfoFromUnits(item)
		result.Items = append(result.Items, pendingItem)
	}

	return result, nil
}

// buildPendingItemInfo 构建待收货物品信息
func (s *purchaseOrderSrv) buildPendingItemInfo(
	item *model.PurchaseOrderItem,
	purchaseNum, arrivalNum, pendingNum float64,
) resp.ReceiptPendingItemInfo {
	pendingItem := resp.ReceiptPendingItemInfo{
		PurchaseOrderItemUuid: item.Uuid,
		MaterialUuid:          item.MaterialUuid,
		MaterialCode:          item.MaterialCode,
		LocaleName:            *language.JsonToLocaleResponse(item.MaterialName),
		PurchaseNum:           purchaseNum,
		ArrivalNum:            arrivalNum,
		Num:                   pendingNum,
		UnitUuid:              item.UnitUuid,
		LocaleUnitName:        *language.JsonToLocaleResponse(item.UnitName),
		UnitConversionRate:    item.UnitConversionRate,
		BaseUnitUuid:          item.BaseUnitUuid,
		LocaleBaseUnitName:    *language.JsonToLocaleResponse(item.BaseUnitName),
		UnitList:              make([]resp.PurchaseOrderItemMaterialUnit, 0),
		Units:                 make([]resp.ReceiptPendingItemUnit, 0),
	}

	// 构建可选单位列表（从 Material 关联获取）
	if item.Material != nil && len(item.Material.NotBaseUnitList) > 0 {
		for _, unit := range item.Material.NotBaseUnitList {
			pendingItem.UnitList = append(pendingItem.UnitList, resp.PurchaseOrderItemMaterialUnit{
				Uuid:           unit.Uuid,
				ConversionRate: unit.ConversionRate,
				LocaleName:     *language.JsonToLocaleResponse(unit.Name),
			})
		}
	}

	// 构建多单位待收货信息
	if len(item.Units) > 0 {
		for _, unit := range item.Units {
			unitPendingNum := unit.Num - unit.ArrivalNum
			if unitPendingNum < 0 {
				unitPendingNum = 0
			}
			pendingItem.Units = append(pendingItem.Units, resp.ReceiptPendingItemUnit{
				UnitUuid:       unit.UnitUuid,
				LocaleUnitName: *language.JsonToLocaleResponse(unit.UnitName),
				ConversionRate: unit.UnitConversionRate,
				PurchaseNum:    unit.Num,
				ArrivalNum:     unit.ArrivalNum,
				Num:            unitPendingNum,
			})
		}
	}

	return pendingItem
}

// buildPendingItemInfoFromUnits 基于 ttpos_purchase_order_item_unit 构建待收货物品信息
func (s *purchaseOrderSrv) buildPendingItemInfoFromUnits(item *model.PurchaseOrderItem) resp.ReceiptPendingItemInfo {
	// 计算总采购数量、已到货数量、待收货数量（基于 Units）
	var totalPurchaseNum, totalArrivalNum, totalPendingNum float64
	for _, unit := range item.Units {
		totalPurchaseNum += unit.Num
		totalArrivalNum += unit.ArrivalNum
		pendingNum := unit.Num - unit.ArrivalNum
		if pendingNum > 0 {
			totalPendingNum += pendingNum
		}
	}

	pendingItem := resp.ReceiptPendingItemInfo{
		PurchaseOrderItemUuid: item.Uuid,
		MaterialUuid:          item.MaterialUuid,
		MaterialCode:          item.MaterialCode,
		LocaleName:            *language.JsonToLocaleResponse(item.MaterialName),
		PurchaseNum:           totalPurchaseNum,
		ArrivalNum:            totalArrivalNum,
		Num:                   totalPendingNum,
		UnitList:              make([]resp.PurchaseOrderItemMaterialUnit, 0),
		Units:                 make([]resp.ReceiptPendingItemUnit, 0),
	}

	// 构建可选单位列表（从 Material 关联获取）
	if item.Material != nil && len(item.Material.NotBaseUnitList) > 0 {
		for _, unit := range item.Material.NotBaseUnitList {
			pendingItem.UnitList = append(pendingItem.UnitList, resp.PurchaseOrderItemMaterialUnit{
				Uuid:           unit.Uuid,
				ConversionRate: unit.ConversionRate,
				LocaleName:     *language.JsonToLocaleResponse(unit.Name),
			})
		}
	}

	// 构建多单位待收货信息（只使用 ttpos_purchase_order_item_unit 数据）
	for _, unit := range item.Units {
		unitPendingNum := unit.Num - unit.ArrivalNum
		if unitPendingNum < 0 {
			unitPendingNum = 0
		}
		pendingItem.Units = append(pendingItem.Units, resp.ReceiptPendingItemUnit{
			UnitUuid:       unit.UnitUuid,
			LocaleUnitName: *language.JsonToLocaleResponse(unit.UnitName),
			ConversionRate: unit.UnitConversionRate,
			PurchaseNum:    unit.Num,
			ArrivalNum:     unit.ArrivalNum,
			Num:            unitPendingNum,
		})
	}

	return pendingItem
}
