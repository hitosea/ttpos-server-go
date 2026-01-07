package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberCouponRepo interface {
	IMemberCouponQuery
	CreateMemberCoupon(memberCoupon *model.MemberCoupon) error                     // 添加会员优惠券
	CreateMemberCouponRecord(memberCouponRecord model.MemberCouponUseRecord) error // 添加会员优惠券使用记录
	DeleteMemberCouponRecord(memberCouponUuid uint64) error                        // 删除会员优惠券使用记录
	Update(uuid uint64, vars map[string]any) error                                 // 更新会员优惠券信息
	VerifyMemberCoupon(uuid uint64) error                                          // 核销会员优惠券
	CancelVerifyMemberCoupon(uuid uint64) error                                    // 取消核销会员优惠券
}

// IMemberCouponQuery 会员优惠券查询接口,查询单独定义一个接口，方便后续扩展缓存
type IMemberCouponQuery interface {
	GetMemberCoupon(opts ...DBOption) (*model.MemberCoupon, error)                // 获取会员优惠券
	GetMemberCouponList(opts ...DBOption) ([]*model.MemberCoupon, error)          // 获取会员优惠券列表
	GetMemberCouponRecord(opts ...DBOption) (*model.MemberCouponUseRecord, error) // 获取会员优惠券
	GetMemberCouponByUuid(uuid uint64) (*model.MemberCoupon, error)               // 根据uuid获取会员优惠券
	GetMembersByUuids(uuids []uint64) ([]*model.MemberCoupon, error)              // 根据uuid列表获取会员优惠券列表
	GetValidMemberCouponList(memberUuid uint64) ([]*model.MemberCoupon, error)    // 获取会员有效期内的优惠券列表
	GetHistoryMemberCouponList(memberUuid uint64) ([]*model.MemberCoupon, error)  // 获取会员历史优惠券列表
}

func NewMemberCouponRepo(db *gorm.DB) IMemberCouponRepo {
	return NewMemberCouponRepoImpl(db)
}

type memberCouponRepo struct {
	db *gorm.DB
}

func NewMemberCouponRepoImpl(db *gorm.DB) IMemberCouponRepo {
	return &memberCouponRepo{db: db}
}

// GetMemberCoupon 获取会员优惠券
func (r *memberCouponRepo) GetMemberCoupon(opts ...DBOption) (*model.MemberCoupon, error) {
	var memberCoupon model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&memberCoupon).Error; err != nil {
		return nil, err
	}
	return &memberCoupon, nil
}

// GetMemberCouponList 获取会员优惠券列表
func (r *memberCouponRepo) GetMemberCouponList(opts ...DBOption) ([]*model.MemberCoupon, error) {
	var memberCoupons []*model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Order("id DESC").Find(&memberCoupons)
	return memberCoupons, nil
}

// GetMemberCouponRecord 获取会员优惠券使用记录
func (r *memberCouponRepo) GetMemberCouponRecord(opts ...DBOption) (*model.MemberCouponUseRecord, error) {
	var memberCouponUseRecord model.MemberCouponUseRecord
	db := r.db.Model(&model.MemberCouponUseRecord{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&memberCouponUseRecord)
	return &memberCouponUseRecord, nil
}

// GetMemberCouponByUuid 根据uuid获取会员优惠券
func (r *memberCouponRepo) GetMemberCouponByUuid(uuid uint64) (*model.MemberCoupon, error) {
	memberCoupon, err := r.GetMemberCoupon(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MarketingCoupon",
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return memberCoupon, nil
}

// GetMembersByUuids 根据uuid列表获取会员优惠券列表
func (r *memberCouponRepo) GetMembersByUuids(uuids []uint64) ([]*model.MemberCoupon, error) {
	var memberCoupons []*model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	if err := db.Where("uuid IN ?", uuids).Preload("MarketingCoupon").Find(&memberCoupons).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberCoupons, nil
}

// GetValidMemberCouponList 获取会员有效期内的优惠券列表
func (r *memberCouponRepo) GetValidMemberCouponList(memberUuid uint64) ([]*model.MemberCoupon, error) {
	return r.GetMemberCouponList(
		CommonRepo.WhereBySoftDelete(),           // 未删除
		CommonRepo.WhereByMemberUuid(memberUuid), // 根据会员UUID查询
		CommonRepo.WhereByStartTimeEndTime(),     // 是否在有效期内
		CommonRepo.WhereByStatus(0),              // 未使用
		CommonRepo.Preload(
			WithPreload{
				Query: "MarketingCoupon",
			},
		), // 预加载营销优惠券
	)
}

// GetHistoryMemberCouponList 获取会员历史优惠券列表
func (r *memberCouponRepo) GetHistoryMemberCouponList(memberUuid uint64) ([]*model.MemberCoupon, error) {
	return r.GetMemberCouponList(
		CommonRepo.WhereBySoftDelete(),           // 未删除
		CommonRepo.WhereByMemberUuid(memberUuid), // 根据会员UUID查询
		func(db *gorm.DB) *gorm.DB {
			now := time.Now().Unix()
			// 已使用或已过期的优惠券
			return db.Where("status = ? OR end_time < ?", 1, now)
		},
	)
}

// CreateMemberCoupon 添加会员优惠券
func (r *memberCouponRepo) CreateMemberCoupon(memberCoupon *model.MemberCoupon) error {
	return r.db.Create(memberCoupon).Error
}

// CreateMemberCouponRecord 添加会员优惠券使用记录
func (r *memberCouponRepo) CreateMemberCouponRecord(memberCouponRecord model.MemberCouponUseRecord) error {
	return r.db.Create(&memberCouponRecord).Error
}

// DeleteMemberCouponRecord 删除会员优惠券使用记录,软删除
func (r *memberCouponRepo) DeleteMemberCouponRecord(memberCouponUuid uint64) error {
	return r.db.Model(&model.MemberCouponUseRecord{}).Where("coupon_uuid = ?", memberCouponUuid).Update("delete_time", time.Now().Unix()).Error
}

// Update 更新会员优惠券信息
func (r *memberCouponRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.MemberCoupon{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// 核销会员优惠券
func (r *memberCouponRepo) VerifyMemberCoupon(uuid uint64) error {
	return r.Update(uuid, map[string]any{
		"status":   1,
		"use_time": time.Now().Unix(),
	})
}

// 取消核销会员优惠券，反结账退还优惠券场景
func (r *memberCouponRepo) CancelVerifyMemberCoupon(uuid uint64) error {
	return r.Update(uuid, map[string]any{
		"status":   0,
		"use_time": 0,
	})
}
