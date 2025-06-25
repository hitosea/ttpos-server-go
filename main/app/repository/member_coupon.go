package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberCouponRepo interface {
	GetMemberCoupon(opts ...DBOption) model.MemberCoupon                            // 获取会员优惠券
	GetMemberCouponList(opts ...DBOption) ([]*model.MemberCoupon, error)            // 获取会员优惠券列表
	GetMemberCouponRecord(opts ...DBOption) (*model.MemberCouponUseRecord, error)   // 获取会员优惠券
	GetMemberCouponByUuid(uuid uint64) (*model.MemberCoupon, error)                 // 根据uuid获取会员优惠券
	GetMembersByUuids(uuids []uint64) ([]*model.MemberCoupon, error)                // 根据uuid列表获取会员优惠券列表
	GetValidMemberCouponList(memberUuid uint64) ([]*model.MemberCoupon, error)      // 获取会员有效期内的优惠券列表
	CreateMemberCoupon(memberCoupon *model.MemberCoupon) error                      // 添加会员优惠券
	CreateMemberCouponRecord(memberCouponRecord *model.MemberCouponUseRecord) error // 添加会员优惠券使用记录
	Update(uuid uint64, vars map[string]any) error                                  // 更新会员优惠券信息
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
func (r *memberCouponRepo) GetMemberCoupon(opts ...DBOption) model.MemberCoupon {
	var memberCoupon model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&memberCoupon)
	return memberCoupon
}

// GetMemberCouponList 获取会员优惠券列表
func (r *memberCouponRepo) GetMemberCouponList(opts ...DBOption) ([]*model.MemberCoupon, error) {
	var memberCoupons []*model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Find(&memberCoupons)
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
	var memberCoupon model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	db.First(&memberCoupon, uuid)
	return &memberCoupon, nil
}

// GetMembersByUuids 根据uuid列表获取会员优惠券列表
func (r *memberCouponRepo) GetMembersByUuids(uuids []uint64) ([]*model.MemberCoupon, error) {
	var memberCoupons []*model.MemberCoupon
	db := r.db.Model(&model.MemberCoupon{}).Scopes(NotDeleted)
	db.Where("uuid IN ?", uuids).Find(&memberCoupons)
	return memberCoupons, nil
}

func (r *memberCouponRepo) GetValidMemberCouponList(memberUuid uint64) ([]*model.MemberCoupon, error) {
	return r.GetMemberCouponList(
		CommonRepo.WhereBySoftDelete(),           // 未删除
		CommonRepo.WhereByMemberUuid(memberUuid), // 根据会员UUID查询
		CommonRepo.WhereByStartTimeEndTime(),     // 是否在有效期内
		CommonRepo.WhereByStatus(0),              // 未使用
		CommonRepo.SortWithCreateTime("desc"),    // 按创建时间降序
		CommonRepo.Preload(
			WithPreload{
				Query: "MarketingCoupon",
			},
		), // 预加载营销优惠券
	)
}

// CreateMemberCoupon 添加会员优惠券
func (r *memberCouponRepo) CreateMemberCoupon(memberCoupon *model.MemberCoupon) error {
	return r.db.Create(memberCoupon).Error
}

// CreateMemberCouponRecord 添加会员优惠券使用记录
func (r *memberCouponRepo) CreateMemberCouponRecord(memberCouponRecord *model.MemberCouponUseRecord) error {
	return r.db.Create(memberCouponRecord).Error
}

// Update 更新会员优惠券信息
func (r *memberCouponRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.MemberCoupon{}).Where("uuid = ?", uuid).Updates(vars).Error
}
