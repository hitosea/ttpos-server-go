package repository

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductCategoryRepo 商品类别
type IProductCategoryRepo interface {
	CreateProductCategory(productCategory model.ProductCategory) (uint64, error)
}

func NewProductCategoryRepo(db *gorm.DB) IProductCategoryRepo {
	return NewProductCategoryRepoImpl(db)
}

// ICategoryRepositorySrv 分类仓库接口
type ICategoryRepositorySrv interface {
	CreateCategory(params req.CreateCategoryRequest) (uint64, error)
}

func NewCategoryRepositoryService(db *gorm.DB) ICategoryRepositorySrv {
	return NewCategoryRepositorySrvImpl(db)
}

// CategoryRepositoryService 分类仓库服务实现
type CategoryRepositoryService struct {
	db *gorm.DB
}

func NewCategoryRepositorySrvImpl(db *gorm.DB) *CategoryRepositoryService {
	return &CategoryRepositoryService{db: db}
}

// CreateCategory 创建分类
func (s *CategoryRepositoryService) CreateCategory(params req.CreateCategoryRequest) (uint64, error) {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 创建多语言名称表
	multiLanguageName := model.MultiLanguageName{
		ThName:   params.Name.TH,
		ZhTwName: params.Name.ZHTW,
		ZhName:   params.Name.ZH,
		EnName:   params.Name.EN,
	}
	nameId, err := NewMultiLanguageNameRepoImpl(tx).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
		tx.Rollback() // 发生错误，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 创建分类
	category := model.ProductCategory{
		ParentUuid:            params.ParentID,
		MultiLanguageNameUuid: nameId,
		Name:                  params.Name.ZH,
		Sort:                  uint(params.Sort),
	}
	id, err := NewProductCategoryRepo(tx).CreateProductCategory(category)
	if err != nil {
		tx.Rollback() // 发生错误，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	return id, nil
}
