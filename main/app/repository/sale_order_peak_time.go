package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ISaleOrderPeakTimeRepo interface {
	Record(recordType string, saleBill *model.SaleBill, refundMoney float64, timezone ...string) error
	GetMaxRecord(timezone string, startTime, endTime uint, cashierUuid uint64, excludeDataManage bool) ([]business_data_resp.PeakHour, error)
	CreateSaleOrderPeakTime(saleOrderPeakTime model.SaleOrderPeakTime) error
}

type saleOrderPeakTimeRepo struct {
	db *gorm.DB
}

func NewSaleOrderPeakTimeRepo(db *gorm.DB) ISaleOrderPeakTimeRepo {
	return &saleOrderPeakTimeRepo{db: db}
}

// Record 记录销售订单高峰时间
// recordType: inc - 增加, dec - 减少
func (r *saleOrderPeakTimeRepo) Record(recordType string, saleBill *model.SaleBill, refundMoney float64, timezone ...string) error {
	tz := string(utils.ZH_TIMEZONE)
	var finishTime int64 = 0
	if len(timezone) > 0 {
		tz = timezone[0]
	}
	if len(timezone) > 1 {
		finishTimeInt, err := strconv.ParseInt(timezone[1], 10, 64)
		if err == nil {
			finishTime = finishTimeInt
		}
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}
	// 将时间戳转换为time.Time
	finishTime = utils.IfInt64(saleBill.FinishTime > 0, saleBill.FinishTime, finishTime)
	tm := time.Unix(finishTime, 0).In(loc)
	// 获取日期时间戳 - 将时间设置为当天的0点0分0秒
	dateTime := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.Location())
	dateTimestamp := dateTime.Unix()
	// 获取小时
	hour := int64(tm.Hour())
	// 查询条件
	condition := map[string]any{
		"date":         dateTimestamp,
		"hour":         hour,
		"cashier_uuid": saleBill.CashierUuid,
	}
	// 查询高峰时段
	var saleOrderPeakTime model.SaleOrderPeakTime
	if err := r.db.Where(condition).First(&saleOrderPeakTime).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithMessage(err)
	}
	// 更新高峰时段
	if saleOrderPeakTime.Uuid > 0 {
		// 产品需求，当遇到订单退款时，订单数不改变，只改变金额（v1.1.1）
		descNum := utils.IfInt(refundMoney > 0, saleOrderPeakTime.Num, saleOrderPeakTime.Num-1)
		descAmount := utils.IfFloat64(refundMoney > 0, saleOrderPeakTime.Amount-refundMoney, saleOrderPeakTime.Amount-saleBill.Amount)
		descNum = utils.IfInt(descNum < 0, 0, descNum)
		descAmount = utils.IfFloat64(descAmount < 0, 0.0, descAmount)
		// 更新高峰时段
		updateData := map[string]any{
			"num":    utils.IfInt(recordType == "inc", saleOrderPeakTime.Num+1, descNum),
			"amount": utils.IfFloat64(recordType == "inc", saleOrderPeakTime.Amount+saleBill.PaymentAmount, descAmount),
		}
		if err := r.db.Model(&model.SaleOrderPeakTime{}).Where(condition).Updates(updateData).Error; err != nil {
			return errors.WithMessage(err)
		}
	} else if recordType == "inc" && saleBill.IsFinish() {
		// 创建高峰时段
		if err := r.db.Model(&model.SaleOrderPeakTime{}).Create(&model.SaleOrderPeakTime{
			Date:        dateTimestamp,
			Hour:        hour,
			CashierUuid: saleBill.CashierUuid,
			Num:         1,
			Amount:      saleBill.PaymentAmount,
		}).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// GetMaxRecord 获取高峰时段
func (r *saleOrderPeakTimeRepo) GetMaxRecord(timezone string, startTime, endTime uint, cashierUuid uint64, excludeDataManage bool) ([]business_data_resp.PeakHour, error) {
	// 获取开始时间和结束时间的时间对象
	startTimeObj := time.Unix(int64(startTime), 0)
	endTimeObj := time.Unix(int64(endTime), 0)
	// 获取开始和结束时间的小时部分
	startHour := int64(startTimeObj.Hour())
	endHour := int64(endTimeObj.Hour())
	// 获取开始日期和结束日期的时间戳（0点0分0秒）
	startDate := time.Date(startTimeObj.Year(), startTimeObj.Month(), startTimeObj.Day(), 0, 0, 0, 0, startTimeObj.Location()).Unix()
	endDate := time.Date(endTimeObj.Year(), endTimeObj.Month(), endTimeObj.Day(), 0, 0, 0, 0, endTimeObj.Location()).Unix()
	// 排除测试营业时段：用 (date + hour * 3600) 作为记录时间点判断
	excludeTestBusinessPeakSQL := "NOT EXISTS (SELECT 1 FROM ttpos_business_status_period bsp " +
		"WHERE bsp.delete_time = 0 " +
		"AND (date + hour * 3600) >= bsp.start_time " +
		"AND (bsp.end_time = 0 OR (date + hour * 3600) <= bsp.end_time))"

	// 定义一个临时结构体来接收查询结果
	type PeakTimeResult struct {
		SumNum int    `gorm:"column:sum_num"`
		Ids    string `gorm:"column:ids"`
	}
	var peakTimeResult PeakTimeResult
	if err := r.db.Model(&model.SaleOrderPeakTime{}).
		Where("(date + hour * 60 * 60) between ? and ?", startDate+startHour*3600, endDate+endHour*3600).
		Where(excludeTestBusinessPeakSQL).
		// Where("cashier_uuid = ?", cashierUuid).
		Group("CONCAT(date,hour)").
		Select("sum(num) as sum_num, group_concat(id) as ids").
		Order("sum_num desc").
		First(&peakTimeResult).Error; err != nil {
		// 如果是记录未找到的错误，返回空数组
		if err == gorm.ErrRecordNotFound {
			return []business_data_resp.PeakHour{}, nil
		}
		// 其他错误则返回错误信息
		return nil, errors.WithMessage(err)
	}
	// 如果没有IDs，返回空数组
	if peakTimeResult.Ids == "" {
		return []business_data_resp.PeakHour{}, nil
	}

	// 使用子查询实现复杂的SQL查询
	subQuery := r.db.Model(&model.SaleOrderPeakTime{}).
		Where("(date + hour * 60 * 60) between ? and ?", startDate+startHour*3600, endDate+endHour*3600).
		Where(excludeTestBusinessPeakSQL).
		Group("CONCAT(date,hour)").
		Select("sum(num) as sum_num, group_concat(id) as ids")
	if err := r.db.Table("(?) as t", subQuery).
		Select("sum_num, group_concat(ids) as ids").
		Where("sum_num = ?", peakTimeResult.SumNum).
		First(&peakTimeResult).Error; err != nil {
		// 如果是记录未找到的错误，返回空数组
		if err == gorm.ErrRecordNotFound {
			return []business_data_resp.PeakHour{}, nil
		}
		// 其他错误则返回错误信息
		return nil, errors.WithMessage(err)
	}

	// 将逗号分隔的ID字符串转换为整数数组
	var maxIdsIds []uint64
	idsStr := strings.Split(peakTimeResult.Ids, ",")
	for _, idStr := range idsStr {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			fmt.Println("解析ID错误:", err)
			continue
		}
		if id == 0 {
			continue
		}
		maxIdsIds = append(maxIdsIds, id)
	}
	//
	var results []*model.SaleOrderPeakTime
	if err := r.db.Model(&model.SaleOrderPeakTime{}).
		Where("id IN ?", maxIdsIds).
		Group("CONCAT(date,hour)").
		Select("date, hour, sum(num) as num, sum(amount) amount").
		Scan(&results).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []business_data_resp.PeakHour{}, nil
		}
		return nil, errors.WithMessage(err)
	}
	//
	var peakHours []business_data_resp.PeakHour
	for _, v := range results {
		startHour := fmt.Sprintf("%02d", v.Hour) // 将整点数补齐为两位数作为起始小时
		endHour := fmt.Sprintf("%02d", v.Hour+1) // 结束小时为起始小时加1
		timePeriod := fmt.Sprintf("%s %s:00-%s:00", utils.SetTimezone(timezone).FormatUnixTime(v.Date, "01/02"), startHour, endHour)
		peakHours = append(peakHours, business_data_resp.PeakHour{
			TimePeriod: timePeriod,
			OrderNum:   v.Num,
			Amount:     v.Amount,
		})
	}
	// 如果 saleBillMap 中有数据，需要从对应的时间段中减去
	if excludeDataManage {
		saleBillMap := make(map[int64]float64)
		saleBillRepo := NewOrderRepoImpl(r.db)
		saleBillList := saleBillRepo.GetSaleBillList(
			CommonRepo.WhereInDataManageSubQuery(r.db, "uuid", CommonRepo.WhereByType(model.DataManageTypeOrder), CommonRepo.WhereBySoftDelete()),
			CommonRepo.Preload(
				WithPreload{
					Query: "SaleOrders",
					Args: []interface{}{
						CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					},
				},
				WithPreload{
					Query: "SaleOrders.PaymentOrders.PaymentMethod",
				},
				WithPreload{
					Query: "SaleOrders.ReturnOrders",
				},
			),
			CommonRepo.WhereBySoftDelete(),
		)
		for _, saleBill := range saleBillList {
			amountDec := decimal.NewFromFloat(saleBill.PaymentAmount)
			for _, saleOrder := range saleBill.SaleOrders {
				for _, returnOrder := range saleOrder.ReturnOrders {
					amountDec = amountDec.Sub(decimal.NewFromFloat(returnOrder.RefundAmount))
				}
			}
			saleBillMap[saleBill.FinishTime] = amountDec.InexactFloat64()
		}
		if len(saleBillMap) > 0 {
			// 遍历 saleBillMap，找到对应的时间段并减少
			for finishTime, amount := range saleBillMap {
				// 在 results 中找到匹配的项（finishTime 在时间段范围内）
				for i, v := range results {
					// 计算时间段的开始和结束时间戳
					timePeriodStart := v.Date + v.Hour*3600
					timePeriodEnd := v.Date + (v.Hour+1)*3600
					// 检查 finishTime 是否在时间段范围内
					if finishTime >= timePeriodStart && finishTime < timePeriodEnd {
						// 找到匹配的时间段，减少 OrderNum 和 Amount
						peakHours[i].OrderNum--
						peakHours[i].Amount = utils.DecimalSub(peakHours[i].Amount, amount)
						// 确保 OrderNum 和 Amount 不为负数
						if peakHours[i].OrderNum < 0 {
							peakHours[i].OrderNum = 0
						}
						if peakHours[i].Amount < 0 {
							peakHours[i].Amount = 0
						}
						break
					}
				}
			}
		}
	}
	//
	return peakHours, nil
}

// CreateSaleOrderPeakTime 创建高峰时段
func (r *saleOrderPeakTimeRepo) CreateSaleOrderPeakTime(saleOrderPeakTime model.SaleOrderPeakTime) error {
	if err := r.db.Create(&saleOrderPeakTime).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
