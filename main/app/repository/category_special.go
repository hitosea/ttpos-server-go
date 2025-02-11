package repository

import (
	"gorm.io/gorm"
)

// ProductSpecialCategoryRepoImpl 商品特殊类别
type ProductSpecialCategoryRepoImpl struct {
	db *gorm.DB
}
