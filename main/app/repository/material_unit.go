package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialUnitRepo 原料单位仓库接口
type IMaterialUnitRepo interface {
	// 基础CRUD操作
	GetMaterialUnitByUuid(uuid uint64, opts ...DBOption) (model.MaterialUnit, error)
	GetMaterialUnitsByUuid(uuid uint64, opts ...DBOption) (model.MaterialUnit, error)
	GetMaterialUnitList(opts ...DBOption) ([]*model.MaterialUnit, error)
	CreateMaterialUnit(materialUnit model.MaterialUnit) (uint64, error)
	GetMaterialUnitListByBaseUnitUuid(baseUnitUuid uint64) ([]*model.MaterialUnit, error)
}

// NewMaterialUnitRepo 创建新的原料单位仓库
func NewMaterialUnitRepo(db *gorm.DB) IMaterialUnitRepo {
	return NewMaterialUnitRepoImpl(db)
}

// NewMaterialUnitRepoImpl 创建新的原料单位仓库实现
func NewMaterialUnitRepoImpl(db *gorm.DB) *MaterialUnitRepoImpl {
	return &MaterialUnitRepoImpl{db: db}
}

type MaterialUnitRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetMaterialUnitByUuid 根据UUID获取原料单位详情
func (r *MaterialUnitRepoImpl) GetMaterialUnitByUuid(uuid uint64, opts ...DBOption) (model.MaterialUnit, error) {
	var materialUnit model.MaterialUnit

	query := r.db.Model(&model.MaterialUnit{}).Where("uuid = ? AND delete_time = ?", uuid, 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.First(&materialUnit).Error; err != nil {
		return model.MaterialUnit{}, errors.WithMessage(err, "查询原料单位详情失败")
	}

	return materialUnit, nil
}

// CreateMaterialUnit 创建原料单位
func (r *MaterialUnitRepoImpl) CreateMaterialUnit(materialUnit model.MaterialUnit) (uint64, error) {
	materialUnit.SetNil()
	if err := r.db.Create(&materialUnit).Error; err != nil {
		return 0, errors.WithMessage(err, "创建原料单位失败")
	}
	return materialUnit.Uuid, nil
}

func (r *MaterialUnitRepoImpl) GetMaterialUnitList(opts ...DBOption) ([]*model.MaterialUnit, error) {
	var materialUnits []*model.MaterialUnit
	query := r.db.Model(&model.MaterialUnit{}).Where("delete_time = ?", 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.Find(&materialUnits).Error; err != nil {
		return nil, errors.WithMessage(err, "查询原料单位列表失败")
	}
	return materialUnits, nil
}

func (r *MaterialUnitRepoImpl) GetMaterialUnitsByUuid(baseUnitUuid uint64, opts ...DBOption) (model.MaterialUnit, error) {
	return r.GetMaterialUnitByUuid(
		baseUnitUuid,
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "Unit",
			},
		),
	)
}

func (r *MaterialUnitRepoImpl) GetMaterialUnitListByBaseUnitUuid(baseUnitUuid uint64) ([]*model.MaterialUnit, error) {
	return r.GetMaterialUnitList(
		CommonRepo.WhereByFromUnitUuid(baseUnitUuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "Unit.MultiLanguageName",
			},
		),
	)
}
