package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderProductRepo interface {
	CreateSaleOrderProduct(model *model.SaleOrderProduct) (uint64, error)
	CreateSaleOrderProductAndBomAndAttribute(model model.SaleOrderProduct) (uint64, error)
	UpdateSaleOrderProduct(model *model.SaleOrderProduct) error
	UpdateSaleOrderProductByMap(uuid uint64, vars map[string]any) error
	UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error
	GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error)
	UpdateSaleOrderProductRecord(model model.SaleOrderProduct) error
	UpdateOrCreateSaleOrderProductRecord(obj model.SaleOrderProduct) error
	CreateSaleOrderProductReasons(saleOrderUuid uint64, saleOrderProductUuid uint64, source string, returnFoodReasons [][2]uint64) error
	DeleteSaleOrderProductReasons(saleOrderUuid uint64, saleOrderProductUuid uint64, source string) error
}

type saleOrderProductRepo struct {
	db *gorm.DB
}

func NewSaleOrderProductRepo(db *gorm.DB) ISaleOrderProductRepo {
	return &saleOrderProductRepo{db: db}
}

// 创建销售订单商品及BOM、属性
func (r *saleOrderProductRepo) CreateSaleOrderProductAndBomAndAttribute(obj model.SaleOrderProduct) (uint64, error) {
	db := r.db
	// 创建销售订单商品
	saleOrderProduct := obj
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

// 创建销售订单商品
func (r *saleOrderProductRepo) CreateSaleOrderProduct(model *model.SaleOrderProduct) (uint64, error) {
	db := r.db
	if err := db.Create(&model).Error; err != nil {
		return 0, err
	}
	return model.Uuid, nil
}

// 更新销售订单商品
func (r *saleOrderProductRepo) UpdateSaleOrderProduct(model *model.SaleOrderProduct) error {
	db := r.db
	if err := db.Model(&model).Updates(model).Error; err != nil {
		return err
	}
	return nil
}

// 更新销售订单商品
func (r *saleOrderProductRepo) UpdateSaleOrderProductByMap(uuid uint64, vars map[string]any) error {
	db := r.db
	if err := db.Model(&model.SaleOrderProduct{}).Where("uuid = ?", uuid).Updates(vars).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProductRecord(obj model.SaleOrderProduct) error {
	obj.SetNil()
	if err := r.db.Model(&model.SaleOrderProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateOrCreateSaleOrderProductRecord(obj model.SaleOrderProduct) error {
	if obj.ID == 0 {
		_, err := r.CreateSaleOrderProductAndBomAndAttribute(obj)
		return err
	}
	// 如果标记商品需要更新才更新该商品
	if obj.GetUpdate() {
		obj.SetNil()
		return r.db.Model(&model.SaleOrderProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
	}
	return nil
}

// 批量更新销售订单商品
func (r *saleOrderProductRepo) UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error {
	db := r.db
	for _, m := range models {
		if err := db.Model(&m).Updates(m).Error; err != nil {
			return err
		}
	}
	return nil
}

// 根据uuid获取销售订单商品
func (r *saleOrderProductRepo) GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error) {
	db := r.db
	var model model.SaleOrderProduct
	if err := db.Where("uuid = ?", uuid).First(&model).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return &model, nil
}

// 批量创建销售订单商品原因
func (r *saleOrderProductRepo) CreateSaleOrderProductReasons(
	saleOrderUuid uint64,
	saleOrderProductUuid uint64,
	source string,
	returnFoodReasons [][2]uint64,
) error {
	if len(returnFoodReasons) == 0 {
		return nil
	}
	db := r.db
	// 构建批量插入数据
	reasons := make([]*model.SaleOrderProductReason, len(returnFoodReasons))
	for i, reason := range returnFoodReasons {
		reasons[i] = &model.SaleOrderProductReason{
			SaleOrderUuid:         saleOrderUuid,
			SaleOrderProductUuid:  saleOrderProductUuid,
			MultiLanguageNameUuid: reason[1],
		}
		if source == constant.ProductReasonTypeReturnFood {
			reasons[i].ReturnFoodReasonUuid = reason[0]
		}
		if source == constant.ProductReasonTypeGift {
			reasons[i].GiftReasonUuid = reason[0]
		}
		if source == constant.ProductReasonTypeFree {
			reasons[i].FreeReasonUuid = reason[0]
		}
	}
	// 批量创建
	return db.Create(&reasons).Error
}

// 批量删除销售订单商品原因
func (r *saleOrderProductRepo) DeleteSaleOrderProductReasons(saleOrderUuid uint64, saleOrderProductUuid uint64, source string) error {
	switch source {
	case constant.ProductReasonTypeReturnFood:
		return r.db.Model(&model.SaleOrderProductReason{}).
			Where("sale_order_uuid = ? and sale_order_product_uuid = ? and return_food_reason_uuid > 0", saleOrderUuid, saleOrderProductUuid).
			Delete(&model.SaleOrderProductReason{}).Error
	case constant.ProductReasonTypeGift:
		return r.db.Model(&model.SaleOrderProductReason{}).
			Where("sale_order_uuid = ? and sale_order_product_uuid = ? and gift_reason_uuid > 0", saleOrderUuid, saleOrderProductUuid).
			Delete(&model.SaleOrderProductReason{}).Error
	case constant.ProductReasonTypeFree:
		return r.db.Model(&model.SaleOrderProductReason{}).
			Where("sale_order_uuid = ? and sale_order_product_uuid = ? and free_reason_uuid > 0", saleOrderUuid, saleOrderProductUuid).
			Delete(&model.SaleOrderProductReason{}).Error
	default:
		return nil
	}
}
