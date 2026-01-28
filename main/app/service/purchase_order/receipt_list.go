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

	// 获取该采购单的所有收货单
	receiptOrders, err := receiptOrderRepo.GetList(
		receiptOrderRepo.WherePurchaseOrderUuid(purchaseOrder.Uuid),
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
		IncludeItems: false,
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

		// 计算是否完成收货
		isCompleted := true
		for _, dnItem := range dn.Items {
			receivedQty := s.calculateReceivedQtyForDNItem(dnItem.ItemCode, relatedReceipts)
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
			IsDeliveryNote:     true,
			DeliveryNoteNo:     dn.Name,
			IsCompleted:        isCompleted,
			SupplierName:       purchaseOrder.SupplierName,
			SupplierErpCode:    purchaseOrder.SupplierErpCode,
			ErpPurchaseOrderNo: purchaseOrder.ErpOrderNo,
			ReceiptOrders:      receiptOrderInfos,
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
			IsDeliveryNote:     false,
			DeliveryNoteNo:     "",
			IsCompleted:        isCompleted,
			SupplierName:       supplierName,
			SupplierErpCode:    supplierCode,
			ErpPurchaseOrderNo: purchaseOrder.ErpOrderNo,
			ReceiptOrders:      receiptOrderInfos,
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
		IsLegacy:           true,
		IsDeliveryNote:     false,
		DeliveryNoteNo:     "",
		IsCompleted:        isCompleted,
		SupplierName:       purchaseOrder.SupplierName,
		SupplierErpCode:    purchaseOrder.SupplierErpCode,
		ErpPurchaseOrderNo: purchaseOrder.ErpOrderNo,
		ReceiptOrders:      receiptOrderInfos,
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
func (s *purchaseOrderSrv) calculateReceivedQtyForDNItem(itemCode string, receipts []model.PurchaseReceiptOrder) float64 {
	totalReceived := 0.0
	for _, receipt := range receipts {
		// 只计算已确认收货的
		if receipt.Status != constant.ReceiptOrderStatusReceived {
			continue
		}
		for _, item := range receipt.Items {
			if item.MaterialCode == itemCode {
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
		IsDeliveryNote:  true,
		DeliveryNoteNo:  dnNo,
		SupplierName:    purchaseOrder.SupplierName,
		SupplierErpCode: purchaseOrder.SupplierErpCode,
		Items:           make([]resp.ReceiptPendingItemInfo, 0),
	}

	companySetting := ctx.GetCompanySetting()
	db := ctx.GetDB()

	// 调用ERP获取DN详情（包含物品）
	erpSrv := erp.NewIErpSrv(s.dbm)
	dnListResp, err := erpSrv.GetDeliveryNoteList(ctx, &delivery_note.GetDeliveryNoteListReq{
		CompanyAbbr:  companySetting.ErpnextHeadquarterAbbr,
		SoNo:         purchaseOrder.ErpSaleOrderNo,
		IncludeItems: true,
	})
	if err != nil {
		return result, err
	}

	// 找到指定的DN
	var targetDN *delivery_note.DeliveryNote
	for _, dn := range dnListResp.DeliveryNoteList {
		if dn.Name == dnNo {
			targetDN = dn
			break
		}
	}
	if targetDN == nil {
		return result, errors.New("未找到指定的DN单据")
	}

	// 构建DN物品编码到物品信息的映射（包含数量和单位）
	type dnItemInfo struct {
		Qty float64
		Uom string
	}
	dnItemMap := make(map[string]dnItemInfo)
	for _, dnItem := range targetDN.Items {
		dnItemMap[dnItem.ItemCode] = dnItemInfo{
			Qty: dnItem.Qty,
			Uom: dnItem.Uom,
		}
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

	// 计算每个物品已收货数量（按DN单位Uom统计）
	receivedQtyMap := make(map[string]float64)
	for _, receipt := range receiptOrders {
		if receipt.Status != constant.ReceiptOrderStatusReceived {
			continue
		}
		for _, item := range receipt.Items {
			dnInfo, exists := dnItemMap[item.MaterialCode]
			if !exists {
				continue
			}
			// 在收货单物品的多单位中查找与DN单位匹配的单位
			if len(item.Units) > 0 {
				for _, unit := range item.Units {
					if unit.ErpnextUom == dnInfo.Uom {
						receivedQtyMap[item.MaterialCode] += unit.Num
					}
				}
			} else if item.ErpnextUom == dnInfo.Uom {
				// 没有多单位时，检查主单位
				receivedQtyMap[item.MaterialCode] += item.Num
			}
		}
	}

	// 构建物品编码到采购单物品的映射
	itemCodeMap := make(map[string]*model.PurchaseOrderItem)
	for i := range purchaseOrder.Items {
		item := &purchaseOrder.Items[i]
		itemCodeMap[item.MaterialCode] = item
	}

	// 遍历DN物品，计算待收货数量
	for _, dnItem := range targetDN.Items {
		purchaseItem, exists := itemCodeMap[dnItem.ItemCode]
		if !exists {
			continue
		}

		// 计算该物品的已收货数量（按DN单位）
		receivedQty := receivedQtyMap[dnItem.ItemCode]
		pendingQty := dnItem.Qty - receivedQty
		if pendingQty <= 0 {
			continue
		}

		// 构建待收货物品信息（数量使用DN单位）
		pendingItem := s.buildPendingItemInfo(purchaseItem, dnItem.Qty, receivedQty, pendingQty)
		result.Items = append(result.Items, pendingItem)
	}

	return result, nil
}

// getSupplierPendingItems 获取供应商直接发货类型的待收货物品
func (s *purchaseOrderSrv) getSupplierPendingItems(
	_ context.Context,
	db *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
	supplierErpCode string,
) (resp.ReceiptPendingItemsResp, error) {
	result := resp.ReceiptPendingItemsResp{
		IsDeliveryNote:  false,
		DeliveryNoteNo:  "",
		SupplierErpCode: supplierErpCode,
		Items:           make([]resp.ReceiptPendingItemInfo, 0),
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
		PendingNum:            pendingNum,
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
				PendingNum:     unitPendingNum,
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
		PendingNum:            totalPendingNum,
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
			PendingNum:     unitPendingNum,
		})
	}

	return pendingItem
}
