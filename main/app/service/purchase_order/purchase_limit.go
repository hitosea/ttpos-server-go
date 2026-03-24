package purchase_order

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
)

// ====================================================================================
// 品牌采购限购校验方法（基于新的限购方案表）
// ====================================================================================

// checkPurchaseLimit 检查品牌采购限购（新方案）
//
// 逻辑：
//   - 查询所有启用的限购方案
//   - 校验当前星期是否在限购周期内
//   - 校验是否包含禁止采购的物品
//   - 校验每日申请次数限制
//   - 校验物品数量限制
//   - 校验物品销售单位是否变更
func (s *purchaseOrderSrv) checkPurchaseLimit(ctx context.Context, order *model.PurchaseOrder) error {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	// 如果没有总部，跳过限购检查
	if headquarterUuid == 0 {
		return nil
	}

	// 1 检查是否包含禁止采购的物品
	disallowedMaterials := s.helper.getDisallowedPurchaseMaterials(ctx, s.dbm, order)
	if len(disallowedMaterials) > 0 {
		if err := s.checkDisallowedPurchase(ctx, order, disallowedMaterials); err != nil {
			return err
		}
	}

	// 2 检查每日申请次数限制
	minDailyLimit := s.helper.getMinDailyLimit(ctx, s.dbm, order)
	if minDailyLimit != -1 {
		if order.IsStorePending() && minDailyLimit > 0 {
			minDailyLimit += 1
		}
		if err := s.checkDailyLimitByScheme(ctx, minDailyLimit); err != nil {
			return err
		}
	}

	// 3 检查物品数量限制（最大和最小采购数量）
	quotaConfigMap := s.helper.getQuotaLimitMap(ctx, s.dbm, order)
	if len(quotaConfigMap) > 0 {
		// 3.1 检查最大采购数量
		if err := s.checkItemLimitByScheme(ctx, order, quotaConfigMap); err != nil {
			return err
		}
		// 3.2 检查最小采购数量
		if err := s.checkMinItemLimitByScheme(ctx, order, quotaConfigMap); err != nil {
			return err
		}
	}

	// 4 检查物品销售单位是否变更（仅在用户未确认时检查）
	if err := s.checkSalesUnitChanged(ctx, order); err != nil {
		return err
	}

	return nil
}

// isWeekdayInScheme 检查当前星期是否在方案的限购周期内
func (s *purchaseOrderSrv) isWeekdayInScheme(currentWeekday int8, weekdaysStr string) bool {
	if weekdaysStr == "" {
		return false
	}

	// 解析星期配置（逗号分隔，如 "1,3,5"）
	weekdayParts := strings.Split(weekdaysStr, ",")
	for _, part := range weekdayParts {
		var weekday int8
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &weekday); err == nil {
			if weekday == currentWeekday {
				return true
			}
		}
	}

	return false
}

// checkDisallowedPurchase 检查是否包含禁止采购的物品
func (s *purchaseOrderSrv) checkDisallowedPurchase(
	ctx context.Context,
	order *model.PurchaseOrder,
	disallowedMaterials []string,
) error {
	lang := ctx.GetLanguage()

	// 构建禁止采购物品编码集合
	disallowedSet := make(map[string]bool)
	for _, code := range disallowedMaterials {
		disallowedSet[code] = true
	}

	// 检查订单中是否包含禁止采购的物品
	var disallowedNames []string
	for _, item := range order.Items {
		if disallowedSet[item.MaterialCode] {
			materialName := language.JsonToLocaleResponse(item.MaterialName).GetLocale(lang)
			disallowedNames = append(disallowedNames, materialName)
		}
	}

	if len(disallowedNames) > 0 {
		return errors.NewWithCode(
			constant.CodeErrorConfirmRefresh,
			fmt.Sprintf(i18n.Translate(lang, "物品 %s 禁止采购，请移除后重试"), strings.Join(disallowedNames, "、")),
		)
	}

	return nil
}

// checkDailyLimitByScheme 检查每日申请次数限制（按方案）
func (s *purchaseOrderSrv) checkDailyLimitByScheme(ctx context.Context, dailyLimit int) error {
	lang := ctx.GetLanguage()
	companySetting := ctx.GetCompanySetting()

	// 使用系统时区计算当天的起止时间戳
	todayStart, todayEnd := utils.SetTimezone(companySetting.GetTimezone()).TodayStartEndUnix()

	// 通过 Repository 统计当天已提交的申请次数
	count, err := repository.NewPurchaseOrderRepo(ctx.GetDB()).CountBrandPurchaseByTimeRange(todayStart, todayEnd)
	if err != nil {
		return errors.WithMessage(errors.New("查询每日申请次数失败"), err.Error())
	}

	// 校验是否超限
	if count >= int64(dailyLimit) {
		return errors.New(fmt.Sprintf(i18n.Translate(lang, "今日申请次数已达上限（%s次），请明天再试"), fmt.Sprintf("%d", dailyLimit)))
	}

	return nil
}

// checkSalesUnitChanged 检查物品销售单位是否已变更
// 比较采购单创建时的单位与物品当前的默认销售单位
func (s *purchaseOrderSrv) checkSalesUnitChanged(
	ctx context.Context,
	order *model.PurchaseOrder,
) error {
	lang := ctx.GetLanguage()

	var changedNames []string
	for _, item := range order.Items {
		if item.Num <= 0 || item.Material == nil {
			continue
		}
		if item.Material.DefaultSalesUnitUuid > 0 {
			// 检查采购单物品的各个单位是否与当前默认销售单位一致
			for _, unit := range item.Units {
				if unit.UnitUuid != item.Material.DefaultSalesUnitUuid {
					materialName := language.JsonToLocaleResponse(item.MaterialName).GetLocale(lang)
					changedNames = append(changedNames, materialName)
					break // 只要有一个单位变更就记录，避免重复
				}
			}
		}
	}

	if len(changedNames) > 0 {
		return errors.NewWithCode(
			constant.CodeErrorConfirmRefresh,
			fmt.Sprintf(i18n.Translate(lang, "物品 %s 单位变动，请检查物品和数量"), strings.Join(changedNames, "、")),
		)
	}

	return nil
}

// checkItemLimitByScheme 检查物品数量限制（按方案）
func (s *purchaseOrderSrv) checkItemLimitByScheme(
	ctx context.Context,
	order *model.PurchaseOrder,
	quotaConfigMap map[string]repository.QuotaLimitConfig,
) error {
	lang := ctx.GetLanguage()

	// 3. 汇总订单中的物品数量（按 MaterialCode 分组）
	type MaterialSummary struct {
		MaterialCode string
		MaterialName string
		TotalQty     float64
	}

	materialSummaryMap := make(map[string]*MaterialSummary)

	// 遍历订单明细
	for _, orderItem := range order.Items {
		// 检查是否在限购配置中，且有最大限购数量
		quotaConfig, inQuota := quotaConfigMap[orderItem.MaterialCode]
		if !inQuota || quotaConfig.QuotaLimit <= 0 {
			continue // 不在限购配置中或无最大限购数量，跳过
		}
		// 获取限购配置单位
		if orderItem.Material == nil {
			continue // Material 未加载，跳过
		}
		quotaUnit := orderItem.Material.GetUnitByUuidForQuotaConfig()
		if quotaUnit == nil {
			continue // 未找到限购配置单位，跳过
		}
		quotaUnitUuid := quotaUnit.Uuid

		materialName := language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(lang)

		// 累加单位申请数量
		for _, unit := range orderItem.Units {
			// 只累加限购单位的数量
			if unit.UnitUuid != quotaUnitUuid {
				return errors.NewWithCode(constant.CodeErrorConfirmRefresh, fmt.Sprintf(i18n.Translate(lang, "物品 %s 超出限购/单位变动，请检查物品和数量"), materialName))
			}
			key := orderItem.MaterialCode
			if summary, exists := materialSummaryMap[key]; exists {
				summary.TotalQty += unit.Num
			} else {
				materialSummaryMap[key] = &MaterialSummary{
					MaterialCode: orderItem.MaterialCode,
					MaterialName: materialName,
					TotalQty:     unit.Num,
				}
			}
		}
	}

	// 4. 将 map 转换为切片并排序
	summaries := make([]*MaterialSummary, 0, len(materialSummaryMap))
	for _, summary := range materialSummaryMap {
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].MaterialCode < summaries[j].MaterialCode
	})

	// 5. 逐个检查物品是否超限
	for _, summary := range summaries {
		// 获取限购限制（前面已经过滤，这里肯定存在）
		quotaLimit := quotaConfigMap[summary.MaterialCode].QuotaLimit
		// 校验是否超限
		if summary.TotalQty > quotaLimit {
			return errors.NewWithCode(constant.CodeErrorConfirmRefresh, fmt.Sprintf(i18n.Translate(lang, "物品 %s 超出限购/单位变动，请检查物品和数量"), summary.MaterialName))
		}
	}

	return nil
}

// checkMinItemLimitByScheme 检查物品最小采购数量限制（按方案）
func (s *purchaseOrderSrv) checkMinItemLimitByScheme(
	ctx context.Context,
	order *model.PurchaseOrder,
	quotaConfigMap map[string]repository.QuotaLimitConfig,
) error {
	lang := ctx.GetLanguage()

	// 1. 定义物品汇总结构
	type MaterialSummary struct {
		MaterialCode string
		MaterialName string
		TotalQty     float64
	}

	materialSummaryMap := make(map[string]*MaterialSummary)

	// 2. 遍历订单明细，汇总物品数量
	for _, orderItem := range order.Items {
		// 检查是否在限购配置中，且有最小采购数量
		quotaConfig, inQuota := quotaConfigMap[orderItem.MaterialCode]
		if !inQuota || quotaConfig.MinQuotaLimit <= 0 {
			continue // 不在配置中或无最小采购数量，跳过
		}
		if orderItem.Material == nil {
			continue // Material 未加载，跳过
		}
		// 获取限购配置单位
		quotaUnit := orderItem.Material.GetUnitByUuidForQuotaConfig()
		if quotaUnit == nil {
			continue // 未找到限购配置单位，跳过
		}
		quotaUnitUuid := quotaUnit.Uuid

		materialName := language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(lang)

		// 累加单位申请数量
		for _, unit := range orderItem.Units {
			// 只累加限购单位的数量
			if unit.UnitUuid != quotaUnitUuid {
				continue
			}
			key := orderItem.MaterialCode
			if summary, exists := materialSummaryMap[key]; exists {
				summary.TotalQty += unit.Num
			} else {
				materialSummaryMap[key] = &MaterialSummary{
					MaterialCode: orderItem.MaterialCode,
					MaterialName: materialName,
					TotalQty:     unit.Num,
				}
			}
		}
	}

	// 3. 逐个检查物品是否低于最小采购数量
	for _, summary := range materialSummaryMap {
		minQuotaLimit := quotaConfigMap[summary.MaterialCode].MinQuotaLimit
		if summary.TotalQty < minQuotaLimit {
			// 错误提示：物品 {name} 申请总数（{actual}），不能小于（{min}），请调整后提交
			return errors.NewWithCode(
				constant.CodeErrorConfirmRefresh,
				fmt.Sprintf(
					i18n.Translate(lang, "物品 %s 申请总数（%s），不能小于（%s），请调整后提交"),
					summary.MaterialName,
					formatQuantity(summary.TotalQty),
					formatQuantity(minQuotaLimit),
				),
			)
		}
	}

	return nil
}

// formatQuantity 格式化数量，去掉不必要的小数位
func formatQuantity(qty float64) string {
	if qty == float64(int64(qty)) {
		return fmt.Sprintf("%d", int64(qty))
	}
	return fmt.Sprintf("%.2f", qty)
}

// checkItemDefaultWarehouse 校验品牌采购物品是否在 ERPNext 中配置了默认发货仓
// 通过查询 ERPNext 物品的 item_defaults，检查是否存在 company 为总部 erpnext_company_abbr 的记录
// 返回未配置默认仓库的物品描述列表（格式：物品名称（内部编码））
func (s *purchaseOrderSrv) checkItemDefaultWarehouse(ctx context.Context, purchaseOrder *model.PurchaseOrder) []string {
	if len(purchaseOrder.Items) == 0 {
		return nil
	}

	// 获取总部的 erpnext_company_abbr
	companySetting := ctx.GetCompanySetting()
	companyAbbr := companySetting.ErpnextCompanyAbbr

	// 收集物品编码（跳过物品已删除或供应商直发的物品，以物品主表为准）
	itemCodes := make([]string, 0, len(purchaseOrder.Items))
	for _, item := range purchaseOrder.Items {
		if item.MaterialCode != "" && item.Material != nil && item.Material.DeliveredBySupplier != 1 {
			itemCodes = append(itemCodes, item.MaterialCode)
		}
	}

	// 从 ERPNext 批量查询物品默认仓库
	erpSrv := erp.NewIErpSrv(s.dbm)
	itemWarehouses, err := erpSrv.GetItemDefaultWarehouses(ctx, itemCodes, companyAbbr, true)
	if err != nil {
		logger.Logger.Warn("checkItemDefaultWarehouse-查询ERPNext物品默认仓库失败",
			zap.Error(err),
			zap.Strings("item_codes", itemCodes),
		)
		return nil
	}

	// 检查未配置默认仓库的物品
	lang := ctx.GetLanguage()
	var noWarehouseItems []string
	for _, item := range purchaseOrder.Items {
		if item.Material == nil || item.Material.DeliveredBySupplier == 1 {
			continue
		}
		if warehouse, ok := itemWarehouses[item.MaterialCode]; !ok || warehouse == "" {
			materialName := language.JsonToLocaleResponse(item.MaterialName).GetLocale(lang)
			code := item.Material.InternalCode
			desc := materialName
			if code != "" {
				desc += "（" + code + "）"
			}
			noWarehouseItems = append(noWarehouseItems, desc)
		}
	}

	return noWarehouseItems
}
