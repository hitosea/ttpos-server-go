package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IMemberRepo interface {
	GetMemberLevels() []model.MemberLevel
	SearchMember(keyword string) []model.Member
	CreateMember(member model.Member) error
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

func (r *MemberRepo) GetMemberLevels() []model.MemberLevel {
	var levels []model.MemberLevel
	r.db.Scopes(UnDelete).Model(&model.MemberLevel{}).Select("uuid, name, priority, create_time").Order("priority asc, create_time asc").Find(&levels)
	return levels
}

func (r *MemberRepo) SearchMember(keyword string) []model.Member {
	var members []model.Member
	keyword = Like(keyword)
	r.db.Scopes(UnDelete).Model(&model.Member{}).Select("uuid, nickname, phone").Where("phone LIKE ? OR uuid LIKE ?", keyword, keyword).Find(&members)
	return members
}

func (r *MemberRepo) CreateMember(member model.Member) error {
	return r.db.Create(&member).Error
}
