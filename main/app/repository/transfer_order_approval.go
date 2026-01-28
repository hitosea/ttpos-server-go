package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITransferOrderApprovalRepo 调拨单审批流程Repository接口
type ITransferOrderApprovalRepo interface {
	Create(approval *model.TransferOrderApproval) error
	CreateBatch(approvals []*model.TransferOrderApproval) error
	Update(approval *model.TransferOrderApproval) error
	UpdateAll(approval *model.TransferOrderApproval) error // 更新所有字段（包括零值）
	Delete(uuid uint64) error
	DeleteByTransferOrderUuid(transferOrderUuid uint64) error
	GetByUuid(uuid uint64) (*model.TransferOrderApproval, error)
	GetListByTransferOrderUuid(transferOrderUuid uint64) ([]*model.TransferOrderApproval, error) // 根据调拨单UUID获取审批流程列表
	GetCurrentApproval(transferOrderUuid uint64, companyUuid uint64) (*model.TransferOrderApproval, error)
	GetNextApproval(transferOrderUuid uint64, currentSequence int) (*model.TransferOrderApproval, error)
}

// TransferOrderApprovalRepoImpl 调拨单审批流程Repository实现
type TransferOrderApprovalRepoImpl struct {
	db *gorm.DB
}

// NewTransferOrderApprovalRepo 创建调拨单审批流程Repository
func NewTransferOrderApprovalRepo(db *gorm.DB) ITransferOrderApprovalRepo {
	return &TransferOrderApprovalRepoImpl{db: db}
}

func (r *TransferOrderApprovalRepoImpl) Create(approval *model.TransferOrderApproval) error {
	return r.db.Create(approval).Error
}

func (r *TransferOrderApprovalRepoImpl) CreateBatch(approvals []*model.TransferOrderApproval) error {
	if len(approvals) == 0 {
		return nil
	}
	return r.db.Create(approvals).Error
}

func (r *TransferOrderApprovalRepoImpl) Update(approval *model.TransferOrderApproval) error {
	return r.db.Model(&model.TransferOrderApproval{}).Where("uuid = ?", approval.Uuid).Updates(approval).Error
}

// UpdateAll 更新所有字段（包括零值）
func (r *TransferOrderApprovalRepoImpl) UpdateAll(approval *model.TransferOrderApproval) error {
	return r.db.Model(&model.TransferOrderApproval{}).Select("*").Where("uuid = ?", approval.Uuid).Updates(approval).Error
}

func (r *TransferOrderApprovalRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.TransferOrderApproval{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

func (r *TransferOrderApprovalRepoImpl) DeleteByTransferOrderUuid(transferOrderUuid uint64) error {
	return r.db.Model(&model.TransferOrderApproval{}).Where("transfer_order_uuid = ?", transferOrderUuid).Update("delete_time", time.Now().Unix()).Error
}

func (r *TransferOrderApprovalRepoImpl) GetByUuid(uuid uint64) (*model.TransferOrderApproval, error) {
	var approval model.TransferOrderApproval
	err := r.db.Model(&model.TransferOrderApproval{}).Where("uuid = ?", uuid).Where("delete_time = ?", constant.NotDeleted).First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// GetListByTransferOrderUuid 根据调拨单UUID获取审批流程列表
func (r *TransferOrderApprovalRepoImpl) GetListByTransferOrderUuid(transferOrderUuid uint64) ([]*model.TransferOrderApproval, error) {
	var approvals []*model.TransferOrderApproval
	err := r.db.Model(&model.TransferOrderApproval{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Where("delete_time = ?", constant.NotDeleted).
		Order("sequence ASC").
		Find(&approvals).Error
	return approvals, err
}

func (r *TransferOrderApprovalRepoImpl) GetCurrentApproval(transferOrderUuid uint64, companyUuid uint64) (*model.TransferOrderApproval, error) {
	var approval model.TransferOrderApproval
	err := r.db.Model(&model.TransferOrderApproval{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Where("approval_company_uuid = ?", companyUuid).
		Where("status = ?", constant.TransferApprovalPending).
		Where("is_required = ?", 1).
		Where("delete_time = ?", constant.NotDeleted).
		Order("sequence ASC").
		First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *TransferOrderApprovalRepoImpl) GetNextApproval(transferOrderUuid uint64, currentSequence int) (*model.TransferOrderApproval, error) {
	var approval model.TransferOrderApproval
	err := r.db.Model(&model.TransferOrderApproval{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Where("sequence > ?", currentSequence).
		Where("status = ?", constant.TransferApprovalPending).
		Where("is_required = ?", 1).
		Where("delete_time = ?", constant.NotDeleted).
		Order("sequence ASC").
		First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}
