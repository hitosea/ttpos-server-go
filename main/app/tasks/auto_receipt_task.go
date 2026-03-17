package tasks

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/purchase_order"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"ttpos-bmp/app/ttpos-erp/api/delivery_note"
	"ttpos-server-go/app/service/rpc/erp"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// shopRuleInfo 门店规则信息
type shopRuleInfo struct {
	RuleUuid         uint64
	WarehouseErpCode string
	DelayDays        int
	HeadquarterUuid  uint64
}

// orderProcessDeps 订单处理依赖项（减少函数参数数量）
type orderProcessDeps struct {
	erpSrv           erp.IErpSrv
	purchaseOrderSrv purchase_order.IPurchaseOrderSrv
	logRepo          repository.IAutoReceiptLogRepo
}

// AutoReceiptTask 品采自动收货定时任务
// 每小时执行一次，检查门店是否到达本地午夜，满足条件则自动收货
type AutoReceiptTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewAutoReceiptTask(dbm *database.DBManager, cache cache.Cache) *AutoReceiptTask {
	return &AutoReceiptTask{dbm: dbm, cache: cache}
}

// Execute 执行定时任务
func (t *AutoReceiptTask) Execute() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("自动收货定时任务panic", zap.Any("error", r))
		}
	}()

	logger.Logger.Info("开始执行自动收货定时任务")

	start := time.Now()
	lock.NewSystemLock().LockUuid(lock.AutoReceiptLock)
	defer lock.NewSystemLock().UnlockUuid(lock.AutoReceiptLock)
	if time.Since(start) > 1*time.Second {
		logger.Logger.Warn("自动收货任务: 其他节点已处理，跳过")
		return
	}

	saasDB := t.dbm.GetDB(constant.DefaultDB)
	if saasDB == nil {
		logger.Logger.Error("自动收货任务: 无法获取saas主库连接")
		return
	}

	// 获取所有启用的规则+门店
	shopRepo := repository.NewAutoReceiptRuleShopRepo(saasDB)
	ruleShops, err := shopRepo.GetAllEnabledWithRules()
	if err != nil {
		logger.Logger.Error("自动收货任务: 查询启用规则失败", zap.Error(err))
		return
	}
	if len(ruleShops) == 0 {
		return
	}

	// 按门店分组规则
	shopRulesMap := groupRulesByShop(ruleShops)

	settingSrv := setting.NewSrvImpl(t.dbm, t.cache)

	for shopUuid, rules := range shopRulesMap {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Logger.Error("自动收货任务: 门店处理panic",
						zap.Uint64("company_uuid", shopUuid), zap.Any("error", r), zap.Stack("stack_trace"))
				}
			}()
			t.processShop(shopUuid, rules, settingSrv, saasDB)
		}()
	}

	logger.Logger.Info("自动收货定时任务执行完成", zap.Int("shop_count", len(shopRulesMap)))
}

// processShop 处理单个门店的自动收货
func (t *AutoReceiptTask) processShop(shopUuid uint64, rules []shopRuleInfo, settingSrv *setting.Srv, saasDB *gorm.DB) {
	shopDB := t.dbm.GetDB(shopUuid)
	// 1. 获取门店 Company 信息（含 CompanySetting）
	companyRepo := repository.NewCompanyRepo(shopDB)
	companyPtr, err := companyRepo.GetCompanyInfoByUuid(shopUuid)
	if err != nil {
		logger.Logger.Error("自动收货任务: 获取门店信息失败",
			zap.Uint64("company_uuid", shopUuid), zap.Error(err))
		return
	}
	company := *companyPtr
	if !company.IsOpenErp() || company.CompanySetting == nil {
		return
	}

	// 2. 从 CompanySetting 获取时区，检查是否到达本地午夜
	timezone := company.CompanySetting.GetTimezone()
	if utils.SetTimezone(timezone).Now().Hour() != 0 {
		return
	}

	logger.Logger.Info("自动收货任务: 门店到达本地午夜",
		zap.Uint64("company_uuid", shopUuid), zap.String("timezone", timezone))

	// 3. 获取门店采购订单（status=2 已通过），预加载明细+单位
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(shopDB)
	orders, err := purchaseOrderRepo.GetList(purchaseOrderRepo.WhereStatus(constant.PurchaseOrderStatusApproved), purchaseOrderRepo.WherePurchaseType(constant.PurchaseTypeBrand), purchaseOrderRepo.WithSimpleItems())
	if err != nil {
		logger.Logger.Error("自动收货任务: 查询采购订单失败",
			zap.Uint64("company_uuid", shopUuid), zap.Error(err))
		return
	}
	if len(orders) == 0 {
		return
	}

	// 4. 创建上下文
	ctx := context.NewContext(
		context.WithCompanyUuid(shopUuid),
		context.WithCompany(company),
		context.WithCompanySetting(*company.CompanySetting),
	)
	ctx.SetDB(shopDB)

	deps := orderProcessDeps{
		erpSrv:           erp.NewIErpSrv(t.dbm),
		purchaseOrderSrv: purchase_order.NewPurchaseOrderSrvImpl(t.dbm, settingSrv),
		logRepo:          repository.NewAutoReceiptLogRepo(saasDB),
	}

	// 5. 解析时区 Location（一次解析，传递给后续方法）
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc, _ = time.LoadLocation(string(utils.ZH_TIMEZONE))
	}

	// 6. 处理每个采购订单
	receiptRepo := repository.NewPurchaseReceiptOrderRepo(shopDB)
	for i := range orders {
		order := &orders[i]
		if order.ErpSaleOrderNo == "" {
			continue
		}

		// 获取已确认收货单（status=1）的明细，用于计算已收数量
		confirmedReceipts, err := receiptRepo.GetList(
			receiptRepo.WherePurchaseOrderUuid(order.Uuid),
			receiptRepo.WhereStatusIn([]int{constant.ReceiptOrderStatusReceived}),
			receiptRepo.WithItems(),
		)
		if err != nil {
			logger.Logger.Error("自动收货任务: 查询已收货记录失败",
				zap.Uint64("company_uuid", shopUuid),
				zap.Uint64("purchase_order_uuid", order.Uuid),
				zap.Error(err))
			continue
		}

		t.processOrder(ctx, order, rules, loc, shopUuid, deps, confirmedReceipts)
	}
}

// processOrder 处理单个采购订单的 DN 自动收货
func (t *AutoReceiptTask) processOrder(
	ctx context.Context,
	order *model.PurchaseOrder,
	rules []shopRuleInfo,
	loc *time.Location,
	shopUuid uint64,
	deps orderProcessDeps,
	confirmedReceipts []model.PurchaseReceiptOrder,
) {
	// 获取 DN 列表
	dnListResp, err := deps.erpSrv.GetDeliveryNoteList(ctx, &delivery_note.GetDeliveryNoteListReq{
		CompanyAbbr:  ctx.GetCompany().CompanySetting.ErpnextHeadquarterAbbr,
		SoNo:         order.ErpSaleOrderNo,
		IncludeItems: true,
	})
	if err != nil {
		logger.Logger.Error("自动收货任务: 获取DN列表失败",
			zap.Uint64("company_uuid", shopUuid),
			zap.String("so_no", order.ErpSaleOrderNo),
			zap.Error(err))
		return
	}
	if dnListResp == nil || len(dnListResp.DeliveryNoteList) == 0 {
		return
	}

	// 4层过滤处理每个 DN
	for _, dn := range dnListResp.DeliveryNoteList {
		// 第1层: DN 状态过滤（只处理 "To Bill" 即 docstatus=1）
		if dn.Status != "To Bill" || dn.Docstatus != 1 {
			continue
		}

		// 第2层: 仓库匹配
		matchedRule := findMatchingRule(dn.SetWarehouse, rules)
		if matchedRule == nil {
			continue
		}

		// 第3层: 时间判断
		if !shouldAutoReceipt(dn.PostingDate, matchedRule.DelayDays, loc) {
			continue
		}

		// 第4层: 按本 DN 已收数量计算待收物品
		arrivedQtyMap := buildArrivedQtyMap(confirmedReceipts, dn.Name)
		pendingItems := calculatePendingItems(dn, order.Items, arrivedQtyMap)
		if len(pendingItems) == 0 {
			continue
		}

		// 满足所有条件，执行自动收货
		t.executeAutoReceipt(ctx, order, dn, pendingItems, matchedRule, shopUuid, loc, deps)
	}
}

// executeAutoReceipt 执行自动收货
func (t *AutoReceiptTask) executeAutoReceipt(
	ctx context.Context,
	order *model.PurchaseOrder,
	dn *delivery_note.DeliveryNote,
	pendingItems []req.PurchaseReceiptItemCreateReq,
	rule *shopRuleInfo,
	shopUuid uint64,
	loc *time.Location,
	deps orderProcessDeps,
) {
	now := time.Now()
	receiptReq := req.PurchaseReceiptCreateReq{
		PurchaseOrderUuid: order.Uuid,
		ReceiveTime:       now.Unix(),
		ReceiptType:       constant.ReceiptTypeInternal,
		Items:             pendingItems,
		IsConfirm:         true,
		DeliveryNoteNo:    dn.Name,
		IsAutoReceipt:     true,
	}

	result, err := deps.purchaseOrderSrv.CreatePurchaseReceiptOrder(ctx, receiptReq)
	if err != nil {
		logger.Logger.Error("自动收货任务: 创建收货单失败",
			zap.Uint64("company_uuid", shopUuid),
			zap.String("dn_name", dn.Name),
			zap.Error(err))
		return
	}

	// 记录日志（使用门店时区的当天0点）
	receiptTime, _ := utils.Timezone(loc.String()).TodayStartEndUnix()
	_, err = deps.logRepo.Create(model.AutoReceiptLog{
		HeadquarterCompanyUuid: rule.HeadquarterUuid,
		RuleUuid:               rule.RuleUuid,
		ShopCompanyUuid:        shopUuid,
		ReceiptOrderUuid:       result.Uuid,
		ReceiptOrderNo:         result.OrderNo,
		ReceiptErpOrderNo:      dn.Name,
		ReceiptTime:            receiptTime,
	})
	if err != nil {
		logger.Logger.Error("自动收货任务: 记录日志失败",
			zap.Uint64("company_uuid", shopUuid), zap.Error(err))
	}

	logger.Logger.Info("自动收货任务: 成功创建收货单",
		zap.Uint64("company_uuid", shopUuid),
		zap.String("dn_name", dn.Name),
		zap.Uint64("receipt_uuid", result.Uuid))
}

// groupRulesByShop 按门店UUID分组规则（从 JOIN 结果转换为门店 → 规则列表映射）
func groupRulesByShop(ruleShops []repository.RuleShopJoinResult) map[uint64][]shopRuleInfo {
	shopRulesMap := make(map[uint64][]shopRuleInfo)
	for _, rs := range ruleShops {
		shopRulesMap[rs.ShopCompanyUuid] = append(shopRulesMap[rs.ShopCompanyUuid], shopRuleInfo{
			RuleUuid:         rs.RuleUuid,
			WarehouseErpCode: rs.WarehouseErpCode,
			DelayDays:        rs.DelayDays,
			HeadquarterUuid:  rs.HeadquarterUuid,
		})
	}
	return shopRulesMap
}

// shouldAutoReceipt 判断 DN 是否达到自动收货条件
func shouldAutoReceipt(dnPostingDate string, delayDays int, loc *time.Location) bool {
	postingDate, err := time.ParseInLocation("2006-01-02", dnPostingDate, loc)
	if err != nil {
		return false
	}
	// 收货截止 = 过账日期 + delayDays+1 天的 0:00
	deadline := postingDate.AddDate(0, 0, delayDays+1)
	return !time.Now().In(loc).Before(deadline)
}

// findMatchingRule 查找匹配仓库的规则
func findMatchingRule(dnWarehouse string, rules []shopRuleInfo) *shopRuleInfo {
	for i := range rules {
		if rules[i].WarehouseErpCode == dnWarehouse {
			return &rules[i]
		}
	}
	return nil
}

// buildArrivedQtyMap 构建指定 DN 的已收数量映射 key="material_code:erpnext_uom"
// 仅统计 DeliveryNoteNo 匹配的已确认收货单
func buildArrivedQtyMap(confirmedReceipts []model.PurchaseReceiptOrder, dnName string) map[string]float64 {
	qtyMap := make(map[string]float64)
	for _, receipt := range confirmedReceipts {
		if receipt.DeliveryNoteNo != dnName {
			continue
		}
		for _, item := range receipt.Items {
			addItemArrivedQty(qtyMap, item)
		}
	}
	return qtyMap
}

// materialUomKey 生成 "material_code:erpnext_uom" 格式的映射键
func materialUomKey(materialCode, uom string) string {
	return fmt.Sprintf("%s:%s", materialCode, uom)
}

// addItemArrivedQty 将单个收货明细的数量累加到映射中
func addItemArrivedQty(qtyMap map[string]float64, item model.PurchaseReceiptOrderItem) {
	if len(item.Units) > 0 {
		// 多单位: 按各单位的 ErpnextUom 分别统计
		for _, unit := range item.Units {
			if unit.ErpnextUom != "" {
				qtyMap[materialUomKey(item.MaterialCode, unit.ErpnextUom)] += unit.Num
			}
		}
		return
	}
	// 单单位: 使用明细自身的 ErpnextUom
	if item.MaterialCode != "" && item.ErpnextUom != "" {
		qtyMap[materialUomKey(item.MaterialCode, item.ErpnextUom)] += item.Num
	}
}

// calculatePendingItems 计算待收物品（DN数量 - 已收数量 > 0）
// 将 DN Item 的 ItemCode 匹配采购订单明细的 MaterialCode，构建收货请求
func calculatePendingItems(dn *delivery_note.DeliveryNote, orderItems []model.PurchaseOrderItem, arrivedQtyMap map[string]float64) []req.PurchaseReceiptItemCreateReq {
	// 按 MaterialCode 索引采购订单明细（便于 O(1) 匹配 DN ItemCode）
	orderItemMap := make(map[string]*model.PurchaseOrderItem, len(orderItems))
	for i := range orderItems {
		if orderItems[i].MaterialCode != "" {
			orderItemMap[orderItems[i].MaterialCode] = &orderItems[i]
		}
	}

	var items []req.PurchaseReceiptItemCreateReq
	for _, dnItem := range dn.Items {
		// 匹配采购订单明细
		orderItem, ok := orderItemMap[dnItem.ItemCode]
		if !ok {
			logger.Logger.Warn("自动收货任务: DN物料在采购订单中未找到，跳过",
				zap.String("dn_name", dn.Name),
				zap.String("item_code", dnItem.ItemCode))
			continue
		}

		// 计算待收数量
		key := materialUomKey(dnItem.ItemCode, dnItem.Uom)
		arrived := arrivedQtyMap[key]
		pendingQty := dnItem.Qty - arrived
		if pendingQty <= 0 {
			continue
		}

		// 查找匹配的单位 UUID（DN Uom → PurchaseOrderItemUnit.ErpnextUom）
		unitUuid, found := findMatchingUnitUuid(orderItem, dnItem.Uom)
		if !found {
			logger.Logger.Warn("自动收货任务: DN物料单位在采购订单中无匹配，跳过",
				zap.String("dn_name", dn.Name),
				zap.String("item_code", dnItem.ItemCode),
				zap.String("dn_uom", dnItem.Uom))
			continue
		}

		items = append(items, req.PurchaseReceiptItemCreateReq{
			PurchaseOrderItemUuid: orderItem.Uuid,
			UnitList: []req.PurchaseReceiptItemMaterialUnitReq{
				{Uuid: unitUuid, Num: pendingQty},
			},
		})
	}
	return items
}

// findMatchingUnitUuid 根据 ErpnextUom 查找采购订单明细中匹配的单位 UUID
func findMatchingUnitUuid(orderItem *model.PurchaseOrderItem, erpnextUom string) (uint64, bool) {
	// 优先从多单位列表中查找
	for _, unit := range orderItem.Units {
		if unit.ErpnextUom == erpnextUom {
			return unit.UnitUuid, true
		}
	}
	// 回退到明细自身的单位（单单位情况）
	if orderItem.ErpnextUom == erpnextUom && orderItem.UnitUuid != 0 {
		return orderItem.UnitUuid, true
	}
	return 0, false
}
