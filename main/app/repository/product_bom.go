package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

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
}

// IProductBomQueryRepo 定义仓库查询接口
type IProductBomQueryRepo interface {
	GetProductBom(opts ...DBOption) (*model.ProductBom, error)
	GetProductBoms(opts ...DBOption) ([]*model.ProductBom, error)
	GetFlavorProductBomByUuid(uuid uint64) (*model.ProductBom, error)
	GetSauceProductBomByUuid(uuid uint64) (*model.ProductBom, error) // 获取小料商品信息
	GetSauceProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error)
	GetFlavorProductBomUuidsByCardUuids(uuids []uint64) ([]uint64, error) // 通过成本卡uuid列表获取规格商品uuid列表
	GetProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error)
	GetProductBomByItemCode(itemCode string) (*model.ProductBom, error)
	GetProductBomByProductBomCardUuid(productBomCardUuid uint64) (*model.ProductBom, error) // 通过成本卡uuid获取商品信息

	WhereProductSauceUuid(uuid uint64) DBOption // 查询条件 商品加料UUID
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

func (r *productBomRepoImpl) GetFlavorProductBomByUuid(uuid uint64) (*model.ProductBom, error) {
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

func (r *productBomRepoImpl) GetSauceProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error) {
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
