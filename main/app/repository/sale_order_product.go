package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderProductRepo interface {
	CreateSaleOrderProduct(model *model.SaleOrderProduct) (uint64, error)
	CreateSaleOrderProductAndBomAndAttribute(model *model.SaleOrderProduct) (uint64, error)
	UpdateSaleOrderProduct(model *model.SaleOrderProduct) error
	UpdateSaleOrderProductByMap(uuid uint64, vars map[string]any) error
	UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error
	GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error)
	UpdateSaleOrderProductOnly(model *model.SaleOrderProduct) error
}

type saleOrderProductRepo struct {
	db *gorm.DB
}

func NewSaleOrderProductRepo(db *gorm.DB) ISaleOrderProductRepo {
	return &saleOrderProductRepo{db: db}
}

// 创建销售订单商品及BOM、属性
func (r *saleOrderProductRepo) CreateSaleOrderProductAndBomAndAttribute(obj *model.SaleOrderProduct) (uint64, error) {
	db := r.db
	// 创建销售订单商品
	saleOrderProduct := *obj
	saleOrderProduct.SetNil()
	if err := db.Model(&model.SaleOrderProduct{}).Create(&saleOrderProduct).Error; err != nil {
		return 0, err
	}
	// 创建BOM
	for _, bom := range obj.SaleOrderProductBoms {
		bom.SaleOrderProductUuid = obj.Uuid
		if err := db.Create(&bom).Error; err != nil {
			return 0, err
		}
	}
	// 创建属性
	for _, attribute := range obj.SaleOrderProductAttributes {
		attribute.SaleOrderProductUuid = obj.Uuid
		if err := db.Create(&attribute).Error; err != nil {
			return 0, err
		}
	}
	return obj.Uuid, nil
}

func (r *saleOrderProductRepo) CreateSaleOrderProduct(model *model.SaleOrderProduct) (uint64, error) {
	db := r.db
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.Uuid, nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProduct(model *model.SaleOrderProduct) error {
	db := r.db
	if err := db.Model(&model).Updates(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProductByMap(uuid uint64, vars map[string]any) error {
	db := r.db
	if err := db.Model(&model.SaleOrderProduct{}).Where("uuid = ?", uuid).Updates(vars).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProductOnly(obj *model.SaleOrderProduct) error {
	db := r.db
	if err := db.Model(&model.SaleOrderProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error {
	db := r.db
	for _, m := range models {
		if err := db.Model(&m).Updates(m).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *saleOrderProductRepo) GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error) {
	db := r.db
	var model model.SaleOrderProduct
	if err := db.Where("uuid = ?", uuid).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}
