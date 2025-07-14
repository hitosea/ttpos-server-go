package saas

import (
	saas_model "ttpos-server-go/app/model/saas"

	"gorm.io/gorm"
)

type IMemberRepo interface {
	CreateMember(member *saas_model.Member) error // 添加会员
	CheckMemberExists(phone string) bool          // 根据手机号检查是否存在
	CheckNicknameExists(nickname string) bool     // 根据昵称检查是否存在
}

func NewMemberRepo(db *gorm.DB) IMemberRepo {
	return NewMemberRepoImpl(db)
}

type memberRepo struct {
	db *gorm.DB
}

func NewMemberRepoImpl(db *gorm.DB) IMemberRepo {
	return &memberRepo{db: db}
}

// CreateMember 添加会员
func (r *memberRepo) CreateMember(member *saas_model.Member) error {
	err := r.db.Create(member).Error
	return err
}

// CheckMemberExists 根据手机号检查是否存在
func (r *memberRepo) CheckMemberExists(phone string) bool {
	var memberUuid uint64
	r.db.Model(&saas_model.Member{}).Where("phone = ?", phone).Select("uuid").Scan(&memberUuid)
	return memberUuid > 0
}

// CheckNicknameExists 根据昵称检查是否存在
func (r *memberRepo) CheckNicknameExists(nickname string) bool {
	var exists uint64
	r.db.Model(&saas_model.Member{}).Where("nickname = ?", nickname).Select("uuid").Scan(&exists)
	return exists > 0
}
