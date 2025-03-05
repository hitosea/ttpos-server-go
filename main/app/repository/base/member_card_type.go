package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMemberCardTypeRepo 会员卡类型
type IMemberCardTypeRepo interface {
	GetMemberCardTypeList() ([]model.MemberCardType, error)
	UpdateMemberCardType(uuid uint, memberCardType model.MemberCardType) error
	CreateMemberCardType(memberCardType model.MemberCardType) (uint64, error)
	DeleteMemberCardType(uuid uint) error
}

func NewMemberCardTypeRepo(db *gorm.DB) IMemberCardTypeRepo {
	return NewMemberCardTypeRepoImpl(db)
}

// NewMemberCardTypeRepoImpl 创建新的商品规格仓库实现
func NewMemberCardTypeRepoImpl(db *gorm.DB) *MemberCardTypeRepo {
	return &MemberCardTypeRepo{db: db}
}

type MemberCardTypeRepo struct {
	db *gorm.DB
}

// GetMemberCardTypeList 获取商品规格列表，排除逻辑删除的规格
func (r *MemberCardTypeRepo) GetMemberCardTypeList() ([]model.MemberCardType, error) {
	var memberCardTypes []model.MemberCardType
	err := r.db.Model(&model.MemberCardType{}).Where("delete_time = ?", 0).Find(&memberCardTypes).Error
	return memberCardTypes, errors.WithMessage(err)
}

// UpdateMemberCardType 更新自助餐客户类型
func (r *MemberCardTypeRepo) UpdateMemberCardType(uuid uint, memberCardType model.MemberCardType) error {
	if err := r.db.Model(&model.MemberCardType{}).Where("uuid = ?", uuid).Updates(memberCardType).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateMemberCardType 创建会员卡类型
func (r *MemberCardTypeRepo) CreateMemberCardType(memberCardType model.MemberCardType) (uint64, error) {
	// 创建会员卡类型
	if err := r.db.Create(&memberCardType).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return memberCardType.Uuid, nil
}

// DeleteMemberCardType 软删除会员卡类型
func (r *MemberCardTypeRepo) DeleteMemberCardType(uuid uint) error {
	return r.db.Model(&model.MemberCardType{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
