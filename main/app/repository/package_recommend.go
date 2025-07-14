package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPackageRecommendRepo 包推荐
type IPackageRecommendRepo interface {
	GetRecommendInfo() (*model.PackageRecommend, error) // 查询推荐信息
}

func NewPackageRecommendRepo(db *gorm.DB) IPackageRecommendRepo {
	return NewPackageRecommendRepoImpl(db)
}

// NewPackageRecommendRepoImpl 创建新的包推荐仓库实现
func NewPackageRecommendRepoImpl(db *gorm.DB) *PackageRecommendRepoImpl {
	return &PackageRecommendRepoImpl{db: db}
}

type PackageRecommendRepoImpl struct {
	db *gorm.DB
}

// GetPackageRecommend 查询推荐
func (r *PackageRecommendRepoImpl) GetPackageRecommend(opts ...DBOption) (*model.PackageRecommend, error) {
	var packageRecommend model.PackageRecommend
	db := r.db.Model(&model.PackageRecommend{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&packageRecommend).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return &packageRecommend, nil
}

// 查询商品当前的商品推荐
func (r *PackageRecommendRepoImpl) GetRecommendInfo() (*model.PackageRecommend, error) {
	packageRecommend, err := r.GetPackageRecommend(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByStatus(constant.PackageRecommendStatusOpen),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取推荐信息失败")
	}
	// 如果没有查到记录，则认为商家未开启推荐
	if packageRecommend == nil {
		return &model.PackageRecommend{Status: constant.PackageRecommendStatusClose}, nil
	}

	return packageRecommend, nil
}
