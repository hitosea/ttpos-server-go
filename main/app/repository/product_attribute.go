package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductAttributeRepo interface {
	GetProductAttribute(opts ...DBOption) (*model.ProductAttribute, error)
	GetProductAttributes(opts ...DBOption) ([]*model.ProductAttribute, error)
	GetProductAttributeByUuid(uuid uint64) (*model.ProductAttribute, error)                      // 根据UUID查询商品属性（软删除过滤）
	GetProductAttributesByUuids(uuids []uint64) ([]*model.ProductAttribute, error)               // 批量根据UUID列表查询商品属性（包含Preload）
	GetProductAttributeWithMultiLanguageName(uuid uint64) (*model.ProductAttribute, error)       // 根据UUID查询商品属性（包含MultiLanguageName）
	GetProductAttributesWithMultiLanguageName(uuids []uint64) ([]*model.ProductAttribute, error) // 批量根据UUID列表查询商品属性（包含MultiLanguageName）
	UpdateNameByMultiLanguageNameUuid(multiLanguageNameUuid uint64, name string) error
	CreateRecord(attr *model.ProductAttribute) error
	CreateBatch(attrs []model.ProductAttribute) error
	UpdateProductAttributeData(data map[string]any, opts ...DBOption) error
	HardDeleteByAttributeGroupUuids(uuids []uint64) error
}

type productAttributeRepoImpl struct {
	db *gorm.DB
}

func NewProductAttributeRepo(db *gorm.DB) IProductAttributeRepo {
	return &productAttributeRepoImpl{db: db}
}

func (r *productAttributeRepoImpl) GetProductAttribute(opts ...DBOption) (*model.ProductAttribute, error) {
	var productAttribute model.ProductAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productAttribute)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productAttribute, nil
}

func (r *productAttributeRepoImpl) GetProductAttributes(opts ...DBOption) ([]*model.ProductAttribute, error) {
	var productAttributes []*model.ProductAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productAttributes)
	if result.Error != nil {
		return nil, result.Error
	}

	return productAttributes, nil
}

// GetProductAttributeByUuid 根据UUID查询商品属性（软删除过滤）
func (r *productAttributeRepoImpl) GetProductAttributeByUuid(uuid uint64) (*model.ProductAttribute, error) {
	attr, err := r.GetProductAttribute(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询商品属性失败")
	}
	return attr, nil
}

func (r *productAttributeRepoImpl) GetProductAttributesByUuids(uuids []uint64) ([]*model.ProductAttribute, error) {
	if len(uuids) == 0 {
		return []*model.ProductAttribute{}, nil
	}
	productAttributes, err := r.GetProductAttributes(
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品属性失败")
	}
	return productAttributes, nil
}

// GetProductAttributeWithMultiLanguageName 根据UUID查询商品属性（包含MultiLanguageName）
func (r *productAttributeRepoImpl) GetProductAttributeWithMultiLanguageName(uuid uint64) (*model.ProductAttribute, error) {
	if uuid == 0 {
		return nil, errors.New("商品属性UUID不能为0")
	}
	attr, err := r.GetProductAttribute(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询商品属性失败")
	}
	return attr, nil
}

// GetProductAttributesWithMultiLanguageName 批量根据UUID列表查询商品属性（包含MultiLanguageName）
func (r *productAttributeRepoImpl) GetProductAttributesWithMultiLanguageName(uuids []uint64) ([]*model.ProductAttribute, error) {
	if len(uuids) == 0 {
		return []*model.ProductAttribute{}, nil
	}
	// 过滤掉 0 值
	validUuids := make([]uint64, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			validUuids = append(validUuids, uuid)
		}
	}
	if len(validUuids) == 0 {
		return []*model.ProductAttribute{}, nil
	}
	productAttributes, err := r.GetProductAttributes(
		CommonRepo.WhereInUuids(validUuids),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品属性失败")
	}
	return productAttributes, nil
}

// CreateRecord 创建属性记录
func (r *productAttributeRepoImpl) CreateRecord(attr *model.ProductAttribute) error {
	return r.db.Create(attr).Error
}

// UpdateNameByMultiLanguageNameUuid 根据多语言名称UUID更新name字段
func (r *productAttributeRepoImpl) UpdateNameByMultiLanguageNameUuid(multiLanguageNameUuid uint64, name string) error {
	return r.db.Model(&model.ProductAttribute{}).Where("multi_language_name_uuid = ?", multiLanguageNameUuid).Update("name", name).Error
}

func (r *productAttributeRepoImpl) CreateBatch(attrs []model.ProductAttribute) error {
	if len(attrs) == 0 {
		return nil
	}
	return r.db.Create(&attrs).Error
}

func (r *productAttributeRepoImpl) UpdateProductAttributeData(data map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.ProductAttribute{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(data).Error
}

func (r *productAttributeRepoImpl) HardDeleteByAttributeGroupUuids(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	return r.db.Where("attribute_group_uuid IN (?)", uuids).Delete(&model.ProductAttribute{}).Error
}
