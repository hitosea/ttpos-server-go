package purchase_order

import (
	"fmt"
	"sort"
	"strings"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"
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
//   - 校验每日申请次数限制
//   - 校验物品数量限制
func (s *purchaseOrderSrv) checkPurchaseLimit(ctx context.Context, order *model.PurchaseOrder) error {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	// 如果没有总部，跳过限购检查
	if headquarterUuid == 0 {
		return nil
	}

	// 1 检查每日申请次数限制
	minDailyLimit := s.helper.getMinDailyLimit(ctx, s.dbm, order)
	if minDailyLimit != -1 {
		if order.IsStorePending() && minDailyLimit > 0 {
			minDailyLimit += 1
		}
		if err := s.checkDailyLimitByScheme(ctx, minDailyLimit); err != nil {
			return err
		}
	}

	// 2 检查物品数量限制
	quotaMap := s.helper.getQuotaLimitMap(ctx, s.dbm, order)
	if len(quotaMap) > 0 {
		if err := s.checkItemLimitByScheme(ctx, order, quotaMap); err != nil {
			return err
		}
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
		return errors.New(i18n.Translate(lang, "今日申请次数已达上限（%s次），请明天再试", fmt.Sprintf("%d", dailyLimit)))
	}

	return nil
}

// checkItemLimitByScheme 检查物品数量限制（按方案）
func (s *purchaseOrderSrv) checkItemLimitByScheme(
	ctx context.Context,
	order *model.PurchaseOrder,
	quotaMap map[string]float64,
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
		// 检查是否在限购配置中
		_, inQuota := quotaMap[orderItem.MaterialCode]
		if !inQuota {
			continue // 不在限购配置中，跳过
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
				return errors.NewWithCode(constant.CodeErrorConfirmRefresh, "订单物品超出限购/单位变动，请检查物品和数量。")
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
		quotaLimit := quotaMap[summary.MaterialCode]
		// 校验是否超限
		if summary.TotalQty > quotaLimit {
			return errors.NewWithCode(constant.CodeErrorConfirmRefresh, "订单物品超出限购/单位变动，请检查物品和数量。")
		}
	}

	return nil
}
