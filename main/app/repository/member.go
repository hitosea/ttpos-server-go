package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IMemberRepo interface {
	GetMemberLevels() []model.MemberLevel       // 获取会员等级
	SearchMember(keyword string) []model.Member // 关键字搜索会员
	CreateMember(member model.Member) error     // 添加会员
	CheckMemberExists(phone string) bool        // 根据手机号检查是否存在
	CheckLevelExists(uuid uint64) bool          // 根据Uuid检查等级是否存在
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
