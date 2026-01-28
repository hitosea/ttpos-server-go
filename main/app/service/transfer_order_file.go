package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// ITransferOrderFileSrv 调拨单附件服务接口
type ITransferOrderFileSrv interface {
	// SaveTransferOrderFiles 保存调拨单附件关联
	SaveTransferOrderFiles(ctx context.Context, transferOrderUuid uint64, fileUuids []uint64) error

	// GetTransferOrderFiles 查询调拨单附件列表
	GetTransferOrderFiles(ctx context.Context, transferOrderUuid uint64) ([]resp.TransferOrderFileInfo, error)

	// DeleteTransferOrderFile 删除调拨单附件
	DeleteTransferOrderFile(ctx context.Context, fileUuid uint64, transferOrderUuid uint64) error

	// DeleteAllTransferOrderFiles 删除调拨单的所有附件
	DeleteAllTransferOrderFiles(ctx context.Context, transferOrderUuid uint64) error

	// ValidateFileLimit 验证附件数量限制
	ValidateFileLimit(ctx context.Context, transferOrderUuid uint64, newFileCount int) error

	// ValidateTransferOrderStatus 验证调拨单状态（待收货状态才能编辑）
	ValidateTransferOrderStatus(ctx context.Context, transferOrderUuid uint64) error
}

// transferOrderFileSrv 调拨单附件服务实现
type transferOrderFileSrv struct {
	dbm *database.DBManager
}

// NewTransferOrderFileSrv 创建调拨单附件服务
func NewTransferOrderFileSrv(dbm *database.DBManager) ITransferOrderFileSrv {
	return NewTransferOrderFileSrvImpl(dbm)
}

// NewTransferOrderFileSrvImpl 创建调拨单附件服务实现
func NewTransferOrderFileSrvImpl(dbm *database.DBManager) ITransferOrderFileSrv {
	return &transferOrderFileSrv{
		dbm: dbm,
	}
}

// SaveTransferOrderFiles 保存调拨单附件关联
func (s *transferOrderFileSrv) SaveTransferOrderFiles(ctx context.Context, transferOrderUuid uint64, fileUuids []uint64) error {
	if len(fileUuids) == 0 {
		return nil
	}

	// 验证附件数量限制
	if err := s.ValidateFileLimit(ctx, transferOrderUuid, len(fileUuids)); err != nil {
		return err
	}

	db := ctx.GetDB()
	repo := repository.NewTransferOrderFileRepo(db)

	// 先删除旧的附件关联
	if err := repo.DeleteByTransferOrderUuid(transferOrderUuid); err != nil {
		return errors.WithMessage(err, "删除旧附件关联失败")
	}

	// 批量创建附件关联
	var files []model.TransferOrderFile
	for idx, fileUuid := range fileUuids {
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}

		files = append(files, model.TransferOrderFile{
			BaseModel: model.BaseModel{
				Uuid: uuid,
			},
			TransferOrderUuid: transferOrderUuid,
			FileUuid:          fileUuid,
			SortOrder:         idx,
		})
	}

	return repo.BatchCreate(files)
}

// GetTransferOrderFiles 查询调拨单附件列表
func (s *transferOrderFileSrv) GetTransferOrderFiles(ctx context.Context, transferOrderUuid uint64) ([]resp.TransferOrderFileInfo, error) {
	db := ctx.GetDB()
	repo := repository.NewTransferOrderFileRepo(db)

	// 查询附件关联（预加载文件信息）
	files, err := repo.GetByTransferOrderUuidWithFiles(transferOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询调拨单附件失败")
	}

	// 转换为响应格式
	var result = make([]resp.TransferOrderFileInfo, 0)
	baseURL := utils.GetBaseURL(ctx.GetGin().Request)

	for _, file := range files {
		if file.File == nil {
			continue
		}

		result = append(result, resp.TransferOrderFileInfo{
			FileUuid:   file.FileUuid,
			FileName:   file.File.RealName,
			FileSize:   int64(file.File.FileSize),
			FileType:   file.File.FileType,
			Extension:  file.File.Extension,
			FilePath:   file.File.GetUrl(baseURL),
			SortOrder:  file.SortOrder,
			CreateTime: int(file.CreateTime),
		})
	}

	return result, nil
}

// DeleteTransferOrderFile 删除调拨单附件
func (s *transferOrderFileSrv) DeleteTransferOrderFile(ctx context.Context, fileUuid uint64, transferOrderUuid uint64) error {
	// 验证调拨单状态
	if err := s.ValidateTransferOrderStatus(ctx, transferOrderUuid); err != nil {
		return err
	}

	db := ctx.GetDB()
	repo := repository.NewTransferOrderFileRepo(db)

	return repo.DeleteByFileUuidAndTransferOrderUuid(fileUuid, transferOrderUuid)
}

// DeleteAllTransferOrderFiles 删除调拨单的所有附件
func (s *transferOrderFileSrv) DeleteAllTransferOrderFiles(ctx context.Context, transferOrderUuid uint64) error {
	db := ctx.GetDB()
	repo := repository.NewTransferOrderFileRepo(db)

	return repo.DeleteByTransferOrderUuid(transferOrderUuid)
}

// ValidateFileLimit 验证附件数量限制
func (s *transferOrderFileSrv) ValidateFileLimit(ctx context.Context, transferOrderUuid uint64, newFileCount int) error {
	if newFileCount > 10 {
		return errors.New("最多支持10个附件")
	}

	return nil
}

// ValidateTransferOrderStatus 验证调拨单状态（待收货状态才能编辑）
func (s *transferOrderFileSrv) ValidateTransferOrderStatus(ctx context.Context, transferOrderUuid uint64) error {
	db := ctx.GetDB()
	transferOrderRepo := repository.NewTransferOrderRepo(db)

	// 查询调拨单
	transferOrder, err := transferOrderRepo.GetByUuid(transferOrderUuid)
	if err != nil {
		return errors.WithMessage(err, "调拨单不存在")
	}

	// 仅待收货状态可以编辑附件
	if transferOrder.Status != constant.TransferOrderStatusReceiving {
		return errors.New("仅待收货状态的调拨单可以修改附件")
	}

	return nil
}
