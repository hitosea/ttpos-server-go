package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberSaleOrderAddressRepo interface {
	GetMemberSaleOrderAddress(opts ...DBOption) (*model.MemberSaleOrderAddress, error)
	GetMemberSaleOrderAddressRecord(uuid uint64) (*model.MemberSaleOrderAddress, error)
	CreateMemberSaleOrderAddress(memberSaleOrderAddress model.MemberSaleOrderAddress) error
}

func NewMemberSaleOrderAddressRepo(db *gorm.DB) IMemberSaleOrderAddressRepo {
	return NewMemberSaleOrderAddressRepoImpl(db)
}

type MemberSaleOrderAddressRepo struct {
	db *gorm.DB
}

func NewMemberSaleOrderAddressRepoImpl(db *gorm.DB) *MemberSaleOrderAddressRepo {
	return &MemberSaleOrderAddressRepo{db: db}
}

func (r *MemberSaleOrderAddressRepo) GetMemberSaleOrderAddress(opts ...DBOption) (*model.MemberSaleOrderAddress, error) {
	var memberSaleOrderAddress model.MemberSaleOrderAddress
	db := r.db.Model(&model.MemberSaleOrderAddress{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&memberSaleOrderAddress).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &memberSaleOrderAddress, nil
}

func (r *MemberSaleOrderAddressRepo) GetMemberSaleOrderAddressRecord(uuid uint64) (*model.MemberSaleOrderAddress, error) {
	memberSaleOrderAddress, err := r.GetMemberSaleOrderAddress(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberSaleOrderAddress, nil
}

func (r *MemberSaleOrderAddressRepo) CreateMemberSaleOrderAddress(memberSaleOrderAddress model.MemberSaleOrderAddress) error {
	memberSaleOrderAddress.SetNil()
	if err := r.db.Create(&memberSaleOrderAddress).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
