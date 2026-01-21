package purchase_order

import (
	"fmt"
	"sort"
	"strings"

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
	lang := ctx.GetLanguage()
	companyUuid := ctx.GetCompanyUuid()
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid

	// 如果没有总部，跳过限购检查
	if headquarterUuid == 0 {
		return nil
	}

	headquarterDb := s.dbm.GetDB(headquarterUuid)

	// 1. 查询所有启用的限购方案（状态=1）
	schemeRepo := repository.NewPurchaseLimitSchemeRepo(headquarterDb)
	schemes, err := schemeRepo.GetActiveSchemes(companyUuid)
	if err != nil {
		return errors.WithMessage(errors.New("查询限购方案失败"), err.Error())
	}

	// 如果没有限购方案，跳过检查
	if len(schemes) == 0 {
		return nil
	}

	// 2. 获取当前星期几（1=周一, 7=周日）
	currentWeekday := int8(utils.SetTimezone(companySetting.GetTimezone()).Now().Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7 // 将周日从 0 转换为 7
	}

	// 3. 遍历所有方案，执行检查
	for _, scheme := range schemes {
		// 3.1 检查当前星期是否在限购周期内
		if !s.isWeekdayInScheme(currentWeekday, scheme.Weekdays) {
			continue // 不在限购周期内，跳过此方案
		}

		// 3.2 检查每日申请次数限制
		if scheme.DailyLimit > 0 {
			if err := s.checkDailyLimitByScheme(ctx, scheme.DailyLimit); err != nil {
				return err
			}
		}

		// 3.3 检查物品数量限制
		if err := s.checkItemLimitByScheme(ctx, order, scheme); err != nil {
			// 获取方案名称
			schemeName := language.JsonToLocaleResponse(scheme.Name).GetLocale(lang)
			return errors.New(fmt.Sprintf(
				i18n.Translate(lang, "限购方案【%s】：%s"),
				schemeName, err.Error(),
			))
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
	scheme *model.PurchaseLimitScheme,
) error {
	lang := ctx.GetLanguage()
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	headquarterDb := s.dbm.GetDB(headquarterUuid)

	// 1. 查询方案的物品配置
	itemRepo := repository.NewPurchaseLimitSchemeItemRepo(headquarterDb)
	schemeItems, err := itemRepo.GetBySchemeUuid(scheme.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("查询限购物品配置失败"), err.Error())
	}

	// 如果方案没有物品配置，跳过检查
	if len(schemeItems) == 0 {
		return nil
	}

	// 2. 构建限购物品映射表（MaterialCode -> QuotaLimit）
	quotaMap := make(map[string]float64)
	for _, item := range schemeItems {
		quotaMap[item.MaterialCode] = item.QuotaLimit
	}

	// 3. 汇总订单中的物品数量（按 MaterialCode 分组）
	type MaterialSummary struct {
		MaterialCode string
		MaterialName string
		TotalQty     float64
	}

	materialSummaryMap := make(map[string]*MaterialSummary)

	// 遍历订单明细
	for _, orderItem := range order.Items {
		materialName := language.JsonToLocaleResponse(orderItem.MaterialName).GetLocale(lang)

		// 如果有多单位，累加所有单位的数量
		if len(orderItem.Units) > 0 {
			for _, unit := range orderItem.Units {
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
		} else {
			// 没有多单位，使用主单位
			key := orderItem.MaterialCode
			if summary, exists := materialSummaryMap[key]; exists {
				summary.TotalQty += orderItem.Num
			} else {
				materialSummaryMap[key] = &MaterialSummary{
					MaterialCode: orderItem.MaterialCode,
					MaterialName: materialName,
					TotalQty:     orderItem.Num,
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
		// 检查是否在限购配置中
		quotaLimit, inQuota := quotaMap[summary.MaterialCode]
		if !inQuota {
			continue // 不在限购配置中，跳过
		}

		// 如果限购数量为 0，表示不限制
		if quotaLimit == 0 {
			continue
		}

		// 校验是否超限
		if summary.TotalQty > quotaLimit {
			return errors.New(fmt.Sprintf(
				i18n.Translate(lang, "物品【%s】本次申请数量 %.2f 已超限（最多 %.2f），请减少数量后重试"),
				summary.MaterialName, summary.TotalQty, quotaLimit,
			))
		}
	}

	return nil
}
