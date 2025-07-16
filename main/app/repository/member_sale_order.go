package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberSaleOrderRepo interface {
	GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error)
	GetMemberSaleOrderRecord(uuid uint64) (*model.MemberSaleOrder, error)
	CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error
}

func NewMemberSaleOrderRepo(db *gorm.DB) IMemberSaleOrderRepo {
	return NewMemberSaleOrderRepoImpl(db)
}

type MemberSaleOrderRepo struct {
	db *gorm.DB
}

func NewMemberSaleOrderRepoImpl(db *gorm.DB) *MemberSaleOrderRepo {
	return &MemberSaleOrderRepo{db: db}
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error) {
	var memberSaleOrder model.MemberSaleOrder
	db := r.db.Model(&model.MemberSaleOrder{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&memberSaleOrder).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &memberSaleOrder, nil
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrderRecord(uuid uint64) (*model.MemberSaleOrder, error) {
	memberSaleOrder, err := r.GetMemberSaleOrder(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "Address",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberSaleOrder, nil
}

func (r *MemberSaleOrderRepo) CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error {
	memberSaleOrder.SetNil()
	if err := r.db.Create(&memberSaleOrder).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
