package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

type IProductPackageAttributeRepo interface {
	GetProductPackageAttribute(opts ...DBOption) (*model.ProductPackageAttribute, error)
	GetProductPackageAttributes(opts ...DBOption) ([]*model.ProductPackageAttribute, error)
	GetProductPackageAttributesByUuids(companyUuid uint64, uuids []uint64) ([]*model.ProductPackageAttribute, error)
	CreateProductPackageAttributes(productPackageAttributes []model.ProductPackageAttribute) error
	DeleteProductPackageAttribute(opts ...DBOption) error
	UpdateProductPackageAttribute(data map[string]any, opts ...DBOption) error
	DestroyProductPackageAttribute(opts ...DBOption) error

	GetProductPackageAttributeGroupCount(attributeUuid uint64) ([]model.ProductPackageAttributeGroupCount, error)
}

type productPackageAttributeRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageAttributeRepo(db *gorm.DB) IProductPackageAttributeRepo {
	return &productPackageAttributeRepoImpl{db: db}
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttribute(opts ...DBOption) (*model.ProductPackageAttribute, error) {
	var productPackageAttribute model.ProductPackageAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackageAttribute)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productPackageAttribute, nil
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttributes(opts ...DBOption) ([]*model.ProductPackageAttribute, error) {
	var productPackageAttributes []*model.ProductPackageAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productPackageAttributes)
	if result.Error != nil {
		return nil, result.Error
	}

	return productPackageAttributes, nil
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttributesByUuids(companyUuid uint64, uuids []uint64) ([]*model.ProductPackageAttribute, error) {
	// 检查是否启用对象存储缓存
	var productPackageAttributes []*model.ProductPackageAttribute
	var err error

	if adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 使用对象存储模块缓存查询
		productPackageAttributes, err = r.getProductPackageAttributesWithCache(companyUuid, uuids)
	} else {
		// 直接查询数据库
		productPackageAttributes, err = r.queryProductPackageAttributes(uuids)
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageAttributes, nil
}

// queryProductPackageAttributes 查询商品包属性列表（包含预加载的关联数据）
// 这是一个私有方法，用于统一查询逻辑，避免代码重复
func (r *productPackageAttributeRepoImpl) queryProductPackageAttributes(uuids []uint64) ([]*model.ProductPackageAttribute, error) {
	productPackageAttributes, err := r.GetProductPackageAttributes(
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.Preload(WithPreload{
			Query: "Attribute.MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageAttributes, nil
}

// getProductPackageAttributesWithCache 使用对象存储模块缓存查询商品包属性列表
func (r *productPackageAttributeRepoImpl) getProductPackageAttributesWithCache(companyUuid uint64, uuids []uint64) ([]*model.ProductPackageAttribute, error) {
	if len(uuids) == 0 {
		return []*model.ProductPackageAttribute{}, nil
	}
	if companyUuid == 0 {
		return nil, errors.New("getProductPackageAttributesWithCache companyUuid cannot be 0")
	}

	// 构建批量查询的 keys
	keys := make([]string, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			keys = append(keys, persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductPackageAttribute, uuid))
		}
	}

	if len(keys) == 0 {
		return []*model.ProductPackageAttribute{}, nil
	}

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.ProductPackageAttribute](cache.Global, 5*time.Minute)

	// 使用批量查询缓存
	batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductPackageAttribute, error) {
		// 缓存未命中时，从数据库查询
		attributes, err := r.queryProductPackageAttributes(uuids)
		if err != nil {
			return nil, err
		}
		// 转换为 map[string]*model.ProductPackageAttribute
		result := make(map[string]*model.ProductPackageAttribute)
		for _, attr := range attributes {
			key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductPackageAttribute, attr.Uuid)
			result[key] = attr
		}
		return result, nil
	})

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryProductPackageAttributes(uuids)
	}

	// 将批量查询结果转换为列表，保持原有顺序
	result := make([]*model.ProductPackageAttribute, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductPackageAttribute, uuid)
			if attr, ok := batchResult[key]; ok && attr != nil {
				result = append(result, attr)
			}
		}
	}

	return result, nil
}

func (r *productPackageAttributeRepoImpl) CreateProductPackageAttributes(productPackageAttributes []model.ProductPackageAttribute) error {
	// 如果productPackageAttributes为空，则不创建
	if len(productPackageAttributes) == 0 {
		return nil
	}
	// 清空关联对象
	list := make([]model.ProductPackageAttribute, 0)
	for _, attribute := range productPackageAttributes {
		attribute.SetNil()
		list = append(list, attribute)
	}

	// 创建product_package_attribute表数据
	if err := r.db.Model(&model.ProductPackageAttribute{}).Create(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageAttributeRepoImpl) DeleteProductPackageAttribute(opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageAttribute{}).Updates(map[string]any{
		"delete_time": time.Now().Unix(),
	}).Error
}

func (r *productPackageAttributeRepoImpl) UpdateProductPackageAttribute(data map[string]any, opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageAttribute{}).Updates(data).Error
}

// DestroyProductPackageAttribute 销毁商品包属性
func (r *productPackageAttributeRepoImpl) DestroyProductPackageAttribute(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ProductPackageAttribute{}).Error
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttributeGroupCount(attributeUuid uint64) ([]model.ProductPackageAttributeGroupCount, error) {
	var productPackageAttributeGroupCountList []model.ProductPackageAttributeGroupCount
	err := r.db.Model(&model.ProductPackageAttribute{}).Select("product_package_attribute_group_uuid, count(1) as related_attribute_uuid_count").
		Scopes(NotDeleted).Where("attribute_uuid = ?", attributeUuid).
		Group("product_package_attribute_group_uuid").Scan(&productPackageAttributeGroupCountList).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageAttributeGroupCountList, nil
}
