// Package cashier 包含收银相关的服务
package cashier

import (
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// ICashierSrv 定义收银服务接口
type ICashierSrv interface {
	// GetProductCategory 获取收银机点餐页面产品类别列表
	GetProductCategory(dbId uint64, language string) (resp.ProductCategory, error)
}

// NewCashierProductCategorySrv 创建新的收银产品类别服务
func NewCashierProductCategorySrv(dbm *database.DBManager) ICashierSrv {
	return NewCashierSrvImpl(dbm)
}

// NewCashierSrvImpl 创建新的收银服务实现
func NewCashierSrvImpl(dbm *database.DBManager) *CashierSrv {
	return &CashierSrv{
		dbm: dbm,
	}
}

// CashierSrv 收银服务结构体
type CashierSrv struct {
	dbm *database.DBManager
}

// GetProductCategory 获取产品类别的实现
func (s *CashierSrv) GetProductCategory(dbId uint64, language string) (resp.ProductCategory, error) {
	db := s.dbm.GetDB(dbId)
	// 查询产品类别表
	productCategoryList, err := repository.NewProductCategoryRepo(db).GetProductCategoryListWithMultiLanguageName()
	if err != nil {
		return resp.ProductCategory{}, err
	}
	// 查询特殊分类表
	productSpecialCategoryList, err := repository.NewProductSpecialCategoryRepo(db).GetProductSpecialCategoryListWithMultiLanguageName()
	if err != nil {
		return resp.ProductCategory{}, err
	}

	var specialCategoryList []resp.SpecialCategory
	for _, productSpecialCategory := range productSpecialCategoryList {
		specialCategoryList = append(specialCategoryList, resp.SpecialCategory{
			Name: productSpecialCategory.Name,
			Id:   productSpecialCategory.Id,
		})
	}

	var categoryList []resp.Category
	categoryMap := make(map[uint]int)
	index := 0
	for _, productCategory := range productCategoryList {
		if productCategory.ParentUuid != 0 {
			category := resp.Category{
				Name:     productCategory.Name,
				Children: nil, // todo
			}
			categoryList = append(categoryList, category)
			categoryMap[productCategory.Id] = index
			index++
		} else {
			categoryList[categoryMap[productCategory.ParentUuid]].Children = append(categoryList[categoryMap[productCategory.ParentUuid]].Children, resp.ChildCategory{
				Name: productCategory.Name,
			})
		}
	}
	return resp.ProductCategory{
		SpecialCategoryList: specialCategoryList,
		CategoryList:        categoryList,
	}, nil
}
