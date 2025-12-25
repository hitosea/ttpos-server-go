package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ProductPackageGroupItemTakeoutRepo struct {
	db *gorm.DB
}

func NewProductPackageGroupItemTakeoutRepo(db *gorm.DB) *ProductPackageGroupItemTakeoutRepo {
	return &ProductPackageGroupItemTakeoutRepo{db: db}
}

// CreateProductPackageGroupItemTakeout 创建外卖套餐子商品价格
func (r *ProductPackageGroupItemTakeoutRepo) CreateProductPackageGroupItemTakeout(item *model.ProductPackageGroupItemTakeout) error {
	return r.db.Create(item).Error
}

// BatchCreateProductPackageGroupItemTakeout 批量创建外卖套餐子商品价格
func (r *ProductPackageGroupItemTakeoutRepo) BatchCreateProductPackageGroupItemTakeout(items []*model.ProductPackageGroupItemTakeout) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// GetProductPackageGroupItemTakeoutList 获取外卖套餐子商品价格列表
func (r *ProductPackageGroupItemTakeoutRepo) GetProductPackageGroupItemTakeoutList(productPackageTakeoutUuid uint64) ([]*model.ProductPackageGroupItemTakeout, error) {
	var items []*model.ProductPackageGroupItemTakeout
	err := r.db.Where("product_package_takeout_uuid = ?", productPackageTakeoutUuid).
		Where("delete_time = 0").
		Find(&items).Error
	return items, err
}

// DeleteByProductPackageTakeoutUuid 根据外卖商品UUID删除所有套餐子商品价格
func (r *ProductPackageGroupItemTakeoutRepo) DeleteByProductPackageTakeoutUuid(productPackageTakeoutUuid uint64) error {
	return r.db.Where("product_package_takeout_uuid = ?", productPackageTakeoutUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// GetByGroupItemUuid 根据套餐子商品UUID获取外卖价格
func (r *ProductPackageGroupItemTakeoutRepo) GetByGroupItemUuid(productPackageTakeoutUuid uint64, groupItemUuid uint64) (*model.ProductPackageGroupItemTakeout, error) {
	var item model.ProductPackageGroupItemTakeout
	err := r.db.Where("product_package_takeout_uuid = ?", productPackageTakeoutUuid).
		Where("product_package_group_item_uuid = ?", groupItemUuid).
		Where("delete_time = 0").
		First(&item).Error
	return &item, err
}

// UpdateAddPrice 更新套餐子商品加价
func (r *ProductPackageGroupItemTakeoutRepo) UpdateAddPrice(uuid uint64, addPrice float64) error {
	return r.db.Model(&model.ProductPackageGroupItemTakeout{}).
		Where("uuid = ?", uuid).
		Update("add_price", addPrice).Error
}

// SoftDelete 软删除套餐子商品价格
func (r *ProductPackageGroupItemTakeoutRepo) SoftDelete(uuid uint64) error {
	return r.db.Model(&model.ProductPackageGroupItemTakeout{}).
		Where("uuid = ?", uuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// DestroyProductPackageGroupItemTakeout 物理删除套餐子商品价格
func (r *ProductPackageGroupItemTakeoutRepo) DestroyProductPackageGroupItemTakeout(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	return r.db.Where("uuid IN ?", uuids).Delete(&model.ProductPackageGroupItemTakeout{}).Error
}
