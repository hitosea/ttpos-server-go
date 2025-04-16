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

	"gorm.io/gorm"
)

type ISaleOrderPeakTimeRepo interface {
	Record(recordType string, saleBill *model.SaleBill, refundMoney float64) error
	GetMaxRecord(timezone string, startTime, endTime uint, cashierUuid uint64) ([]business_data_resp.PeakHour, error)
}

type saleOrderPeakTimeRepo struct {
	db *gorm.DB
}

func NewSaleOrderPeakTimeRepo(db *gorm.DB) ISaleOrderPeakTimeRepo {
	return &saleOrderPeakTimeRepo{db: db}
}

// Record 记录销售订单高峰时间
// recordType: inc - 增加, dec - 减少
func (r *saleOrderPeakTimeRepo) Record(recordType string, saleBill *model.SaleBill, refundMoney float64) error {
	// 获取完成时间的时间对象
	finishTime := time.Unix(saleBill.FinishTime, 0)
	// 获取日期时间戳 - 将时间设置为当天的0点0分0秒
	dateTime := time.Date(finishTime.Year(), finishTime.Month(), finishTime.Day(), 0, 0, 0, 0, finishTime.Location())
	dateTimestamp := dateTime.Unix()
	// 获取小时
	hour := int64(finishTime.Hour())
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
			"amount": utils.IfFloat64(recordType == "inc", saleOrderPeakTime.Amount+saleBill.Amount, descAmount),
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
			Amount:      saleBill.Amount,
		}).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// 获取高峰时段
func (r *saleOrderPeakTimeRepo) GetMaxRecord(timezone string, startTime, endTime uint, cashierUuid uint64) ([]business_data_resp.PeakHour, error) {
	// 获取开始时间和结束时间的时间对象
	startTimeObj := time.Unix(int64(startTime), 0)
	endTimeObj := time.Unix(int64(endTime), 0)
	// 获取开始和结束时间的小时部分
	startHour := int64(startTimeObj.Hour())
	endHour := int64(endTimeObj.Hour())
	// 获取开始日期和结束日期的时间戳（0点0分0秒）
	startDate := time.Date(startTimeObj.Year(), startTimeObj.Month(), startTimeObj.Day(), 0, 0, 0, 0, startTimeObj.Location()).Unix()
	endDate := time.Date(endTimeObj.Year(), endTimeObj.Month(), endTimeObj.Day(), 0, 0, 0, 0, endTimeObj.Location()).Unix()
	// 定义一个临时结构体来接收查询结果
	type PeakTimeResult struct {
		SumNum int    `gorm:"column:sum_num"`
		Ids    string `gorm:"column:ids"`
	}
	var peakTimeResult PeakTimeResult
	if err := r.db.Model(&model.SaleOrderPeakTime{}).
		Where("(date + hour * 60 * 60) between ? and ?", startDate+startHour*3600, endDate+endHour*3600).
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
	//
	return peakHours, nil
}
