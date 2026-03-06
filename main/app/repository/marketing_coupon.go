package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMarketingCouponRepo interface {
	IMarketingCouponQuery
	DecreaseCouponQuantity(couponId uint64, memberUuid uint64, activityUuid uint64) error        // 减少优惠券数量
	CreateMarketingCouponRecord(marketingCouponRecord model.MarketingCouponRecord) error         // 添加营销优惠券使用记录
	CreateMemberCouponRecord(marketingCouponUuid uint64, memberUuid uint64, leftCount int) error // 创建会员优惠券记录（核销扣减）
	CreateCommonCouponRecord(marketingCouponUuid uint64, leftCount int) error                    // 创建通用优惠券记录（核销扣减）
	CreateCommonCouponRecordCancel(marketingCouponUuid uint64, leftCount int) error              // 创建通用优惠券记录（反结账退还）
	UpdateCommonCouponCount(marketingCouponUuid uint64) error                                    // 通用优惠券核销，数量减1
	UpdateCommonCouponCountCancel(marketingCouponUuid uint64) error                              // 通用优惠券取消核销，数量加1，反结账场景
	SoftDeleteByHeadquarterExcludeUuids(headquarterUuid uint64, excludeUuids []uint64) error     // 软删除总部优惠券（排除指定uuid）
	FindHeadquarterByUuids(uuids []uint64) ([]model.MarketingCoupon, error)                      // 根据uuid列表查找总部优惠券
	UnscopedDeleteByUuid(uuid uint64) error                                                      // 硬删除优惠券
	CreateCoupon(coupon *model.MarketingCoupon) error                                            // 创建优惠券
	UnscopedDeleteByHeadquarter(headquarterUuid uint64) error                                    // 硬删除指定总部的所有优惠券
	FindAllHeadquarter() ([]model.MarketingCoupon, error)                                        // 查找所有总部优惠券
}

// IMarketingCouponQuery 营销优惠券查询接口,查询单独定义一个接口，方便后续扩展缓存
type IMarketingCouponQuery interface {
	GetCoupon(opts ...DBOption) (*model.MarketingCoupon, error)    // 获取优惠券
	GetCoupons(opts ...DBOption) ([]*model.MarketingCoupon, error) //获取店铺优惠券列表
	GetValidCouponList() ([]*model.MarketingCoupon, error)         // 获取商家还可用的none类型优惠卷（所有人可用的）
	GetCouponByUuid(uuid uint64) (*model.MarketingCoupon, error)   // 获取优惠券
	GetCouponLeftCount(uuid uint64) (int, error)                   // 查询营销优惠券的剩余数量
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
			CouponUuid:   couponId,
			MemberUuid:   memberUuid,
			ActivityUuid: activityUuid,
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

// GetCoupon 获取优惠券
func (r *MarketingCouponRepo) GetCoupon(opts ...DBOption) (*model.MarketingCoupon, error) {
	coupons, err := r.GetCoupons(opts...)
	if err != nil {
		return nil, err
	}
	if len(coupons) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return coupons[0], nil
}

// GetCoupons 获取店铺优惠券列表
func (r *MarketingCouponRepo) GetCoupons(opts ...DBOption) ([]*model.MarketingCoupon, error) {
	coupons := make([]*model.MarketingCoupon, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&coupons)
	if result.Error != nil {
		return nil, result.Error
	}

	return coupons, nil
}

// GetValidCouponList 获取商家还可用的none类型优惠卷（所有人可用的）
func (r *MarketingCouponRepo) GetValidCouponList() ([]*model.MarketingCoupon, error) {
	coupons, err := r.GetCoupons(
		CommonRepo.WhereByStatus(1),                                   // 已启用
		CommonRepo.WhereBySoftDelete(),                                // 未删除
		CommonRepo.WhereByCountGtZero(),                               // 优惠券数量大于0
		CommonRepo.WhereByRequirement(constant.CouponRequirementNone), // 任何人可用
		CommonRepo.WhereByValidTime(),                                 // 是否在有效期内
	)
	if err != nil {
		return nil, err
	}
	return coupons, nil
}

// GetCouponByUuid 获取优惠券信息
func (r *MarketingCouponRepo) GetCouponByUuid(uuid uint64) (*model.MarketingCoupon, error) {
	coupon, err := r.GetCoupon(
		CommonRepo.WhereByUuid(uuid),
	)
	if err != nil {
		return nil, err
	}
	return coupon, nil
}

// 查询营销优惠券的剩余数量
func (r *MarketingCouponRepo) GetCouponLeftCount(uuid uint64) (int, error) {
	var count int
	err := r.db.Model(&model.MarketingCoupon{}).Where("uuid = ?", uuid).Select("count").Scan(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CreateMarketingCouponRecord 创建营销优惠券记录
func (r *MarketingCouponRepo) CreateMarketingCouponRecord(marketingCouponRecord model.MarketingCouponRecord) error {
	return r.db.Create(&marketingCouponRecord).Error
}

// CreateMemberCouponRecord 创建会员优惠券记录（核销扣减）
func (r *MarketingCouponRepo) CreateMemberCouponRecord(marketingCouponUuid uint64, memberUuid uint64, leftCount int) error {
	serialNo, err := r.generateSerialNo(r.db)
	if err != nil {
		return err
	}
	marketingCouponRecord := model.MarketingCouponRecord{
		CouponUuid: marketingCouponUuid,
		SerialNo:   serialNo,
		MemberUuid: memberUuid,
		Type:       constant.CouponRecordTypeUsed, // 核销扣减
		Count:      1,
		LeftCount:  leftCount,
	}
	return r.db.Create(&marketingCouponRecord).Error
}

// CreateCommonCouponRecord 创建通用优惠券记录（核销扣减）
func (r *MarketingCouponRepo) CreateCommonCouponRecord(marketingCouponUuid uint64, leftCount int) error {
	serialNo, err := r.generateSerialNo(r.db)
	if err != nil {
		return err
	}
	marketingCouponRecord := model.MarketingCouponRecord{
		CouponUuid: marketingCouponUuid,
		SerialNo:   serialNo,
		Type:       constant.CouponRecordTypeUsed, // 核销扣减
		Count:      1,
		LeftCount:  leftCount,
	}
	return r.db.Create(&marketingCouponRecord).Error
}

// CreateCommonCouponRecordCancel 创建通用优惠券记录（反结账退还）
func (r *MarketingCouponRepo) CreateCommonCouponRecordCancel(marketingCouponUuid uint64, leftCount int) error {
	serialNo, err := r.generateSerialNo(r.db)
	if err != nil {
		return err
	}
	marketingCouponRecord := model.MarketingCouponRecord{
		CouponUuid: marketingCouponUuid,
		SerialNo:   serialNo,
		Type:       constant.CouponRecordTypeReverseSettle, // 反结账退还
		Count:      1,
		LeftCount:  leftCount,
	}
	return r.db.Create(&marketingCouponRecord).Error
}

// UpdateCommonCouponCount 更新通用优惠券数量
func (r *MarketingCouponRepo) UpdateCommonCouponCount(marketingCouponUuid uint64) error {
	return r.db.Model(&model.MarketingCoupon{}).Where("uuid = ?", marketingCouponUuid).Update("count", gorm.Expr("count - 1")).Error
}

// UpdateCommonCouponCountCancel 更新通用优惠券数量
func (r *MarketingCouponRepo) UpdateCommonCouponCountCancel(marketingCouponUuid uint64) error {
	return r.db.Model(&model.MarketingCoupon{}).Where("uuid = ?", marketingCouponUuid).Update("count", gorm.Expr("count + 1")).Error
}

// SoftDeleteByHeadquarterExcludeUuids 软删除总部优惠券（排除指定uuid）
func (r *MarketingCouponRepo) SoftDeleteByHeadquarterExcludeUuids(headquarterUuid uint64, excludeUuids []uint64) error {
	return r.db.Model(&model.MarketingCoupon{}).
		Where("headquarter_uuid = ? AND delete_time = 0 AND uuid NOT IN (?)", headquarterUuid, excludeUuids).
		Update("delete_time", time.Now().Unix()).Error
}

// FindHeadquarterByUuids 根据uuid列表查找总部优惠券
func (r *MarketingCouponRepo) FindHeadquarterByUuids(uuids []uint64) ([]model.MarketingCoupon, error) {
	coupons := make([]model.MarketingCoupon, 0)
	err := r.db.Model(&model.MarketingCoupon{}).
		Where("headquarter_uuid = 0 AND delete_time = 0 AND uuid IN (?)", uuids).
		Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}

// UnscopedDeleteByUuid 硬删除优惠券
func (r *MarketingCouponRepo) UnscopedDeleteByUuid(uuid uint64) error {
	return r.db.Unscoped().Where("uuid = ?", uuid).Delete(&model.MarketingCoupon{}).Error
}

// CreateCoupon 创建优惠券
func (r *MarketingCouponRepo) CreateCoupon(coupon *model.MarketingCoupon) error {
	return r.db.Create(coupon).Error
}

// UnscopedDeleteByHeadquarter 硬删除指定总部的所有优惠券
func (r *MarketingCouponRepo) UnscopedDeleteByHeadquarter(headquarterUuid uint64) error {
	return r.db.Unscoped().Where("headquarter_uuid = ?", headquarterUuid).Delete(&model.MarketingCoupon{}).Error
}

// FindAllHeadquarter 查找所有总部优惠券
func (r *MarketingCouponRepo) FindAllHeadquarter() ([]model.MarketingCoupon, error) {
	coupons := make([]model.MarketingCoupon, 0)
	err := r.db.Model(&model.MarketingCoupon{}).
		Where("headquarter_uuid = 0 AND delete_time = 0").
		Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}
