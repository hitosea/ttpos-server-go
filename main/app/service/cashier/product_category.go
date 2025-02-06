// 包含收银相关的服务
package cashier

import (
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// 定义收银服务接口
type CashierServiceInterface interface {
	// 获取收银机点餐页面产品类别列表
	GetProductCategory(dbId uint, language string) (resp.ProductCategory, error)
}

// 创建新的收银产品类别服务
func NewCashierProductCategoryService(dbm *database.DBManager) CashierServiceInterface {
	return NewCashierServiceImpl(dbm)
}

// 创建新的收银服务实现
func NewCashierServiceImpl(dbm *database.DBManager) *CashierService {
	return &CashierService{
		dbm: dbm,
	}
}

// 收银服务结构体
type CashierService struct {
	dbm *database.DBManager
}

// 获取产品类别的实现
func (s *CashierService) GetProductCategory(dbId uint, language string) (resp.ProductCategory, error) {
	db := s.dbm.GetDB(dbId)
	// 查询产品类别表
	productCategoryList, err := repository.NewProductCategoryRepository(db).GetProductCategoryListWithMultiLanguageName()
	if err != nil {
		return resp.ProductCategory{}, err
	}
	// 查询特殊分类表
	productSpecialCategoryList, err := repository.NewProductSpecialCategoryRepository(db).GetProductSpecialCategoryListWithMultiLanguageName()
	if err != nil {
		return resp.ProductCategory{}, err
	}

	specialCategoryList := []resp.SpecialCategory{}
	for _, productSpecialCategory := range productSpecialCategoryList {
		specialCategoryList = append(specialCategoryList, resp.SpecialCategory{
			Name: productSpecialCategory.Name,
			Id:   productSpecialCategory.Id,
		})
	}

	categoryList := []resp.Category{}
	categoryMap := make(map[uint]int)
	index := 0
	for _, productCategory := range productCategoryList {
		if productCategory.ParentUuid != 0 {
			category := resp.Category{
				Name:     productCategory.Name,
				Key:      productCategory.CategoryKey,
				Children: nil, // todo
			}
			categoryList = append(categoryList, category)
			categoryMap[productCategory.Id] = index
			index++
		} else {
			categoryList[categoryMap[productCategory.ParentUuid]].Children = append(categoryList[categoryMap[productCategory.ParentUuid]].Children, resp.ChildCategory{
				Name: productCategory.Name,
				Key:  productCategory.CategoryKey,
			})
		}
	}
	return resp.ProductCategory{
		SpecialCategoryList: specialCategoryList,
		CategoryList:        categoryList,
	}, nil
}
