package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"github.com/shopspring/decimal"
)

// CountUserAnalysis 统计用户分析数据
func (r *StatisticsRepo) CountUserAnalysis(startTime, endTime int64, language string, enableNationality bool, enableCashierOrder bool, enableTableOrder bool, opts ...DBOption) (*model.UserAnalysisRepoResult, error) {
	result := &model.UserAnalysisRepoResult{
		Nationality:  []model.UserAnalysisItemRepo{},
		OrderSource:  []model.UserAnalysisItemRepo{},
		DeskSource:   []model.UserAnalysisItemRepo{},
		DiningMethod: []model.UserAnalysisItemRepo{},
	}

	var err error
	db := r.db
	// 先应用 opts，但需要在每个查询中明确使用表别名
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsSaleTable := prefix + "statistics_sale"
	saleBillTable := prefix + "sale_bill"
	nationalityTable := prefix + "nationality"
	orderSourceTable := prefix + "order_source"
	multiLanguageNameTable := prefix + "multi_language_name"

	// 构建语言字段名
	langField := getLanguageField(language)

	// 构建数据管理过滤条件（使用子查询，避免字段歧义）
	dataManageSubQuery := r.db.Model(&model.DataManage{}).
		Select("data_uuid").
		Where("type = ? AND delete_time = ?", model.DataManageTypeOrder, constant.NotDeleted)

	// 1. 按国籍统计（仅在开启国籍功能时统计）
	if enableNationality {
		// 统计国籍
		var nationalityResults []struct {
			NationalityUuid uint64
			Name            string
			OrderCount      int64
		}

		// 重新构建查询，应用 opts 并明确使用表别名
		queryDb := r.db
		for _, opt := range opts {
			queryDb = opt(queryDb)
		}
		err = queryDb.Table(statisticsSaleTable+" AS ss").
			Select("ss.nationality_uuid, COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown') AS name, COUNT(DISTINCT ss.sale_bill_uuid) AS order_count").
			Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid").
			Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON n.multi_language_name_uuid = mln.uuid").
			Where("ss.complete_time >= ? AND ss.complete_time <= ?", startTime, endTime).
			Where("ss.nationality_uuid > 0").
			Where("ss.sale_bill_uuid NOT IN (?)", dataManageSubQuery).
			Group("ss.nationality_uuid, name").
			Order("order_count ASC").
			Scan(&nationalityResults).Error

		if err != nil {
			return nil, err
		}

		// 计算总订单数
		var totalOrderCount int64
		for _, item := range nationalityResults {
			totalOrderCount += item.OrderCount
		}

		// 计算占比并转换
		for _, item := range nationalityResults {
			var percentage decimal.Decimal
			if totalOrderCount > 0 {
				percentage = decimal.NewFromInt(item.OrderCount).
					Div(decimal.NewFromInt(totalOrderCount)).
					Mul(decimal.NewFromInt(100)).
					Round(2)
			}
			result.Nationality = append(result.Nationality, model.UserAnalysisItemRepo{
				Name:       item.Name,
				OrderCount: item.OrderCount,
				Percentage: percentage,
			})
		}
	}

	// 2. 按点餐方式来源统计（仅点餐订单，仅在开启点餐功能时统计）
	if enableCashierOrder {
		var orderSourceResults []struct {
			OrderSourceUuid uint64
			Name            string
			OrderCount      int64
		}

		// 重新构建查询，应用 opts 并明确使用表别名
		queryDb2 := r.db
		for _, opt := range opts {
			queryDb2 = opt(queryDb2)
		}
		err = queryDb2.Table(statisticsSaleTable+" AS ss").
			Select("ss.order_source_uuid, "+
				"CASE "+
				"WHEN ss.order_source_uuid = 0 THEN '店内' "+
				"WHEN ss.order_source_uuid > 0 THEN COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown Source') "+
				"END AS name, "+
				"COUNT(DISTINCT ss.sale_bill_uuid) AS order_count").
			Joins("LEFT JOIN "+saleBillTable+" AS sb ON ss.sale_bill_uuid = sb.uuid AND sb.delete_time = ?", constant.NotDeleted).
			Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid").
			Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON os.multi_language_name_uuid = mln.uuid").
			Where("ss.complete_time >= ? AND ss.complete_time <= ?", startTime, endTime).
			Where("sb.bill_type = ?", constant.SaleBillTypeInstant).
			Where("ss.sale_bill_uuid NOT IN (?)", dataManageSubQuery).
			Group("ss.order_source_uuid, name").
			Order("order_count ASC").
			Scan(&orderSourceResults).Error

		if err != nil {
			return nil, err
		}

		// 计算总订单数
		var totalOrderSourceCount int64
		for _, item := range orderSourceResults {
			totalOrderSourceCount += item.OrderCount
		}

		// 计算占比并转换
		for _, item := range orderSourceResults {
			var percentage decimal.Decimal
			if totalOrderSourceCount > 0 {
				percentage = decimal.NewFromInt(item.OrderCount).
					Div(decimal.NewFromInt(totalOrderSourceCount)).
					Mul(decimal.NewFromInt(100)).
					Round(2)
			}
			result.OrderSource = append(result.OrderSource, model.UserAnalysisItemRepo{
				Name:       item.Name,
				OrderCount: item.OrderCount,
				Percentage: percentage,
			})
		}
	}

	// 3. 按桌台方式来源统计（仅桌台订单，且 source > 0，仅在开启桌台功能时统计）
	if enableTableOrder {
		var deskSourceResults []struct {
			Source     uint
			Name       string
			OrderCount int64
		}

		// 重新构建查询，应用 opts 并明确使用表别名
		queryDb3 := r.db
		for _, opt := range opts {
			queryDb3 = opt(queryDb3)
		}
		err = queryDb3.Table(statisticsSaleTable+" AS ss").
			Select("ss.source, "+
				"CASE "+
				"WHEN ss.source = ? THEN '收银机' "+
				"WHEN ss.source = ? THEN '点餐助手' "+
				"WHEN ss.source = ? THEN '平板' "+
				"WHEN ss.source = ? THEN 'H5' "+
				"ELSE 'Not Recorded' "+
				"END AS name, "+
				"COUNT(DISTINCT ss.sale_bill_uuid) AS order_count",
				constant.SaleBillSourceCashier,
				constant.SaleBillSourceAssistant,
				constant.SaleBillSourceTablet,
				constant.SaleBillSourceH5).
			Joins("LEFT JOIN "+saleBillTable+" AS sb ON ss.sale_bill_uuid = sb.uuid AND sb.delete_time = ?", constant.NotDeleted).
			Where("ss.complete_time >= ? AND ss.complete_time <= ?", startTime, endTime).
			Where("sb.bill_type = ?", constant.SaleBillTypeDesk).
			Where("ss.source > ?", constant.SaleBillSourceDefault).
			Where("ss.sale_bill_uuid NOT IN (?)", dataManageSubQuery).
			Group("ss.source, name").
			Order("order_count ASC").
			Scan(&deskSourceResults).Error

		if err != nil {
			return nil, err
		}

		// 计算总订单数
		var totalDeskSourceCount int64
		for _, item := range deskSourceResults {
			totalDeskSourceCount += item.OrderCount
		}

		// 计算占比并转换
		for _, item := range deskSourceResults {
			var percentage decimal.Decimal
			if totalDeskSourceCount > 0 {
				percentage = decimal.NewFromInt(item.OrderCount).
					Div(decimal.NewFromInt(totalDeskSourceCount)).
					Mul(decimal.NewFromInt(100)).
					Round(2)
			}
			result.DeskSource = append(result.DeskSource, model.UserAnalysisItemRepo{
				Name:       item.Name,
				OrderCount: item.OrderCount,
				Percentage: percentage,
			})
		}
	}

	// 4. 按用餐方式统计（点餐+桌台）
	var diningMethodResults []struct {
		DiningMethod uint
		Name         string
		OrderCount   int64
	}

	// 重新构建查询，应用 opts 并明确使用表别名
	queryDb4 := r.db
	for _, opt := range opts {
		queryDb4 = opt(queryDb4)
	}
	err = queryDb4.Table(statisticsSaleTable+" AS ss").
		Select("CASE "+
			"WHEN sb.bill_type = ? THEN 0 "+
			"ELSE sb.dining_method "+
			"END AS dining_method, "+
			"CASE "+
			"WHEN sb.bill_type = ? THEN '店内用餐' "+
			"WHEN sb.dining_method = 0 THEN '店内用餐' "+
			"WHEN sb.dining_method = 1 THEN '打包' "+
			"ELSE 'Not Recorded' "+
			"END AS name, "+
			"COUNT(DISTINCT ss.sale_bill_uuid) AS order_count",
			constant.SaleBillTypeDesk, constant.SaleBillTypeDesk).
		Joins("LEFT JOIN "+saleBillTable+" AS sb ON ss.sale_bill_uuid = sb.uuid AND sb.delete_time = ?", constant.NotDeleted).
		Where("ss.complete_time >= ? AND ss.complete_time <= ?", startTime, endTime).
		Where("sb.bill_type IN (?, ?)", constant.SaleBillTypeInstant, constant.SaleBillTypeDesk).
		Where("ss.sale_bill_uuid NOT IN (?)", dataManageSubQuery).
		Group("dining_method, name").
		Order("order_count ASC").
		Scan(&diningMethodResults).Error

	if err != nil {
		return nil, err
	}

	// 计算总订单数
	var totalDiningMethodCount int64
	for _, item := range diningMethodResults {
		totalDiningMethodCount += item.OrderCount
	}

	// 计算占比并转换
	for _, item := range diningMethodResults {
		var percentage decimal.Decimal
		if totalDiningMethodCount > 0 {
			percentage = decimal.NewFromInt(item.OrderCount).
				Div(decimal.NewFromInt(totalDiningMethodCount)).
				Mul(decimal.NewFromInt(100)).
				Round(2)
		}
		result.DiningMethod = append(result.DiningMethod, model.UserAnalysisItemRepo{
			Name:       item.Name,
			OrderCount: item.OrderCount,
			Percentage: percentage,
		})
	}

	return result, nil
}

// getLanguageField 获取语言字段名（用于 multi_language_name 表）
func getLanguageField(language string) string {
	langMap := map[string]string{
		"zh":    "zh_name",
		"th":    "th_name",
		"en":    "en_name",
		"zhtw":  "zh_tw_name",
		"zh_tw": "zh_tw_name",
		"ja":    "ja_name",
		"ko":    "ko_name",
		"my":    "my_name",
		"tr":    "tr_name",
		"sv":    "sv_name",
	}
	if field, ok := langMap[language]; ok {
		return field
	}
	return "en_name" // 默认返回英文名称
}
