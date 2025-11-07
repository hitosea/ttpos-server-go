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

// IPurchaseReceiptFileSrv 收货单附件服务接口
type IPurchaseReceiptFileSrv interface {
	// SaveReceiptFiles 保存收货单附件关联
	SaveReceiptFiles(ctx context.Context, receiptOrderUuid uint64, fileUuids []uint64) error

	// GetReceiptFiles 查询收货单附件列表
	GetReceiptFiles(ctx context.Context, receiptOrderUuid uint64) ([]resp.ReceiptFileInfo, error)

	// DeleteReceiptFile 删除收货单附件
	DeleteReceiptFile(ctx context.Context, fileUuid uint64, receiptOrderUuid uint64) error

	// DeleteAllReceiptFiles 删除收货单的所有附件
	DeleteAllReceiptFiles(ctx context.Context, receiptOrderUuid uint64) error

	// ValidateFileLimit 验证附件数量限制
	ValidateFileLimit(ctx context.Context, receiptOrderUuid uint64, newFileCount int) error

	// ValidateReceiptStatus 验证收货单状态（草稿状态才能编辑）
	ValidateReceiptStatus(ctx context.Context, receiptOrderUuid uint64) error
}

// purchaseReceiptFileSrv 收货单附件服务实现
type purchaseReceiptFileSrv struct {
	dbm *database.DBManager
}

// NewPurchaseReceiptFileSrv 创建收货单附件服务
func NewPurchaseReceiptFileSrv(dbm *database.DBManager) IPurchaseReceiptFileSrv {
	return NewPurchaseReceiptFileSrvImpl(dbm)
}

// NewPurchaseReceiptFileSrvImpl 创建收货单附件服务实现
func NewPurchaseReceiptFileSrvImpl(dbm *database.DBManager) IPurchaseReceiptFileSrv {
	return &purchaseReceiptFileSrv{
		dbm: dbm,
	}
}

// SaveReceiptFiles 保存收货单附件关联
func (s *purchaseReceiptFileSrv) SaveReceiptFiles(ctx context.Context, receiptOrderUuid uint64, fileUuids []uint64) error {
	if len(fileUuids) == 0 {
		return nil
	}

	// 验证附件数量限制
	if err := s.ValidateFileLimit(ctx, receiptOrderUuid, len(fileUuids)); err != nil {
		return err
	}

	db := ctx.GetDB()
	repo := repository.NewPurchaseReceiptFileRepo(db)

	// 批量创建附件关联
	var files []model.PurchaseReceiptFile
	for idx, fileUuid := range fileUuids {
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}

		files = append(files, model.PurchaseReceiptFile{
			BaseModel: model.BaseModel{
				Uuid: uuid,
			},
			ReceiptOrderUuid: receiptOrderUuid,
			FileUuid:         fileUuid,
			SortOrder:        idx,
		})
	}

	return repo.BatchCreate(files)
}

// GetReceiptFiles 查询收货单附件列表
func (s *purchaseReceiptFileSrv) GetReceiptFiles(ctx context.Context, receiptOrderUuid uint64) ([]resp.ReceiptFileInfo, error) {
	db := ctx.GetDB()
	repo := repository.NewPurchaseReceiptFileRepo(db)

	// 查询附件关联（预加载文件信息）
	files, err := repo.GetByReceiptOrderUuidWithFiles(receiptOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询收货单附件失败")
	}

	// 转换为响应格式
	var result = make([]resp.ReceiptFileInfo, 0)
	baseURL := utils.GetBaseURL(ctx.GetGin().Request)

	for _, file := range files {
		if file.File == nil {
			continue
		}

		result = append(result, resp.ReceiptFileInfo{
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

// DeleteReceiptFile 删除收货单附件
func (s *purchaseReceiptFileSrv) DeleteReceiptFile(ctx context.Context, fileUuid uint64, receiptOrderUuid uint64) error {
	// 验证收货单状态
	if err := s.ValidateReceiptStatus(ctx, receiptOrderUuid); err != nil {
		return err
	}

	db := ctx.GetDB()
	repo := repository.NewPurchaseReceiptFileRepo(db)

	return repo.DeleteByFileUuidAndReceiptOrderUuid(fileUuid, receiptOrderUuid)
}

// DeleteAllReceiptFiles 删除收货单的所有附件
func (s *purchaseReceiptFileSrv) DeleteAllReceiptFiles(ctx context.Context, receiptOrderUuid uint64) error {
	db := ctx.GetDB()
	repo := repository.NewPurchaseReceiptFileRepo(db)

	return repo.DeleteByReceiptOrderUuid(receiptOrderUuid)
}

// ValidateFileLimit 验证附件数量限制
func (s *purchaseReceiptFileSrv) ValidateFileLimit(ctx context.Context, receiptOrderUuid uint64, newFileCount int) error {
	if newFileCount > 10 {
		return errors.New("最多支持10个附件")
	}

	db := ctx.GetDB()
	repo := repository.NewPurchaseReceiptFileRepo(db)

	// 查询当前附件数量
	currentCount, err := repo.CountByReceiptOrderUuid(receiptOrderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询附件数量失败")
	}

	if currentCount+int64(newFileCount) > 10 {
		return errors.New("最多支持10个附件")
	}

	return nil
}

// ValidateReceiptStatus 验证收货单状态（草稿状态才能编辑）
func (s *purchaseReceiptFileSrv) ValidateReceiptStatus(ctx context.Context, receiptOrderUuid uint64) error {
	db := ctx.GetDB()
	receiptRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 查询收货单
	receiptOrder, err := receiptRepo.GetByUuid(receiptOrderUuid)
	if err != nil {
		return errors.WithMessage(err, "收货单不存在")
	}

	// 仅草稿状态可以编辑附件
	if receiptOrder.Status != constant.ReceiptOrderStatusPending {
		return errors.New("仅草稿状态的收货单可以修改附件")
	}

	return nil
}
