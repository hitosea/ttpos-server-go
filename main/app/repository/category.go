package repository

import (
	"strings"

	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductCategoryRepo 商品类别
type IProductCategoryRepo interface {
	CreateProductCategory(productCategory model.ProductCategory) (uint64, error)
	UpdateProductCategory(id uint, productCategory model.ProductCategory) error
	GetProductCategory(opts ...DBOption) (*model.ProductCategory, error)
}

func NewProductCategoryRepo(db *gorm.DB) IProductCategoryRepo {
	return NewProductCategoryRepoImpl(db)
}

// ICategoryRepositorySrv 分类仓库接口
type ICategoryRepositorySrv interface {
	CreateCategory(params req.CreateCategoryRequest) (uint64, error)
	GetCategoryUuidByNameOptimized(name string) (uint64, error) // 根据分类名称获取分类UUID（支持多语言）
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
	nameId, err := NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(multiLanguageName)
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

func (s *CategoryRepositoryService) UpdateProductCategory(uuid uint64, productCategory model.ProductCategory) error {
	return s.db.Model(&model.ProductCategory{}).Where("uuid = ?", uuid).Updates(productCategory).Error
}

// GetCategoryUuidByNameOptimized 支持分级分类路径查询 (例如: "主分类/子分类")
func (s *CategoryRepositoryService) GetCategoryUuidByNameOptimized(name string) (uint64, error) {
	var categories []model.ProductCategory

	// 预加载多语言名称，然后在内存中过滤
	err := s.db.Model(&model.ProductCategory{}).
		Preload("MultiLanguageName").
		Where("delete_time = ?", 0).
		Order("status desc").
		Find(&categories).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}

	// 使用 "/" 分割分类路径，类似 PHP 的 explode('/', $value)
	categoryPath := strings.Split(name, "/")
	categoryId, err := s.findCategoryByName(categories, categoryPath[0], 0)
	if err != nil {
		return 0, err
	}
	// 如果只有一级分类
	if len(categoryPath) == 1 {
		return categoryId, nil
	}

	return s.findCategoryByName(categories, categoryPath[1], categoryId)
}

// findCategoryByName 在指定父分类下查找分类名称（支持多语言）
func (s *CategoryRepositoryService) findCategoryByName(categories []model.ProductCategory, name string, parentUuid uint64) (uint64, error) {
	// 在内存中查找匹配的分类
	for _, category := range categories {
		// 检查父分类是否匹配
		if category.ParentUuid != parentUuid {
			continue
		}
		// 然后检查多语言字段
		if category.MultiLanguageName.Uuid != 0 {
			names := category.MultiLanguageName.GetNames()
			if names.ZH == name || names.ZHTW == name || names.EN == name ||
				names.TH == name || names.MY == name || names.JA == name ||
				names.KO == name || names.TR == name || names.SV == name {
				return category.Uuid, nil
			}
		}
	}

	return 0, nil
}

func (r *ProductCategoryRepoImpl) GetProductCategory(opts ...DBOption) (*model.ProductCategory, error) {
	var productCategory model.ProductCategory
	db := r.db.Model(&model.ProductCategory{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&productCategory).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &productCategory, nil
}
