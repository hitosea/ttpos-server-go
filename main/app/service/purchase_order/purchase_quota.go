package purchase_order

import (
	"fmt"

	"gorm.io/gorm"

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
// 品牌采购限购校验方法
// ====================================================================================

// checkDailySubmitLimit 检查每日申请次数限制
//
// 逻辑：
//   - 从总部（headquarter_uuid）获取全局配置（ttpos_setting 表，key=purchase_brand_daily_limit）
//   - -1 表示不限制
//   - 统计范围：当天（使用店铺时区）已提交的品牌采购申请（status != 0，排除草稿）
//   - 按门店维度统计
func (s *purchaseOrderSrv) checkDailySubmitLimit(ctx context.Context, companyUuid uint64) error {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid

	// 如果没有总部，说明自己就是总部，或者不进行总部限制
	if headquarterUuid == 0 {
		return nil
	}

	// 通过 Repository 读取配置
	headquarterDb := s.dbm.GetDB(headquarterUuid)
	setting := repository.NewSettingRepo(headquarterDb).GetByKey(model.SettingKeyPurchaseBrandDailyLimit)

	// 获取限制值
	dailyLimit := setting.GetPurchaseBrandDailyLimit()

	// -1 表示不限制
	if dailyLimit == -1 {
		return nil
	}

	// 使用系统时区计算当天的起止时间戳
	todayStart, todayEnd := utils.SetTimezone("").TodayStartEndUnix()

	// 通过 Repository 统计当天已提交的申请次数
	count, err := repository.NewPurchaseOrderRepo(ctx.GetDB()).CountBrandPurchaseByTimeRange(todayStart, todayEnd)
	if err != nil {
		return errors.WithMessage(errors.New("查询每日申请次数失败"), err.Error())
	}

	// 校验是否超限
	if count >= int64(dailyLimit) {
		return errors.New(i18n.Translate(ctx.GetLanguage(), "今日申请次数已达上限（%s次），请明天再试", fmt.Sprintf("%d", dailyLimit)))
	}

	return nil
}

// checkPurchaseQuota 检查品牌采购物品限购
//
// 逻辑：
//   - 遍历订单明细，逐个检查是否有限购配置
//   - 查询限购配置时支持门店维度过滤（apply_to_all_shops 或 purchase_quota_config_shop 表）
//   - 校验单位是否匹配限购配置的单位
//   - 根据 period_type 统计已使用额度：
//   - period_type=0: 按天统计（当天已使用额度）
//   - period_type=1: 按月统计（本月已使用额度）
//   - 判断是否超限：已用额度 + 本次申请 > 限购额度
func (s *purchaseOrderSrv) checkPurchaseQuota(ctx context.Context, order *model.PurchaseOrder) error {
	lang := ctx.GetLanguage()
	companyUuid := ctx.GetCompanyUuid()
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	if headquarterUuid == 0 {
		return nil
	}
	headquarterDb := s.dbm.GetDB(headquarterUuid)

	repo := repository.NewPurchaseQuotaConfigRepo(headquarterDb)

	// 遍历订单明细，逐个检查限购
	// 步骤1: 先汇总相同物品的数量（按 MaterialCode + ErpnextUom 分组）
	type MaterialUnitKey struct {
		MaterialCode string
		ErpnextUom   string
	}

	// 用于汇总的 map
	materialSummary := make(map[MaterialUnitKey]struct {
		TotalQty     float64
		MaterialName string
	})

	// 遍历所有物品，累加数量
	for _, item := range order.Items {
		materialName := language.JsonToLocaleResponse(item.MaterialName).GetLocale(lang)

		// 如果有多单位，遍历每个单位进行累加
		if len(item.Units) > 0 {
			for _, unit := range item.Units {
				key := MaterialUnitKey{
					MaterialCode: item.MaterialCode,
					ErpnextUom:   unit.ErpnextUom,
				}
				summary := materialSummary[key]
				summary.TotalQty += unit.Num
				summary.MaterialName = materialName
				materialSummary[key] = summary
			}
		} else {
			// 如果没有多单位，使用主单位
			key := MaterialUnitKey{
				MaterialCode: item.MaterialCode,
				ErpnextUom:   item.ErpnextUom,
			}
			summary := materialSummary[key]
			summary.TotalQty += item.Num
			summary.MaterialName = materialName
			materialSummary[key] = summary
		}
	}

	// 步骤2: 对汇总后的物品进行限购验证
	for key, summary := range materialSummary {
		// 查询该物品的限购配置
		config, err := repo.GetByMaterialCodeAndShop(key.MaterialCode, companyUuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 无限购配置，跳过整个物品
				continue
			}
			return errors.WithMessage(errors.New("获取品牌采购限购配置失败"), err.Error())
		}

		// 使用汇总后的数量进行限购检查
		if err := s.checkSingleUnitQuotaWithConfig(ctx, lang, config, key.MaterialCode, summary.MaterialName, key.ErpnextUom, summary.TotalQty, order); err != nil {
			return err
		}
	}

	return nil
}

// checkSingleUnitQuotaWithConfig 使用已查询的配置检查单个单位的限购
func (s *purchaseOrderSrv) checkSingleUnitQuotaWithConfig(
	ctx context.Context,
	lang string,
	config *model.PurchaseQuotaConfig,
	materialCode string,
	materialName string,
	unitCode string,
	quantity float64,
	order *model.PurchaseOrder,
) error {
	// 如果单位编码为空，跳过检查
	if unitCode == "" {
		return nil
	}

	// 1. 强制性单位校验：配置了限购的物品，必须使用限购单位
	if unitCode != config.UnitCode {
		return errors.New(i18n.Translate(lang, "物品 %s 只能使用 %s 单位进行采购", materialName, config.UnitCode))
	}

	// 2. 根据周期类型查询已使用额度
	var usedQty float64
	var err error
	if config.PeriodType == constant.PurchaseQuotaPeriodTypeDaily {
		// 按天限购：统计今天已使用额度
		usedQty, err = s.getDailyUsedQuota(ctx, materialCode, unitCode, order.Uuid)
	} else {
		// 按月限购：统计本月已使用额度
		usedQty, err = s.getMonthlyUsedQuota(ctx, materialCode, unitCode, order.Uuid)
	}
	if err != nil {
		return errors.WithMessage(errors.New("查询品牌采购已使用额度失败"), err.Error())
	}

	// 3. 校验是否超限
	if usedQty+quantity > config.QuotaLimit {
		return errors.New(fmt.Sprintf(
			i18n.Translate(lang, "本次申请 %s 数量已超限（最多 %v），请减少物品数量后重试"),
			materialName, config.QuotaLimit,
		))
	}

	return nil
}

// getDailyUsedQuota 查询今日已使用额度（实时统计）
//
// 统计逻辑：
//   - 统计范围：当天（使用店铺时区，00:00:00 - 23:59:59）
//   - 订单状态：待审核(1)、已通过(2)、待总部审核(5)、已完成(7)
//   - 排除状态：草稿(0)、已驳回(3)、全部收货(6)
//   - 排除当前单据：避免重复统计
//   - 按门店维度统计
func (s *purchaseOrderSrv) getDailyUsedQuota(
	ctx context.Context,
	materialCode string,
	unitCode string,
	excludeOrderUuid uint64,
) (float64, error) {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	if headquarterUuid == 0 {
		return 0, nil
	}
	headquarterDb := s.dbm.GetDB(headquarterUuid)

	// 使用系统时区计算今天的起止时间戳
	todayStart, todayEnd := utils.SetTimezone("").TodayStartEndUnix()

	// 通过 Repository 统计已使用额度
	return repository.NewPurchaseOrderItemRepo(headquarterDb).SumBrandPurchaseByTimeRange(
		materialCode,
		unitCode,
		todayStart,
		todayEnd,
		excludeOrderUuid,
	)
}

// getMonthlyUsedQuota 查询本月已使用额度（实时统计）
//
// 统计逻辑：
//   - 统计范围：当前自然月（使用系统时区，格式：2026-01）
//   - 订单状态：排除草稿(0)
//   - 排除当前单据：避免重复统计
//   - 按门店维度统计
func (s *purchaseOrderSrv) getMonthlyUsedQuota(
	ctx context.Context,
	materialCode string,
	unitCode string,
	excludeOrderUuid uint64,
) (float64, error) {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	if headquarterUuid == 0 {
		return 0, nil
	}
	headquarterDb := s.dbm.GetDB(headquarterUuid)

	// 使用系统时区计算当前月份（格式：2026-01）
	currentMonth := utils.SetTimezone("").Now().Format("2006-01")

	// 通过 Repository 统计已使用额度
	return repository.NewPurchaseOrderItemRepo(headquarterDb).SumBrandPurchaseByMonth(
		materialCode,
		unitCode,
		currentMonth,
		excludeOrderUuid,
	)
}
