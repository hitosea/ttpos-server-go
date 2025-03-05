package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMemberLevelRepo 会员等级
type IMemberLevelRepo interface {
	GetMemberLevelList() ([]model.MemberLevel, error)
	UpdateMemberLevel(uuid uint, memberLevel model.MemberLevel) error
	CreateMemberLevel(memberLevel model.MemberLevel) (uint64, error)
	DeleteMemberLevel(uuid uint) error
}

func NewMemberLevelRepo(db *gorm.DB) IMemberLevelRepo {
	return NewMemberLevelRepoImpl(db)
}

// NewMemberLevelRepoImpl 创建新的商品规格仓库实现
func NewMemberLevelRepoImpl(db *gorm.DB) *MemberLevelRepo {
	return &MemberLevelRepo{db: db}
}

type MemberLevelRepo struct {
	db *gorm.DB
}

// GetMemberLevelList 获取商品规格列表，排除逻辑删除的规格
func (r *MemberLevelRepo) GetMemberLevelList() ([]model.MemberLevel, error) {
	var memberLevels []model.MemberLevel
	err := r.db.Model(&model.MemberLevel{}).Where("delete_time = ?", 0).Find(&memberLevels).Error
	return memberLevels, err
}

// UpdateMemberLevel 更新自助餐客户类型
func (r *MemberLevelRepo) UpdateMemberLevel(uuid uint, memberLevel model.MemberLevel) error {
	if err := r.db.Model(&model.MemberLevel{}).Where("uuid = ?", uuid).Updates(memberLevel).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateMemberLevel 创建会员等级
func (r *MemberLevelRepo) CreateMemberLevel(memberLevel model.MemberLevel) (uint64, error) {
	// 创建会员等级
	if err := r.db.Create(&memberLevel).Error; err != nil {
		return 0, err
	}
	return memberLevel.Uuid, nil
}

// DeleteMemberLevel 软删除会员等级
func (r *MemberLevelRepo) DeleteMemberLevel(uuid uint) error {
	return r.db.Model(&model.MemberLevel{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
