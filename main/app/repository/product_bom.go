package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductBomRepo interface {
	IProductBomQueryRepo
	CreateProductBom(productBom model.ProductBom) (*model.ProductBom, error)
	UpdateProductBomStockNum(warehouseOutFormItems []*model.WarehouseOutFormItem) error // 更新规格商品或小料的库存数量
	UpdateProductBoms(productBoms []*model.ProductBom) error                            // 更新ProductBom
	CreateProductBoms(productBoms []model.ProductBom) error                             // 创建ProductBom
}

// IProductBomQueryRepo 定义仓库查询接口
type IProductBomQueryRepo interface {
	GetProductBom(opts ...DBOption) (*model.ProductBom, error)
	GetProductBoms(opts ...DBOption) ([]*model.ProductBom, error)
	GetFlavorProductBomByUuid(uuid uint64) (*model.ProductBom, error)
	GetSauceProductBomByUuid(uuid uint64) (*model.ProductBom, error) // 获取小料商品信息
	GetSauceProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error)
	GetProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error)

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
				Query: "FlavorMaterials.Material",
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
				Query: "ProductSauce.SauceMaterials.Material",
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

// UpdateProductBomStockNum 更新规格商品或小料的库存数量
func (r *productBomRepoImpl) UpdateProductBomStockNum(warehouseOutFormItems []*model.WarehouseOutFormItem) error {
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		if warehouseOutFormItem.IsProductBom() {
			if err := r.db.Model(&model.ProductBom{}).Where("uuid = ?", warehouseOutFormItem.ProductBomUuid).Update("stock_num", warehouseOutFormItem.ProductBom.StockNum).Error; err != nil {
				return err
			}
		} else if warehouseOutFormItem.IsMaterial() {
			if err := r.db.Model(&model.Material{}).Where("uuid = ?", warehouseOutFormItem.MaterialUuid).Update("stock_num", warehouseOutFormItem.Material.StockNum).Error; err != nil {
				return err
			}
		}
	}
	return nil
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
