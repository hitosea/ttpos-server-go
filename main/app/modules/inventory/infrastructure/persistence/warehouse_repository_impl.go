package inventory

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/app/modules/inventory/domain/specification"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
	appRepo "ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// WarehouseRepositoryImpl 仓库仓储实现
type WarehouseRepositoryImpl struct {
	dbm *database.DBManager
}

// NewWarehouseRepository 创建仓库仓储
func NewWarehouseRepository(dbm *database.DBManager) repository.IWarehouseRepository {
	return &WarehouseRepositoryImpl{
		dbm: dbm,
	}
}

// Save 保存仓库
func (r *WarehouseRepositoryImpl) Save(ctx context.Context, warehouse *entity.Warehouse) error {
	db := r.getDB(ctx)
	po := r.toPO(warehouse)

	if warehouse.Uuid() == 0 {
		// 生成UUID
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}
		po.Uuid = uuid
		warehouse.SetUuid(uuid)

		// 创建
		if err := db.Create(po).Error; err != nil {
			return errors.WithMessage(err, "创建仓库失败")
		}
	} else {
		// 更新
		if err := db.Model(&model.Warehouse{}).Where("uuid = ?", po.Uuid).Updates(po).Error; err != nil {
			return errors.WithMessage(err, "更新仓库失败")
		}
	}

	return nil
}

// FindByUuid 根据UUID查找仓库
func (r *WarehouseRepositoryImpl) FindByUuid(ctx context.Context, uuid uint64) (*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	po, err := repo.GetByUuid(uuid,
		repo.WithMultiLanguageName(),
		repo.WithItems(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询仓库失败")
	}

	return r.toDomain(po), nil
}

// FindByCode 根据编码查找仓库
func (r *WarehouseRepositoryImpl) FindByCode(ctx context.Context, code valueobject.WarehouseCode) (*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	po, err := repo.GetByCode(code.Value(),
		repo.WithMultiLanguageName(),
		repo.WithItems(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询仓库失败")
	}

	return r.toDomain(po), nil
}

// FindByErpCode 根据ERP编码查找仓库
func (r *WarehouseRepositoryImpl) FindByErpCode(ctx context.Context, erpCode string) (*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	po, err := repo.GetByErpCode(erpCode,
		repo.WithMultiLanguageName(),
		repo.WithItems(),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询仓库失败")
	}

	return r.toDomain(po), nil
}

// Remove 删除仓库
func (r *WarehouseRepositoryImpl) Remove(ctx context.Context, uuid uint64) error {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	if err := repo.Delete(uuid); err != nil {
		return errors.WithMessage(err, "删除仓库失败")
	}

	return nil
}

// ExistsCode 检查编码是否存在
func (r *WarehouseRepositoryImpl) ExistsCode(ctx context.Context, code valueobject.WarehouseCode, excludeUuid uint64) (bool, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	exists, err := repo.IsCodeExists(code.Value(), excludeUuid)
	if err != nil {
		return false, errors.WithMessage(err, "检查仓库编码失败")
	}

	return exists, nil
}

// FindDefault 查找默认仓库
func (r *WarehouseRepositoryImpl) FindDefault(ctx context.Context) (*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	po, err := repo.GetDefaultWarehouse()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询默认仓库失败")
	}

	return r.toDomain(po), nil
}

// FindAllNormal 查找所有普通仓库
func (r *WarehouseRepositoryImpl) FindAllNormal(ctx context.Context) ([]*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	pos, err := repo.GetNormalWarehouse()
	if err != nil {
		return nil, errors.WithMessage(err, "查询普通仓库失败")
	}

	warehouses := make([]*entity.Warehouse, 0, len(pos))
	for _, po := range pos {
		warehouses = append(warehouses, r.toDomain(&po))
	}

	return warehouses, nil
}

// FindTransit 查找在途仓库
func (r *WarehouseRepositoryImpl) FindTransit(ctx context.Context) (*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	po, err := repo.GetTransitWarehouse()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询在途仓库失败")
	}

	return r.toDomain(po), nil
}

// FindByHeadquarterUuid 根据总部UUID查找仓库
func (r *WarehouseRepositoryImpl) FindByHeadquarterUuid(ctx context.Context, headquarterUuid uint64) ([]*entity.Warehouse, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	pos, err := repo.Get(repo.WhereHeadquarterUuid(headquarterUuid))
	if err != nil {
		return nil, errors.WithMessage(err, "查询仓库失败")
	}

	warehouses := make([]*entity.Warehouse, 0, len(pos))
	for _, po := range pos {
		warehouses = append(warehouses, r.toDomain(&po))
	}

	return warehouses, nil
}

// FindWithPagination 分页查询仓库
func (r *WarehouseRepositoryImpl) FindWithPagination(
	ctx context.Context,
	spec repository.WarehouseSpecification,
	pageNo, pageSize int,
) ([]*entity.Warehouse, int64, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	// 构建查询选项
	var opts []appRepo.DBOption

	// 转换规格为查询选项
	if querySpec, ok := spec.(*specification.WarehouseQuerySpec); ok {
		if querySpec.Keyword() != "" {
			opts = append(opts, repo.WhereNameOrCodeLike(querySpec.Keyword()))
		}
		if querySpec.WarehouseType() != nil {
			opts = append(opts, repo.WhereType(querySpec.WarehouseType().String()))
		}
		if querySpec.Status() != nil {
			opts = append(opts, repo.WhereStatus(querySpec.Status().ToInt()))
		}
		if querySpec.IsHeadquarter() != nil {
			opts = append(opts, repo.WhereIsHeadquarter(*querySpec.IsHeadquarter()))
		}
		if querySpec.HeadquarterUuid() != nil {
			opts = append(opts, repo.WhereHeadquarterUuid(*querySpec.HeadquarterUuid()))
		}
	}

	// 排序
	opts = append(opts, repo.OrderByUpdateTime(true))

	// 分页查询
	pos, total, err := repo.GetListWithPagination(pageNo, pageSize, opts...)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "查询仓库列表失败")
	}

	warehouses := make([]*entity.Warehouse, 0, len(pos))
	for _, po := range pos {
		warehouses = append(warehouses, r.toDomain(&po))
	}

	return warehouses, total, nil
}

// SetAsDefault 设置默认仓库
func (r *WarehouseRepositoryImpl) SetAsDefault(ctx context.Context, uuid uint64) error {
	db := r.getDB(ctx)
	repo := appRepo.NewWarehouseRepo(db)

	if err := repo.UpdateIsDefault(uuid); err != nil {
		return errors.WithMessage(err, "设置默认仓库失败")
	}

	return nil
}

// toPO 转换为持久化对象
func (r *WarehouseRepositoryImpl) toPO(warehouse *entity.Warehouse) *model.Warehouse {
	localeResponse := warehouse.Name().ToLocaleResponse()
	nameJson := localeResponse.ToJson()

	return &model.Warehouse{
		BaseModel: model.BaseModel{
			Uuid:       warehouse.Uuid(),
			CreateTime: warehouse.CreateTime(),
			UpdateTime: warehouse.UpdateTime(),
		},
		Name:                  nameJson,
		MultiLanguageNameUuid: warehouse.Name().Uuid(),
		Type:                  warehouse.WarehouseType().String(),
		Code:                  warehouse.Code().Value(),
		Status:                warehouse.Status().ToInt(),
		Contact:               warehouse.ContactInfo().Contact(),
		Phone:                 warehouse.ContactInfo().Phone(),
		Address:               warehouse.ContactInfo().Address(),
		IsDefault:             boolToInt(warehouse.IsDefault()),
		ErpCode:               warehouse.ErpCode(),
		HeadquarterUuid:       warehouse.HeadquarterUuid(),
	}
}

// toDomain 转换为领域对象
func (r *WarehouseRepositoryImpl) toDomain(po *model.Warehouse) *entity.Warehouse {
	// 转换多语言名称
	var localeName dto.LocaleResponse
	if po.MultiLanguageName != nil {
		localeName = po.MultiLanguageName.GetNames()
	} else {
		localeName = *language.JsonToLocaleResponse(po.Name)
	}

	name := valueobject.NewMultiLanguageNameWithUuid(po.MultiLanguageNameUuid, localeName)
	warehouseType := valueobject.NewWarehouseType(po.Type)
	code, _ := valueobject.NewWarehouseCode(po.Code)
	status := valueobject.NewWarehouseStatus(po.Status)
	contactInfo := valueobject.NewContactInfo(po.Contact, po.Phone, po.Address)

	hasItems := po.Items != nil && len(po.Items) > 0

	return entity.ReconstructWarehouse(
		po.Uuid,
		name,
		warehouseType,
		code,
		status,
		contactInfo,
		intToBool(po.IsDefault),
		po.ErpCode,
		po.HeadquarterUuid,
		po.CreateTime,
		po.UpdateTime,
		hasItems,
	)
}

// getDB 获取数据库连接
func (r *WarehouseRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	return ctx.GetDB()
}

// 辅助函数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i == 1
}
