package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderProductRepo interface {
	ISaleOrderProductQueryRepo
	CreateSaleOrderProduct(model *model.SaleOrderProduct) (uint64, error)
	CreateSaleOrderProductAndBomAndAttribute(model model.SaleOrderProduct) (uint64, error)
	UpdateSaleOrderProduct(model *model.SaleOrderProduct) error
	UpdateSaleOrderProductByMap(uuid uint64, vars map[string]any) error
	UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error
	UpdateSaleOrderProductRecord(model model.SaleOrderProduct) error
	UpdateOrCreateSaleOrderProductRecord(obj model.SaleOrderProduct) error
	CreateSaleOrderProductReasons(saleOrderUuid uint64, saleOrderProductUuid uint64, source string, returnFoodReasons [][2]uint64) error
	DeleteSaleOrderProductReasons(saleOrderUuid uint64, saleOrderProductUuid uint64, source string) error
	DeleteSaleOrderProductList(models []*model.SaleOrderProduct) error // 批量删除销售订单商品。delete_time赋值为当前时间
	DeleteSaleOrderProductBySaleBillUuid(saleBillUuid uint64) error    // 根据销售账单uuid删除销售订单商品。delete_time赋值为当前时间
	Update(data map[string]interface{}, opts ...DBOption) error        // 更新订单商品
}

// ISaleOrderProductQueryRepo 销售订单商品查询

type ISaleOrderProductQueryRepo interface {
	GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error)
	GetProductPackageDetail(saleBillUuid uint64, saleOrderUuid uint64, productPackageUuid uint64) ([]*model.SaleOrderProduct, error) // 获取商品选购详情
	GetSaleOrderProducts(opts ...DBOption) ([]*model.SaleOrderProduct, error)                                                        // 根据销售订单uuid获取销售订单商品

	GetSaleOrderProductsByPackageUuid(packageUuid uint64) ([]*model.SaleOrderProduct, error) // 根据套餐uuid获取套餐下所有子商品
}

type saleOrderProductRepo struct {
	db *gorm.DB
}

func NewSaleOrderProductRepo(db *gorm.DB) ISaleOrderProductRepo {
	return &saleOrderProductRepo{db: db}
}

func (r *saleOrderProductRepo) GetSaleOrderProducts(opts ...DBOption) ([]*model.SaleOrderProduct, error) {
	var saleOrderProducts []*model.SaleOrderProduct
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&saleOrderProducts)
	if result.Error != nil {
		return saleOrderProducts, result.Error
	}

	return saleOrderProducts, nil
}

// CreateSaleOrderProductAndBomAndAttribute 创建销售订单商品及BOM、属性
func (r *saleOrderProductRepo) CreateSaleOrderProductAndBomAndAttribute(obj model.SaleOrderProduct) (uint64, error) {
	db := r.db
	// 创建销售订单商品
	saleOrderProduct := obj
	saleOrderProduct.SetNil()
	if err := db.Model(&model.SaleOrderProduct{}).Create(&saleOrderProduct).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	// 创建BOM
	for _, bom := range obj.SaleOrderProductBoms {
		bom.SaleOrderProductUuid = obj.Uuid

		//  如果是新建，这两个字段要清空
		bom.ID = 0
		bom.Uuid = 0

		if err := db.Create(&bom).Error; err != nil {
			return 0, errors.WithMessage(err)
		}
	}
	// 创建属性
	for _, attribute := range obj.SaleOrderProductAttributes {
		attribute.SaleOrderProductUuid = obj.Uuid
		if err := db.Create(&attribute).Error; err != nil {
			return 0, errors.WithMessage(err)
		}
	}
	return obj.Uuid, nil
}

// CreateSaleOrderProduct 创建销售订单商品
func (r *saleOrderProductRepo) CreateSaleOrderProduct(model *model.SaleOrderProduct) (uint64, error) {
	db := r.db
	if err := db.Create(&model).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return model.Uuid, nil
}

// UpdateSaleOrderProduct 更新销售订单商品
func (r *saleOrderProductRepo) UpdateSaleOrderProduct(model *model.SaleOrderProduct) error {
	db := r.db
	if err := db.Model(&model).Updates(model).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateSaleOrderProductByMap 更新销售订单商品
func (r *saleOrderProductRepo) UpdateSaleOrderProductByMap(uuid uint64, vars map[string]any) error {
	db := r.db
	if err := db.Model(&model.SaleOrderProduct{}).Where("uuid = ?", uuid).Updates(vars).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateSaleOrderProductRecord(obj model.SaleOrderProduct) error {
	obj.SetNil()
	if err := r.db.Model(&model.SaleOrderProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *saleOrderProductRepo) UpdateOrCreateSaleOrderProductRecord(obj model.SaleOrderProduct) error {
	if obj.ID == 0 {
		_, err := r.CreateSaleOrderProductAndBomAndAttribute(obj)
		return errors.WithMessage(err)
	}
	// 如果标记商品需要更新才更新该商品
	if obj.GetUpdate() {
		obj.SetNil()
		return r.db.Model(&model.SaleOrderProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
	}
	return nil
}

// UpdateSaleOrderProductList 批量更新销售订单商品
func (r *saleOrderProductRepo) UpdateSaleOrderProductList(models []*model.SaleOrderProduct) error {
	db := r.db

	// 构建新的对象。 为了防止在加购并送厨的场景下，加购的商品不会被更新
	list := make([]model.SaleOrderProduct, 0)
	for _, m := range models {
		// 如果对象没有主键，则跳过
		if m.NoPrimaryKey() {
			continue
		}
		list = append(list, *m)
	}

	for _, model := range list {
		model.SetNil()
		// 如果对象没有主键，则跳过
		if model.NoPrimaryKey() {
			continue
		}
		if err := db.Model(&model).Select("*").Updates(&model).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// GetSaleOrderProductByUuid 根据uuid获取销售订单商品
func (r *saleOrderProductRepo) GetSaleOrderProductByUuid(uuid uint64) (*model.SaleOrderProduct, error) {
	db := r.db
	var model model.SaleOrderProduct
	if err := db.Where("uuid = ?", uuid).First(&model).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return &model, nil
}

// CreateSaleOrderProductReasons 批量创建销售订单商品原因
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

// DeleteSaleOrderProductReasons 批量删除销售订单商品原因
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

// DeleteSaleOrderProductList 批量删除销售订单商品。delete_time赋值为当前时间
func (r *saleOrderProductRepo) DeleteSaleOrderProductList(models []*model.SaleOrderProduct) error {
	uuids := make([]uint64, 0)
	for _, model := range models {
		uuids = append(uuids, model.Uuid)
	}
	now := time.Now().Unix()
	return r.db.Model(&model.SaleOrderProduct{}).Where("uuid in (?)", uuids).Update("delete_time", now).Error
}

// DeleteSaleOrderProductBySaleBillUuid 根据销售账单uuid删除销售订单商品。delete_time赋值为当前时间
func (r *saleOrderProductRepo) DeleteSaleOrderProductBySaleBillUuid(saleBillUuid uint64) error {
	now := time.Now().Unix()
	return r.db.Model(&model.SaleOrderProduct{}).Where("sale_bill_uuid = ?", saleBillUuid).Update("delete_time", now).Error
}

// Update 更新
func (r *saleOrderProductRepo) Update(data map[string]interface{}, opts ...DBOption) error {
	db := r.db.Model(&model.SaleOrderProduct{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Updates(data).Error

	return errors.WithMessage(err)
}

func (r *saleOrderProductRepo) GetProductPackageDetail(saleBillUuid uint64, saleOrderUuid uint64, productPackageUuid uint64) ([]*model.SaleOrderProduct, error) {
	models, err := r.GetSaleOrderProducts(
		CommonRepo.WhereBySaleBillUuid(saleBillUuid),
		CommonRepo.WhereBySaleOrderUuid(saleOrderUuid),
		CommonRepo.WhereByProductPackageUuid(productPackageUuid),
		CommonRepo.WhereByH5OrderUuid(0), // 未下单的h5商品
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrderProductAttributes",
			},
			WithPreload{
				Query: "SaleOrderProductBoms",
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (r *saleOrderProductRepo) GetSaleOrderProductsByPackageUuid(packageUuid uint64) ([]*model.SaleOrderProduct, error) {
	db := r.db
	var models []*model.SaleOrderProduct
	if err := db.Where("package_uuid = ?", packageUuid).Scopes(NotDeleted).Find(&models).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return models, nil
}
