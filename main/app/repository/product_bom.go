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

type IProductBomRepo interface {
	IProductBomQueryRepo
	CreateProductBom(productBom model.ProductBom) (*model.ProductBom, error)
	UpdateProductBom(data map[string]any, opts ...DBOption) error
	UpdateProductBoms(productBoms []*model.ProductBom) error                                       // 更新ProductBom
	CreateProductBoms(productBoms []model.ProductBom) error                                        // 创建ProductBom
	UpdateProductBomCard(productBomUuid uint64, productBomCardUuid uint64, stockNum float64) error // 更新规格商品的成本卡
	GetProductBomCardByUuid(productBomUuid uint64) (*model.ProductBomCard, error)                  // 获取成本卡
	AddActualSaleNum(productBomUuid uint64, saleNum float64) error                                 // 增加实际销量
	SubActualSaleNum(productBomUuid uint64, saleNum float64) error                                 // 减少实际销量
	DestroyProductBom(opts ...DBOption) error
}

// IProductBomQueryRepo 定义仓库查询接口
type IProductBomQueryRepo interface {
	GetProductBom(opts ...DBOption) (*model.ProductBom, error)
	GetProductBoms(opts ...DBOption) ([]*model.ProductBom, error)
	GetFlavorProductBomByUuid(companyUuid uint64, uuid uint64) (*model.ProductBom, error)
	GetSauceProductBomByUuid(uuid uint64) (*model.ProductBom, error) // 获取小料商品信息
	GetSauceProductBomsByUuids(companyUuid uint64, uuids []uint64) ([]*model.ProductBom, error)
	GetFlavorProductBomUuidsByCardUuids(uuids []uint64) ([]uint64, error) // 通过成本卡uuid列表获取规格商品uuid列表
	GetProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error)
	GetProductBomByItemCode(itemCode string) (*model.ProductBom, error)
	GetProductBomByProductBomCardUuid(productBomCardUuid uint64) (*model.ProductBom, error) // 通过成本卡uuid获取商品信息
	GetProductBomsByHasCard() ([]*model.ProductBom, error)                                  // 获取有成本卡的product_bom
	GetProductPackageUuidByBomUuid(productBomUuid uint64) (uint64, error)                   // 通过商品bom uuid获取套餐product_package_uuid
	GetProductBomUuidByProductPackageUuid(productPackageUuid uint64) (uint64, error)        // 通过套餐product_package_uuid获取商品bom uuid
	WhereProductSauceUuid(uuid uint64) DBOption                                             // 查询条件 商品加料UUID
}

type productBomRepoImpl struct {
	db *gorm.DB
}

func NewProductBomRepo(db *gorm.DB) IProductBomRepo {
	return &productBomRepoImpl{db: db}
}

func (r *productBomRepoImpl) CreateProductBom(productBom model.ProductBom) (*model.ProductBom, error) {
	productBom.SetNil()
	if err := r.db.Create(&productBom).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productBom, nil
}

func (r *productBomRepoImpl) UpdateProductBom(data map[string]any, opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.ProductBom{}).Updates(data).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productBomRepoImpl) GetProductBom(opts ...DBOption) (*model.ProductBom, error) {
	var productBom model.ProductBom
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productBom)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productBom, nil
}

func (r *productBomRepoImpl) GetProductBoms(opts ...DBOption) ([]*model.ProductBom, error) {
	productBoms := make([]*model.ProductBom, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productBoms)
	if result.Error != nil {
		return nil, result.Error
	}

	return productBoms, nil
}

func (r *productBomRepoImpl) GetFlavorProductBomByUuid(companyUuid uint64, uuid uint64) (*model.ProductBom, error) {
	// 检查是否启用对象存储缓存
	var productBom *model.ProductBom
	var err error

	if adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 使用对象存储模块缓存查询
		productBom, err = r.getFlavorProductBomWithCache(companyUuid, uuid)
	} else {
		// 直接查询数据库
		productBom, err = r.queryFlavorProductBom(uuid)
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

// queryFlavorProductBom 查询规格商品 ProductBom（包含预加载的关联数据）
// 这是一个私有方法，用于统一查询逻辑，避免代码重复
func (r *productBomRepoImpl) queryFlavorProductBom(uuid uint64) (*model.ProductBom, error) {
	productBom, err := r.GetProductBom(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "FlavorMaterials",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "FlavorMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "ProductPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductPackage.ProductUnit",
			},
			WithPreload{
				Query: "ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "ProductSauce.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

// getFlavorProductBomWithCache 使用对象存储模块缓存查询规格商品 ProductBom（包含预加载的关联数据）
func (r *productBomRepoImpl) getFlavorProductBomWithCache(companyUuid uint64, uuid uint64) (*model.ProductBom, error) {
	if uuid == 0 {
		return nil, errors.New("getFlavorProductBomWithCache uuid cannot be 0")
	}
	if companyUuid == 0 {
		return nil, errors.New("getFlavorProductBomWithCache companyUuid cannot be 0")
	}

	// 构建缓存 key（使用 companyUuid 而不是 context）
	key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductBomFlavor, uuid)

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.ProductBom](cache.Global, 5*time.Minute)

	// 使用缓存查询
	result, err := cacheLayer.GET(key, func() (*model.ProductBom, error) {
		// 缓存未命中时，从数据库查询（包含所有预加载）
		return r.queryFlavorProductBom(uuid)
	})

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryFlavorProductBom(uuid)
	}

	return result, nil
}

// GetSauceProductBomByUuid 获取小料商品信息
func (r *productBomRepoImpl) GetSauceProductBomByUuid(uuid uint64) (*model.ProductBom, error) {
	productBom, err := r.GetProductBom(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(WithPreload{
			Query: "ProductSauce.MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

func (r *productBomRepoImpl) GetSauceProductBomsByUuids(companyUuid uint64, uuids []uint64) ([]*model.ProductBom, error) {
	// 检查是否启用对象存储缓存
	var productBoms []*model.ProductBom
	var err error

	if adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 使用对象存储模块缓存查询
		productBoms, err = r.getSauceProductBomsWithCache(companyUuid, uuids)
	} else {
		// 直接查询数据库
		productBoms, err = r.querySauceProductBoms(uuids)
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBoms, nil
}

// querySauceProductBoms 查询小料商品 ProductBom 列表（包含预加载的关联数据）
// 这是一个私有方法，用于统一查询逻辑，避免代码重复
func (r *productBomRepoImpl) querySauceProductBoms(uuids []uint64) ([]*model.ProductBom, error) {
	productBoms, err := r.GetProductBoms(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductSauce.SauceMaterials.Material.WarehouseItems",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBoms, nil
}

// getSauceProductBomsWithCache 使用对象存储模块缓存查询小料商品 ProductBom 列表
func (r *productBomRepoImpl) getSauceProductBomsWithCache(companyUuid uint64, uuids []uint64) ([]*model.ProductBom, error) {
	if len(uuids) == 0 {
		return []*model.ProductBom{}, nil
	}
	if companyUuid == 0 {
		return nil, errors.New("getSauceProductBomsWithCache companyUuid cannot be 0")
	}

	// 构建批量查询的 keys
	keys := make([]string, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			keys = append(keys, persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductBomSauce, uuid))
		}
	}

	if len(keys) == 0 {
		return []*model.ProductBom{}, nil
	}

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.ProductBom](cache.Global, 5*time.Minute)

	// 使用批量查询缓存
	batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductBom, error) {
		// 缓存未命中时，从数据库查询
		boms, err := r.querySauceProductBoms(uuids)
		if err != nil {
			return nil, err
		}
		// 转换为 map[string]*model.ProductBom
		result := make(map[string]*model.ProductBom)
		for _, bom := range boms {
			key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductBomSauce, bom.Uuid)
			result[key] = bom
		}
		return result, nil
	})

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.querySauceProductBoms(uuids)
	}

	// 将批量查询结果转换为列表，保持原有顺序
	result := make([]*model.ProductBom, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductBomSauce, uuid)
			if bom, ok := batchResult[key]; ok && bom != nil {
				result = append(result, bom)
			}
		}
	}

	return result, nil
}

func (r *productBomRepoImpl) GetProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error) {
	productBoms, err := r.GetProductBoms(
		CommonRepo.WhereInUuids(uuids),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBoms, nil
}

func (r *productBomRepoImpl) UpdateProductBoms(productBoms []*model.ProductBom) error {
	if len(productBoms) == 0 {
		return nil
	}
	list := make([]model.ProductBom, 0)
	for _, productBom := range productBoms {
		bom := *productBom
		bom.SetNil()
		list = append(list, bom)
	}
	if err := r.db.Model(&model.ProductBom{}).Save(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productBomRepoImpl) CreateProductBoms(productBoms []model.ProductBom) error {
	// 如果productBoms为空，则不创建
	if len(productBoms) == 0 {
		return nil
	}
	// 清空关联对象
	list := make([]model.ProductBom, 0)
	for _, productBom := range productBoms {
		productBom.SetNil()
		list = append(list, productBom)
	}

	// 创建product_bom表数据
	if err := r.db.Model(&model.ProductBom{}).Create(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productBomRepoImpl) WhereProductSauceUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_sauce_uuid = ?", uuid)
	}
}

// 更新规格商品的成本卡，修改成本卡uuid和库存数量
func (r *productBomRepoImpl) UpdateProductBomCard(productBomUuid uint64, productBomCardUuid uint64, stockNum float64) error {
	if err := r.db.Model(&model.ProductBom{}).Where("uuid = ?", productBomUuid).Updates(map[string]interface{}{
		"product_bom_card_uuid": productBomCardUuid,
		"stock_num":             stockNum,
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productBomRepoImpl) GetProductBomCardByUuid(productBomUuid uint64) (*model.ProductBomCard, error) {
	productBom, err := r.GetProductBom(
		CommonRepo.WhereByUuid(productBomUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductBomCard.RelatedMaterials.Material.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductBomCard.RelatedMaterials.Material.NotBaseUnitList.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductBomCard.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom.ProductBomCard, nil
}

func (r *productBomRepoImpl) AddActualSaleNum(productBomUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.ProductBom{}).Where("uuid = ?", productBomUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num + ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productBomRepoImpl) SubActualSaleNum(productBomUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.ProductBom{}).Where("uuid = ?", productBomUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num - ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// DestroyProductBom 销毁商品bom
func (r *productBomRepoImpl) DestroyProductBom(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ProductBom{}).Error
}

// 通过成本卡uuid列表获取规格商品列表
func (r *productBomRepoImpl) GetFlavorProductBomUuidsByCardUuids(uuids []uint64) ([]uint64, error) {
	var productBomUuids []uint64
	err := r.db.Model(&model.ProductBom{}).Where("product_bom_card_uuid IN ?", uuids).Where("delete_time = 0").Pluck("uuid", &productBomUuids).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBomUuids, nil
}

// 根据商品编码获取商品信息
func (r *productBomRepoImpl) GetProductBomByItemCode(itemCode string) (*model.ProductBom, error) {
	var productBom model.ProductBom
	err := r.db.Model(&model.ProductBom{}).Where("erp_code = ?", itemCode).Where("delete_time = 0").
		Preload("ProductPackage.MultiLanguageName").
		First(&productBom).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productBom, nil
}

func (r *productBomRepoImpl) GetProductBomByProductBomCardUuid(productBomCardUuid uint64) (*model.ProductBom, error) {
	var productBom model.ProductBom
	err := r.db.Model(&model.ProductBom{}).Where("product_bom_card_uuid = ?", productBomCardUuid).Where("delete_time = 0").
		First(&productBom).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productBom, nil
}

func (r *productBomRepoImpl) GetProductBomsByHasCard() ([]*model.ProductBom, error) {
	var productBoms []*model.ProductBom
	err := r.db.Model(&model.ProductBom{}).Where("product_bom_card_uuid <> 0").Where("delete_time = 0").
		Preload("ProductBomCard.RelatedMaterials.Material.WarehouseItems").
		Find(&productBoms).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBoms, nil
}

func (r *productBomRepoImpl) GetProductPackageUuidByBomUuid(productBomUuid uint64) (uint64, error) {
	var productPackageUuid uint64
	err := r.db.Model(&model.ProductBom{}).Where("uuid = ?", productBomUuid).
		Pluck("product_package_uuid", &productPackageUuid).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return productPackageUuid, nil
}

func (r *productBomRepoImpl) GetProductBomUuidByProductPackageUuid(productPackageUuid uint64) (uint64, error) {
	var productBomUuid uint64
	err := r.db.Model(&model.ProductBom{}).Where("product_package_uuid = ?", productPackageUuid).
		Pluck("uuid", &productBomUuid).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return productBomUuid, nil
}
