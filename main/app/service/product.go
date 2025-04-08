package service

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp/product_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// IProductSrv 定义产品服务接口
type IProductSrv interface {
	GetProductList(ctx context.Context, req req.ProductListReq) (product_resp.ProductListWithPaginationResp, error) // 获取产品列表
	GetProductCategoryList(dbId uint64) (product_resp.ProductCategoryListResp, error)                               // 获取产品类别列表
}

type productSrv struct {
	dbm       *database.DBManager // 数据库管理器
	localeSrv ILocaleSrv          // 多语言名称服务
}

func NewProductSrv(dbm *database.DBManager, localeSrv ILocaleSrv) IProductSrv {
	return NewProductSrvImpl(dbm, localeSrv)
}

func NewProductSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv) IProductSrv {
	return &productSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
	}
}

// GetProductList 获取产品列表
func (s *productSrv) GetProductList(ctx context.Context, req req.ProductListReq) (product_resp.ProductListWithPaginationResp, error) {
	dbId := ctx.GetDbId()
	// 获取产品列表
	commonRepo := repository.NewCommonRepo()
	sourceMap := map[string]repository.DBOption{
		constant.SourceCashier:   commonRepo.WhereByIsShowCashier(1),
		constant.SourceAssistant: commonRepo.WhereByIsShowAssistant(1),
		constant.SourceTablet:    commonRepo.WhereByIsShowTablet(1),
		constant.SourceKitchen:   commonRepo.WhereByIsShowKitchen(1),
		constant.SourceH5:        commonRepo.WhereByIsShowH5(1),
	}
	productRepo := repository.NewProductRepo(s.dbm.GetDB(dbId))
	var dbOptions []repository.DBOption
	if option, ok := sourceMap[ctx.GetSource()]; ok {
		dbOptions = append(dbOptions, option)
	}
	dbOptions = append(dbOptions, commonRepo.WhereByStatus(1), commonRepo.WhereBySoftDelete(), commonRepo.SortWithSort("ASC"), commonRepo.SortWithID("DESC"))
	products, total, err := productRepo.GetProductListWithPagination(
		req.PageNo,
		req.PageSize,
		dbOptions...,
	)

	// 处理错误
	if err != nil {
		return product_resp.ProductListWithPaginationResp{}, errors.WithMessage(err, "获取产品列表失败")
	}

	// 返回响应对象
	return product_resp.ProductListWithPaginationResp{
		List: s.formatProducts(ctx, products),
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *productSrv) formatProducts(ctx context.Context, products []model.ProductPackage) []product_resp.Product {
	// 转换为响应对象
	list := make([]product_resp.Product, 0, len(products))
	for _, product := range products {
		flavors := make([]product_resp.ProductFlavor, 0, len(product.ProductBoms))                                   // 商品规格
		sauces := make([]product_resp.ProductSauce, 0, len(product.ProductBoms))                                     // 商品小料
		attributeGroups := make([]product_resp.ProductAttributeGroup, 0, len(product.ProductPackageAttributeGroups)) // 商品属性组
		var prices []float64                                                                                         // 保存所有价格，用于计算最低价格

		// 商品规格、加料
		if len(product.ProductBoms) > 0 {
			for _, productBom := range product.ProductBoms {
				if productBom.IsDelete() {
					continue
				}
				if productBom.IsFlavor() {
					flavors = append(flavors, product_resp.ProductFlavor{
						Uuid:       productBom.Uuid,
						LocaleName: s.localeSrv.GetLocaleNames(productBom.ProductFlavor.MultiLanguageName),
						Price:      productBom.Price,
						StockNum:   int(productBom.GetStockNum()),
						Barcode:    productBom.BarcodeValue,
					})
					if len(prices) == 0 {
						prices = append(prices, productBom.Price)
					} else {
						if prices[0] > productBom.Price {
							prices[0] = productBom.Price
						}
					}
				}
				if productBom.IsSauce() {
					sauces = append(sauces, product_resp.ProductSauce{
						Uuid:              productBom.Uuid,
						LocaleName:        productBom.ProductSauce.MultiLanguageName.GetNames(),
						Price:             productBom.Price,
						IsDefaultSelected: productBom.IsDefaultSelect == 1,
						StockNum:          int(productBom.GetStockNum()),
					})
				}
			}
		}

		// 商品属性组
		if len(product.ProductPackageAttributeGroups) > 0 {
			for _, group := range product.ProductPackageAttributeGroups {
				attributeValues := make([]product_resp.ProductAttributeValue, 0, len(group.ProductPackageAttributes)) // 商品属性值
				for _, attribute := range group.ProductPackageAttributes {
					attributeValues = append(attributeValues, product_resp.ProductAttributeValue{
						Uuid:              attribute.Uuid,
						LocaleName:        attribute.Attribute.MultiLanguageName.GetNames(),
						IsDefaultSelected: attribute.IsDefaultSelected == 1,
					})
				}
				attributeGroups = append(attributeGroups, product_resp.ProductAttributeGroup{
					Uuid:       group.ProductAttributeGroupUuid,
					LocaleName: group.ProductAttributeGroup.MultiLanguageName.GetNames(),
					IsMust:     group.IsMust == 1,
					MaxSelect:  group.MaxSelection,
					Attributes: product_resp.ProductAttributeValueList{
						List: attributeValues,
					},
				})
			}
		}

		image := product.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
		// 添加到列表
		minPrice := float64(0)
		if len(prices) > 0 {
			minPrice = slices.Min(prices)
		}
		list = append(list, product_resp.Product{
			Uuid:                product.Uuid,
			Image:               image,
			LocaleName:          product.MultiLanguageName.GetNames(),
			Unit:                product.ProductUnit.MultiLanguageName.GetNames(),
			Price:               minPrice,
			LimitNum:            product.LimitNum,
			CategoryUuid:        product.CategoryUuid,
			FirstCategoryUuid:   product.ProductCategory.GetFirstCategoryUuid(),
			SpecialCategoryUuid: product.SpecialCategoryUuid,
			Flavors: product_resp.ProductFlavorList{
				List: flavors,
			},
			Sauces: product_resp.ProductSauceList{
				List: sauces,
			},
			AttributeGroups: product_resp.ProductAttributeGroupList{
				List: attributeGroups,
			},
			Describe: product.Describe,
		})
	}
	return list
}

// GetProductCategoryList 获取产品类别列表
func (s *productSrv) GetProductCategoryList(dbId uint64) (product_resp.ProductCategoryListResp, error) {
	// 获取产品类别列表
	categories, err := repository.NewProductRepo(s.dbm.GetDB(dbId)).GetProductCategoryList(
		repository.NewCommonRepo().Preload(
			repository.WithPreload{
				Query: "MultiLanguageName",
			},
		),
		repository.NewCommonRepo().WhereBySoftDelete(),
		repository.NewCommonRepo().WhereByStatus(1),
		repository.NewCommonRepo().SortWithIsSpecial("DESC"),
		repository.NewCommonRepo().SortWithSort("ASC"),
		repository.NewCommonRepo().SortWithID("DESC"),
	)
	if err != nil {
		return product_resp.ProductCategoryListResp{}, errors.WithMessage(err, "获取产品类别列表失败")
	}

	// 根据parent_uuid分组转换为响应对象
	list := make([]product_resp.ProductCategory, 0, len(categories))
	for _, category := range categories {
		if category.ParentUuid == 0 {
			children := make([]product_resp.ProductCategory, 0)
			for _, child := range categories {
				if child.ParentUuid == category.Uuid {
					children = append(children, product_resp.ProductCategory{
						Uuid:        child.Uuid,
						LocaleName:  s.localeSrv.GetLocaleNames(child.MultiLanguageName),
						ParentUuid:  child.ParentUuid,
						IsSpecial:   child.IsSpecial == 1,
						CategoryKey: child.CategoryKey,
						Children: product_resp.ProductCategoryListResp{
							List: make([]product_resp.ProductCategory, 0),
						},
					})
				}
			}
			list = append(list, product_resp.ProductCategory{
				Uuid:        category.Uuid,
				LocaleName:  s.localeSrv.GetLocaleNames(category.MultiLanguageName),
				ParentUuid:  category.ParentUuid,
				IsSpecial:   category.IsSpecial == 1,
				CategoryKey: category.CategoryKey,
				Children: product_resp.ProductCategoryListResp{
					List: children,
				},
			})
		}
	}

	// 返回响应对象
	return product_resp.ProductCategoryListResp{
		List: list,
	}, nil
}
