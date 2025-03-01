package repository

import (
	"errors"
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

type IMemberRepo interface {
	WithMemberLevel() DBOption    // Member 预加载会员等级
	WithMemberCard() DBOption     // Member 预加载会员卡
	WithMemberCardType() DBOption // Member 预加载会员卡.卡类型

	WhereUuid(uuid uint64) DBOption // Uuid 条件

	GetMember(opts ...DBOption) model.Member    // 获取会员
	GetMemberLevels() []model.MemberLevel       // 获取会员等级
	SearchMember(keyword string) []model.Member // 关键字搜索会员
	CheckMemberExists(phone string) bool        // 根据手机号检查是否存在
	CheckLevelExists(uuid uint64) bool          // 根据Uuid检查等级是否存在

	CreateMember(member model.Member) error        // 添加会员
	Update(uuid uint64, vars map[string]any) error // 更新会员信息

	GetMemberInfoForSaleOrder(ctx context.Context, memberUuid uint64) (*model.Member, error) // 获取会员信息，用于销售订单结账时
}

func NewMemberRepo(db *gorm.DB) IMemberRepo {
	return NewMemberRepoImpl(db)
}

type MemberRepo struct {
	db *gorm.DB
}

func NewMemberRepoImpl(db *gorm.DB) *MemberRepo {
	return &MemberRepo{db: db}
}

// GetMemberLevels 获取会员等级
func (r *MemberRepo) GetMemberLevels() []model.MemberLevel {
	var levels []model.MemberLevel
	r.db.Model(&model.MemberLevel{}).Scopes(NotDeleted).Select("uuid, name, priority, create_time").Order("priority asc, create_time asc").Find(&levels)
	return levels
}

// SearchMember 关键字搜索会员
func (r *MemberRepo) SearchMember(keyword string) []model.Member {
	var members []model.Member
	keyword = Like(keyword)
	r.db.Model(&model.Member{}).Scopes(NotDeleted).Select("uuid, nickname, phone").Where("phone LIKE ? OR uuid LIKE ?", keyword, keyword).Find(&members)
	return members
}

// CreateMember 添加会员
func (r *MemberRepo) CreateMember(member model.Member) error {
	return r.db.Create(&member).Error
}

// CheckMemberExists 根据手机号检查是否存在
func (r *MemberRepo) CheckMemberExists(phone string) bool {
	var memberUuid uint64
	r.db.Model(&model.Member{}).Scopes(NotDeleted).Where("phone = ?", phone).Select("uuid").Scan(&memberUuid)
	return memberUuid > 0
}

// CheckLevelExists 根据uuid检查等级是否存在
func (r *MemberRepo) CheckLevelExists(uuid uint64) bool {
	var exists uint64
	r.db.Model(&model.MemberLevel{}).Scopes(NotDeleted).Where("uuid = ?", uuid).Select("uuid").Scan(&exists)
	return exists > 0
}

// GetMember 查询会员
func (r *MemberRepo) GetMember(opts ...DBOption) model.Member {
	var member model.Member
	db := r.db.Model(&model.Member{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&member)
	return member
}

// Update 更新会员信息
func (r *MemberRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// WithMemberLevel Member 预加载会员等级
func (r *MemberRepo) WithMemberLevel() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberLevel")
	}
}

// WithMemberCard Member 预加载会员卡
func (r *MemberRepo) WithMemberCard() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard")
	}
}

// WithMemberCardType Member 预加载会员卡.卡类型
func (r *MemberRepo) WithMemberCardType() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard.MemberCardType")
	}
}

// WhereUuid Uuid 条件
func (r *MemberRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *MemberRepo) GetMemberInfoForSaleOrder(ctx context.Context, memberUuid uint64) (*model.Member, error) {
	member := r.GetMember(r.WhereUuid(memberUuid), r.WithMemberCardType(), r.WithMemberLevel())
	if member.Uuid == 0 {
		return nil, errors.New("会员不存在")
	}
	return &member, nil
}
