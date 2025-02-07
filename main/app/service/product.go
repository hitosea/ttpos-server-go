package service

import (
	"errors"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IProductSrv 定义收银服务接口
type IProductSrv interface {
	GetProductList(dbId uint, req req.ProductListReq) (resp.ProductListWithPaginationResp, error) // 获取收银机点餐页面产品类别列表
}

// 收银服务结构体
type productSrv struct {
	dbm       *database.DBManager // 数据库管理器
	localeSrv ILocaleSrv          // 多语言名称服务
}

// NewProductSrv 创建新的收银产品类别服务
func NewProductSrv(dbm *database.DBManager, localeSrv ILocaleSrv) IProductSrv {
	return NewProductSrvImpl(dbm, localeSrv)
}

// NewProductSrvImpl 创建新的收银服务实现
func NewProductSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv) IProductSrv {
	return &productSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
	}
}

// GetProductList 获取收银机点餐页面产品类别列表
func (s *productSrv) GetProductList(dbId uint, req req.ProductListReq) (resp.ProductListWithPaginationResp, error) {
	// 构建查询条件
	filters := make(map[string]interface{})
	if req.Keyword != "" {
		filters["name"] = req.Keyword
	}
	if req.CategoryID > 0 {
		filters["category_id"] = req.CategoryID
	}

	// 获取总数
	products, total, err := repository.NewProductRepo(s.dbm.GetDB(dbId)).GetProductListWithPagination(
		req.PageNo,
		req.PageSize,
		repository.NewCommonRepo().WhereLikeByName(req.Keyword),
	)
	if err != nil {
		return resp.ProductListWithPaginationResp{}, errors.New("获取产品列表失败")
	}

	// 转换为响应对象
	productList := make([]resp.Product, 0, len(products))
	for _, product := range products {
		productList = append(productList, resp.Product{
			ID:    product.ID,
			Image: product.ImageUrl,
			Name:  s.localeSrv.GetLocaleNames(product.MultiLanguageName),
			Unit:  s.localeSrv.GetLocaleNames(product.ProductUnit.MultiLanguageName),
		})
	}

	// 返回响应对象
	return resp.ProductListWithPaginationResp{
		List: productList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
