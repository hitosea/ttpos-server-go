package service

import (
	"encoding/json"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp/product_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IProductSrv 定义产品服务接口
type IProductSrv interface {
	GetProductList(ctx context.Context, req req.ProductListReq) (product_resp.ProductListWithPaginationResp, error)               // 获取产品列表
	GetProductCategoryList(dbId uint64) (product_resp.ProductCategoryListResp, error)                                             // 获取产品类别列表（销售端）
	GetProductUnitList(ctx context.Context, req req.ProductUnitListReq) (product_resp.ProductUnitListResp, error)                 // 获取产品单位列表
	GetProductUnit(ctx context.Context, req req.ProductUnitReq) (product_resp.ProductUnitDetail, error)                           // 获取产品单位详情
	GetProductRecommendList(ctx context.Context, req req.ProductRecommendListReq) (*product_resp.ProductRecommendListResp, error) // 获取产品推荐列表
	SearchProducts(ctx context.Context, req req.ProductSearchReq) (*product_resp.ProductSearchResp, error)                        // 搜索商品

	GetProductShopCategoryList(ctx context.Context, req req.ProductShopCategoryListReq) (product_resp.ProductShopCategoryListResp, error) // 获取产品类别列表（商家端）
	SortProductShopCategory(ctx context.Context, req req.ProductShopCategorySortReq) error                                                // 排序产品分类
	AddProductShopCategory(ctx context.Context, req req.ProductShopCategoryAddReq) error                                                  // 添加产品分类
	EditProductShopCategory(ctx context.Context, req req.ProductShopCategoryEditReq) error                                                // 编辑产品分类
	DeleteProductShopCategory(ctx context.Context, req req.ProductShopCategoryReq) error                                                  // 删除产品分类

	AddProductUnit(ctx context.Context, req req.ProductUnitAddReq) error   // 添加产品单位
	EditProductUnit(ctx context.Context, req req.ProductUnitEditReq) error // 编辑产品单位
	DeleteProductUnit(ctx context.Context, req req.ProductUnitReq) error   // 删除产品单位
	SortProductUnit(ctx context.Context, req req.ProductUnitSortReq) error // 排序产品单位

	GetProductSauceList(ctx context.Context, req req.ProductSauceListReq) (product_resp.ProductSauceListResp, error) // 获取商品加料列表
	GetProductSauce(ctx context.Context, req req.ProductSauceReq) (product_resp.ProductSauceDetail, error)           // 获取商品加料详情
	AddProductSauce(ctx context.Context, req req.ProductSauceAddReq) error                                           // 添加商品加料
	EditProductSauce(ctx context.Context, req req.ProductSauceEditReq) error                                         // 编辑商品加料
	DeleteProductSauce(ctx context.Context, req req.ProductSauceReq) error                                           // 删除商品加料
	SortProductSauce(ctx context.Context, req req.ProductSauceSortReq) error                                         // 排序商品加料

	GetProductAttributeGroupList(ctx context.Context, req req.ProductAttributeGroupListReq) (product_resp.ProductAttributeGroupListResp, error) // 获取商品属性分组列表
	GetProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupReq) (product_resp.ProductAttributeGroupDetail, error)           // 获取商品属性分组详情
	AddProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupAddReq) error                                                    // 添加商品属性分组，商品属性一起添加
	EditProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupEditReq) error                                                  // 编辑商品属性分组，商品属性一起编辑
	DeleteProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupReq) error                                                    // 删除商品属性分组
	SortProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupSortReq) error                                                  // 排序商品属性分组
	SortProductAttribute(ctx context.Context, req req.ProductAttributeSortReq) error                                                            // 排序商品属性

	GetProductFlavorList(ctx context.Context, req req.ProductFlavorListReq) (product_resp.ProductFlavorListResp, error) // 获取商品规格列表
	GetProductFlavor(ctx context.Context, req req.ProductFlavorReq) (product_resp.ProductFlavorDetailResp, error)       // 获取商品规格详情
	AddProductFlavor(ctx context.Context, req req.ProductFlavorAddReq) error                                            // 添加商品规格
	EditProductFlavor(ctx context.Context, req req.ProdudctFlavorEditReq) error                                         // 编辑商品规格
	DeleteProductFlavor(ctx context.Context, req req.ProductFlavorDeleteReq) error                                      // 删除商品规格
	SortProductFlavor(ctx context.Context, req req.ProductFlavorSortReq) error                                          // 排序商品规格
}

type productSrv struct {
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv          // 多语言名称服务
	settingSrv setting.ISrv
}

func NewProductSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IProductSrv {
	return NewProductSrvImpl(dbm, localeSrv, settingSrv)
}

func NewProductSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IProductSrv {
	return &productSrv{
		dbm:        dbm,
		localeSrv:  localeSrv,
		settingSrv: settingSrv,
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
		constant.SourceMember:    commonRepo.WhereByIsShowMember(1),
	}
	productRepo := repository.NewProductRepo(s.dbm.GetDB(dbId))
	var dbOptions []repository.DBOption
	if option, ok := sourceMap[ctx.GetSource()]; ok {
		dbOptions = append(dbOptions, option)
	}

	// 如果查询推荐商品
	if len(req.RecommendProductPackageUuids) > 0 {
		dbOptions = append(dbOptions, commonRepo.WhereInUuids(req.RecommendProductPackageUuids))
	}

	dbOptions = append(dbOptions, commonRepo.WhereByStatus(1), commonRepo.WhereBySoftDelete(), commonRepo.SortWithSort("ASC"), commonRepo.SortWithID("DESC"))
	if req.IsMember {
		// 会员端查询商品列表，预加载外送税
		dbOptions = append(dbOptions, commonRepo.Preload(
			repository.WithPreload{
				Query: "TakeoutTax",
			},
		))
	}

	products, total, err := productRepo.GetProductListWithPagination(
		req.PageNo,
		req.PageSize,
		dbOptions...,
	)

	// 处理错误
	if err != nil {
		return product_resp.ProductListWithPaginationResp{}, errors.WithMessage(err, "获取产品列表失败")
	}

	// 如果是会员端查询商品列表
	if req.IsMember {
		// 获取外送折扣率
		// 获取门店业务设置
		businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
		if err != nil {
			return product_resp.ProductListWithPaginationResp{}, errors.WithMessage(err, "获取门店业务设置失败")
		}
		// 获取外送折扣率
		deliveryPriceRatio := businessSetting.GetDeliveryPriceRatio()

		taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
		if err != nil {
			return product_resp.ProductListWithPaginationResp{}, errors.WithMessage(err, "获取门店设置失败")
		}
		taxFeeType := taxRateSetting.GetTaxFeeType()

		// 返回响应对象
		return product_resp.ProductListWithPaginationResp{
			List: FormatProducts(ctx, products, WithTakeoutDiscountRate(deliveryPriceRatio, taxFeeType)),
			Meta: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
		}, nil
	}

	// 返回响应对象
	return product_resp.ProductListWithPaginationResp{
		List: FormatProducts(ctx, products),
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

type FormatProductsFn func(opts *FormatProductsOption)

type FormatProductsOption struct {
	IsMember            bool    // 是否是会员端查询商品列表
	TakeoutDiscountRate float64 // 外送端折扣率. 取值范围1%-300% 即 0.01-3
	TaxFeeType          uint8   // 税费类型
}

// 设置参数，会员端查询商品列表，外送端折扣率
func WithTakeoutDiscountRate(rate float64, taxFeeType uint8) FormatProductsFn {
	return func(opts *FormatProductsOption) {
		opts.TakeoutDiscountRate = rate
		opts.TaxFeeType = taxFeeType // constant.TaxFeeTypeNoTax
		opts.IsMember = true
	}
}

// FormatProducts 格式化产品列表
func FormatProducts(ctx context.Context, products []model.ProductPackage, options ...FormatProductsFn) []product_resp.Product {
	var option FormatProductsOption
	for _, fn := range options {
		fn(&option)
	}

	// 转换为响应对象
	list := make([]product_resp.Product, 0, len(products))
	for _, product := range products {
		image := product.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
		unit := product.ProductUnit.MultiLanguageName.GetNames()

		if product.ProductType == constant.ProductTypePackage {
			packageGroupList := make([]product_resp.ProductPackageGroup, 0)
			for _, group := range product.ProductPackageGroups {
				productList := make([]product_resp.PackageProductDetail, 0)
				for _, item := range group.ProductPackageGroupItems {
					flavor := getFlavor(item.ProductBom)                       // 商品规格
					attributeGroups := getAttributeGroups(item.ProductPackage) // 商品属性组
					productDetail := product_resp.PackageProductDetail{
						Detail: product_resp.Product{
							Uuid:       item.ProductBom.Uuid,
							LocaleName: item.ProductPackage.MultiLanguageName.GetNames(),
							Image:      image,
							Unit:       item.ProductPackage.ProductUnit.MultiLanguageName.GetNames(),
							Price:      0, // 商品价格，套餐内目前是0元
							Flavors: product_resp.ProductFlavorList{
								List: []product_resp.ProductFlavor{flavor},
							},
							AttributeGroups: product_resp.ProductAttributeGroupList{
								List: attributeGroups,
							},
							Sauces: product_resp.ProductSauceList{
								List: make([]product_resp.ProductSauce, 0),
							},
							Describe: product.Describe,
						},
						Num: item.Num,
					}
					productDetail.CanEdit = productDetail.GetCanEdit() // 是否可以编辑

					productList = append(productList, productDetail)
				}

				packageGroup := product_resp.ProductPackageGroup{
					Uuid:       group.Uuid,
					LocaleName: group.MultiLanguageName.GetNames(),
					Num:        len(productList),
					Products: product_resp.ProductList{
						List: productList,
					},
				}
				packageGroup.IsFull = packageGroup.GetIsFull() // 是否选满
				packageGroupList = append(packageGroupList, packageGroup)
			}

			// 套餐的默认规格
			flavors := make([]product_resp.ProductFlavor, 0)
			for _, productBom := range product.ProductBoms {
				if productBom.IsDelete() {
					continue
				}
				flavors = append(flavors, product_resp.ProductFlavor{
					Uuid:       productBom.Uuid,
					LocaleName: product.MultiLanguageName.GetNames(),
					Price:      productBom.Price,
					StockNum:   int(productBom.GetStockNum()),
					Barcode:    productBom.BarcodeValue,
				})
			}

			list = append(list, product_resp.Product{
				Uuid:              product.Uuid,
				Image:             image,
				LocaleName:        product.MultiLanguageName.GetNames(),
				Unit:              unit,
				Price:             flavors[0].Price,
				LimitNum:          product.LimitNum,
				CategoryUuid:      product.CategoryUuid,
				FirstCategoryUuid: product.ProductCategory.GetFirstCategoryUuid(),
				Describe:          product.Describe,
				IsShowKitchen:     product.IsShowKitchen,
				ProductType:       product.ProductType,
				Flavors: product_resp.ProductFlavorList{
					List: flavors,
				},
				Sauces: product_resp.ProductSauceList{
					List: make([]product_resp.ProductSauce, 0),
				},
				AttributeGroups: product_resp.ProductAttributeGroupList{
					List: make([]product_resp.ProductAttributeGroup, 0),
				},
				PackageGroupList: &product_resp.ProductPackageGroupList{
					List: packageGroupList,
				},
			})
		} else {
			flavors := make([]product_resp.ProductFlavor, 0, len(product.ProductBoms))                                   // 商品规格
			sauces := make([]product_resp.ProductSauce, 0, len(product.ProductBoms))                                     // 商品小料
			attributeGroups := make([]product_resp.ProductAttributeGroup, 0, len(product.ProductPackageAttributeGroups)) // 商品属性组
			var prices []float64                                                                                         // 保存所有价格，用于计算最低价格

			// 商品规格、加料
			if len(product.ProductBoms) > 0 {
				taxRate := product.TakeoutTax.TaxRate
				takeoutDiscountRate := option.TakeoutDiscountRate // 外送端折扣率

				for _, productBom := range product.ProductBoms {
					if productBom.IsDelete() {
						continue
					}
					if productBom.IsFlavor() {
						flavor := product_resp.ProductFlavor{
							Uuid:       productBom.Uuid,
							LocaleName: productBom.ProductFlavor.MultiLanguageName.GetNames(),
							Price:      productBom.Price,
							StockNum:   int(productBom.GetStockNum()),
							Barcode:    productBom.BarcodeValue,
						}
						// 如果是会员端查询商品列表，需要获取在会员端的该商品规格价格
						// 会员端商品规格价格=原商品规格价*外送商品折扣率 + 税费。 税费=原商品规格价*外送商品折扣率*外送的税率
						// 故，会员端商品规格价格=原商品规格价*外送商品折扣率 * (1 + 外送的税率)
						if option.IsMember {
							flavor.Price = calculateTakeoutProductPrice(productBom.Price, takeoutDiscountRate, taxRate, option.TaxFeeType)
						}
						flavors = append(flavors, flavor)
						if len(prices) == 0 {
							prices = append(prices, productBom.Price)
						} else {
							if prices[0] > productBom.Price {
								prices[0] = productBom.Price
							}
						}
					}
					if productBom.IsSauce() {
						sauce := product_resp.ProductSauce{
							Uuid:              productBom.Uuid,
							LocaleName:        productBom.ProductSauce.MultiLanguageName.GetNames(),
							Price:             productBom.Price,
							IsDefaultSelected: productBom.IsDefaultSelect == 1,
							StockNum:          int(productBom.GetStockNum()),
						}
						// 如果是会员端查询商品列表，需要获取在会员端的该商品小料价格
						// 会员端商品小料价格=原商品小料价*外送商品折扣率 + 税费。 税费=原商品小料价*外送商品折扣率*外送的税率
						// 故，会员端商品小料价格=原商品小料价*外送商品折扣率 * (1 + 外送的税率)
						if option.IsMember {
							sauce.Price = calculateTakeoutProductPrice(productBom.Price, takeoutDiscountRate, taxRate, option.TaxFeeType)
						}
						sauces = append(sauces, sauce)
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

			// 添加到列表
			minPrice := float64(0)
			if len(prices) > 0 {
				minPrice = slices.Min(prices)
			}
			if option.IsMember {
				minPrice = calculateTakeoutProductPrice(minPrice, option.TakeoutDiscountRate, product.TakeoutTax.TaxRate, option.TaxFeeType)
			}
			list = append(list, product_resp.Product{
				Uuid:                product.Uuid,
				Image:               image,
				LocaleName:          product.MultiLanguageName.GetNames(),
				Unit:                unit,
				Price:               minPrice,
				NumType:             product.NumType,
				LimitNum:            product.LimitNum,
				CategoryUuid:        product.CategoryUuid,
				FirstCategoryUuid:   product.ProductCategory.GetFirstCategoryUuid(),
				SpecialCategoryUuid: product.SpecialCategoryUuid,
				Flavors: product_resp.ProductFlavorList{
					List: flavors,
				},
				Sauces: product_resp.ProductSauceList{
					List:      sauces,
					IsMust:    product.SauceRequired == 1,
					MaxSelect: int(product.SauceMaxSelection),
				},
				AttributeGroups: product_resp.ProductAttributeGroupList{
					List: attributeGroups,
				},
				Describe:      product.Describe,
				IsShowKitchen: product.IsShowKitchen,
			})
		}
	}
	return list
}

func getFlavor(productBom *model.ProductBom) product_resp.ProductFlavor {
	if productBom.IsDelete() {
		return product_resp.ProductFlavor{}
	}
	if productBom.IsFlavor() {
		flavor := product_resp.ProductFlavor{
			Uuid:       productBom.Uuid,
			LocaleName: productBom.ProductFlavor.MultiLanguageName.GetNames(),
			Price:      productBom.Price,
			StockNum:   int(productBom.GetStockNum()),
			Barcode:    productBom.BarcodeValue,
		}
		return flavor
	}
	return product_resp.ProductFlavor{}
}

func getAttributeGroups(product *model.ProductPackage) []product_resp.ProductAttributeGroup {
	attributeGroups := make([]product_resp.ProductAttributeGroup, 0)
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
	return attributeGroups
}

// 外送端商品价格计算
// originPrice 原商品小料价、原商品规格价
// takeoutDiscountRate 外送商品折扣率 取值范围是1%-300%，即0.01-3
// takeoutTaxRate 外送的税率 取值范围是0%-100%，即0-1
func calculateTakeoutProductPrice(originPrice float64, takeoutDiscountRate float64, takeoutTaxRate float64, TaxFeeType uint8) float64 {
	// 商品未含税时
	if TaxFeeType == constant.TaxFeeTypeNoTax {
		// 未含税商品价格
		unTaxPrice := originPrice
		// 涨价后的未含税商品金额
		unTaxPriceAfterDiscount := decimal.NewFromFloat(unTaxPrice).Mul(decimal.NewFromFloat(takeoutDiscountRate)).Round(2).InexactFloat64()
		// 涨价后的税费
		taxFee := decimal.NewFromFloat(unTaxPriceAfterDiscount).Mul(decimal.NewFromFloat(takeoutTaxRate)).Round(2).InexactFloat64()
		// 涨价后会员端显示的价格
		price := unTaxPriceAfterDiscount + taxFee
		return decimal.NewFromFloat(price).Round(2).InexactFloat64()
	}

	// 商品已含税时
	if TaxFeeType == constant.TaxFeeTypeTax {
		// 当商品已含税时，显示售价=商品规格价*外送商品折扣率
		return decimal.NewFromFloat(originPrice).Mul(decimal.NewFromFloat(takeoutDiscountRate)).Round(2).InexactFloat64()
	}

	// 不收取税费时
	if TaxFeeType == constant.TaxFeeTypeNone {
		// 不收取税费时，显示售价=商品规格价*外送商品折扣率
		return decimal.NewFromFloat(originPrice).Mul(decimal.NewFromFloat(takeoutDiscountRate)).Round(2).InexactFloat64()
	}

	// 默认按不收取税费处理
	return decimal.NewFromFloat(originPrice).Mul(decimal.NewFromFloat(takeoutDiscountRate)).Round(2).InexactFloat64()
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
		repository.NewCommonRepo().SortWithCreateTime("DESC"),
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
				if child.ParentUuid != 0 && child.ParentUuid == category.Uuid {
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

// GetProductShopCategoryList 获取产品类别列表（商家端）
func (s *productSrv) GetProductShopCategoryList(ctx context.Context, req req.ProductShopCategoryListReq) (product_resp.ProductShopCategoryListResp, error) {
	dbId := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(dbId)
	// 查询/关联
	opts := []repository.DBOption{
		commonRepo.Preload(
			repository.WithPreload{
				Query: "MultiLanguageName",
			},
		),
		productRepo.WhereCategoryKey(""),
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithIsSpecial("DESC"),
		commonRepo.SortWithSort("ASC"),
		commonRepo.SortWithID("DESC"),
	}
	// 搜索关键词
	if req.Keyword != nil && *req.Keyword != "" {
		opts = append(opts, commonRepo.WhereLikeByName(*req.Keyword))
	}
	// 父级分类
	if req.ParentUuid != nil {
		opts = append(opts, productRepo.WhereParentUuid(*req.ParentUuid))
	}
	// 是否特色分类
	if req.IsSpecial != nil {
		if *req.IsSpecial {
			opts = append(opts, productRepo.WhereByIsSpecial(1))
		} else {
			opts = append(opts, productRepo.WhereByIsSpecial(0))
		}
	}
	// 获取产品类别列表
	categories, err := repository.NewProductRepo(dbId).GetProductCategoryList(opts...)
	if err != nil {
		return product_resp.ProductShopCategoryListResp{}, errors.WithMessage(err, "获取分类列表失败")
	}

	// 根据parent_uuid分组转换为响应对象
	list := make([]product_resp.ProductShopCategory, 0, len(categories))
	for _, category := range categories {
		if category.ParentUuid == 0 {
			children := make([]product_resp.ProductShopCategory, 0)
			for _, child := range categories {
				if child.ParentUuid != 0 && child.ParentUuid == category.Uuid {
					children = append(children, product_resp.ProductShopCategory{
						Uuid:       child.Uuid,
						LocaleName: s.localeSrv.GetLocaleNames(child.MultiLanguageName),
						ParentUuid: child.ParentUuid,
						IsSpecial:  child.IsSpecial == 1,
						Sort:       child.Sort,
						Status:     child.Status,
						Children: product_resp.ProductShopCategoryListResp{
							List: make([]product_resp.ProductShopCategory, 0),
						},
					})
				}
			}
			list = append(list, product_resp.ProductShopCategory{
				Uuid:       category.Uuid,
				LocaleName: s.localeSrv.GetLocaleNames(category.MultiLanguageName),
				ParentUuid: category.ParentUuid,
				IsSpecial:  category.IsSpecial == 1,
				Sort:       category.Sort,
				Status:     category.Status,
				Children: product_resp.ProductShopCategoryListResp{
					List: children,
				},
			})
		}
	}

	// 返回响应对象
	return product_resp.ProductShopCategoryListResp{
		List: list,
	}, nil
}

// SortProductShopCategory 排序产品分类
func (s *productSrv) SortProductShopCategory(ctx context.Context, req req.ProductShopCategorySortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	productCategoryUuids := make([]uint64, 0, len(req.List))
	for _, item := range req.List {
		productCategoryUuids = append(productCategoryUuids, item.Uuid)
	}
	productCategories, _ := productRepo.GetProductCategoryCount(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuidIn(productCategoryUuids),
	)
	if productCategories != int64(len(productCategoryUuids)) {
		return errors.New("分类不存在")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.List {
			tx.Model(&model.ProductCategory{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			})
		}
		return nil
	})
	return err
}

// AddProductShopCategory 添加产品分类
func (s *productSrv) AddProductShopCategory(ctx context.Context, addReq req.ProductShopCategoryAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: "category",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}
	if addReq.Status != 0 && addReq.Status != 1 {
		return errors.New("分类状态不正确")
	}
	if addReq.IsSpecial && addReq.ParentUuid != 0 {
		return errors.New("特色分类不能有一级分类")
	}
	if addReq.ParentUuid != 0 {
		count, err := productRepo.GetProductCategoryCount(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereUuid(addReq.ParentUuid),
			productRepo.WhereParentUuid(0),
		)
		if err != nil {
			return errors.WithMessage(err, "获取一级分类失败")
		}
		if count == 0 {
			return errors.New("一级分类不存在")
		}

	}
	maxSort, err := productRepo.GetProductCategoryMaxSort(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereParentUuid(addReq.ParentUuid),
		productRepo.WhereByIsSpecial(utils.IfUint(addReq.IsSpecial, 1, 0)),
	)
	if err != nil {
		return errors.WithMessage(err, "获取一级分类最大排序失败")
	}
	sort := uint(maxSort + 1)

	err = db.Transaction(func(tx *gorm.DB) error {
		// 保存多语言名称
		multiLanguageName := model.MultiLanguageName{
			ZhName:   addReq.LocaleName.ZH,
			ThName:   addReq.LocaleName.TH,
			EnName:   addReq.LocaleName.EN,
			ZhTwName: addReq.LocaleName.ZHTW,
			JaName:   addReq.LocaleName.JA,
			KoName:   addReq.LocaleName.KO,
			MyName:   addReq.LocaleName.MY,
			TrName:   addReq.LocaleName.TR,
			SvName:   addReq.LocaleName.SV,
		}
		err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return err
		}
		// 保存产品分类
		productCategory := model.ProductCategory{
			ParentUuid:            addReq.ParentUuid,
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Name:                  addReq.LocaleName.ToJson(),
			Sort:                  sort,
			IsSpecial:             utils.IfInt(addReq.IsSpecial, 1, 0),
			Status:                addReq.Status,
		}
		err = tx.Model(&model.ProductCategory{}).Create(&productCategory).Error
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Logger.Error("添加产品分类失败", zap.Any("func", "AddProductShopCategory"), zap.Any("params", addReq), zap.Error(err))
		return errors.WithMessage(err, "添加产品分类失败")
	}

	return nil
}

// EditProductShopCategory 编辑产品分类
func (s *productSrv) EditProductShopCategory(ctx context.Context, editReq req.ProductShopCategoryEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	productCategory, err := productRepo.GetProductCategory(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(editReq.Uuid),
	)
	if err != nil {
		return errors.WithMessage(err, "获取分类详情失败")
	}
	if productCategory.Uuid == 0 {
		return errors.New("分类不存在")
	}
	if productCategory.IsSpecial == 1 && editReq.ParentUuid != 0 {
		return errors.New("特色分类不能修改为二级分类")
	}
	if productCategory.ParentUuid == 0 && editReq.ParentUuid != 0 {
		return errors.New("一级分类不能修改为二级分类")
	}
	if productCategory.ParentUuid != 0 && editReq.ParentUuid == 0 {
		return errors.New("二级分类不能修改为一级分类")
	}
	if editReq.Status != 0 && editReq.Status != 1 {
		return errors.New("分类状态不正确")
	}
	if editReq.ParentUuid != 0 {
		count, err := productRepo.GetProductCategoryCount(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereUuid(editReq.ParentUuid),
			productRepo.WhereParentUuid(0),
		)
		if err != nil {
			return errors.WithMessage(err, "获取一级分类失败")
		}
		if count == 0 {
			return errors.New("一级分类不存在")
		}
	}
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: "category",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productCategory.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    editReq.LocaleName.ZH,
			"th_name":    editReq.LocaleName.TH,
			"en_name":    editReq.LocaleName.EN,
			"zh_tw_name": editReq.LocaleName.ZHTW,
			"ja_name":    editReq.LocaleName.JA,
			"ko_name":    editReq.LocaleName.KO,
			"my_name":    editReq.LocaleName.MY,
			"tr_name":    editReq.LocaleName.TR,
			"sv_name":    editReq.LocaleName.SV,
		}).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.ProductCategory{}).Where("uuid = ?", editReq.Uuid).Updates(map[string]any{
			"parent_uuid": editReq.ParentUuid,
			"name":        editReq.LocaleName.ToJson(),
			"status":      editReq.Status,
		}).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("编辑产品分类失败", zap.Any("func", "EditProductShopCategory"), zap.Any("params", editReq), zap.Error(err))
		return errors.WithMessage(err, "编辑分类失败")
	}

	return nil
}

// DeleteProductShopCategory 删除产品分类
func (s *productSrv) DeleteProductShopCategory(ctx context.Context, deleteReq req.ProductShopCategoryReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	productCategory, err := productRepo.GetProductCategory(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(deleteReq.Uuid),
	)
	if productCategory.Uuid == 0 || err != nil {
		return errors.New("分类不存在")
	}
	opts := []repository.DBOption{
		commonRepo.WhereBySoftDelete(),
	}
	if productCategory.IsSpecial == 1 {
		opts = append(opts, productRepo.WhereSpecialCategoryUuid(deleteReq.Uuid))
	} else {
		opts = append(opts, productRepo.WhereCategoryUuid(deleteReq.Uuid))
	}
	productCount, err := productRepo.GetProductCount(opts...)
	if err != nil {
		return errors.WithMessage(err, "获取商品数量失败")
	}
	if productCount > 0 {
		return errors.New("分类下有商品，不能删除")
	}
	if productCategory.ParentUuid == 0 && productCategory.IsSpecial == 0 {
		categoryCount, err := productRepo.GetProductCategoryCount(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereParentUuid(deleteReq.Uuid),
		)
		if err != nil {
			return errors.WithMessage(err, "获取子分类数量失败")
		}
		if categoryCount > 0 {
			return errors.New("分类下有子分类，不能删除")
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.ProductCategory{}).Where("uuid = ?", deleteReq.Uuid).Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Logger.Error("删除分类失败", zap.Any("func", "DeleteProductShopCategory"), zap.Any("params", deleteReq), zap.Error(err))
		return errors.WithMessage(err, "删除分类失败")
	}

	return nil
}

// GetProductRecommendList 获取产品推荐列表
func (s *productSrv) GetProductRecommendList(ctx context.Context, request req.ProductRecommendListReq) (*product_resp.ProductRecommendListResp, error) {
	// 获取商机推荐信息
	packageRecommend, err := repository.NewPackageRecommendRepo(s.dbm.GetDB(ctx.GetDbId())).GetRecommendInfo()
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return &product_resp.ProductRecommendListResp{
				IsOpen: false,
			}, nil
		}
		return nil, errors.WithMessage(err, "获取商机推荐信息失败")
	}
	if !packageRecommend.IsOpen() {
		return &product_resp.ProductRecommendListResp{
			IsOpen: false,
		}, nil
	}

	recommendProducts, err := s.ParseRecommendInfo(packageRecommend)
	if err != nil {
		return nil, errors.WithMessage(err, "解析推荐商品失败")
	}

	var recommendProductUuids []uint64
	productsSortMap := make(map[uint64]int) // 商品排序map, key是商品uuid, value是排序值
	for _, recommendProduct := range recommendProducts {
		// string转换为uint64
		uuid, err := utils.StringToUint64(recommendProduct.Uuid)
		if err != nil {
			ctx.Log().Error("字符串转换为uint64失败", zap.String("uuid", recommendProduct.Uuid), zap.Error(err))
			continue // 跳过错误,不显示该错误的推荐商品
		}
		recommendProductUuids = append(recommendProductUuids, uuid)
		sort, err := strconv.Atoi(recommendProduct.Sort)
		if err != nil {
			sort = 0 // 默认排序为0
		}
		productsSortMap[uuid] = sort
	}
	// 获取产品推荐列表
	products, err := s.GetProductList(ctx, req.ProductListReq{
		PageReq: dto.PageReq{
			PageNo:   1,
			PageSize: 100,
		},
		RecommendProductPackageUuids: recommendProductUuids,
		IsMember:                     true,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "获取产品推荐列表失败")
	}

	// 将商品排序map转换为商品列表
	for index, product := range products.List {
		products.List[index].Sort = productsSortMap[product.Uuid]
	}

	// 按照sort排序, sort是字符串类型, 小的在前
	sort.Slice(products.List, func(i, j int) bool {
		return products.List[i].Sort < products.List[j].Sort
	})

	// 返回响应对象
	return &product_resp.ProductRecommendListResp{
		List:   products.List,
		Title:  packageRecommend.Title,
		IsOpen: packageRecommend.IsOpen(),
	}, nil
}

// 解析商品推荐中的JSON字符串
func (s *productSrv) ParseRecommendInfo(packageRecommend *model.ProductPackageRecommend) ([]ProductItemInfo, error) {
	var packageRecommendArray []ProductItemInfo
	if err := json.Unmarshal([]byte(packageRecommend.ProductPackages), &packageRecommendArray); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 按照sort排序, sort是字符串类型, 小的在前
	sort.Slice(packageRecommendArray, func(i, j int) bool {
		return packageRecommendArray[i].Sort < packageRecommendArray[j].Sort
	})
	return packageRecommendArray, nil
}

type ProductItemInfo struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
	Sort string `json:"sort"`
}

// SearchProducts 搜索商品
func (s *productSrv) SearchProducts(ctx context.Context, req req.ProductSearchReq) (*product_resp.ProductSearchResp, error) {
	dbId := ctx.GetDbId()
	// 获取产品列表
	commonRepo := repository.NewCommonRepo()
	sourceMap := map[string]repository.DBOption{
		constant.SourceCashier:   commonRepo.WhereByIsShowCashier(1),
		constant.SourceAssistant: commonRepo.WhereByIsShowAssistant(1),
		constant.SourceTablet:    commonRepo.WhereByIsShowTablet(1),
		constant.SourceKitchen:   commonRepo.WhereByIsShowKitchen(1),
		constant.SourceH5:        commonRepo.WhereByIsShowH5(1),
		constant.SourceMember:    commonRepo.WhereByIsShowMember(1),
	}
	productRepo := repository.NewProductRepo(s.dbm.GetDB(dbId))
	var dbOptions []repository.DBOption
	if option, ok := sourceMap[ctx.GetSource()]; ok {
		dbOptions = append(dbOptions, option)
	}

	// 会员端查询商品列表，预加载外送税
	dbOptions = append(dbOptions, commonRepo.Preload(
		repository.WithPreload{
			Query: "TakeoutTax",
		},
	))

	// 添加搜索条件
	dbOptions = append(dbOptions,
		commonRepo.WhereByStatus(1),
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereLikeByName(req.Keyword), // 根据关键词搜索商品名称
		commonRepo.SortWithSort("ASC"),
		commonRepo.SortWithID("DESC"),
	)

	// 获取商品列表，不分页
	products, _, err := productRepo.GetProductListWithPagination(
		1,    // 第一页
		1000, // 足够大的页面大小，实际不分页
		dbOptions...,
	)

	// 处理错误
	if err != nil {
		return nil, errors.WithMessage(err, "搜索商品失败")
	}

	// 如果是会员端查询商品列表
	if req.IsMember {
		// 获取外送折扣率
		// 获取门店业务设置
		businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
		if err != nil {
			return nil, errors.WithMessage(err, "获取门店业务设置失败")
		}
		// 获取外送折扣率
		deliveryPriceRatio := businessSetting.GetDeliveryPriceRatio()

		taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
		if err != nil {
			return nil, errors.WithMessage(err, "获取门店设置失败")
		}
		taxFeeType := taxRateSetting.GetTaxFeeType()

		// 返回响应对象
		return &product_resp.ProductSearchResp{
			List: FormatProducts(ctx, products, WithTakeoutDiscountRate(deliveryPriceRatio, taxFeeType)),
		}, nil
	}

	// 返回响应对象
	return &product_resp.ProductSearchResp{
		List: FormatProducts(ctx, products),
	}, nil
}

// GetProductUnitList 获取产品单位列表
func (s *productSrv) GetProductUnitList(ctx context.Context, req req.ProductUnitListReq) (product_resp.ProductUnitListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	language := ctx.GetLanguage()
	productRepo := repository.NewProductRepo(db)
	units, total, err := productRepo.PaginateGetProductUnitList(
		req.PageNo,
		req.PageSize,
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductPackages(),
		productRepo.WithProductPackagesMultiLanguageName(),
	)
	if err != nil {
		return product_resp.ProductUnitListResp{}, errors.WithMessage(err, "获取产品单位列表失败")
	}

	productUnitList := make([]product_resp.ProductUnitItem, 0, len(units))
	for _, unit := range units {
		productUnitList = append(productUnitList, product_resp.ProductUnitItem{
			Uuid:                unit.Uuid,
			Name:                unit.MultiLanguageName.GetNameByLang(language),
			Sort:                unit.Sort,
			ProductPackageCount: unit.ProductPackageCount,
		})
	}

	// 返回响应对象
	return product_resp.ProductUnitListResp{
		List: productUnitList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *productSrv) GetProductUnit(ctx context.Context, getUnitReq req.ProductUnitReq) (product_resp.ProductUnitDetail, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	language := ctx.GetLanguage()

	productRepo := repository.NewProductRepo(db)
	unit, err := productRepo.GetProductUnit(
		productRepo.WhereUuid(getUnitReq.Uuid),
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductPackages(),
		productRepo.WithProductPackagesMultiLanguageName(),
	)
	if err != nil {
		return product_resp.ProductUnitDetail{}, errors.WithMessage(err, "获取产品单位详情失败")
	}

	productPackages := make([]product_resp.ProductUnitProductPackage, 0, len(unit.ProductPackages))
	for _, productPackage := range unit.ProductPackages {
		productPackages = append(productPackages, product_resp.ProductUnitProductPackage{
			Uuid: productPackage.Uuid,
			Name: productPackage.MultiLanguageName.GetNameByLang(language),
		})
	}

	productUnit := product_resp.ProductUnitDetail{
		Uuid:       unit.Uuid,
		LocaleName: unit.MultiLanguageName.GetNames(),
		ProductPackages: product_resp.ProductUnitProductPackageList{
			List: productPackages,
		},
	}

	return productUnit, nil
}

func (s *productSrv) AddProductUnit(ctx context.Context, addReq req.ProductUnitAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: "unit",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}
	// 获取当前最大的排序值
	var maxSort int
	db.Model(&model.ProductUnit{}).Select("MAX(sort)").Scan(&maxSort)

	err := db.Transaction(func(tx *gorm.DB) error {
		// 保存多语言名称
		multiLanguageName := model.MultiLanguageName{
			ZhName:   addReq.LocaleName.ZH,
			ThName:   addReq.LocaleName.TH,
			EnName:   addReq.LocaleName.EN,
			ZhTwName: addReq.LocaleName.ZHTW,
			JaName:   addReq.LocaleName.JA,
			KoName:   addReq.LocaleName.KO,
			MyName:   addReq.LocaleName.MY,
			TrName:   addReq.LocaleName.TR,
			SvName:   addReq.LocaleName.SV,
		}
		err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return errors.WithMessage(err, "保存多语言名称失败")
		}
		// 保存产品单位
		productUnit := model.ProductUnit{
			Sort:                  maxSort + 1,
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Name:                  addReq.LocaleName.ToJson(),
		}
		err = tx.Model(&model.ProductUnit{}).Create(&productUnit).Error
		if err != nil {
			return errors.WithMessage(err, "保存产品单位失败")
		}

		// 修改商品的单位UUID
		for _, productPackageUuid := range addReq.ProductPackageUuids {
			err = tx.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			}).Error
			if err != nil {
				return errors.WithMessage(err, "修改商品的单位UUID失败")
			}
		}
		return nil
	})
	return err
}

func (s *productSrv) EditProductUnit(ctx context.Context, editUnitReq req.ProductUnitEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	// 获取产品单位
	productUnit, err := productRepo.GetProductUnit(productRepo.WhereUuid(editUnitReq.Uuid), productRepo.WithMultiLanguageName())
	if err != nil {
		return errors.WithMessage(err, "单位不存在")
	}

	if productUnit.MultiLanguageNameUuid == 0 {
		return errors.New("单位名称不存在")
	}

	// 检查名称是否存在
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editUnitReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editUnitReq.Uuid,
		Source: "unit",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 修改多语言名称
		err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productUnit.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    editUnitReq.LocaleName.ZH,
			"th_name":    editUnitReq.LocaleName.TH,
			"en_name":    editUnitReq.LocaleName.EN,
			"zh_tw_name": editUnitReq.LocaleName.ZHTW,
			"ja_name":    editUnitReq.LocaleName.JA,
			"ko_name":    editUnitReq.LocaleName.KO,
			"my_name":    editUnitReq.LocaleName.MY,
			"tr_name":    editUnitReq.LocaleName.TR,
			"sv_name":    editUnitReq.LocaleName.SV,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改多语言名称失败")
		}
		// 修改产品单位
		err = tx.Model(&model.ProductUnit{}).Where("uuid = ?", editUnitReq.Uuid).Updates(map[string]any{
			"name": editUnitReq.LocaleName.ToJson(),
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改产品单位失败")
		}
		// 修改商品的单位UUID, 如果商品的单位UUID是当前单位，则修改为0
		err = tx.Model(&model.ProductPackage{}).Where("unit_uuid = ?", editUnitReq.Uuid).Updates(map[string]any{
			"unit_uuid": 0,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改商品的单位UUID失败")
		}
		// 修改商品的单位UUID
		if len(editUnitReq.ProductPackageUuids) > 0 {
			// 修改商品的单位UUID
			err = tx.Model(&model.ProductPackage{}).Where("uuid in (?)", editUnitReq.ProductPackageUuids).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			}).Error
			if err != nil {
				return errors.WithMessage(err, "修改商品的单位UUID失败")
			}
		}
		return nil
	})
	return err
}

func (s *productSrv) DeleteProductUnit(ctx context.Context, deleteUnitReq req.ProductUnitReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	// 获取产品单位
	productUnit, err := productRepo.GetProductUnit(
		productRepo.WhereUuid(deleteUnitReq.Uuid),
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductPackages(),
	)
	if err != nil {
		return errors.WithMessage(err, "单位不存在")
	}
	if productUnit.MultiLanguageNameUuid == 0 {
		return errors.New("单位名称不存在")
	}
	// 是否关联商品
	if len(productUnit.ProductPackages) > 0 {
		return errors.New("该单位下存在商品，不允许删除")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除产品单位
		err := tx.Model(&model.ProductUnit{}).Where("uuid = ?", deleteUnitReq.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除产品单位失败")
		}
		// 删除多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productUnit.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除多语言名称失败")
		}
		return nil
	})
	return err
}

func (s *productSrv) SortProductUnit(ctx context.Context, sortReq req.ProductUnitSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productUnitUuids := make([]uint64, 0, len(sortReq.List))
	for _, item := range sortReq.List {
		productUnitUuids = append(productUnitUuids, item.Uuid)
	}
	productUnits, _ := productRepo.GetProductUnitCount(productRepo.WhereUuidIn(productUnitUuids))
	if productUnits != int64(len(productUnitUuids)) {
		return errors.New("单位不存在")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, item := range sortReq.List {
			err := tx.Model(&model.ProductUnit{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			}).Error
			if err != nil {
				return errors.WithMessage(err, "排序产品单位失败")
			}
		}
		return nil
	})
	return err
}

func (s *productSrv) GetProductSauceList(ctx context.Context, sauceListReq req.ProductSauceListReq) (product_resp.ProductSauceListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productSauceList, total, err := productRepo.PaginateGetProductSauceList(sauceListReq.PageNo, sauceListReq.PageSize, productRepo.WithMultiLanguageName())
	if err != nil {
		return product_resp.ProductSauceListResp{}, errors.WithMessage(err, "获取商品加料列表失败")
	}
	productSauceListResp := make([]product_resp.ProductSauceItem, 0, len(productSauceList))
	for _, productSauce := range productSauceList {
		productSauceListResp = append(productSauceListResp, product_resp.ProductSauceItem{
			Uuid:                productSauce.Uuid,
			Name:                productSauce.MultiLanguageName.GetNameByLang(language),
			Sort:                productSauce.Sort,
			ProductPackageCount: productSauce.ProductPackageCount,
		})
	}
	return product_resp.ProductSauceListResp{
		List: productSauceListResp,
		Meta: dto.PageResponse{
			PageNo:   sauceListReq.PageNo,
			PageSize: sauceListReq.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *productSrv) GetProductSauce(ctx context.Context, sauceReq req.ProductSauceReq) (product_resp.ProductSauceDetail, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productSauce, err := productRepo.GetProductSauce(
		productRepo.WhereUuid(sauceReq.Uuid),
		productRepo.WithMultiLanguageName(),
		productRepo.WithActiveProductBoms(),
		productRepo.WithActiveProductBomsProductPackages(),
		productRepo.WithActiveProductBomsProductPackagesMultiLanguageName(),
	)
	if err != nil {
		return product_resp.ProductSauceDetail{}, errors.WithMessage(err, "获取商品加料详情失败")
	}

	productPackages := make([]product_resp.ProductSauceProductPackage, 0, len(productSauce.ProductBoms))
	for _, productBom := range productSauce.ProductBoms {
		productPackages = append(productPackages, product_resp.ProductSauceProductPackage{
			Uuid: productBom.ProductPackage.Uuid,
			Name: productBom.ProductPackage.MultiLanguageName.GetNameByLang(language),
		})
	}
	return product_resp.ProductSauceDetail{
		Uuid:       productSauce.Uuid,
		Price:      productSauce.Price,
		LocaleName: productSauce.MultiLanguageName.GetNames(),
		ProductPackages: product_resp.ProductSauceProductPackageList{
			List: productPackages,
		},
	}, nil
}

func (s *productSrv) AddProductSauce(ctx context.Context, addReq req.ProductSauceAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: "sauce",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	productRepo := repository.NewProductRepo(db)
	// 检查商品是否存在
	productPackages, err := productRepo.GetProductPackageListByUuids(addReq.ProductPackageUuids)
	if err != nil {
		return errors.WithMessage(err, "商品不存在")
	}
	if len(productPackages) != len(addReq.ProductPackageUuids) {
		return errors.New("商品不存在")
	}

	// 获取当前最大的排序值
	var maxSort int
	db.Model(&model.ProductSauce{}).Select("MAX(sort)").Scan(&maxSort)

	err = db.Transaction(func(tx *gorm.DB) error {
		// 保存多语言名称
		multiLanguageName := model.MultiLanguageName{
			ZhName:   addReq.LocaleName.ZH,
			ThName:   addReq.LocaleName.TH,
			EnName:   addReq.LocaleName.EN,
			ZhTwName: addReq.LocaleName.ZHTW,
			JaName:   addReq.LocaleName.JA,
			KoName:   addReq.LocaleName.KO,
			MyName:   addReq.LocaleName.MY,
			TrName:   addReq.LocaleName.TR,
			SvName:   addReq.LocaleName.SV,
		}
		err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return errors.WithMessage(err, "保存多语言名称失败")
		}
		// 保存商品加料
		productSauce := model.ProductSauce{
			Sort:                  maxSort + 1,
			Price:                 addReq.Price,
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Name:                  addReq.LocaleName.ToJson(),
		}
		err = tx.Model(&model.ProductSauce{}).Create(&productSauce).Error
		if err != nil {
			return errors.WithMessage(err, "保存商品加料失败")
		}

		var boms []model.ProductBom
		for _, productPackageUuid := range addReq.ProductPackageUuids {
			// 创建商品BOM
			boms = append(boms, model.ProductBom{
				Name:               addReq.LocaleName.ToJson(),
				Price:              addReq.Price,
				StockNum:           99999999,
				IsOpenStock:        1,
				Status:             1,
				ProductPackageUuid: productPackageUuid,
				ProductSauceUuid:   productSauce.Uuid,
			})
		}
		err = repository.NewProductBomRepo(tx).CreateProductBoms(boms)
		if err != nil {
			return errors.WithMessage(err, "保存商品BOM失败")
		}
		return nil
	})
	return err
}

func (s *productSrv) EditProductSauce(ctx context.Context, editReq req.ProductSauceEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	// 判断名称是否存在
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: "sauce",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	// 获取商品加料
	productSauce, err := productRepo.GetProductSauce(
		productRepo.WhereUuid(editReq.Uuid),
		productRepo.WithMultiLanguageName(),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品加料详情失败")
	}

	if productSauce.MultiLanguageNameUuid == 0 {
		return errors.New("商品加料名称不存在")
	}

	// 检查商品是否存在
	productPackages, err := productRepo.GetProductPackageListByUuids(editReq.ProductPackageUuids)
	if err != nil {
		return errors.WithMessage(err, "商品不存在")
	}
	if len(productPackages) != len(editReq.ProductPackageUuids) {
		return errors.New("商品不存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		name := editReq.LocaleName.ToJson()
		// 修改多语言名称
		err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productSauce.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    editReq.LocaleName.ZH,
			"th_name":    editReq.LocaleName.TH,
			"en_name":    editReq.LocaleName.EN,
			"zh_tw_name": editReq.LocaleName.ZHTW,
			"ja_name":    editReq.LocaleName.JA,
			"ko_name":    editReq.LocaleName.KO,
			"my_name":    editReq.LocaleName.MY,
			"tr_name":    editReq.LocaleName.TR,
			"sv_name":    editReq.LocaleName.SV,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改多语言名称失败")
		}
		// 修改商品加料
		err = tx.Model(&model.ProductSauce{}).Where("uuid = ?", editReq.Uuid).Updates(map[string]any{
			"name":  name,
			"price": editReq.Price,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改商品加料失败")
		}
		if len(editReq.ProductPackageUuids) > 0 {
			bomRepo := repository.NewProductBomRepo(tx)
			boms, _ := bomRepo.GetProductBoms(bomRepo.WhereProductSauceUuid(editReq.Uuid), repository.NotDeleted)
			var bomProductPackageUuids []uint64
			// boms中product_sauce_uuid不是editReq.Uuid的记录，删除
			for _, bom := range boms {
				if !slices.Contains(editReq.ProductPackageUuids, bom.ProductPackageUuid) {
					tx.Model(&model.ProductBom{}).Where("uuid = ?", bom.Uuid).Update("delete_time", time.Now().Unix())
				}
				bomProductPackageUuids = append(bomProductPackageUuids, bom.ProductPackageUuid)
			}
			// editReq.ProductPackageUuids 中，不在bomProductPackageUuids中的记录，创建新的bom
			var newBoms []model.ProductBom
			for _, productPackageUuid := range editReq.ProductPackageUuids {
				if !slices.Contains(bomProductPackageUuids, productPackageUuid) {
					newBoms = append(newBoms, model.ProductBom{
						Name:               name,
						Price:              editReq.Price,
						StockNum:           99999999,
						IsOpenStock:        1,
						Status:             1,
						ProductPackageUuid: productPackageUuid,
						ProductSauceUuid:   editReq.Uuid,
					})
				}
			}
			err = bomRepo.CreateProductBoms(newBoms)
			if err != nil {
				return errors.WithMessage(err, "保存商品BOM失败")
			}
		} else {
			// 删除bom中sauce_uuid为当前商品加料的记录
			tx.Model(&model.ProductBom{}).Where("product_sauce_uuid = ?", editReq.Uuid).Update("delete_time", time.Now().Unix())
		}
		return nil
	})
	return err
}

func (s *productSrv) DeleteProductSauce(ctx context.Context, deleteReq req.ProductSauceReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productSauce, err := productRepo.GetProductSauce(
		productRepo.WhereUuid(deleteReq.Uuid),
		productRepo.WithMultiLanguageName(),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品加料详情失败")
	}
	if productSauce.MultiLanguageNameUuid == 0 {
		return errors.New("商品加料名称不存在")
	}
	// bom是否存在
	bomRepo := repository.NewProductBomRepo(db)
	boms, _ := bomRepo.GetProductBoms(bomRepo.WhereProductSauceUuid(deleteReq.Uuid), repository.NotDeleted)
	if len(boms) > 0 {
		return errors.New("该加料下存在商品，不允许删除")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除商品加料
		err := tx.Model(&model.ProductSauce{}).Where("uuid = ?", deleteReq.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除商品加料失败")
		}
		// 删除多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productSauce.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除多语言名称失败")
		}
		return nil
	})
	return err
}

func (s *productSrv) SortProductSauce(ctx context.Context, sortReq req.ProductSauceSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productSauceUuids := make([]uint64, 0, len(sortReq.List))
	for _, item := range sortReq.List {
		productSauceUuids = append(productSauceUuids, item.Uuid)
	}
	productSauceCount, _ := productRepo.GetProductSauceCount(productRepo.WhereUuidIn(productSauceUuids))
	if productSauceCount != int64(len(productSauceUuids)) {
		return errors.New("商品加料不存在")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, item := range sortReq.List {
			err := tx.Model(&model.ProductSauce{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			}).Error
			if err != nil {
				return errors.WithMessage(err, "排序商品加料失败")
			}
		}
		return nil
	})
	return err
}

func (s *productSrv) GetProductAttributeGroupList(ctx context.Context, req req.ProductAttributeGroupListReq) (product_resp.ProductAttributeGroupListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productAttributeGroupList, total, err := productRepo.PaginateGetProductAttributeGroupList(
		req.PageNo, req.PageSize,
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductAttributes(),
		productRepo.WithProductAttributesMultiLanguageName(),
	)
	if err != nil {
		return product_resp.ProductAttributeGroupListResp{}, errors.WithMessage(err, "获取商品属性分组列表失败")
	}
	productAttributeGroupListResp := make([]product_resp.ProductAttributeGroupItem, 0, len(productAttributeGroupList))
	for _, productAttributeGroup := range productAttributeGroupList {
		attributeNames := make([]string, 0, len(productAttributeGroup.ProductAttributes))
		for _, productAttribute := range productAttributeGroup.ProductAttributes {
			attributeNames = append(attributeNames, productAttribute.MultiLanguageName.GetNameByLang(language))
		}
		productAttributeGroupListResp = append(productAttributeGroupListResp, product_resp.ProductAttributeGroupItem{
			Uuid:          productAttributeGroup.Uuid,
			Name:          productAttributeGroup.MultiLanguageName.GetNameByLang(language),
			AttributeName: strings.Join(attributeNames, "、"),
		})
	}
	return product_resp.ProductAttributeGroupListResp{
		List: productAttributeGroupListResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *productSrv) GetProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupReq) (product_resp.ProductAttributeGroupDetail, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productAttributeGroup, err := productRepo.GetProductAttributeGroup(
		productRepo.WithMultiLanguageName(),                                                                                    // 多语言
		productRepo.WithProductAttributes(),                                                                                    // 商品属性
		productRepo.WithProductAttributesMultiLanguageName(),                                                                   // 商品属性多语言
		productRepo.WithProductAttributesProductPackageAttributes(),                                                            // 商品属性关联的产品包属性
		productRepo.WithProductAttributesProductPackageAttributesProductPackageAttributeGroup(),                                // 商品属性关联的产品包属性组
		productRepo.WithProductAttributesProductPackageAttributesProductPackageAttributeGroupProductPackage(),                  // 商品属性关联的产品包属性组关联的产品包
		productRepo.WithProductAttributesProductPackageAttributesProductPackageAttributeGroupProductPackageMultiLanguageName(), // 商品属性关联的产品包属性组关联的产品包多语言名称
		productRepo.WhereUuid(req.Uuid),
	)
	if err != nil {
		return product_resp.ProductAttributeGroupDetail{}, errors.WithMessage(err, "获取商品属性分组详情失败")
	}

	productAttributeGroupResp := product_resp.ProductAttributeGroupDetail{
		Uuid:       productAttributeGroup.Uuid,
		LocaleName: productAttributeGroup.MultiLanguageName.GetNames(),
	}

	productAttributeList := make([]product_resp.ProductAttribute, 0, len(productAttributeGroup.ProductAttributes))
	for _, productAttribute := range productAttributeGroup.ProductAttributes {
		productPackageList := make([]product_resp.ProductAttributeProductPackage, 0, len(productAttribute.ProductPackageAttributes))
		for _, productPackageAttribute := range productAttribute.ProductPackageAttributes {
			productPackageList = append(productPackageList, product_resp.ProductAttributeProductPackage{
				Uuid: productPackageAttribute.ProductPackageAttributeGroup.ProductPackage.Uuid,
				Name: productPackageAttribute.ProductPackageAttributeGroup.ProductPackage.MultiLanguageName.GetNameByLang(language),
			})
		}
		productAttributeList = append(productAttributeList, product_resp.ProductAttribute{
			Uuid:       productAttribute.Uuid,
			LocaleName: productAttribute.MultiLanguageName.GetNames(),
			ProductPackages: product_resp.ProductAttributeProductPackageList{
				List: productPackageList,
			},
		})
	}

	productAttributeGroupResp.Attributes = product_resp.ProductAttributes{
		List: productAttributeList,
	}

	return productAttributeGroupResp, nil
}

// GetProductFlavorList 获取商品规格列表
func (s *productSrv) GetProductFlavorList(ctx context.Context, req req.ProductFlavorListReq) (product_resp.ProductFlavorListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	opts := []repository.DBOption{
		productRepo.WithMultiLanguageName(),
	}
	if req.Keyword != "" {
		opts = append(opts, commonRepo.WhereLike("ttpos_product_flavor.name", req.Keyword))
	}

	productFlavorList, total, err := productRepo.PaginateGetProductFlavorList(
		req.PageNo, req.PageSize,
		opts...,
	)
	if err != nil {
		return product_resp.ProductFlavorListResp{}, errors.WithMessage(err, "获取商品规格列表失败")
	}
	productFlavorListResp := make([]product_resp.ProductFlavorItemResp, 0, len(productFlavorList))
	for _, productFlavor := range productFlavorList {
		productFlavorListResp = append(productFlavorListResp, product_resp.ProductFlavorItemResp{
			Uuid:                productFlavor.Uuid,
			LocaleName:          productFlavor.MultiLanguageName.GetNames(),
			Sort:                productFlavor.Sort,
			ProductPackageCount: productFlavor.ProductPackageCount,
		})
	}
	return product_resp.ProductFlavorListResp{
		List: productFlavorListResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetProductFlavor 获取商品规格详情
func (s *productSrv) GetProductFlavor(ctx context.Context, flavorReq req.ProductFlavorReq) (product_resp.ProductFlavorDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productFlavor, err := productRepo.GetProductFlavor(
		productRepo.WhereUuid(flavorReq.Uuid),
		productRepo.WithMultiLanguageName(),
		productRepo.WithActiveProductBoms(
			productRepo.WhereProductSauceUuid(0),
		),
		productRepo.WithActiveProductBomsProductPackages(),
		productRepo.WithActiveProductBomsProductPackagesMultiLanguageName(),
	)
	if err != nil {
		return product_resp.ProductFlavorDetailResp{}, errors.WithMessage(err, "获取商品规格详情失败")
	}

	productPackages := make([]product_resp.ProductFlavorProductPackageResp, 0, len(productFlavor.ProductBoms))
	for _, productBom := range productFlavor.ProductBoms {
		productPackages = append(productPackages, product_resp.ProductFlavorProductPackageResp{
			Uuid:    productBom.ProductPackage.Uuid,
			BomUuid: productBom.Uuid,
			Name:    productBom.ProductPackage.MultiLanguageName.GetNameByLang(language),
			Price:   productBom.Price,
		})
	}
	return product_resp.ProductFlavorDetailResp{
		Uuid:       productFlavor.Uuid,
		LocaleName: productFlavor.MultiLanguageName.GetNames(),
		ProductPackageList: product_resp.ProductFlavorProductPackageListResp{
			List: productPackages,
		},
	}, nil
}

func (s *productSrv) AddProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	// 遍历req.ProductAttributes，判断商品是否存在
	productPackageUuids := []uint64{}
	for _, productAttribute := range req.ProductAttributes {
		productPackageUuids = append(productPackageUuids, productAttribute.ProductPackageUuids...)
	}
	// 去重
	productPackageUuids = slices.Compact(productPackageUuids)
	if len(productPackageUuids) > 0 {
		// 判断商品是否存在
		productPackageList, err := productRepo.GetProductPackageListByUuids(productPackageUuids)
		if err != nil {
			return errors.WithMessage(err, "商品不存在")
		}
		if len(productPackageList) != len(productPackageUuids) {
			return errors.New("商品不存在")
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// 保存多语言
		multiLanguageName := model.MultiLanguageName{
			ZhName:   req.LocaleName.ZH,
			ThName:   req.LocaleName.TH,
			EnName:   req.LocaleName.EN,
			ZhTwName: req.LocaleName.ZHTW,
			JaName:   req.LocaleName.JA,
			KoName:   req.LocaleName.KO,
			MyName:   req.LocaleName.MY,
			TrName:   req.LocaleName.TR,
			SvName:   req.LocaleName.SV,
		}
		err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return errors.WithMessage(err, "保存多语言名称失败")
		}
		// 添加商品属性分组
		productAttributeGroup := model.ProductAttributeGroup{
			Name:                  req.LocaleName.ToJson(),
			MultiLanguageNameUuid: multiLanguageName.Uuid,
		}
		err = tx.Model(&model.ProductAttributeGroup{}).Create(&productAttributeGroup).Error
		if err != nil {
			return errors.WithMessage(err, "保存商品属性分组失败")
		}

		// 每个商品关联多个属性
		productPackageMapAttributeUuids := make(map[uint64][]uint64)

		// 遍历req.ProductAttributes，添加商品属性
		for _, productAttribute := range req.ProductAttributes {
			// 保存多语言
			multiLanguageName := model.MultiLanguageName{
				ZhName:   productAttribute.LocaleName.ZH,
				ThName:   productAttribute.LocaleName.TH,
				EnName:   productAttribute.LocaleName.EN,
				ZhTwName: productAttribute.LocaleName.ZHTW,
				JaName:   productAttribute.LocaleName.JA,
				KoName:   productAttribute.LocaleName.KO,
				MyName:   productAttribute.LocaleName.MY,
				TrName:   productAttribute.LocaleName.TR,
				SvName:   productAttribute.LocaleName.SV,
			}
			err = tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
			if err != nil {
				return errors.WithMessage(err, "保存多语言名称失败")
			}
			// 添加商品属性
			productAttributeModel := model.ProductAttribute{
				Name:                  productAttribute.LocaleName.ToJson(),
				MultiLanguageNameUuid: multiLanguageName.Uuid,
				AttributeGroupUuid:    productAttributeGroup.Uuid,
			}
			err = tx.Model(&model.ProductAttribute{}).Create(&productAttributeModel).Error
			if err != nil {
				return errors.WithMessage(err, "保存商品属性失败")
			}

			// 按照product_package_uuid分组，将product_attribute_uuid保存到productPackageMapAttributeUuids中
			for _, productPackageUuid := range productAttribute.ProductPackageUuids {
				productPackageMapAttributeUuids[productPackageUuid] = append(productPackageMapAttributeUuids[productPackageUuid], productAttributeModel.Uuid)
			}
		}
		// 遍历每个商品，添加商品属性组和商品属性
		for productPackageUuid, attributeUuids := range productPackageMapAttributeUuids {
			// 关联商品属性组
			productPackageAttributeGroup := model.ProductPackageAttributeGroup{
				ProductPackageUuid:        productPackageUuid,
				ProductAttributeGroupUuid: productAttributeGroup.Uuid,
			}
			err = tx.Model(&model.ProductPackageAttributeGroup{}).Create(&productPackageAttributeGroup).Error
			if err != nil {
				return errors.WithMessage(err, "保存商品包属性组失败")
			}

			// 关联多个属性，添加到product_package_attribute表中
			productPackageAttributeList := make([]model.ProductPackageAttribute, 0, len(attributeUuids))
			for _, attributeUuid := range attributeUuids {
				productPackageAttributeList = append(productPackageAttributeList, model.ProductPackageAttribute{
					ProductPackageAttributeGroupUuid: productPackageAttributeGroup.Uuid,
					AttributeUuid:                    attributeUuid,
				})
			}
			err = tx.Model(&model.ProductPackageAttribute{}).Create(productPackageAttributeList).Error
			if err != nil {
				return errors.WithMessage(err, "保存商品包属性失败")
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "添加属性失败")
	}
	return nil
}

// AddProductFlavor 添加商品规格
func (s *productSrv) AddProductFlavor(ctx context.Context, addReq req.ProductFlavorAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: "flavor",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	maxSort, err := productRepo.GetProductFlavorMaxSort(
		commonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品规格最大排序失败")
	}
	sort := int(maxSort + 1)

	err = db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)
		// 保存多语言名称
		multiLanguageName := model.MultiLanguageName{
			ZhName:   addReq.LocaleName.ZH,
			ThName:   addReq.LocaleName.TH,
			EnName:   addReq.LocaleName.EN,
			ZhTwName: addReq.LocaleName.ZHTW,
			JaName:   addReq.LocaleName.JA,
			KoName:   addReq.LocaleName.KO,
			MyName:   addReq.LocaleName.MY,
			TrName:   addReq.LocaleName.TR,
			SvName:   addReq.LocaleName.SV,
		}
		err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return err
		}
		// 保存产品规格
		productFlavor := model.ProductFlavor{
			Name:                  addReq.LocaleName.ToJson(),
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Sort:                  sort,
		}
		err = tx.Model(&model.ProductFlavor{}).Create(&productFlavor).Error
		if err != nil {
			return err
		}

		var productBoms []model.ProductBom
		if len(addReq.List) > 0 {
			for _, item := range addReq.List {
				if item.Price < 0 {
					return errors.New("商品价格不能小于0")
				}
				// 判断商品是否存在
				productPackage, err := productRepo.GetProduct(
					productRepo.WhereUuid(item.Uuid),
					productRepo.WhereProductType(constant.ProductTypeProduct),
					commonRepo.WhereBySoftDelete(),
				)
				if productPackage.ID == 0 || err != nil {
					return errors.WithMessage(err, "商品不存在")
				}
				uuid, _ := utils.GetID()
				productBoms = append(productBoms, model.ProductBom{
					BaseModel: model.BaseModel{
						Uuid:       uuid,
						CreateTime: time.Now().Unix(),
						UpdateTime: time.Now().Unix(),
					},
					Price:              item.Price,
					Name:               addReq.LocaleName.ToJson(),
					Status:             int(productPackage.Status),
					ProductFlavorUuid:  productFlavor.Uuid,
					ProductPackageUuid: productPackage.Uuid,
				})
			}
		}

		if len(productBoms) > 0 {
			err := tx.Create(&productBoms).Error
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("添加商品规格失败", zap.Any("func", "AddProductFlavor"), zap.Any("params", addReq), zap.Error(err))
		return errors.WithMessage(err, "添加商品规格失败")
	}

	return nil
}

func (s *productSrv) EditProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productAttributeGroup, err := productRepo.GetProductAttributeGroup(
		productRepo.WhereUuid(req.Uuid),
	)
	if err != nil {
		return errors.WithMessage(err, "属性组不存在")
	}

	var attributeUuids []uint64
	for _, productAttribute := range req.ProductAttributes {
		if productAttribute.Uuid != 0 {
			attributeUuids = append(attributeUuids, productAttribute.Uuid)
		}
	}
	productAttributes, err := productRepo.GetProductAttributes(
		productRepo.WhereUuidIn(attributeUuids),
	)
	if err != nil {
		return errors.WithMessage(err, "属性不存在")
	}
	if len(productAttributes) != len(attributeUuids) {
		return errors.New("属性不存在")
	}

	// 遍历req.ProductAttributes，判断商品是否存在
	newProductPackageUuids := []uint64{}
	for _, productAttribute := range req.ProductAttributes {
		newProductPackageUuids = append(newProductPackageUuids, productAttribute.ProductPackageUuids...)
	}
	// 去重
	newProductPackageUuids = slice.Unique(newProductPackageUuids)
	if len(newProductPackageUuids) > 0 {
		// 判断商品是否存在
		productPackageList, err := productRepo.GetProductPackageListByUuids(newProductPackageUuids)
		if err != nil {
			return errors.WithMessage(err, "商品不存在")
		}
		if len(productPackageList) != len(newProductPackageUuids) {
			return errors.New("商品不存在")
		}
	}

	// 原商品包属性组、商品包属性
	productPackageAttributeGroups, err := productRepo.GetProductPackageAttributeGroups(
		productRepo.WhereProductAttributeGroupUuid(productAttributeGroup.Uuid),
		productRepo.WithProductPackageAttributes(),
	)
	if err != nil {
		return errors.WithMessage(err, "商品包属性组不存在")
	}

	type Attribute struct {
		ProductPackageAttributeGroupUuid uint64
		AttributeUuids                   []uint64
	}
	// 原商品包属性组
	oldProductPackageUuids := []uint64{}

	// 原商品包属性组map
	productPackageAttributeGroupMap := make(map[uint64]Attribute)
	for _, productPackageAttributeGroup := range productPackageAttributeGroups {
		oldProductPackageUuids = append(oldProductPackageUuids, productPackageAttributeGroup.ProductPackageUuid)

		productAttributeUuids := make([]uint64, 0)
		for _, productPackageAttribute := range productPackageAttributeGroup.ProductPackageAttributes {
			productAttributeUuids = append(productAttributeUuids, productPackageAttribute.AttributeUuid)
		}

		productPackageAttributeGroupMap[productPackageAttributeGroup.ProductPackageUuid] = Attribute{
			ProductPackageAttributeGroupUuid: productPackageAttributeGroup.Uuid,
			AttributeUuids:                   productAttributeUuids,
		}
	}

	deletingProductPackageUuids := slice.Difference(oldProductPackageUuids, newProductPackageUuids)
	err = db.Transaction(func(tx *gorm.DB) error {

		// 更新属性组语言
		err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productAttributeGroup.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    req.LocaleName.ZH,
			"th_name":    req.LocaleName.TH,
			"en_name":    req.LocaleName.EN,
			"zh_tw_name": req.LocaleName.ZHTW,
			"ja_name":    req.LocaleName.JA,
			"ko_name":    req.LocaleName.KO,
			"my_name":    req.LocaleName.MY,
			"tr_name":    req.LocaleName.TR,
			"sv_name":    req.LocaleName.SV,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "更新语言失败")
		}
		for k, productAttribute := range req.ProductAttributes {
			if productAttribute.Uuid != 0 {
				err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productAttribute.Uuid).Updates(map[string]any{
					"zh_name":    productAttribute.LocaleName.ZH,
					"th_name":    productAttribute.LocaleName.TH,
					"en_name":    productAttribute.LocaleName.EN,
					"zh_tw_name": productAttribute.LocaleName.ZHTW,
					"ja_name":    productAttribute.LocaleName.JA,
					"ko_name":    productAttribute.LocaleName.KO,
					"my_name":    productAttribute.LocaleName.MY,
					"tr_name":    productAttribute.LocaleName.TR,
					"sv_name":    productAttribute.LocaleName.SV,
				}).Error
				if err != nil {
					return errors.WithMessage(err, "更新语言失败")
				}
			} else {
				// 添加语言
				multiLanguageName := model.MultiLanguageName{
					ZhName:   productAttribute.LocaleName.ZH,
					ThName:   productAttribute.LocaleName.TH,
					EnName:   productAttribute.LocaleName.EN,
					ZhTwName: productAttribute.LocaleName.ZHTW,
					JaName:   productAttribute.LocaleName.JA,
					KoName:   productAttribute.LocaleName.KO,
					MyName:   productAttribute.LocaleName.MY,
					TrName:   productAttribute.LocaleName.TR,
					SvName:   productAttribute.LocaleName.SV,
				}
				err = tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
				if err != nil {
					return errors.WithMessage(err, "添加语言失败")
				}
				productAttributeModel := model.ProductAttribute{
					Name:                  productAttribute.LocaleName.ToJson(),
					MultiLanguageNameUuid: multiLanguageName.Uuid,
					AttributeGroupUuid:    productAttributeGroup.Uuid,
				}
				err = tx.Model(&model.ProductAttribute{}).Create(&productAttributeModel).Error
				if err != nil {
					return errors.WithMessage(err, "添加商品属性失败")
				}
				req.ProductAttributes[k].Uuid = productAttributeModel.Uuid
			}
		}

		// 传递的参数中，商品包和商品属性关联关系
		productPackageAttributeMaps := make(map[uint64][]uint64)
		for _, productAttribute := range req.ProductAttributes {
			for _, productPackageUuid := range productAttribute.ProductPackageUuids {
				if _, ok := productPackageAttributeMaps[productPackageUuid]; !ok {
					productPackageAttributeMaps[productPackageUuid] = []uint64{productAttribute.Uuid}
				} else {
					productPackageAttributeMaps[productPackageUuid] = append(productPackageAttributeMaps[productPackageUuid], productAttribute.Uuid)
				}
			}
		}

		if len(deletingProductPackageUuids) > 0 {
			// 删除商品包属性组
			err := tx.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_uuid IN (?)", deletingProductPackageUuids).Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(err, "删除商品包属性组失败")
			}
		}

		// 遍历传递的参数，对比原商品包属性组和新的商品包属性组
		for productPackageUuid, attributeUuids := range productPackageAttributeMaps {
			if _, ok := productPackageAttributeGroupMap[productPackageUuid]; !ok { // 新增的商品包属性组
				// 添加商品包属性组
				productPackageAttributeGroup := model.ProductPackageAttributeGroup{
					ProductPackageUuid:        productPackageUuid,
					ProductAttributeGroupUuid: productAttributeGroup.Uuid,
				}
				err = tx.Model(&model.ProductPackageAttributeGroup{}).Create(&productPackageAttributeGroup).Error
				if err != nil {
					return errors.WithMessage(err, "添加商品包属性组失败")
				}

				productPackageAttributeList := make([]model.ProductPackageAttribute, 0, len(attributeUuids))
				for _, attributeUuid := range attributeUuids {
					productPackageAttributeList = append(productPackageAttributeList, model.ProductPackageAttribute{
						ProductPackageAttributeGroupUuid: productPackageAttributeGroup.Uuid,
						AttributeUuid:                    attributeUuid,
					})
				}
				err = tx.Model(&model.ProductPackageAttribute{}).Create(productPackageAttributeList).Error
				if err != nil {
					return errors.WithMessage(err, "添加商品包属性失败")
				}
			} else { // 已有的商品包属性组

				deletingAttributeUuids := slice.Difference(productPackageAttributeGroupMap[productPackageUuid].AttributeUuids, attributeUuids)
				if len(deletingAttributeUuids) > 0 { // 删除的商品包属性
					err = tx.Model(&model.ProductPackageAttribute{}).Where("product_package_attribute_group_uuid = ? AND attribute_uuid IN (?)", productPackageAttributeGroupMap[productPackageUuid].ProductPackageAttributeGroupUuid, deletingAttributeUuids).Update("delete_time", time.Now().Unix()).Error
					if err != nil {
						return errors.WithMessage(err, "删除商品包属性失败")
					}

					// 删除商品属性
					err = tx.Model(&model.ProductAttribute{}).Where("uuid IN (?)", deletingAttributeUuids).Update("delete_time", time.Now().Unix()).Error
					if err != nil {
						return errors.WithMessage(err, "删除商品属性失败")
					}
				}

				addingAttributeUuids := slice.Difference(attributeUuids, productPackageAttributeGroupMap[productPackageUuid].AttributeUuids)
				if len(addingAttributeUuids) > 0 { // 新增的商品包属性
					productPackageAttributeList := make([]model.ProductPackageAttribute, 0, len(addingAttributeUuids))
					for _, attributeUuid := range addingAttributeUuids {
						productPackageAttributeList = append(productPackageAttributeList, model.ProductPackageAttribute{
							ProductPackageAttributeGroupUuid: productPackageAttributeGroupMap[productPackageUuid].ProductPackageAttributeGroupUuid,
							AttributeUuid:                    attributeUuid,
						})
					}
					err = tx.Model(&model.ProductPackageAttribute{}).Create(productPackageAttributeList).Error
					if err != nil {
						return errors.WithMessage(err, "添加商品包属性失败")
					}
				}
			}
		}
		return nil
	})
	return err
}

// EditProductFlavor 编辑商品规格
func (s *productSrv) EditProductFlavor(ctx context.Context, editReq req.ProdudctFlavorEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: "flavor",
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)

		flavor, err := productRepo.GetProductFlavor(
			productRepo.WhereUuid(editReq.Uuid),
			commonRepo.WhereBySoftDelete(),
		)
		if err != nil || flavor.ID == 0 {
			return errors.WithMessage(err, "获取商品规格详情失败")
		}

		// 更新多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", flavor.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    editReq.LocaleName.ZH,
			"th_name":    editReq.LocaleName.TH,
			"en_name":    editReq.LocaleName.EN,
			"zh_tw_name": editReq.LocaleName.ZHTW,
			"ja_name":    editReq.LocaleName.JA,
			"ko_name":    editReq.LocaleName.KO,
			"my_name":    editReq.LocaleName.MY,
			"tr_name":    editReq.LocaleName.TR,
			"sv_name":    editReq.LocaleName.SV,
		}).Error
		if err != nil {
			return err
		}
		// 更新商品规格
		err = tx.Model(&model.ProductFlavor{}).Where("uuid = ?", editReq.Uuid).Updates(map[string]any{
			"name": editReq.LocaleName.ToJson(),
		}).Error
		if err != nil {
			return err
		}
		// 更新关联商品包
		for _, item := range editReq.List {
			if item.IsDelete {
				// 删除商品BOM
				err := tx.Model(&model.ProductBom{}).Where("uuid = ?", item.BomUuid).Updates(map[string]any{
					"delete_time": time.Now().Unix(),
				}).Error
				if err != nil {
					return err
				}
			} else {
				// 判断商品是否存在
				productPackage, err := productRepo.GetProduct(
					productRepo.WhereUuid(item.Uuid),
					productRepo.WhereProductType(constant.ProductTypeProduct),
					commonRepo.WhereBySoftDelete(),
				)
				if productPackage.ID == 0 || err != nil {
					return errors.WithMessage(err, "商品不存在")
				}
				if item.BomUuid == 0 {
					// 新增商品BOM
					err := tx.Create(&model.ProductBom{
						Price:              item.Price,
						Name:               editReq.LocaleName.ToJson(),
						Status:             int(productPackage.Status),
						ProductFlavorUuid:  flavor.Uuid,
						ProductPackageUuid: item.Uuid,
					}).Error
					if err != nil {
						return err
					}
				} else {
					// 编辑商品BOM
					err := tx.Model(&model.ProductBom{}).Where("uuid = ?", item.BomUuid).Updates(map[string]any{
						"price":                item.Price,
						"name":                 editReq.LocaleName.ToJson(),
						"status":               int(productPackage.Status),
						"product_package_uuid": item.Uuid,
					}).Error
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("编辑商品规格失败", zap.Any("func", "EditProductFlavor"), zap.Any("params", editReq), zap.Error(err))
		return errors.WithMessage(err, "编辑商品规格失败")
	}

	return nil
}

func (s *productSrv) DeleteProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productAttributeGroup, err := productRepo.GetProductAttributeGroup(
		productRepo.WhereUuid(req.Uuid),
		productRepo.WithProductPackageAttributes(),
	)
	if err != nil {
		return errors.WithMessage(err, "属性组不存在")
	}

	productAttributeUuids := []uint64{}
	for _, productAttribute := range productAttributeGroup.ProductAttributes {
		productAttributeUuids = append(productAttributeUuids, productAttribute.Uuid)
	}

	err = db.Transaction(func(tx *gorm.DB) error {

		// 删除商品属性组
		err = tx.Model(&model.ProductAttributeGroup{}).Where("uuid = ?", productAttributeGroup.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除商品属性组失败")
		}

		// 删除商品属性
		err = tx.Model(&model.ProductAttribute{}).Where("attribute_group_uuid = ?", productAttributeGroup.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除商品属性失败")
		}

		// 删除商品包属性组
		err = tx.Model(&model.ProductPackageAttributeGroup{}).Where("product_attribute_group_uuid = ?", productAttributeGroup.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除商品属性组失败")
		}

		// 删除商品包属性
		err = tx.Model(&model.ProductPackageAttribute{}).Where("attribute_uuid IN (?)", productAttributeUuids).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(err, "删除商品包属性失败")
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "删除属性组失败")
	}
	return nil
}

// DeleteProductFlavor 删除商品规格
func (s *productSrv) DeleteProductFlavor(ctx context.Context, deleteReq req.ProductFlavorDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	// 获取商品规格详情
	productFlavor, err := productRepo.GetProductFlavor(
		productRepo.WhereUuid(deleteReq.Uuid),
		commonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品规格详情失败")
	}

	// 判断商品规格是否关联了商品
	count, _ := productRepo.GetProductBomCount(
		commonRepo.WhereByProductFlavorUuid(productFlavor.Uuid),
		commonRepo.WhereBySoftDelete(),
	)
	if count > 0 {
		return errors.New("该规格已经关联了商品，不可删除")
	}

	// 软删除商品规格
	err = db.Model(&model.ProductFlavor{}).Where("uuid = ?", productFlavor.Uuid).Updates(map[string]any{
		"delete_time": time.Now().Unix(),
	}).Error
	if err != nil {
		return errors.WithMessage(err, "删除商品规格失败")
	}

	return nil
}

func (s *productSrv) SortProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	productAttributeGroupUuids := []uint64{}
	for _, productAttributeGroup := range req.List {
		productAttributeGroupUuids = append(productAttributeGroupUuids, productAttributeGroup.Uuid)
	}
	productAttributeGroups, err := productRepo.GetProductAttributeGroups(
		productRepo.WhereUuidIn(productAttributeGroupUuids),
	)
	if err != nil {
		return errors.WithMessage(err, "获取属性组失败")
	}
	if len(productAttributeGroups) != len(productAttributeGroupUuids) {
		return errors.New("属性组不存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, productAttributeGroup := range req.List {
			err = tx.Model(&model.ProductAttributeGroup{}).Where("uuid = ?", productAttributeGroup.Uuid).Update("sort", productAttributeGroup.Sort).Error
			if err != nil {
				return errors.WithMessage(err, "排序属性组失败")
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "排序属性组失败")
	}
	return nil
}

func (s *productSrv) SortProductAttribute(ctx context.Context, req req.ProductAttributeSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	productAttributeUuids := []uint64{}
	for _, productAttribute := range req.List {
		productAttributeUuids = append(productAttributeUuids, productAttribute.Uuid)
	}
	productAttributes, err := productRepo.GetProductAttributes(
		productRepo.WhereProductAttributeGroupUuid(req.ProductAttributeGroupUuid),
		productRepo.WhereUuidIn(productAttributeUuids),
	)
	if err != nil {
		return errors.WithMessage(err, "获取属性失败")
	}
	if len(productAttributes) != len(productAttributeUuids) {
		return errors.New("属性不存在")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, productAttribute := range req.List {
			err = tx.Model(&model.ProductAttribute{}).Where("uuid = ?", productAttribute.Uuid).Update("sort", productAttribute.Sort).Error
			if err != nil {
				return errors.WithMessage(err, "排序属性失败")
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "排序属性失败")
	}
	return nil
}

// SortProductFlavor 排序商品规格
func (s *productSrv) SortProductFlavor(ctx context.Context, req req.ProductFlavorSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	productFlavorUuids := make([]uint64, 0, len(req.List))
	for _, item := range req.List {
		productFlavorUuids = append(productFlavorUuids, item.Uuid)
	}
	productFlavorCount, _ := productRepo.GetProductFlavorCount(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuidIn(productFlavorUuids),
	)
	if productFlavorCount != int64(len(productFlavorUuids)) {
		return errors.New("规格不存在")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.List {
			tx.Model(&model.ProductFlavor{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			})
		}
		return nil
	})
	return err
}
