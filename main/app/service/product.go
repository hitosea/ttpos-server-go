package service

import (
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/database"
)

// 定义收银服务接口
type ProductServiceInterface interface {
	// 获取收银机点餐页面产品类别列表
	GetProduct(dbId uint, language string) (resp.Product, error)
}

// 创建新的收银产品类别服务
func NewProductService(dbm *database.DBManager) ProductServiceInterface {
	return NewProductServiceImpl(dbm)
}

// 创建新的收银服务实现
func NewProductServiceImpl(dbm *database.DBManager) *ProductService {
	return &ProductService{
		dbm: dbm,
	}
}

// 收银服务结构体
type ProductService struct {
	dbm *database.DBManager
}

// 获取收银机点餐页面产品类别列表
func (s *ProductService) GetProduct(dbId uint, language string) (resp.Product, error) {
	return resp.Product{}, nil
}
