package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberAddressRepo interface {
	Create(address model.MemberAddress) (model.MemberAddress, error)                        // 创建会员地址
	GetMemberAddressList(opts ...DBOption) ([]*model.MemberAddress, error)                  // 获取会员地址列表
	GetMemberAddress(opts ...DBOption) (*model.MemberAddress, error)                        // 获取会员地址
	PaginateGet(page, pageSize int, opts ...DBOption) ([]model.MemberAddress, int64, error) // 分页获取会员地址列表
	GetMemberAddressByUuid(uuid uint64) (*model.MemberAddress, error)                       // 获取会员地址
	GetMemberAddressByMemberUuid(memberUuid uint64) ([]*model.MemberAddress, error)         // 获取会员地址
	Update(address model.MemberAddress) (model.MemberAddress, error)                        // 更新会员地址
	Delete(uuid uint64) error                                                               // 删除会员地址
	UpdateIsDefault(memberUuid uint64) error                                                // 将该会员的所有地址设置为非默认
}

func NewMemberAddressRepo(db *gorm.DB) IMemberAddressRepo {
	return NewMemberAddressRepoImpl(db)
}

type MemberAddressRepo struct {
	db *gorm.DB
}

func NewMemberAddressRepoImpl(db *gorm.DB) *MemberAddressRepo {
	return &MemberAddressRepo{db: db}
}

// Create 创建会员地址
func (r *MemberAddressRepo) Create(address model.MemberAddress) (model.MemberAddress, error) {
	err := r.db.Model(&model.MemberAddress{}).Create(&address).Error
	return address, errors.WithMessage(err)
}

// GetMemberAddress 获取会员地址
func (r *MemberAddressRepo) GetMemberAddress(opts ...DBOption) (*model.MemberAddress, error) {
	var address model.MemberAddress
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&address).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &address, nil
}

// GetMemberAddressList 获取会员地址列表
func (r *MemberAddressRepo) GetMemberAddressList(opts ...DBOption) ([]*model.MemberAddress, error) {
	var addresses []*model.MemberAddress
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&addresses)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}
	return addresses, nil
}

// PaginateGet 分页获取会员地址列表
func (r *MemberAddressRepo) PaginateGet(page, pageSize int, opts ...DBOption) ([]model.MemberAddress, int64, error) {
	var addresses []model.MemberAddress
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	total := int64(0)
	db.Model(&model.MemberAddress{}).Where(db).Count(&total)

	db.Order("is_default desc, create_time desc").Where("delete_time = 0").Offset((page - 1) * pageSize).Limit(pageSize).Find(&addresses)
	return addresses, total, nil
}

// GetMemberAddressByUuid 获取会员地址
func (r *MemberAddressRepo) GetMemberAddressByUuid(uuid uint64) (*model.MemberAddress, error) {
	memberAddress, err := r.GetMemberAddress(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "Member",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberAddress, nil
}

// GetMemberAddressByMemberUuid 获取会员地址
func (r *MemberAddressRepo) GetMemberAddressByMemberUuid(memberUuid uint64) ([]*model.MemberAddress, error) {
	var addresses []*model.MemberAddress
	err := r.db.Model(&model.MemberAddress{}).
		Where("delete_time = 0").
		Where("member_uuid = ?", memberUuid).
		Order("is_default desc, create_time desc").
		Find(&addresses).Error
	return addresses, errors.WithMessage(err)
}

// Update 更新会员地址
func (r *MemberAddressRepo) Update(address model.MemberAddress) (model.MemberAddress, error) {
	err := r.db.Model(&model.MemberAddress{}).Where("uuid = ?", address.Uuid).Updates(&address).Error
	return address, errors.WithMessage(err)
}

// Delete 删除会员地址
func (r *MemberAddressRepo) Delete(uuid uint64) error {
	err := r.db.Model(&model.MemberAddress{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
	return errors.WithMessage(err)
}

// UpdateIsDefault 将该会员的所有地址设置为非默认
func (r *MemberAddressRepo) UpdateIsDefault(memberUuid uint64) error {
	err := r.db.Model(&model.MemberAddress{}).
		Where("member_uuid = ?", memberUuid).
		Update("is_default", 0).Error
	return errors.WithMessage(err)
}
