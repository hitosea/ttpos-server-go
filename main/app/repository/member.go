package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IMemberRepo interface {
	GetMemberLevels() []model.MemberLevel              // 获取会员等级
	SearchMember(keyword string) []model.Member        // 关键字搜索会员
	CreateMember(member model.Member) error            // 添加会员
	CheckMemberExists(phone string) bool               // 根据手机号检查是否存在
	CheckLevelExists(uuid uint64) bool                 // 根据Uuid检查等级是否存在
	GetByUuid(uuid uint64, withs ...With) model.Member // 根据Uuid查询会员
	Update(uuid uint64, vars map[string]any) error     // 更新会员信息

	WithMemberLevel() With    // Member 预加载会员等级
	WithMemberCard() With     // Member 预加载会员卡
	WithMemberCardType() With // Member 预加载会员卡.卡类型
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
	r.db.Scopes(NotDeleted).Model(&model.MemberLevel{}).Select("uuid, name, priority, create_time").Order("priority asc, create_time asc").Find(&levels)
	return levels
}

// SearchMember 关键字搜索会员
func (r *MemberRepo) SearchMember(keyword string) []model.Member {
	var members []model.Member
	keyword = Like(keyword)
	r.db.Scopes(NotDeleted).Model(&model.Member{}).Select("uuid, nickname, phone").Where("phone LIKE ? OR uuid LIKE ?", keyword, keyword).Find(&members)
	return members
}

// CreateMember 添加会员
func (r *MemberRepo) CreateMember(member model.Member) error {
	return r.db.Create(&member).Error
}

// CheckMemberExists 根据手机号检查是否存在
func (r *MemberRepo) CheckMemberExists(phone string) bool {
	var memberUuid uint64
	r.db.Scopes(NotDeleted).Model(&model.Member{}).Where("phone = ?", phone).Select("uuid").Scan(&memberUuid)
	return memberUuid > 0
}

// CheckLevelExists 根据uuid检查等级是否存在
func (r *MemberRepo) CheckLevelExists(uuid uint64) bool {
	var exists uint64
	r.db.Scopes(NotDeleted).Model(&model.MemberLevel{}).Where("uuid = ?", uuid).Select("uuid").Scan(&exists)
	return exists > 0
}

// GetByUuid 根据uuid查询会员
func (r *MemberRepo) GetByUuid(uuid uint64, withs ...With) model.Member {
	var member model.Member
	handleWiths(r.db, withs).Scopes(NotDeleted).Model(&model.Member{}).Where("uuid = ?", uuid).Debug().First(&member)
	return member
}

// Update 更新会员信息
func (r *MemberRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// WithMemberLevel Member 预加载会员等级
func (r *MemberRepo) WithMemberLevel() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberLevel")
	}
}

// WithMemberCard Member 预加载会员卡
func (r *MemberRepo) WithMemberCard() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard")
	}
}

// WithMemberCardType Member 预加载会员卡.卡类型
func (r *MemberRepo) WithMemberCardType() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard.MemberCardType")
	}
}
