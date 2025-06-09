package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMarketingCouponRepo interface {
	// DecreaseCouponQuantity 减少优惠券数量
	DecreaseCouponQuantity(couponId uint64, memberUuid uint64, activityUuid uint64) error
}

// MarketingCouponRepo 营销优惠券仓库
type MarketingCouponRepo struct {
	db *gorm.DB
}

// NewMarketingCouponRepo 创建营销优惠券仓库
func NewMarketingCouponRepo(db *gorm.DB) IMarketingCouponRepo {
	return &MarketingCouponRepo{
		db: db,
	}
}

// generateSerialNo 生成记录编号
func (r *MarketingCouponRepo) generateSerialNo(tx *gorm.DB) (string, error) {
	// 获取当前时间部分
	timePart := time.Now().Format("060102150405")

	// 查询今天最大的记录编号
	var maxRecord model.MarketingCouponRecord
	err := tx.Where("serial_no LIKE ?", timePart+"%").Order("serial_no DESC").First(&maxRecord).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 获取序号部分
	var serialNum int
	if err == gorm.ErrRecordNotFound {
		// 如果没有记录，从0000开始
		serialNum = 0
	} else {
		// 从现有记录中提取序号并加1
		serialNum = int(maxRecord.SerialNo[len(timePart):][0]-'0')*1000 +
			int(maxRecord.SerialNo[len(timePart):][1]-'0')*100 +
			int(maxRecord.SerialNo[len(timePart):][2]-'0')*10 +
			int(maxRecord.SerialNo[len(timePart):][3]-'0') + 1
		// 如果超过9999，循环回0000
		if serialNum > 9999 {
			serialNum = 0
		}
	}

	// 生成新的记录编号
	return fmt.Sprintf("%s%04d", timePart, serialNum), nil
}

// DecreaseCouponQuantity 减少优惠券数量
func (r *MarketingCouponRepo) DecreaseCouponQuantity(couponId uint64, memberUuid uint64, activityUuid uint64) error {
	// 开启事务
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询优惠券信息
		var coupon model.MarketingCoupon
		if err := tx.Where("uuid = ? AND delete_time = 0", couponId).First(&coupon).Error; err != nil {
			return err
		}
		// 2. 检查优惠券数量是否足够
		if coupon.Count <= 0 {
			return gorm.ErrRecordNotFound
		}
		// 3. 更新优惠券数量
		if err := tx.Model(&model.MarketingCoupon{}).
			Where("uuid = ? AND delete_time = 0", couponId).
			Update("count", gorm.Expr("count - ?", 1)).Error; err != nil {
			return err
		}
		// 4. 生成记录编号
		serialNo, err := r.generateSerialNo(tx)
		if err != nil {
			return err
		}
		// 5. 创建使用记录
		record := model.MarketingCouponRecord{
			BaseModel: model.BaseModel{
				Uuid:       uint64(time.Now().UnixNano()),
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			CouponUuid:   int64(couponId),
			MemberUuid:   int64(memberUuid),
			ActivityUuid: int64(activityUuid),
			Type:         constant.CouponRecordTypeBonus, // 调整扣减
			Count:        1,
			LeftCount:    coupon.Count - 1,
			SerialNo:     serialNo,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		//
		return nil
	})
}
