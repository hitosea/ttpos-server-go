package service

import (
	"encoding/json"
	"fmt"
	"math"
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
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IProductSrv 定义产品服务接口
type IProductSrv interface {
	GetProductList(ctx context.Context, req req.ProductListReq) (product_resp.ProductListWithPaginationResp, error)               // 获取产品列表（销售端）
	GetProductCategoryList(dbId uint64) (product_resp.ProductCategoryListResp, error)                                             // 获取产品类别列表（销售端）
	GetProductUnitList(ctx context.Context, req req.ProductUnitListReq) (product_resp.ProductUnitListResp, error)                 // 获取产品单位列表
	GetProductUnit(ctx context.Context, req req.ProductUnitReq) (product_resp.ProductUnitDetail, error)                           // 获取产品单位详情
	GetProductRecommendList(ctx context.Context, req req.ProductRecommendListReq) (*product_resp.ProductRecommendListResp, error) // 获取产品推荐列表
	SearchProducts(ctx context.Context, req req.ProductSearchReq) (*product_resp.ProductSearchResp, error)                        // 搜索商品

	GetProductShopCategoryList(ctx context.Context, req req.ProductShopCategoryListReq) (product_resp.ProductShopCategoryListResp, error) // 获取产品类别列表（商家端）
	GetProductShopCategory(ctx context.Context, req req.ProductShopCategoryReq) (product_resp.ProductShopCategoryDetailResp, error)       // 获取产品类别详情
	SortProductShopCategory(ctx context.Context, req req.ProductShopCategorySortReq) error                                                // 排序产品分类
	AddProductShopCategory(ctx context.Context, req req.ProductShopCategoryAddReq) error                                                  // 添加产品分类
	EditProductShopCategory(ctx context.Context, req req.ProductShopCategoryEditReq) error                                                // 编辑产品分类
	DeleteProductShopCategory(ctx context.Context, req req.ProductShopCategoryDeleteReq) error                                            // 删除产品分类

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
	DeleteProductAttribute(ctx context.Context, req req.ProductAttributeDeleteReq) error                                                        // 删除商品属性

	GetProductFlavorList(ctx context.Context, req req.ProductFlavorListReq) (product_resp.ProductFlavorListResp, error) // 获取商品规格列表
	GetProductFlavor(ctx context.Context, req req.ProductFlavorReq) (product_resp.ProductFlavorDetailResp, error)       // 获取商品规格详情
	AddProductFlavor(ctx context.Context, req req.ProductFlavorAddReq) error                                            // 添加商品规格
	EditProductFlavor(ctx context.Context, req req.ProductFlavorEditReq) error                                          // 编辑商品规格
	DeleteProductFlavor(ctx context.Context, req req.ProductFlavorDeleteReq) error                                      // 删除商品规格
	SortProductFlavor(ctx context.Context, req req.ProductFlavorSortReq) error                                          // 排序商品规格

	// 导入商品
	ImportProductList(ctx context.Context, req req.ProductImportListReq) (product_resp.ProductImportResp, error) // 导入商品列表
	ImportProduct(ctx context.Context, req req.ProductImportReq) error                                           // 导入商品

	GetProductSingleList(ctx context.Context, req req.ProductSingleListReq) (*product_resp.ProductSingleListResp, error) // 获取单规格商品列表

	GetProductShopList(ctx context.Context, req req.ProductShopListReq) (*product_resp.ProductShopListResp, error)                                                                                                     // 获取商品列表（商家端）
	SortProductShopList(ctx context.Context, req req.SortProductShopListReq) error                                                                                                                                     // 排序商品列表
	GetProductDetail(ctx context.Context, req req.ProductDetailReq) (*product_resp.ProductDetailResp, error)                                                                                                           // 获取商品详情
	ProductShopStatus(ctx context.Context, req req.ProductShopStatusReq) error                                                                                                                                         // 修改商品状态
	ProductTaxList(ctx context.Context) product_resp.ProductTaxListResp                                                                                                                                                // 获取商品税类列表
	AddProductShop(ctx context.Context, req req.ProductShopAddReq) error                                                                                                                                               // 添加商品
	EditProductShop(ctx context.Context, req req.ProductShopEditReq) (*product_resp.ProductEditResp, []string, error)                                                                                                  // 编辑商品
	DeleteProductShop(ctx context.Context, req req.ProductShopDeleteReq) (*product_resp.ProductDeleteResp, error)                                                                                                      // 删除商品
	ProductShopChangePrice(ctx context.Context, req req.ProductShopChangePriceReq) error                                                                                                                               // 商品改价
	AddProductPackage(ctx context.Context, tx *gorm.DB, req req.ProductShopAddReq, price float64) (*AddProductPackageRes, error)                                                                                       // 添加商品包
	EditProductPackage(ctx context.Context, tx *gorm.DB, req req.ProductShopEditReq, price float64) (*AddProductPackageRes, error)                                                                                     // 编辑商品包
	SaveProductPackageBom(ctx context.Context, tx *gorm.DB, productPackageUuid uint64, unitUuid uint64, templateItemCode string, flavorListResult CheckProductFlavorResult, sauceResult CheckProductSauceResult) error // 保存商品bom
	SaveProductPackageAttribute(tx *gorm.DB, param []CheckProductAttributeGroupParam, productPackageUuid uint64) error                                                                                                 // 保存商品属性
	SaveProductPackageGroup(tx *gorm.DB, groupList []CheckProductPackageGroupResult, productPackageUuid uint64) error                                                                                                  // 保存商品套餐组
}

type productSrv struct {
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv          // 多语言名称服务
	settingSrv setting.ISrv        // 设置服务
	cache      cache.Cache         // 缓存
}

func NewProductSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, cache cache.Cache) IProductSrv {
	return NewProductSrvImpl(dbm, localeSrv, settingSrv, cache)
}

func NewProductSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, cache cache.Cache) IProductSrv {
	return &productSrv{
		dbm:        dbm,
		localeSrv:  localeSrv,
		settingSrv: settingSrv,
		cache:      cache,
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
					StockNum:   productBom.GetStockNum(),
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
							StockNum:   productBom.GetStockNum(),
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
							StockNum:          productBom.GetStockNum(),
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
			StockNum:   productBom.GetStockNum(),
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
	language := ctx.GetLanguage()
	// 查询/关联
	opts := []repository.DBOption{
		productRepo.WithMultiLanguageName(),
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
						Name:       child.MultiLanguageName.GetNameByLang(language),
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
				Name:       category.MultiLanguageName.GetNameByLang(language),
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

// GetProductShopCategory 获取产品类别详情
func (s *productSrv) GetProductShopCategory(ctx context.Context, req req.ProductShopCategoryReq) (product_resp.ProductShopCategoryDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productCategory, err := productRepo.GetProductCategory(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(req.Uuid),
		productRepo.WithMultiLanguageName(),
	)
	if err != nil {
		return product_resp.ProductShopCategoryDetailResp{}, errors.WithMessage(err, "获取分类详情失败")
	}
	if productCategory.Uuid == 0 {
		return product_resp.ProductShopCategoryDetailResp{}, errors.New("分类不存在")
	}

	parentName := ""
	if productCategory.ParentUuid != 0 {
		parentCategory, err := productRepo.GetProductCategory(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereUuid(productCategory.ParentUuid),
			productRepo.WithMultiLanguageName(),
		)
		if err != nil {
			return product_resp.ProductShopCategoryDetailResp{}, errors.WithMessage(err, "获取父级分类详情失败")
		}
		if parentCategory.Uuid == 0 {
			return product_resp.ProductShopCategoryDetailResp{}, errors.New("父级分类不存在")
		}
		parentName = parentCategory.MultiLanguageName.GetNameByLang(language)
	}

	// 获取商品数量和子级数量
	var productCount, childCount int64
	opts := []repository.DBOption{
		commonRepo.WhereBySoftDelete(),
	}
	if productCategory.IsSpecial == 1 {
		opts = append(opts, productRepo.WhereSpecialCategoryUuid(productCategory.Uuid))
	} else {
		opts = append(opts, productRepo.WhereCategoryUuid(productCategory.Uuid))
	}
	productCount, _ = productRepo.GetProductCount(opts...)

	if productCategory.ParentUuid == 0 && productCategory.IsSpecial == 0 {
		categoryCount, _ := productRepo.GetProductCategoryCount(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereParentUuid(productCategory.Uuid),
		)
		childCount = categoryCount
	}

	return product_resp.ProductShopCategoryDetailResp{
		Uuid:         productCategory.Uuid,
		LocaleName:   productCategory.MultiLanguageName.GetNames(),
		ParentUuid:   productCategory.ParentUuid,
		ParentName:   parentName,
		Sort:         productCategory.Sort,
		Status:       productCategory.Status,
		ProductCount: productCount,
		ChildCount:   childCount,
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

	sorts := make(map[uint64]int)
	for _, item := range req.List {
		if item.Sort == 0 {
			return errors.New("排序不能为0")
		}
		sorts[item.Uuid] = item.Sort
	}
	err := productRepo.BatchUpdateSort(&model.ProductCategory{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序分类失败"), err.Error())
	}
	return nil
}

// AddProductShopCategory 添加产品分类
func (s *productSrv) AddProductShopCategory(ctx context.Context, addReq req.ProductShopCategoryAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 50) {
			return errors.New("名称长度不能超过50")
		}
	}
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceCategory,
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
		// 删除缓存
		tag := cryptor.Md5String(fmt.Sprintf("category%d%d%d", ctx.GetCompanyUuid(), utils.IfInt(!addReq.IsSpecial, 1, 0), 1))
		err = cache.NewTaggedCache(s.cache).TagClear(tag)
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
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 50) {
			return errors.New("名称长度不能超过50")
		}
	}
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: constant.CheckNameSourceCategory,
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
		// 删除缓存
		tag := cryptor.Md5String(fmt.Sprintf("category%d%d%d", ctx.GetCompanyUuid(), utils.IfInt(productCategory.IsSpecial == 0, 1, 0), 1))
		err = cache.NewTaggedCache(s.cache).TagClear(tag)
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

// DeleteProductShop 删除商品
func (s *productSrv) DeleteProductShop(ctx context.Context, req req.ProductShopDeleteReq) (*product_resp.ProductDeleteResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productBomRepo := repository.NewProductBomRepo(db)
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(db)
	setting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return nil, errors.New("获取公司设置失败")
	}

	product, err := productRepo.GetProduct(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(req.Uuid),
		productRepo.WithProductBoms(
			commonRepo.WhereBySoftDelete(),
		),
	)
	if product.Uuid == 0 || err != nil {
		return nil, errors.New("商品不存在")
	}

	productPackageGroupItems, err := productPackageGroupRepo.GetProductPackageGroupItems(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByRelatedUuid(req.Uuid),
		productPackageGroupRepo.WithProductPackageGroup(commonRepo.WhereBySoftDelete()),
		productPackageGroupRepo.WithProductPackageGroupProduct(commonRepo.WhereBySoftDelete()),
		productPackageGroupRepo.WithProductPackageGroupProductMultiLanguageName(commonRepo.WhereBySoftDelete()),
	)
	if err != nil {
		return nil, errors.New("获取商品套餐组商品失败")
	}
	packageNames := []string{}
	for _, item := range productPackageGroupItems {
		if item.ProductPackageGroup != nil {
			if !slices.Contains(packageNames, item.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage())) {
				packageNames = append(packageNames, item.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage()))
			}
		}
	}
	if len(packageNames) > 0 {
		return &product_resp.ProductDeleteResp{List: packageNames}, errors.New("商品已关联如下套餐，暂时无法删除，请先修改套餐")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		productPackageAttributeGroupRepo := repository.NewProductPackageAttributeGroupRepo(tx)
		productPackageAttributeRepo := repository.NewProductPackageAttributeRepo(tx)
		warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
		// 删除商品包
		err := tx.Model(&model.ProductPackage{}).Where("uuid = ?", req.Uuid).Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
		if err != nil {
			return err
		}
		// 删除商品包关联语言包
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", product.MultiLanguageNameUuid).Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
		if err != nil {
			return err
		}
		for _, productBom := range product.ProductBoms {
			// 删除商品包关联的商品BOM
			err = productBomRepo.UpdateProductBom(map[string]any{"delete_time": time.Now().Unix()}, commonRepo.WhereByUuid(productBom.Uuid))
			if err != nil {
				return err
			}
			if setting.SaleStock == 1 {
				// 如果是规格或者套餐，则删除出库
				if productBom.IsFlavor() || product.ProductType == constant.ProductTypePackage {
					// 删除出库
					outFormUuid, _ := utils.GetID()
					warehouseForm := model.WarehouseOutForm{
						BaseModel: model.BaseModel{
							Uuid: outFormUuid,
						},
						FormNo:       warehouseFormRepo.GenerateWarehouseOutFormNo(setting.Timezone),
						Scene:        constant.WarehouseOutFormSceneDelete,
						Status:       constant.WarehouseOutFormStatusSuccess,
						OperatorUuid: ctx.GetStaffUuid(),
					}
					err = warehouseFormRepo.CreateWarehouseOutFormRecord(warehouseForm)
					if err != nil {
						return errors.WithMessage(err, "保存出库单失败")
					}
					warehouseOutFormItem := model.WarehouseOutFormItem{
						Num:                  productBom.StockNum,
						Scene:                constant.WarehouseOutFormSceneDelete,
						Status:               1,
						ReduceStock:          constant.WarehouseOutFormItemReduceStockSuccess,
						WarehouseOutFormUuid: outFormUuid,
						ProductBomUuid:       productBom.Uuid,
					}
					err = warehouseFormRepo.CreateWarehouseOutFormItemRecord(warehouseOutFormItem)
					if err != nil {
						return errors.WithMessage(err, "保存出库单明细失败")
					}
				}
			}
		}

		groups, err := productPackageAttributeGroupRepo.GetProductPackageAttributeGroups(
			commonRepo.WhereByProductPackageUuid(req.Uuid),
			commonRepo.WhereBySoftDelete(),
		)
		if err != nil {
			return err
		}
		for _, group := range groups {
			err = productPackageAttributeGroupRepo.DeleteProductPackageAttributeGroup(
				commonRepo.WhereByUuid(group.Uuid),
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return err
			}
			err = productPackageAttributeRepo.DeleteProductPackageAttribute(
				commonRepo.WhereByProductPackageAttributeGroupUuid(group.Uuid),
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("删除商品失败", zap.Any("func", "DeleteProductShop"), zap.Any("params", req), zap.Error(err))
		return nil, errors.New("删除商品失败")
	}

	return nil, nil
}

// DeleteProductShopCategory 删除产品分类
func (s *productSrv) DeleteProductShopCategory(ctx context.Context, deleteReq req.ProductShopCategoryDeleteReq) error {
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
		return errors.New("该分类已经关联了商品，不可删除")
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
			return errors.New("该分类存在二级分类，不可删除")
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除分类
		err := tx.Model(&model.ProductCategory{}).Where("uuid = ?", deleteReq.Uuid).Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
		if err != nil {
			return err
		}
		// 删除多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productCategory.MultiLanguageNameUuid).Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
		if err != nil {
			return err
		}
		// 重新排序
		productRepo := repository.NewProductRepo(tx)
		productCategories := make([]model.ProductCategory, 0)
		if productCategory.IsSpecial == 1 {
			productCategories, _ = productRepo.GetProductCategoryList(
				commonRepo.WhereBySoftDelete(),
				productRepo.WhereByIsSpecial(1),
			)
		} else {
			productCategories, _ = productRepo.GetProductCategoryList(
				commonRepo.WhereBySoftDelete(),
				productRepo.WhereParentUuid(productCategory.ParentUuid),
			)
		}
		sorts := make(map[uint64]int)
		for i, productCategory := range productCategories {
			sorts[productCategory.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductCategory{}, sorts)
		if err != nil {
			return errors.WithMessage(errors.New("重新排序分类失败"), err.Error())
		}
		// 删除缓存
		tag := cryptor.Md5String(fmt.Sprintf("category%d%d%d", ctx.GetCompanyUuid(), utils.IfInt(productCategory.IsSpecial == 0, 1, 0), 1))
		err = cache.NewTaggedCache(s.cache).TagClear(tag)
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
		return product_resp.ProductUnitListResp{}, errors.WithMessage(errors.New("获取单位列表失败"), err.Error())
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

// GetProductUnit 获取产品单位详情
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
		return product_resp.ProductUnitDetail{}, errors.WithMessage(errors.New("获取单位详情失败"), err.Error())
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

// AddProductUnit 添加产品单位
func (s *productSrv) AddProductUnit(ctx context.Context, addReq req.ProductUnitAddReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceUnit,
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}
	// 获取当前最大的排序值
	var maxSort int
	db.Model(&model.ProductUnit{}).Scopes(repository.NotDeleted).Select("ifnull(max(sort), 0)").Scan(&maxSort)

	// 保存产品单位
	productUnit := model.ProductUnit{
		Sort: maxSort + 1,
		Name: addReq.LocaleName.ToJson(),
	}
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
			return errors.WithMessage(errors.New("保存名称多语言失败"), err.Error())
		}
		productUnit.MultiLanguageNameUuid = multiLanguageName.Uuid
		err = tx.Model(&model.ProductUnit{}).Create(&productUnit).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存单位失败"), err.Error())
		}

		// 修改商品的单位UUID
		for _, productPackageUuid := range addReq.ProductPackageUuids {
			err = tx.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			}).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存关联商品失败"), err.Error())
			}
		}
		return nil
	})

	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()

	// 开启了ERP，并且是TTPOS站点，同步到ERPNext
	if company.IsOpenErp() && companySetting.IsTtposSite() {
		enName, err := s.getEnName(ctx, addReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		erpUom := company.Name + "-" + enName
		err = erp.NewIErpSrv(s.dbm).SaveUom(ctx.GetContext(), req.SaveUomReq{
			SiteCode:          companySetting.ErpnextSiteCode,
			CompanyAbbr:       companySetting.ErpnextCompanyAbbr,
			Branch:            companySetting.ErpnextBranchName,
			UomName:           erpUom,
			AliasName:         enName,
			MustBeWholeNumber: true,
		})
		if err != nil {
			return errors.WithMessage(errors.New("同步单位到erp失败"), err.Error())
		}
		err = db.Model(&model.ProductUnit{}).Where("uuid = ?", productUnit.Uuid).Update("erpnext_uom", erpUom).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存erp单位失败"), err.Error())
		}

		// 修改商品的单位UUID
		for _, productPackageUuid := range addReq.ProductPackageUuids {
			// 同步更新商品到erp
			if ctx.GetCompany().IsOpenErp() {
				productPackageRepo := repository.NewProductPackageRepo(db)
				productPackage, errGetProductPackage := productPackageRepo.GetProductPackage(
					repository.CommonRepo.WhereByUuid(productPackageUuid),
					repository.CommonRepo.Preload(
						repository.WithPreload{
							Query: "ProductBoms",
						},
					),
				)
				if errGetProductPackage != nil {
					return errors.WithMessage(errGetProductPackage, "获取商品包失败")
				}
				for _, productBom := range productPackage.ProductBoms {
					multiLanguageName := model.NewMultiLanguageName(productBom.Name)
					enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					erpSrv := erp.NewIErpSrv(s.dbm)
					_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
						ItemName: enName,
						StockUom: erpUom,
						ItemCode: productBom.ErpCode,
					})
					if errErp != nil {
						return errors.WithMessage(errErp, "同步商品到erp失败")
					}
				}
			}
		}
	}

	return err
}

func (s *productSrv) getEnName(ctx context.Context, locale dto.LocaleResponse) (string, error) {
	enName := locale.EN
	if enName != "" {
		return enName, nil
	}
	companySetting := ctx.GetCompanySetting()
	defaultLanguage := companySetting.GetDefaultLanguage()
	res, err := utils.NewTranslateClient().Translate(ctx.GetContext(), []utils.TranslateItem{
		{
			Lang:    defaultLanguage,
			Content: locale.GetLocale(defaultLanguage),
		},
	})
	if err != nil {
		return "", errors.WithMessage(err, "翻译失败")
	}
	return res.Data[0].En, nil
}

// EditProductUnit 编辑产品单位
func (s *productSrv) EditProductUnit(ctx context.Context, editUnitReq req.ProductUnitEditReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editUnitReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	// 获取产品单位
	productUnit, err := productRepo.GetProductUnit(productRepo.WhereUuid(editUnitReq.Uuid), productRepo.WithMultiLanguageName())
	if err != nil {
		return errors.WithMessage(errors.New("单位不存在"), err.Error())
	}

	if productUnit.MultiLanguageNameUuid == 0 {
		return errors.New("单位名称不存在")
	}

	// 检查名称是否存在
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editUnitReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editUnitReq.Uuid,
		Source: constant.CheckNameSourceUnit,
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
			return errors.WithMessage(errors.New("保存名称多语言失败"), err.Error())
		}
		// 修改产品单位
		err = tx.Model(&model.ProductUnit{}).Where("uuid = ?", editUnitReq.Uuid).Updates(map[string]any{
			"name": editUnitReq.LocaleName.ToJson(),
		}).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存单位失败"), err.Error())
		}

		// 修改商品的单位UUID, 如果商品的单位UUID是当前单位，则修改为0
		err = tx.Model(&model.ProductPackage{}).Where("unit_uuid = ?", editUnitReq.Uuid).Updates(map[string]any{
			"unit_uuid": 0,
		}).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存关联商品失败"), err.Error())
		}
		// 修改商品的单位UUID
		if len(editUnitReq.ProductPackageUuids) > 0 {
			// 同步更新商品到erp
			if ctx.GetCompany().IsOpenErp() {
				productPackageRepo := repository.NewProductPackageRepo(tx)
				productPackages, errGetProductPackage := productPackageRepo.GetProductPackageList(
					repository.CommonRepo.WhereInUuids(editUnitReq.ProductPackageUuids),
					repository.CommonRepo.Preload(
						repository.WithPreload{
							Query: "ProductBoms",
						},
					),
				)
				if errGetProductPackage != nil {
					return errors.WithMessage(errGetProductPackage, "获取商品包失败")
				}
				for i := range productPackages {
					productPackage := productPackages[i]
					for _, productBom := range productPackage.ProductBoms {
						if productBom.IsFlavor() {
							if productPackage.IsPackage() {
								// 同步套餐到erp
								multiLanguageName := model.NewMultiLanguageName(productBom.Name)
								enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								erpSrv := erp.NewIErpSrv(s.dbm)
								_, errErp := erpSrv.AddPackage(ctx, req.PackageAddErpReq{
									ItemName: enName,
									StockUom: productUnit.ErpnextUom,
									ItemCode: productBom.ErpCode,
								})
								if errErp != nil {
									return errors.WithMessage(errErp, "同步商品到erp失败")
								}
							} else if productPackage.IsProduct() {
								// 同步商品到erp
								multiLanguageName := model.NewMultiLanguageName(productBom.Name)
								enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								erpSrv := erp.NewIErpSrv(s.dbm)
								_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
									ItemName: enName,
									StockUom: productUnit.ErpnextUom,
									ItemCode: productBom.ErpCode,
								})
								if errErp != nil {
									return errors.WithMessage(errErp, "同步商品到erp失败")
								}
							}
						}
					}
				}
			}
			// 修改商品的单位UUID
			err = tx.Model(&model.ProductPackage{}).Where("uuid in (?)", editUnitReq.ProductPackageUuids).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			}).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存关联商品失败"), err.Error())
			}
		}
		return nil
	})

	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	// 开启了ERP，并且是TTPOS站点，同步到ERPNext
	if company.IsOpenErp() && companySetting.IsTtposSite() {
		enName, err := s.getEnName(ctx, editUnitReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		err = erp.NewIErpSrv(s.dbm).SaveUom(ctx.GetContext(), req.SaveUomReq{
			SiteCode:          companySetting.ErpnextSiteCode,
			CompanyAbbr:       companySetting.ErpnextCompanyAbbr,
			Branch:            companySetting.ErpnextBranchName,
			UomName:           productUnit.ErpnextUom,
			AliasName:         enName,
			MustBeWholeNumber: true,
		})
		if err != nil {
			return errors.WithMessage(errors.New("同步单位到erp失败"), err.Error())
		}
	}

	return err
}

// DeleteProductUnit 删除产品单位
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
		return errors.WithMessage(errors.New("单位不存在"), err.Error())
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
			return errors.WithMessage(errors.New("删除单位失败"), err.Error())
		}
		// 删除多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productUnit.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除名称多语言失败"), err.Error())
		}
		// 重新排序
		productRepo := repository.NewProductRepo(tx)
		productUnits, _ := productRepo.GetProductUnitList()
		sorts := make(map[uint64]int)
		for i, productUnit := range productUnits {
			sorts[productUnit.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductUnit{}, sorts)
		if err != nil {
			return errors.WithMessage(errors.New("重新排序单位失败"), err.Error())
		}
		return nil
	})

	return err
}

// SortProductUnit 排序产品单位
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

	sorts := make(map[uint64]int)
	for _, item := range sortReq.List {
		sorts[item.Uuid] = item.Sort
	}
	err := productRepo.BatchUpdateSort(&model.ProductUnit{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序单位失败"), err.Error())
	}

	return nil
}

// GetProductSauceList 获取商品加料列表
func (s *productSrv) GetProductSauceList(ctx context.Context, sauceListReq req.ProductSauceListReq) (product_resp.ProductSauceListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	language := ctx.GetLanguage()

	productSauceList, total, err := productRepo.PaginateGetProductSauceList(sauceListReq.PageNo, sauceListReq.PageSize, productRepo.WithMultiLanguageName())
	if err != nil {
		return product_resp.ProductSauceListResp{}, errors.WithMessage(errors.New("获取加料列表失败"), err.Error())
	}
	productSauceListResp := make([]product_resp.ProductSauceItem, 0, len(productSauceList))
	for _, productSauce := range productSauceList {
		var productBomCardName dto.LocaleResponse
		if productSauce.ProductBomCardUuid > 0 {
			productBomCardName, err = repository.NewProductBomCardRepo(db).GetProductBomCardName(productSauce.ProductBomCardUuid)
			if err != nil {
				return product_resp.ProductSauceListResp{}, errors.WithMessage(errors.New("获取成本卡名称失败"), err.Error())
			}
		}
		productSauceListResp = append(productSauceListResp, product_resp.ProductSauceItem{
			Uuid:                productSauce.Uuid,
			Name:                productSauce.MultiLanguageName.GetNameByLang(language),
			Price:               productSauce.Price,
			Sort:                productSauce.Sort,
			ProductPackageCount: productSauce.ProductPackageCount,
			ProductBomCardUuid:  productSauce.ProductBomCardUuid,
			ProductBomCardName:  productBomCardName,
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

// GetProductSauce 获取商品加料详情
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
		return product_resp.ProductSauceDetail{}, errors.WithMessage(errors.New("获取加料详情失败"), err.Error())
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

// AddProductSauce 添加商品加料
func (s *productSrv) AddProductSauce(ctx context.Context, addReq req.ProductSauceAddReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceSauce,
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	productRepo := repository.NewProductRepo(db)
	// 检查商品是否存在
	productPackages, err := productRepo.GetProductPackageListByUuids(addReq.ProductPackageUuids)
	if err != nil {
		return errors.WithMessage(errors.New("商品不存在"), err.Error())
	}
	if len(productPackages) != len(addReq.ProductPackageUuids) {
		return errors.New("商品不存在")
	}

	// 获取当前最大的排序值
	var maxSort int
	db.Model(&model.ProductSauce{}).Scopes(repository.NotDeleted).Select("ifnull(max(sort), 0)").Scan(&maxSort)

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
			return errors.WithMessage(errors.New("保存名称多语言失败"), err.Error())
		}

		// 同步新增加料到erp
		erpCode := ""
		if ctx.GetCompany().IsOpenErp() {
			multiLanguageName := model.NewMultiLanguageName(addReq.LocaleName.ToJson())
			enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
			if err != nil {
				return errors.WithMessage(err, "翻译失败")
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
				ItemName: enName,
				StockUom: "Nos",
			})
			if errErp != nil {
				return errors.WithMessage(errErp, "同步新增加料到erp失败")
			}
			erpCode = itemInfo.ItemCode
		}

		// 保存商品加料
		productSauce := model.ProductSauce{
			Sort:                  maxSort + 1,
			Price:                 *addReq.Price,
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Name:                  addReq.LocaleName.ToJson(),
			ErpCode:               erpCode,
		}
		err = tx.Model(&model.ProductSauce{}).Create(&productSauce).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存加料失败"), err.Error())
		}

		var boms []model.ProductBom
		for _, productPackageUuid := range addReq.ProductPackageUuids {
			// 创建商品BOM
			boms = append(boms, model.ProductBom{
				Name:               addReq.LocaleName.ToJson(),
				Price:              *addReq.Price,
				StockNum:           99999999,
				IsOpenStock:        1,
				Status:             1,
				ProductPackageUuid: productPackageUuid,
				ProductSauceUuid:   productSauce.Uuid,
			})
		}
		err = repository.NewProductBomRepo(tx).CreateProductBoms(boms)
		if err != nil {
			return errors.WithMessage(errors.New("保存关联商品失败"), err.Error())
		}
		return nil
	})
	return err
}

// EditProductSauce 编辑商品加料
func (s *productSrv) EditProductSauce(ctx context.Context, editReq req.ProductSauceEditReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	// 判断名称是否存在
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: constant.CheckNameSourceSauce,
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
		return errors.WithMessage(errors.New("获取加料详情失败"), err.Error())
	}

	if productSauce.MultiLanguageNameUuid == 0 {
		return errors.New("加料名称不存在")
	}

	// 检查商品是否存在
	productPackages, err := productRepo.GetProductPackageListByUuids(editReq.ProductPackageUuids)
	if err != nil {
		return errors.WithMessage(errors.New("商品不存在"), err.Error())
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
			return errors.WithMessage(errors.New("保存名称多语言失败"), err.Error())
		}
		// 修改商品加料
		err = tx.Model(&model.ProductSauce{}).Where("uuid = ?", editReq.Uuid).Updates(map[string]any{
			"name":  name,
			"price": editReq.Price,
		}).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存加料失败"), err.Error())
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
						Price:              *editReq.Price,
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
				return errors.WithMessage(errors.New("保存关联商品失败"), err.Error())
			}
		} else {
			// 删除bom中sauce_uuid为当前商品加料的记录
			err := tx.Model(&model.ProductBom{}).Where("product_sauce_uuid = ?", editReq.Uuid).Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(errors.New("删除关联商品失败"), err.Error())
			}
		}
		return nil
	})
	return err
}

// DeleteProductSauce 删除商品加料
func (s *productSrv) DeleteProductSauce(ctx context.Context, deleteReq req.ProductSauceReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productSauce, err := productRepo.GetProductSauce(
		productRepo.WhereUuid(deleteReq.Uuid),
		productRepo.WithMultiLanguageName(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("获取加料详情失败"), err.Error())
	}
	if productSauce.MultiLanguageNameUuid == 0 {
		return errors.New("加料名称不存在")
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
			return errors.WithMessage(errors.New("删除加料失败"), err.Error())
		}
		// 删除多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productSauce.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除名称多语言失败"), err.Error())
		}

		// 重新排序
		productRepo := repository.NewProductRepo(tx)
		productSauceList, _ := productRepo.GetProductSauceList()
		sorts := make(map[uint64]int)
		for i, productSauce := range productSauceList {
			sorts[productSauce.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductSauce{}, sorts)
		if err != nil {
			return errors.WithMessage(errors.New("重新排序加料失败"), err.Error())
		}
		return nil
	})

	return err
}

// SortProductSauce 排序商品加料
func (s *productSrv) SortProductSauce(ctx context.Context, sortReq req.ProductSauceSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productSauceUuids := make([]uint64, 0, len(sortReq.List))
	for _, item := range sortReq.List {
		productSauceUuids = append(productSauceUuids, item.Uuid)
	}
	productSauceCount, _ := productRepo.GetProductSauceCount(productRepo.WhereUuidIn(productSauceUuids))
	if productSauceCount != int64(len(productSauceUuids)) {
		return errors.New("加料不存在")
	}

	sorts := make(map[uint64]int)
	for _, item := range sortReq.List {
		sorts[item.Uuid] = item.Sort
	}
	err := productRepo.BatchUpdateSort(&model.ProductSauce{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序加料失败"), err.Error())
	}
	return nil
}

// GetProductAttributeGroupList 获取商品属性分组列表
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
		return product_resp.ProductAttributeGroupListResp{}, errors.WithMessage(errors.New("获取属性组列表失败"), err.Error())
	}
	productAttributeGroupListResp := make([]product_resp.ProductAttributeGroupItem, 0, len(productAttributeGroupList))
	for _, productAttributeGroup := range productAttributeGroupList {
		attributeNames := make([]string, 0, len(productAttributeGroup.ProductAttributes))

		attributeList := make([]product_resp.ProductAttributeGroupAttributeItem, 0, len(productAttributeGroup.ProductAttributes))
		for _, productAttribute := range productAttributeGroup.ProductAttributes {
			attributeNames = append(attributeNames, productAttribute.MultiLanguageName.GetNameByLang(language))
			attributeList = append(attributeList, product_resp.ProductAttributeGroupAttributeItem{
				Uuid:       productAttribute.Uuid,
				LocaleName: productAttribute.MultiLanguageName.GetNames(),
			})
		}
		productAttributeGroupListResp = append(productAttributeGroupListResp, product_resp.ProductAttributeGroupItem{
			Uuid:          productAttributeGroup.Uuid,
			Name:          productAttributeGroup.MultiLanguageName.GetNameByLang(language),
			AttributeName: strings.Join(attributeNames, "、"),
			Attributes:    attributeList,
			Sort:          productAttributeGroup.Sort,
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

// GetProductAttributeGroup 获取商品属性分组详情
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
		return product_resp.ProductAttributeGroupDetail{}, errors.WithMessage(errors.New("获取属性组详情失败"), err.Error())
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
			Sort: productAttribute.Sort,
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
	language := ctx.GetLanguage()

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
		return product_resp.ProductFlavorListResp{}, errors.WithMessage(errors.New("获取规格列表失败"), err.Error())
	}
	productFlavorListResp := make([]product_resp.ProductFlavorItemResp, 0, len(productFlavorList))
	for _, productFlavor := range productFlavorList {
		productFlavorListResp = append(productFlavorListResp, product_resp.ProductFlavorItemResp{
			Uuid:                productFlavor.Uuid,
			Name:                productFlavor.MultiLanguageName.GetNameByLang(language),
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
		return product_resp.ProductFlavorDetailResp{}, errors.WithMessage(errors.New("获取规格详情失败"), err.Error())
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

// AddProductAttributeGroup 添加商品属性组
func (s *productSrv) AddProductAttributeGroup(ctx context.Context, addReq req.ProductAttributeGroupAddReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("属性组名称不能为空")
	}
	for _, productAttribute := range addReq.ProductAttributes {
		if !productAttribute.LocaleName.CheckRequiredLocale(storeLanguages) {
			return errors.New("属性值名称不能为空")
		}
	}
	// 检查名称是否存在
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceAttributeGroup,
		Names:  names,
	})
	if exists {
		return errors.New("属性组名称已存在")
	}
	for _, productAttribute := range addReq.ProductAttributes {
		names := checkService.MakeCheckNameList(ctx, productAttribute.LocaleName)
		exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
			Source: constant.CheckNameSourceAttribute,
			Names:  names,
		})
		if exists {
			return errors.New("属性值名称已存在")
		}
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	// 遍历req.ProductAttributes，判断商品是否存在
	productPackageUuids := []uint64{}
	for _, productAttribute := range addReq.ProductAttributes {
		productPackageUuids = append(productPackageUuids, productAttribute.ProductPackageUuids...)
	}
	// 去重
	productPackageUuids = slices.Compact(productPackageUuids)
	if len(productPackageUuids) > 0 {
		// 判断商品是否存在
		productPackageList, err := productRepo.GetProductPackageListByUuids(productPackageUuids)
		if err != nil {
			return errors.WithMessage(errors.New("商品不存在"), err.Error())
		}
		if len(productPackageList) != len(productPackageUuids) {
			return errors.New("商品不存在")
		}
	}

	uuidLocaleMap := make(map[uint64]dto.LocaleResponse)

	// 获取当前最大的排序值
	var maxSort int
	db.Model(&model.ProductAttributeGroup{}).Scopes(repository.NotDeleted).Select("ifnull(max(sort), 0)").Scan(&maxSort)

	productAttributeGroup := model.ProductAttributeGroup{
		Name: addReq.LocaleName.ToJson(),
		Sort: maxSort + 1,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		// 保存多语言
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
			return errors.WithMessage(errors.New("保存属性组名称多语言失败"), err.Error())
		}
		// 添加商品属性分组
		productAttributeGroup.MultiLanguageNameUuid = multiLanguageName.Uuid
		err = tx.Model(&model.ProductAttributeGroup{}).Create(&productAttributeGroup).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存属性组失败"), err.Error())
		}

		// 每个商品关联多个属性
		productPackageMapAttributeUuids := make(map[uint64][]uint64)

		// 遍历req.ProductAttributes，添加商品属性
		for i, productAttribute := range addReq.ProductAttributes {
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
				return errors.WithMessage(errors.New("保存属性值名称多语言失败"), err.Error())
			}
			// 添加商品属性
			productAttributeModel := model.ProductAttribute{
				Name:                  productAttribute.LocaleName.ToJson(),
				MultiLanguageNameUuid: multiLanguageName.Uuid,
				AttributeGroupUuid:    productAttributeGroup.Uuid,
				Sort:                  i + 1,
			}
			err = tx.Model(&model.ProductAttribute{}).Create(&productAttributeModel).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存属性失败"), err.Error())
			}
			uuidLocaleMap[productAttributeModel.Uuid] = productAttribute.LocaleName
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
				return errors.WithMessage(errors.New("保存属性组关联商品失败"), err.Error())
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
				return errors.WithMessage(errors.New("保存属性关联商品失败"), err.Error())
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("添加属性组失败"), err.Error())
	}

	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	// 开启了ERP，并且是TTPOS站点，同步到ERPNext
	if company.IsOpenErp() && companySetting.IsTtposSite() {
		attributeValueList := []req.SaveAttributeValueReq{}
		uuidErpValueNameMap := make(map[uint64]string)
		for attributeUuid, locale := range uuidLocaleMap {
			enValueName, err := s.getEnName(ctx, locale)
			if err != nil {
				return errors.WithMessage(errors.New("翻译失败"), err.Error())
			}
			erpValueName := company.Name + "-" + enValueName
			uuidErpValueNameMap[attributeUuid] = erpValueName
			attributeValueList = append(attributeValueList, req.SaveAttributeValueReq{
				AttributeValue: erpValueName,
				Abbr:           enValueName,
			})
		}
		enGroupName, err := s.getEnName(ctx, addReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		erpGroupName := company.Name + "-" + enGroupName
		err = erp.NewIErpSrv(s.dbm).SaveAttribute(ctx.GetContext(), req.SaveAttributeReq{
			SiteCode:           companySetting.ErpnextSiteCode,
			CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
			Branch:             companySetting.ErpnextBranchName,
			AttributeName:      erpGroupName,
			AliasName:          enGroupName,
			AttributeValueList: attributeValueList,
		})
		if err != nil {
			return errors.WithMessage(errors.New("同步属性组到erp失败"), err.Error())
		}
		// 保存属性组 erpnext_attribute_group_name
		err = db.Model(&model.ProductAttributeGroup{}).Where("uuid = ?", productAttributeGroup.Uuid).Update("erpnext_attribute_group_name", erpGroupName).Error
		if err != nil {
			return errors.WithMessage(errors.New("保存erp属性组失败"), err.Error())
		}
		// 保存属性值 erpnext_attribute_value
		for attributeUuid, erpValueName := range uuidErpValueNameMap {
			err = db.Model(&model.ProductAttribute{}).Where("uuid = ?", attributeUuid).Update("erpnext_attribute_value", erpValueName).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存erp属性值失败"), err.Error())
			}
		}
	}

	return nil
}

// AddProductFlavor 添加商品规格
func (s *productSrv) AddProductFlavor(ctx context.Context, addReq req.ProductFlavorAddReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceFlavor,
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	maxSort, err := productRepo.GetProductFlavorMaxSort(
		commonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("获取最大排序失败"), err.Error())
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

				// 同步新增商品到erp
				erpCode := ""
				if ctx.GetCompany().IsOpenErp() {
					multiLanguageName := model.NewMultiLanguageName(addReq.LocaleName.ToJson())
					enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
					if errGetUnit != nil {
						return errors.WithMessage(errGetUnit, "获取商品单位失败")
					}

					erpSrv := erp.NewIErpSrv(s.dbm)
					itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
						ItemName:          enName,
						StockUom:          productUnit.ErpnextUom,
						TemplateItemCode:  productPackage.ErpCode,
						ItemSpecification: enName,
					})
					if errErp != nil {
						return errors.WithMessage(errErp, "同步商品到erp失败")
					}
					erpCode = itemInfo.ItemCode
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
					ErpCode:            erpCode,
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

// EditProductAttributeGroup 编辑商品属性组
func (s *productSrv) EditProductAttributeGroup(ctx context.Context, editReq req.ProductAttributeGroupEditReq) error {
	var relatedProductUuid bool
	// 检查多语言
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("属性组名称不能为空")
	}
	for _, productAttribute := range editReq.ProductAttributes {
		if !relatedProductUuid && len(productAttribute.ProductPackageUuids) > 0 {
			relatedProductUuid = true
		}
		if !productAttribute.LocaleName.CheckRequiredLocale(storeLanguages) {
			return errors.New("属性值名称不能为空")
		}
	}

	// 检查名称是否已存在
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: constant.CheckNameSourceAttributeGroup,
		Names:  names,
	})
	if exists {
		return errors.New("属性组名称已存在")
	}
	for _, attribute := range editReq.ProductAttributes {
		names := checkService.MakeCheckNameList(ctx, attribute.LocaleName)
		exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
			Uuid:   attribute.Uuid,
			Source: constant.CheckNameSourceAttribute,
			Names:  names,
		})
		if exists {
			return errors.New("属性值名称已存在")
		}
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	// 检查属性组是否存在
	attributeGroup, err := productRepo.GetProductAttributeGroup(
		productRepo.WhereUuid(editReq.Uuid),
		productRepo.WithProductAttributes(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("属性组不存在"), err.Error())
	}

	// 检查传递的属性值是否存在
	var attributeUuids []uint64
	for _, attribute := range editReq.ProductAttributes {
		if attribute.Uuid != 0 {
			attributeUuids = append(attributeUuids, attribute.Uuid)
		}
	}
	attributes, err := productRepo.GetProductAttributes(
		productRepo.WhereAttributeGroupUuid(attributeGroup.Uuid),
		productRepo.WhereUuidIn(attributeUuids),
	)
	if err != nil {
		return errors.WithMessage(errors.New("属性值不存在"), err.Error())
	}
	if len(attributes) != len(attributeUuids) {
		return errors.New("属性值参数错误")
	}

	// 要删掉的属性值
	var deletingAttributeUuids []uint64
	for _, attribute := range attributeGroup.ProductAttributes {
		if !slices.Contains(attributeUuids, attribute.Uuid) {
			deletingAttributeUuids = append(deletingAttributeUuids, attribute.Uuid)
		}
	}

	// 检查所有属性关联的商品包是否存在
	productPackageUuids := []uint64{}
	for _, productAttribute := range editReq.ProductAttributes {
		productPackageUuids = append(productPackageUuids, productAttribute.ProductPackageUuids...)
	}
	// 去重，判断商品是否存在
	productPackageUuids = slice.Unique(productPackageUuids)
	if len(productPackageUuids) > 0 {
		productPackageList, err := productRepo.GetProductPackageListByUuids(productPackageUuids)
		if err != nil {
			return errors.WithMessage(errors.New("商品不存在"), err.Error())
		}
		if len(productPackageList) != len(productPackageUuids) {
			return errors.New("关联商品参数错误")
		}
	}

	// 原商品包属性组、商品包属性
	productPackageAttributeGroups, err := productRepo.GetProductPackageAttributeGroups(
		productRepo.WhereProductAttributeGroupUuid(attributeGroup.Uuid),
		productRepo.WithProductPackageAttributes(),
	)

	type ProductPackageInfo struct {
		ProductPackageAttributeGroupUuid uint64
		AttributeUuids                   []uint64
	}

	// 整理关系，每个商品包和属性组的关系
	productPackageInfoMap := make(map[uint64]ProductPackageInfo)
	for _, productPackageAttributeGroup := range productPackageAttributeGroups {
		oldAttributeUuids := make([]uint64, 0)
		for _, productPackageAttribute := range productPackageAttributeGroup.ProductPackageAttributes {
			oldAttributeUuids = append(oldAttributeUuids, productPackageAttribute.AttributeUuid)
		}
		productPackageInfoMap[productPackageAttributeGroup.ProductPackageUuid] = ProductPackageInfo{
			ProductPackageAttributeGroupUuid: productPackageAttributeGroup.Uuid,
			AttributeUuids:                   oldAttributeUuids,
		}
	}

	// 属性值uuid和多语言uuid映射
	uuidAttributeMap := make(map[uint64]model.ProductAttribute)
	for _, attribute := range attributes {
		uuidAttributeMap[attribute.Uuid] = attribute
	}

	uuidLocaleMap := make(map[uint64]dto.LocaleResponse)

	err = db.Transaction(func(tx *gorm.DB) error {

		// 删除属性值 product_attribute.Uuid in deletingAttributeUuids
		// 删除商品包属性 product_package_attribute.attribute_uuid in deletingAttributeUuids
		if len(deletingAttributeUuids) > 0 {
			err := tx.Model(&model.ProductAttribute{}).Where("uuid IN (?)", deletingAttributeUuids).Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(errors.New("删除属性值失败"), err.Error())
			}
			err = tx.Model(&model.ProductPackageAttribute{}).Where("attribute_uuid IN (?)", deletingAttributeUuids).Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(errors.New("删除属性值关联商品失败"), err.Error())
			}
		}

		// 删除 product_package_attribute_group 和 product_package_attribute 的数据
		if len(productPackageAttributeGroups) > 0 {
			var deletingProductPackageAttributeGroupUuids []uint64
			if len(productPackageUuids) == 0 { // 取消了所有商品和属性组的关系
				for _, productPackageAttributeGroup := range productPackageAttributeGroups {
					deletingProductPackageAttributeGroupUuids = append(deletingProductPackageAttributeGroupUuids, productPackageAttributeGroup.Uuid)
				}
			} else { // 可能取消了部分商品和属性组的关系
				for _, productPackageAttributeGroup := range productPackageAttributeGroups {
					if !slices.Contains(productPackageUuids, productPackageAttributeGroup.ProductPackageUuid) {
						deletingProductPackageAttributeGroupUuids = append(deletingProductPackageAttributeGroupUuids, productPackageAttributeGroup.Uuid)
					}
				}
			}
			if len(deletingProductPackageAttributeGroupUuids) > 0 {
				err := tx.Model(&model.ProductPackageAttributeGroup{}).Where("uuid IN (?)", deletingProductPackageAttributeGroupUuids).Update("delete_time", time.Now().Unix()).Error
				if err != nil {
					return errors.WithMessage(errors.New("删除属性组关联商品失败"), err.Error())
				}
				err = tx.Model(&model.ProductPackageAttribute{}).Where("product_package_attribute_group_uuid IN (?)", deletingProductPackageAttributeGroupUuids).Update("delete_time", time.Now().Unix()).Error
				if err != nil {
					return errors.WithMessage(errors.New("删除属性关联商品失败"), err.Error())
				}
			}
		}

		// 更新属性组多语言
		err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", attributeGroup.MultiLanguageNameUuid).Updates(map[string]any{
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
			return errors.WithMessage(errors.New("保存属性组名称多语言失败"), err.Error())
		}
		for k, productAttribute := range editReq.ProductAttributes {
			if productAttribute.Uuid != 0 { // 更新属性值多语言
				uuidLocaleMap[productAttribute.Uuid] = productAttribute.LocaleName
				err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", uuidAttributeMap[productAttribute.Uuid].MultiLanguageNameUuid).Updates(map[string]any{
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
					return errors.WithMessage(errors.New("保存属性值名称多语言失败"), err.Error())
				}
			} else { // 新增属性值多语言
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
					return errors.WithMessage(errors.New("添加属性值名称多语言失败"), err.Error())
				}
				productAttributeModel := model.ProductAttribute{
					Name:                  productAttribute.LocaleName.ToJson(),
					MultiLanguageNameUuid: multiLanguageName.Uuid,
					AttributeGroupUuid:    attributeGroup.Uuid,
					Sort:                  productAttribute.Sort,
				}
				err = tx.Model(&model.ProductAttribute{}).Create(&productAttributeModel).Error
				if err != nil {
					return errors.WithMessage(errors.New("保存属性值失败"), err.Error())
				}
				editReq.ProductAttributes[k].Uuid = productAttributeModel.Uuid
				uuidLocaleMap[productAttributeModel.Uuid] = productAttribute.LocaleName
			}
		}

		if relatedProductUuid {
			productPackageUuidToAttributeUuids := s.getAttributeUuidListByProductPackageUuid(editReq.ProductAttributes)
			var newProductPackageAttributes []model.ProductPackageAttribute
			for productPackageUuid, attributeUuids := range productPackageUuidToAttributeUuids {
				if productPackageInfo, ok := productPackageInfoMap[productPackageUuid]; !ok { // 新增的关联商品包
					newProductPackageAttributeGroup := model.ProductPackageAttributeGroup{
						ProductPackageUuid:        productPackageUuid,
						ProductAttributeGroupUuid: attributeGroup.Uuid,
					}
					err := tx.Model(&model.ProductPackageAttributeGroup{}).Create(&newProductPackageAttributeGroup).Error
					if err != nil {
						return errors.WithMessage(errors.New("添加属性组关联商品失败"), err.Error())
					}
					for _, attributeUuid := range attributeUuids {
						newProductPackageAttributes = append(newProductPackageAttributes, model.ProductPackageAttribute{
							ProductPackageAttributeGroupUuid: newProductPackageAttributeGroup.Uuid,
							AttributeUuid:                    attributeUuid,
						})
					}
				} else { // 已有的关联商品包，可能关联了新的属性
					addingAttributeUuids := slice.Difference(attributeUuids, productPackageInfo.AttributeUuids)
					if len(addingAttributeUuids) > 0 {
						for _, attributeUuid := range addingAttributeUuids {
							newProductPackageAttributes = append(newProductPackageAttributes, model.ProductPackageAttribute{
								ProductPackageAttributeGroupUuid: productPackageInfo.ProductPackageAttributeGroupUuid,
								AttributeUuid:                    attributeUuid,
							})
						}
					}
				}
			}
			if len(newProductPackageAttributes) > 0 {
				err := tx.Model(&model.ProductPackageAttribute{}).Create(&newProductPackageAttributes).Error
				if err != nil {
					return errors.WithMessage(errors.New("添加属性关联商品失败"), err.Error())
				}
			}
		}

		productRepo := repository.NewProductRepo(tx)
		productAttributes, _ := productRepo.GetProductAttributes(productRepo.WhereAttributeGroupUuid(attributeGroup.Uuid))
		sorts := make(map[uint64]int)
		for i, productAttribute := range productAttributes {
			sorts[productAttribute.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductAttribute{}, sorts)
		if err != nil {
			return errors.WithMessage(err, "重新排序商品属性失败")
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("保存属性组失败"), err.Error())
	}

	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	// 开启了ERP，并且是TTPOS站点，同步到ERPNext
	if company.IsOpenErp() && companySetting.IsTtposSite() {
		var valueList []req.SaveAttributeValueReq
		uuidErpValueNameMap := make(map[uint64]string)
		// 构建 valueList
		for uuid, locale := range uuidLocaleMap {
			enValueName, err := s.getEnName(ctx, locale)
			if err != nil {
				return errors.WithMessage(errors.New("翻译失败"), err.Error())
			}
			var erpValueName string
			if v, ok := uuidAttributeMap[uuid]; !ok { // 新增属性值
				erpValueName := company.Name + "-" + enValueName
				uuidErpValueNameMap[uuid] = erpValueName
			} else { // 已存在属性值，使用erpnext_attribute_value
				erpValueName = v.ErpnextAttributeValue
			}
			valueList = append(valueList, req.SaveAttributeValueReq{
				AttributeValue: erpValueName,
				Abbr:           enValueName, // 总是获取最新别名
			})
		}
		// 新属性组别名
		enGroupName, err := s.getEnName(ctx, editReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		err = erp.NewIErpSrv(s.dbm).SaveAttribute(ctx.GetContext(), req.SaveAttributeReq{
			SiteCode:           companySetting.ErpnextSiteCode,
			CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
			Branch:             companySetting.ErpnextBranchName,
			AttributeName:      attributeGroup.ErpnextAttributeGroupName,
			AliasName:          enGroupName,
			AttributeValueList: valueList,
		})
		if err != nil {
			return errors.WithMessage(errors.New("同步属性值到erp失败"), err.Error())
		}
		for uuid, erpValueName := range uuidErpValueNameMap {
			err = db.Model(&model.ProductAttribute{}).Where("uuid = ?", uuid).Update("erpnext_attribute_value", erpValueName).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存erp属性值失败"), err.Error())
			}
		}
	}

	return err
}

func (s *productSrv) getAttributeUuidListByProductPackageUuid(productAttributes []req.ProductAttributeGroupEditProductAttributeReq) map[uint64][]uint64 {
	// 使用map存储每个ProductPackageUuid对应的AttributeUuid集合（用于去重）
	productPackageUuidToAttributeUuids := make(map[uint64]map[uint64]bool)

	// 遍历所有productAttributes
	for _, productAttribute := range productAttributes {
		// 遍历每个productAttribute的ProductPackageUuids
		for _, productPackageUuid := range productAttribute.ProductPackageUuids {
			// 如果该ProductPackageUuid还没有记录，则初始化
			if productPackageUuidToAttributeUuids[productPackageUuid] == nil {
				productPackageUuidToAttributeUuids[productPackageUuid] = make(map[uint64]bool)
			}
			// 将AttributeUuid添加到对应ProductPackageUuid的集合中（使用map去重）
			productPackageUuidToAttributeUuids[productPackageUuid][productAttribute.Uuid] = true
		}
	}

	// 将map转换为切片，得到每个ProductPackageUuid关联的AttributeUuid列表
	result := make(map[uint64][]uint64)
	for productPackageUuid, attributeUuids := range productPackageUuidToAttributeUuids {
		uuidList := make([]uint64, 0, len(attributeUuids))
		for uuid := range attributeUuids {
			uuidList = append(uuidList, uuid)
		}
		result[productPackageUuid] = uuidList
	}

	return result
}

// EditProductFlavor 编辑商品规格
func (s *productSrv) EditProductFlavor(ctx context.Context, editReq req.ProductFlavorEditReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, editReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   editReq.Uuid,
		Source: constant.CheckNameSourceFlavor,
		Names:  names,
	})
	if exists {
		return errors.New("名称已存在")
	}

	language := ctx.GetLanguage()
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(db)

	productNames := []string{}
	packageNames := []string{}
	for _, item := range editReq.List {
		if item.IsDelete {
			productBom, _ := productRepo.GetProductBom(
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByUuid(item.BomUuid),
				productRepo.WithProductPackage(commonRepo.WhereBySoftDelete()),
				productRepo.WithProductPackageMultiLanguageName(commonRepo.WhereBySoftDelete()),
				productRepo.WithProductPackageProductBom(
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereGtProductFlavorUuid(0),
				),
			)
			if productBom.ID != 0 && productBom.ProductPackage.Uuid != 0 {
				productPackage := productBom.ProductPackage
				if len(productPackage.ProductBoms) == 1 {
					if !slices.Contains(productNames, productPackage.MultiLanguageName.GetNameByLang(language)) {
						productNames = append(productNames, productPackage.MultiLanguageName.GetNameByLang(language))
					}
				}
				packageItems, _ := productPackageGroupRepo.GetProductPackageGroupItems(
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereByProductBomUuid(productBom.Uuid),
					productPackageGroupRepo.WithProductPackageGroup(
						commonRepo.WhereBySoftDelete(),
					),
					productPackageGroupRepo.WithProductPackageGroupProduct(
						commonRepo.WhereBySoftDelete(),
					),
					productPackageGroupRepo.WithProductPackageGroupProductMultiLanguageName(
						commonRepo.WhereBySoftDelete(),
					),
				)
				for _, packageItem := range packageItems {
					if packageItem.ProductPackageGroup != nil {
						if !slices.Contains(packageNames, packageItem.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(language)) {
							packageNames = append(packageNames, packageItem.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(language))
						}
					}
				}

			}
		}
	}

	if len(productNames) > 0 {
		return errors.NewWithReplace("%s只有一个规格，不可删除", []string{strings.Join(productNames, "、")})
	}
	if len(packageNames) > 0 {
		return errors.NewWithReplace("套餐【%s】已使用此规格，不可删除", []string{strings.Join(packageNames, "、")})
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
					// 同步新增商品到erp
					erpCode := ""
					if ctx.GetCompany().IsOpenErp() {
						multiLanguageName := model.NewMultiLanguageName(editReq.LocaleName.ToJson())
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}

						erpSrv := erp.NewIErpSrv(s.dbm)
						itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
							ItemName:          enName,
							StockUom:          productUnit.ErpnextUom,
							TemplateItemCode:  productPackage.ErpCode,
							ItemSpecification: enName,
						})
						if errErp != nil {
							return errors.WithMessage(errErp, "同步商品到erp失败")
						}
						erpCode = itemInfo.ItemCode
					}
					// 新增商品BOM
					err := tx.Create(&model.ProductBom{
						Price:              item.Price,
						Name:               editReq.LocaleName.ToJson(),
						ErpCode:            erpCode,
						Status:             int(productPackage.Status),
						ProductFlavorUuid:  flavor.Uuid,
						ProductPackageUuid: item.Uuid,
					}).Error
					if err != nil {
						return err
					}
				} else {
					// 同步更新商品到erp
					if ctx.GetCompany().IsOpenErp() {
						multiLanguageName := model.NewMultiLanguageName(editReq.LocaleName.ToJson())
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						productBomRepo := repository.NewProductBomRepo(tx)
						productBom, errGetBom := productBomRepo.GetProductBom(commonRepo.WhereByUuid(item.BomUuid))
						if errGetBom != nil {
							return errors.WithMessage(errGetBom, "获取商品bom失败")
						}

						erpSrv := erp.NewIErpSrv(s.dbm)
						_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
							ItemName:          enName,
							StockUom:          productUnit.ErpnextUom,
							ItemCode:          productBom.ErpCode,
							TemplateItemCode:  productPackage.ErpCode,
							ItemSpecification: enName,
						})
						if errErp != nil {
							return errors.WithMessage(errErp, "同步商品到erp失败")
						}
					}
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

// DeleteProductAttribute 删除商品属性
func (s *productSrv) DeleteProductAttribute(ctx context.Context, req req.ProductAttributeDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productAttribute, err := productRepo.GetProductAttribute(
		productRepo.WhereUuid(req.Uuid),
		productRepo.WithProductPackageAttributes(),
	)
	if err != nil || productAttribute.ID == 0 {
		return errors.WithMessage(errors.New("属性值不存在"), err.Error())
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除商品属性
		err = tx.Model(&model.ProductAttribute{}).Where("uuid = ?", productAttribute.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除属性值失败"), err.Error())
		}

		// 删除属性值多语言
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productAttribute.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除名称多语言失败"), err.Error())
		}

		err = tx.Model(&model.ProductPackageAttribute{}).Where("attribute_uuid = ?", productAttribute.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除属性关联商品失败"), err.Error())
		}

		// 属性关联的商品，标记删除
		var productPackageAttributeGroupUuids []uint64
		for _, productPackageAttribute := range productAttribute.ProductPackageAttributes {
			if !slices.Contains(productPackageAttributeGroupUuids, productPackageAttribute.ProductPackageAttributeGroupUuid) {
				productPackageAttributeGroupUuids = append(productPackageAttributeGroupUuids, productPackageAttribute.ProductPackageAttributeGroupUuid)
			}
		}
		if len(productPackageAttributeGroupUuids) > 0 {
			productPackageAttributeRepo := repository.NewProductPackageAttributeRepo(tx)
			relatedAttributeUuidCountList, err := productPackageAttributeRepo.GetProductPackageAttributeGroupCount(productAttribute.Uuid)
			if err != nil {
				return errors.WithMessage(errors.New("获取属性值关联商品失败"), err.Error())
			}
			for _, relatedAttributeUuidCount := range relatedAttributeUuidCountList {
				if relatedAttributeUuidCount.RelatedAttributeUuidCount == 0 {
					err = tx.Model(&model.ProductPackageAttributeGroup{}).Where("uuid = ?", relatedAttributeUuidCount.ProductPackageAttributeGroupUuid).Update("delete_time", time.Now().Unix()).Error
					if err != nil {
						return errors.WithMessage(errors.New("删除属性值关联商品失败"), err.Error())
					}
				}
			}
		}

		// 重新排序属性值
		productRepo := repository.NewProductRepo(tx)
		productAttributes, _ := productRepo.GetProductAttributes(productRepo.WhereAttributeGroupUuid(productAttribute.AttributeGroupUuid))
		sorts := make(map[uint64]int)
		for i, productAttribute := range productAttributes {
			sorts[productAttribute.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductAttribute{}, sorts)
		if err != nil {
			return errors.WithMessage(err, "重新排序商品属性值失败")
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("删除属性值失败"), err.Error())
	}

	return nil
}

// DeleteProductAttributeGroup 删除商品属性组
func (s *productSrv) DeleteProductAttributeGroup(ctx context.Context, req req.ProductAttributeGroupReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	productAttributeGroup, err := productRepo.GetProductAttributeGroup(
		productRepo.WhereUuid(req.Uuid),
		productRepo.WithProductAttributes(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("属性组不存在"), err.Error())
	}

	if len(productAttributeGroup.ProductAttributes) > 0 {
		return errors.New("该属性组有属性值，不可删除")
	}

	err = db.Transaction(func(tx *gorm.DB) error {

		// 删除属性组
		err = tx.Model(&model.ProductAttributeGroup{}).Where("uuid = ?", productAttributeGroup.Uuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除属性组失败"), err.Error())
		}

		// 删除属性组多语言
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productAttributeGroup.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除名称多语言失败"), err.Error())
		}

		// 重新排序属性组
		productRepo := repository.NewProductRepo(tx)
		productAttributeGroupList, _ := productRepo.GetProductAttributeGroups()
		sorts := make(map[uint64]int)
		for i, productAttributeGroup := range productAttributeGroupList {
			sorts[productAttributeGroup.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductAttributeGroup{}, sorts)
		if err != nil {
			return errors.WithMessage(errors.New("重新排序属性值失败"), err.Error())
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("删除属性组失败"), err.Error())
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
		return errors.WithMessage(errors.New("获取规格详情失败"), err.Error())
	}

	// 判断商品规格是否关联了商品
	productBomCount, _ := productRepo.GetProductBomCount(
		commonRepo.WhereByProductFlavorUuid(productFlavor.Uuid),
		commonRepo.WhereBySoftDelete(),
	)
	if productBomCount > 0 {
		return errors.New("该规格已经关联了商品，不可删除")
	}

	db.Transaction(func(tx *gorm.DB) error {
		// 软删除商品规格
		err = tx.Model(&model.ProductFlavor{}).Where("uuid = ?", productFlavor.Uuid).Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除规格失败"), err.Error())
		}

		// 删除规格名称多语言
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productFlavor.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix()).Error
		if err != nil {
			return errors.WithMessage(errors.New("删除名称多语言失败"), err.Error())
		}

		// 重新排序
		productRepo := repository.NewProductRepo(tx)
		productFlavors, _ := productRepo.GetProductFlavorList()
		sorts := make(map[uint64]int)
		for i, productFlavor := range productFlavors {
			sorts[productFlavor.Uuid] = i + 1
		}
		err = productRepo.BatchUpdateSort(&model.ProductFlavor{}, sorts)
		if err != nil {
			return errors.WithMessage(errors.New("重新排序规格失败"), err.Error())
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("删除规格失败"), err.Error())
	}

	return nil
}

// SortProductAttributeGroup 排序商品属性组
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

	sorts := make(map[uint64]int)
	for _, item := range req.List {
		sorts[item.Uuid] = item.Sort
	}
	err = productRepo.BatchUpdateSort(&model.ProductAttributeGroup{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序属性组失败"), err.Error())
	}
	return nil
}

// SortProductAttribute 排序商品属性
func (s *productSrv) SortProductAttribute(ctx context.Context, req req.ProductAttributeSortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)
	productAttributeUuids := []uint64{}
	for _, productAttribute := range req.List {
		productAttributeUuids = append(productAttributeUuids, productAttribute.Uuid)
	}
	productAttributes, err := productRepo.GetProductAttributes(
		productRepo.WhereAttributeGroupUuid(req.ProductAttributeGroupUuid),
		productRepo.WhereUuidIn(productAttributeUuids),
	)
	if err != nil {
		return errors.WithMessage(errors.New("获取属性值失败"), err.Error())
	}
	if len(productAttributes) != len(productAttributeUuids) {
		return errors.New("属性值不存在")
	}

	sorts := make(map[uint64]int)
	for _, item := range req.List {
		sorts[item.Uuid] = item.Sort
	}
	err = productRepo.BatchUpdateSort(&model.ProductAttribute{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序属性值失败"), err.Error())
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

	sorts := make(map[uint64]int)
	for _, item := range req.List {
		if item.Sort == 0 {
			return errors.New("排序不能为0")
		}
		sorts[item.Uuid] = item.Sort
	}
	err := productRepo.BatchUpdateSort(&model.ProductFlavor{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序规格失败"), err.Error())
	}
	return nil
}

// ImportProductList 导入商品列表
func (s *productSrv) ImportProductList(ctx context.Context, req req.ProductImportListReq) (product_resp.ProductImportResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取公司设置
	companySetting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return product_resp.ProductImportResp{}, err
	}
	// 获取语言
	language := ctx.GetLanguage()

	// 初始化返回
	var productImportResp product_resp.ProductImportResp
	productImportResp.List = make([]product_resp.ProductImportListItem, 0, len(req.List))
	for _, item := range req.List {
		// 复制商品信息
		products := product_resp.ProductImportListItem{}
		copier.Copy(&products, item)

		// 获取分类ID
		categoryUuid, err := repository.NewCategoryRepositoryService(db).GetCategoryUuidByNameOptimized(item.CategoryName)
		if err != nil {
			return product_resp.ProductImportResp{}, err
		}
		// 获取单位ID
		unitUuid, err := base.NewProductUnitRepo(db).GetProductUnitUuidByNameOptimized(item.ProductUnit)
		if err != nil {
			return product_resp.ProductImportResp{}, err
		}
		// 获取规格ID
		skuUuid, err := base.NewProductFlavorRepo(db).GetProductFlavorUuidByNameOptimized(item.SkuName)
		if err != nil {
			return product_resp.ProductImportResp{}, err
		}
		// 获取堂食税类ID
		taxUuid, err := repository.NewTaxRepo(db).GetTaxCategoryUuidByNameOptimized(item.ProductRatingTaxType)
		if err != nil {
			return product_resp.ProductImportResp{}, err
		}
		// 获取外带税类ID
		takeoutTaxUuid, err := repository.NewTaxRepo(db).GetTaxCategoryUuidByNameOptimized(item.ProductTakeoutTaxType)
		if err != nil {
			return product_resp.ProductImportResp{}, err
		}
		// 设置分类ID、单位ID、规格ID、堂食税类ID、外带税类ID
		products.CategoryUuid = categoryUuid
		products.UnitUuid = unitUuid
		products.SkuUuid = skuUuid
		products.DineTaxUuid = taxUuid
		products.TakeoutTaxUuid = takeoutTaxUuid
		// 处理数量计算方法
		// 按小数计价，不在助手、平板、扫码端显示
		if item.NumType == 2 && (products.IsShowTablet || products.IsShowAssistant || products.IsShowH5 || products.IsShowDelivery) {
			return product_resp.ProductImportResp{}, errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "按小数计价只能显示到收银机和厨显"))
		}
		// 未配置外送渠道，无法选择在外送显示
		if companySetting.DeliveryStatus != 1 && products.IsShowDelivery {
			return product_resp.ProductImportResp{}, errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "未配置外送渠道，无法选择在外送显示"))
		}
		// 处理商品名称
		products.LocaleName = item.LocaleName
		products.IsShowCashier = strings.Contains(item.Shows, "1")
		products.IsShowTablet = strings.Contains(item.Shows, "2")
		products.IsShowKitchen = strings.Contains(item.Shows, "3")
		products.IsShowAssistant = strings.Contains(item.Shows, "4")
		products.IsShowH5 = strings.Contains(item.Shows, "5")
		products.IsShowDelivery = strings.Contains(item.Shows, "6")
		// 验证是否已经存在
		products.LocaleNameIsExist = repository.NewProductRepo(db).CheckMultiLanguageNameExist(item.LocaleName)
		// 验证条形码存在性检查
		products.BarcodeIsExist = repository.NewProductRepo(db).CheckBarcodeExist(item.Barcode, 0)
		// 添加到列表
		productImportResp.List = append(productImportResp.List, products)
	}

	// 获取分类列表
	res, err := s.GetProductCategoryList(ctx.GetDbId())
	if err != nil {
		return product_resp.ProductImportResp{}, errors.WithMessage(err, "获取分类列表失败")
	}
	productImportResp.CategoryList = res.List

	// 获取单位列表
	unitList, err := base.NewProductUnitRepo(db).GetProductUnitList()
	if err != nil {
		return product_resp.ProductImportResp{}, errors.WithMessage(err, "获取单位列表失败")
	}
	for _, unit := range unitList {
		productImportResp.UnitList = append(productImportResp.UnitList, product_resp.ProductImportUnitListItem{
			Uuid:       unit.Uuid,
			LocaleName: unit.MultiLanguageName.GetNames(),
		})
	}

	// 获取规格列表
	skuList, err := base.NewProductFlavorRepo(db).GetProductFlavorList()
	if err != nil {
		return product_resp.ProductImportResp{}, errors.WithMessage(err, "获取规格列表失败")
	}
	for _, sku := range skuList {
		productImportResp.SkuList = append(productImportResp.SkuList, product_resp.ProductImportSkuListItem{
			Uuid:       sku.Uuid,
			LocaleName: sku.MultiLanguageName.GetNames(),
		})
	}

	// 获取税类列表
	taxList, err := repository.NewTaxRepo(db).GetTaxCategoryList()
	if err != nil {
		return product_resp.ProductImportResp{}, errors.WithMessage(err, "获取税类列表失败")
	}
	for _, tax := range taxList {
		productImportResp.TaxList = append(productImportResp.TaxList, product_resp.ProductImportTaxListItem{
			Uuid: tax.Uuid,
			Name: tax.Name,
		})
	}

	return productImportResp, nil
}

// ImportProduct 导入商品
func (s *productSrv) ImportProduct(ctx context.Context, reqs req.ProductImportReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	language := ctx.GetLanguage()

	// 验证条形码是否重复
	duplicateRows := reqs.GetBarcodeDuplicateRows()
	if len(duplicateRows) > 0 {
		return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(duplicateRows[0]) + "]: " + i18n.Translate(language, "商品条码不能重复"))
	}

	for _, item := range reqs.List {
		// 验证是否已经存在
		productNameIsExist := repository.NewProductRepo(db).CheckMultiLanguageNameExist(item.LocaleName)
		if !productNameIsExist.IsNull() {
			return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "商品名称已存在"))
		}
		// 验证条形码存在性检查
		if repository.NewProductRepo(db).CheckBarcodeExist(item.Barcode, 0) {
			return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "商品条码已存在"))
		}
		// 处理显示
		item.NumType = utils.IfInt(item.NumType == 1, 0, 1)
	}

	// 多规格合并
	lists := make(map[string]req.ProductShopAddReq)
	for _, item := range reqs.List {
		md5Key := item.LocaleName.GetMd5()
		sku := []req.ProductShopAddFlavorReq{}
		if _, exists := lists[md5Key]; exists {
			sku = lists[md5Key].Flavors
		}
		sku = append(sku, req.ProductShopAddFlavorReq{
			Uuid:         item.SkuUuid,
			Price:        item.ProductPrice,
			BarcodeValue: item.Barcode,
		})
		if _, exists := lists[md5Key]; !exists {
			lists[md5Key] = req.ProductShopAddReq{
				Type:         constant.ProductTypeProduct,
				LocaleName:   item.LocaleName,
				CategoryUuid: item.CategoryUuid,
				UnitUuid:     item.UnitUuid,
				Tax: req.ProductShopAddTaxReq{
					DineUuid:    item.DineTaxUuid,
					TakeoutUuid: item.TakeoutTaxUuid,
				},
				Status:          item.ProductStatus,
				NumType:         item.NumType,
				DeductStockType: item.DeductStockType,
				Show: req.ProductShopAddShowReq{
					IsShowCashier:   item.IsShowCashier,
					IsShowTablet:    item.IsShowTablet,
					IsShowKitchen:   item.IsShowKitchen,
					IsShowAssistant: item.IsShowAssistant,
					IsShowH5:        item.IsShowH5,
					IsShowDelivery:  item.IsShowDelivery,
				},
				Discount: req.ProductShopAddDiscountReq{
					IsEnableMemberDiscount:  item.IsEnableGrade,
					IsEnableOverallDiscount: item.OpenOverallDiscount,
				},
				Row: item.Row,
			}
		}

		// 更新规格列表
		temp := lists[md5Key]
		temp.Flavors = append(temp.Flavors, sku...)
		lists[md5Key] = temp
	}

	// 错误提示
	errorMessages := make([]string, 0)
	for _, item := range lists {
		err := s.AddProductShop(ctx, item)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("行[%d]: %s", item.Row, err.Error()))
		}
	}
	if len(errorMessages) > 0 {
		return errors.New(i18n.Translate(language, "商品导入失败") + ": " + strings.Join(errorMessages, "; "))
	}

	return nil
}

// GetProductSingleList 获取单规格商品列表
func (s *productSrv) GetProductSingleList(ctx context.Context, req req.ProductSingleListReq) (*product_resp.ProductSingleListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	opts := []repository.DBOption{
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductCategory(),
		productRepo.WithProductCategoryMultiLanguageName(),
		productRepo.WithProductPackageImageFile(),
		productRepo.WithProductBoms(
			commonRepo.WhereBySoftDelete(),
		),
		productRepo.WithProductBomsProductFlavorMultiLanguageName(),
	}

	// 搜索商品名称
	if req.Keyword != nil {
		opts = append(opts, commonRepo.WhereLike("name", *req.Keyword))
	}
	// 商品分类
	if req.CategoryUuid != nil {
		opts = append(opts, productRepo.WhereCategoryUuid(*req.CategoryUuid))
	}

	// 获取商品列表
	productPackages, total, err := productRepo.PaginateGetProductShopList(
		req.PageNo, req.PageSize, opts...,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品列表失败")
	}

	productList := make([]product_resp.ProductSingleListItemResp, 0, len(productPackages))
	for _, productPackage := range productPackages {
		for _, productBom := range productPackage.ProductBoms {
			if productBom.IsFlavor() {
				var productBomCardName dto.LocaleResponse
				if productBom.ProductBomCardUuid > 0 {
					productBomCardName, err = repository.NewProductBomCardRepo(db).GetProductBomCardName(productBom.ProductBomCardUuid)
					if err != nil {
						return nil, errors.WithMessage(err, "获取成本卡失败")
					}
				}
				productItem := product_resp.ProductSingleListItemResp{
					Uuid:               productBom.Uuid,
					Name:               productPackage.MultiLanguageName.GetNames(),
					FlavorName:         productBom.ProductFlavor.MultiLanguageName.GetNames(),
					CategoryUuid:       productPackage.CategoryUuid,
					ProductBomCardUuid: productBom.ProductBomCardUuid,
					ProductBomCardName: productBomCardName,
				}
				productList = append(productList, productItem)
			}

		}

	}

	productResp := product_resp.ProductSingleListResp{
		List: productList,
		Meta: dto.PageResponse{
			Total:    total,
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
	}

	return &productResp, nil
}

// GetProductShopList 获取商品列表（商家端）
func (s *productSrv) GetProductShopList(ctx context.Context, req req.ProductShopListReq) (*product_resp.ProductShopListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	opts := []repository.DBOption{
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductCategory(),
		productRepo.WithProductCategoryMultiLanguageName(),
		productRepo.WithProductPackageImageFile(),
		productRepo.WithProductBoms(
			commonRepo.WhereBySoftDelete(),
			commonRepo.SortWithID("DESC"),
		),
		productRepo.WithProductBomsProductFlavor(),
		productRepo.WithProductBomsProductFlavorMultiLanguageName(),
		productRepo.WithProductPackageAttributeGroup(),
		productRepo.WithProductUnit(),
		productRepo.WithProductUnitMultiLanguageName(),
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithSort("ASC"),
		commonRepo.SortWithID("DESC"),
	}

	// 搜索商品名称
	if req.Keyword != nil && *req.Keyword != "" {
		opts = append(opts, commonRepo.WhereLike("name", *req.Keyword))
	}
	// 商品类型
	if req.Type != nil && *req.Type != "" {
		typ, _ := convertor.ToInt(*req.Type)
		opts = append(opts, productRepo.WhereProductType(uint8(typ)))
	}
	// 商品状态
	if req.Status != nil && *req.Status != "" {
		status, _ := convertor.ToInt(*req.Status)
		opts = append(opts, commonRepo.WhereByStatus(uint(status)))
	}

	// 商品标签
	if req.Tag != nil && *req.Tag != "" {
		tagList := strings.Split(*req.Tag, ",")
		for _, tag := range tagList {
			switch tag {
			case "0":
				opts = append(opts, productRepo.WhereHasMultipleSpec())
			case "1":
				opts = append(opts, productRepo.WhereHasAttribute())
			case "2":
				opts = append(opts, productRepo.WhereHasSauce())
			}
		}
	}
	// 商品分类
	if req.CategoryUuid != nil && *req.CategoryUuid != "" {
		categoryUuid, _ := convertor.ToInt(*req.CategoryUuid)
		opts = append(opts, productRepo.WhereCategoryUuid(uint64(categoryUuid)))
	}

	// 获取商品列表
	productPackages, total, err := productRepo.PaginateGetProductShopList(
		req.PageNo, req.PageSize, opts...,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品列表失败")
	}

	productList := make([]product_resp.ProductShopListItemResp, 0, len(productPackages))
	for _, productPackage := range productPackages {
		// 格式化商品信息
		minPrice := 0.0
		maxPrice := 0.0
		specCount := 0
		specStockNum := 0.0
		IsMultipleSpec := false
		IsAttribute := len(productPackage.ProductPackageAttributeGroups) > 0
		IsSauce := false
		flavors := make([]product_resp.ProductShopListItemFlavorItemResp, 0)
		for _, productBom := range productPackage.ProductBoms {
			if productPackage.ProductType == constant.ProductTypeProduct {
				// 商品-规格
				if productBom.ProductFlavorUuid > 0 {
					if minPrice == 0 {
						minPrice = productBom.Price
					} else {
						minPrice = utils.IfFloat64(productBom.Price <= minPrice, productBom.Price, minPrice)
					}
					if maxPrice == 0 {
						maxPrice = productBom.Price
					} else {
						maxPrice = utils.IfFloat64(productBom.Price >= maxPrice, productBom.Price, maxPrice)
					}
					specCount++
					if productBom.StockNum > 0 {
						specStockNum = productBom.StockNum
					}
					flavors = append(flavors, product_resp.ProductShopListItemFlavorItemResp{
						Uuid:       productBom.Uuid,
						LocaleName: productBom.ProductFlavor.MultiLanguageName.GetNames(),
						Price:      productBom.Price,
					})
				}
				// 商品-加料
				if productBom.ProductSauceUuid > 0 {
					IsSauce = true
				}
			} else {
				minPrice = productBom.Price
				maxPrice = productBom.Price
				specStockNum = productBom.StockNum
				// 套餐
				if productBom.ProductFlavorUuid == 0 && productBom.ProductSauceUuid == 0 {
					flavors = append(flavors, product_resp.ProductShopListItemFlavorItemResp{
						Uuid:       productBom.Uuid,
						LocaleName: productPackage.MultiLanguageName.GetNames(),
						Price:      productBom.Price,
					})
				}
			}
		}
		if specCount > 1 {
			IsMultipleSpec = true
		}
		productItem := product_resp.ProductShopListItemResp{
			Uuid:       productPackage.Uuid,
			LocaleName: productPackage.MultiLanguageName.GetNames(),
			Image:      productPackage.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
			Tag: product_resp.ProductShopListItemTagResp{
				IsMultipleSpec: IsMultipleSpec,
				IsAttribute:    IsAttribute,
				IsSauce:        IsSauce,
			},
			MinPrice:            minPrice,
			MaxPrice:            maxPrice,
			Unit:                product_resp.ProductShopListItemUnitResp{LocaleName: productPackage.ProductUnit.MultiLanguageName.GetNames()},
			CategoryUuid:        productPackage.CategoryUuid,
			SpecialCategoryUuid: productPackage.SpecialCategoryUuid,
			Status:              int(productPackage.Status),
			IsSoldOut:           specStockNum <= 0,
			ProductType:         int(productPackage.ProductType),
			Sort:                int(productPackage.Sort),
			Flavors: product_resp.ProductShopListItemFlavorListResp{
				List: flavors,
			},
			NumType: productPackage.NumType,
		}

		productList = append(productList, productItem)
	}

	productResp := product_resp.ProductShopListResp{
		List: productList,
		Meta: dto.PageResponse{
			Total:    total,
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
	}

	return &productResp, nil
}

// SortProductShopList 排序商品列表
func (s *productSrv) SortProductShopList(ctx context.Context, req req.SortProductShopListReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	// 获取分类
	productCategory, err := productRepo.GetProductCategory(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(req.CategoryUuid),
	)
	if err != nil {
		return errors.WithMessage(err, "获取分类失败")
	}

	// 查询分类下的商品
	productPackages, err := productRepo.GetProductShopList(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereCategoryUuid(productCategory.Uuid),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品列表失败")
	}

	// 排序商品
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.List {
			productPackage, ok := slice.FindBy(productPackages, func(index int, productPackage model.ProductPackage) bool {
				return productPackage.Uuid == item.Uuid
			})
			if !ok {
				return errors.New("商品不存在")
			}
			if item.Sort == 0 {
				return errors.New("排序不能为0")
			}
			err := tx.Model(&model.ProductPackage{}).Where("uuid = ?", productPackage.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			}).Error
			if err != nil {
				return errors.WithMessage(err)
			}
		}

		return nil
	})

	return err
}

// GetProductDetail 获取商品详情
func (s *productSrv) GetProductDetail(ctx context.Context, req req.ProductDetailReq) (*product_resp.ProductDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	productPackage, err := repository.NewProductRepo(db).GetProductDetail(req.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品失败")
	}

	productDetailResp := product_resp.ProductDetailResp{
		ProductType:  productPackage.ProductType,
		Uuid:         productPackage.Uuid,
		LocaleName:   productPackage.MultiLanguageName.GetNames(),
		CategoryUuid: productPackage.CategoryUuid,
		CategoryName: productPackage.ProductCategory.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		UnitUuid:     productPackage.ProductUnit.Uuid,
		UnitName:     productPackage.ProductUnit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		Price:        &productPackage.Price,

		TakeoutTaxUuid: productPackage.TakeoutTax.Uuid,
		TakeoutTaxName: productPackage.TakeoutTax.Name,
		DineTaxUuid:    productPackage.DineTax.Uuid,
		DineTaxName:    productPackage.DineTax.Name,

		Status:          productPackage.Status,
		ImageFileUuid:   productPackage.ImageFileUuid,
		Image:           productPackage.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
		NumType:         &productPackage.NumType,
		DeductStockType: productPackage.DeductStockType,

		IsShowCashier:   productPackage.GetIsShowCashier(),
		IsShowTablet:    productPackage.GetIsShowTablet(),
		IsShowKitchen:   productPackage.GetIsShowKitchen(),
		IsShowAssistant: productPackage.GetIsShowAssistant(),
		IsShowH5:        productPackage.GetIsShowH5(),
		IsShowDelivery:  productPackage.GetIsShowDelivery(),

		OpenDiscount:        productPackage.GetOpenDiscount(),
		OpenOverallDiscount: productPackage.GetOpenOverallDiscount(),

		SauceRequired:     productPackage.SauceRequired == 1,
		SauceMaxSelection: productPackage.SauceMaxSelection,

		Flavors: product_resp.ProductFlavorList{
			List: productPackage.GetRespFlavorList(),
		},
		Sauces: product_resp.ProductSauceList{
			List:      productPackage.GetRespSaucesList(),
			IsMust:    productPackage.GetSauceRequired(),
			MaxSelect: int(productPackage.SauceMaxSelection),
		},
		AttributeGroups: product_resp.ProductAttributeGroupList{
			List: productPackage.GetRespAttributeGroupList(),
		},
		PackageSubProductGroups: product_resp.ProductPackageSubProductGroupList{
			List: productPackage.GetRespPackageSubProductGroupList(),
		},
	}

	return &productDetailResp, nil
}

// ProductShopStatus 修改商品状态
func (s *productSrv) ProductShopStatus(ctx context.Context, req req.ProductShopStatusReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	productPackage, err := productRepo.GetProduct(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(req.Uuid),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品失败")
	}
	if productPackage.ID == 0 {
		return errors.New("商品不存在")
	}

	if req.Status == nil {
		return errors.New("商品状态不能为空")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)
		err = tx.Model(&model.ProductPackage{}).Select("status").Where("uuid = ?", req.Uuid).Updates(map[string]any{
			"status": req.Status,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改商品状态失败")
		}
		if productPackage.ProductType == constant.ProductTypeProduct && *req.Status == 0 {
			productPackageGroupItems, err := productPackageGroupRepo.GetProductPackageGroupItems(
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByRelatedUuid(productPackage.Uuid),
				productPackageGroupRepo.WithProductPackageGroup(
					commonRepo.WhereBySoftDelete(),
				),
				productPackageGroupRepo.WithProductPackageGroupProduct(
					commonRepo.WhereBySoftDelete(),
				),
			)
			if err != nil {
				return errors.WithMessage(err, "获取商品套餐组商品失败")
			}
			for _, item := range productPackageGroupItems {
				if item.ProductPackageGroup != nil && item.ProductPackageGroup.ProductPackage != nil {
					err = tx.Model(&model.ProductPackage{}).Select("status").Where("uuid = ?", item.ProductPackageGroup.ProductPackage.Uuid).Updates(map[string]any{
						"status": 0,
					}).Error
					if err != nil {
						return errors.WithMessage(err, "修改商品套餐状态失败")
					}
					err = tx.Model(&model.ProductBom{}).Select("status").Where("product_package_uuid = ?", item.ProductPackageGroup.ProductPackage.Uuid).Updates(map[string]any{
						"status": 0,
					}).Error
					if err != nil {
						return errors.WithMessage(err, "修改商品套餐组商品状态失败")
					}
				}
			}

		}

		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "修改商品状态失败")
	}

	return nil
}

// AddProductShop 添加商品
func (s *productSrv) AddProductShop(ctx context.Context, req req.ProductShopAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productCheckSrv := NewProductCheckSrv(s.dbm, s.localeSrv, s.settingSrv)

	// 检查商品类型
	if err := productCheckSrv.CheckProductType(req.Type); err != nil {
		return errors.WithMessage(err, "检查商品类型失败")
	}
	// 检查商品名称
	if err := productCheckSrv.CheckProductName(ctx, 0, req.LocaleName); err != nil {
		return errors.WithMessage(err, "检查商品名称失败")
	}
	// 检查商品分类
	if err := productCheckSrv.CheckProductCategory(db, req.CategoryUuid); err != nil {
		return errors.WithMessage(err, "检查商品分类失败")
	}
	// 检查商品单位
	if err := productCheckSrv.CheckProductUnique(db, req.UnitUuid); err != nil {
		return errors.WithMessage(err, "检查商品单位失败")
	}
	// 商品专用检查
	flavorListResult := CheckProductFlavorResult{}
	sauceListResult := CheckProductSauceResult{}
	attributeListResult := []CheckProductAttributeGroupParam{}
	packageResult := CheckProductPackageResult{}
	if req.Type == constant.ProductTypeProduct {
		// 商品规格, 必填
		var flavors []CheckProductFlavorParam
		for _, flavor := range req.Flavors {
			flavors = append(flavors, CheckProductFlavorParam{
				Uuid:         flavor.Uuid,
				Price:        flavor.Price,
				BarcodeValue: flavor.BarcodeValue,
			})
		}
		result, err := productCheckSrv.CheckProductFlavor(db, flavors)
		if err != nil {
			return err
		}
		flavorListResult = *result
		flavorListResult.Status = req.Status
		flavorListResult.StockNum = 99999999
		// 商品属性, 可选
		if len(req.Attributes) > 0 {
			var attributes []CheckProductAttributeGroupParam
			for _, attribute := range req.Attributes {
				var attributeParams []CheckProductAttributeParam
				for _, attribute := range attribute.Attributes {
					attributeParams = append(attributeParams, CheckProductAttributeParam{
						Uuid:              attribute.Uuid,
						IsDefaultSelected: attribute.IsDefaultSelected,
					})
				}
				attributes = append(attributes, CheckProductAttributeGroupParam{
					Uuid:         attribute.Uuid,
					IsMust:       attribute.IsMust,
					IsOpenInput:  attribute.IsOpenInput,
					MaxSelection: attribute.MaxSelection,
					Attributes:   attributeParams,
				})
			}
			result, err := productCheckSrv.CheckProductAttribute(db, attributes)
			if err != nil {
				return errors.WithMessage(err, "检查商品属性失败")
			}
			attributeListResult = result
		}
		// 商品加料, 可选
		if len(req.Sauce.Sauces) > 0 {
			var sauceListParam []CheckProductSauceItemParam
			for _, sauceReq := range req.Sauce.Sauces {
				sauceListParam = append(sauceListParam, CheckProductSauceItemParam{
					Uuid:              sauceReq.Uuid,
					IsDefaultSelected: sauceReq.IsDefaultSelected,
				})
			}
			result, err := productCheckSrv.CheckProductSauce(db, CheckProductSauceParam{
				IsMust:       req.Sauce.IsMust,
				IsOpenInput:  req.Sauce.IsOpenInput,
				MaxSelection: req.Sauce.MaxSelection,
				Sauces:       sauceListParam,
			})
			if err != nil {
				return err
			}
			sauceListResult = *result
			sauceListResult.Status = req.Status
		}
	} else {
		// 添加套餐
		var groups []CheckProductPackageGroupParam
		for _, group := range req.Package.Groups {
			var products []CheckProductPackageGroupProductParam
			for _, product := range group.Products {
				products = append(products, CheckProductPackageGroupProductParam{
					BomUuid: product.BomUuid,
					Num:     product.Num,
					Sort:    product.Sort,
				})
			}
			groups = append(groups, CheckProductPackageGroupParam{
				LocaleName: group.LocaleName,
				Products:   products,
			})
		}
		result, err := productCheckSrv.CheckProductPackage(ctx, db, CheckProductPackageParam{
			Price:  req.Package.Price,
			Groups: groups,
		})
		if err != nil {
			return err
		}
		packageResult = *result
		flavorListResult = CheckProductFlavorResult{
			MinPrice: packageResult.Price,
			MaxPrice: packageResult.Price,
			StockNum: packageResult.StockNum,
			Status:   req.Status,
			Flavors: []CheckProductFlavorItemResult{
				{
					Name:  req.LocaleName.ToJson(),
					Price: packageResult.Price,
				},
			},
			IsPackage: true,
		}
	}
	// 商品税类
	if err := productCheckSrv.CheckProductTax(ctx, db, CheckProductTaxParam{
		DineUuid:    req.Tax.DineUuid,
		TakeoutUuid: req.Tax.TakeoutUuid,
	}); err != nil {
		return err
	}
	// 商品状态
	if err := productCheckSrv.CheckProductStatus(req.Status); err != nil {
		return err
	}
	// 商品图片
	if req.ImageFileUuid != 0 {
		if err := productCheckSrv.CheckProductImage(ctx, db, req.ImageFileUuid); err != nil {
			return err
		}
	}
	// 商品计价方式
	if err := productCheckSrv.CheckProductNumType(req.NumType); err != nil {
		return err
	}
	// 商品库存计算方式
	if err := productCheckSrv.CheckProductDeductStockType(req.DeductStockType); err != nil {
		return err
	}
	// 商品显示设置
	if err := productCheckSrv.CheckProductShow(CheckProductShowParam{
		IsShowCashier:   req.Show.IsShowCashier,
		IsShowTablet:    req.Show.IsShowTablet,
		IsShowKitchen:   req.Show.IsShowKitchen,
		IsShowAssistant: req.Show.IsShowAssistant,
		IsShowH5:        req.Show.IsShowH5,
		IsShowDelivery:  req.Show.IsShowDelivery,
	}); err != nil {
		return err
	}
	// 商品会员折扣
	if err := productCheckSrv.CheckProductMemberDiscount(req.Discount.IsEnableMemberDiscount); err != nil {
		return err
	}
	// 商品整单折扣
	if err := productCheckSrv.CheckProductOverallDiscount(req.Discount.IsEnableOverallDiscount); err != nil {
		return err
	}

	// 添加商品
	err := db.Transaction(func(tx *gorm.DB) error {

		// 添加商品包
		productPackageRes, err := s.AddProductPackage(ctx, tx, req, flavorListResult.MinPrice)
		if err != nil {
			return err
		}
		productPackageUuid := productPackageRes.Uuid
		erpCode := productPackageRes.ErpCode
		// 保存商品bom
		err = s.SaveProductPackageBom(ctx, tx, productPackageUuid, req.UnitUuid, erpCode, flavorListResult, sauceListResult)
		if err != nil {
			return err
		}
		// 商品属性
		err = s.SaveProductPackageAttribute(tx, attributeListResult, productPackageUuid)
		if err != nil {
			return err
		}

		// 套餐商品组
		if req.Type == constant.ProductTypePackage {
			err = s.SaveProductPackageGroup(tx, packageResult.Groups, productPackageUuid)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("添加商品失败", zap.Any("func", "AddProductShop"), zap.Any("params", req), zap.Error(err))
		return errors.WithMessage(err, "添加商品失败")
	}

	return nil
}

// EditProductShop 编辑商品
func (s *productSrv) EditProductShop(ctx context.Context, req req.ProductShopEditReq) (*product_resp.ProductEditResp, []string, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productCheckSrv := NewProductCheckSrv(s.dbm, s.localeSrv, s.settingSrv)
	productRepo := repository.NewProductRepo(db)
	productBomRepo := repository.NewProductBomRepo(db)
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(db)

	// 检查商品类型
	if err := productCheckSrv.CheckProductType(req.Type); err != nil {
		return nil, nil, err
	}
	// 检查商品名称
	if err := productCheckSrv.CheckProductName(ctx, req.Uuid, req.LocaleName); err != nil {
		return nil, nil, err
	}
	// 检查商品分类
	if err := productCheckSrv.CheckProductCategory(db, req.CategoryUuid); err != nil {
		return nil, nil, err
	}
	// 检查商品单位
	if err := productCheckSrv.CheckProductUnique(db, req.UnitUuid); err != nil {
		return nil, nil, err
	}
	// 商品专用检查
	flavorListResult := CheckProductFlavorResult{}
	sauceListResult := CheckProductSauceResult{}
	attributeListResult := []CheckProductAttributeGroupParam{}
	packageResult := CheckProductPackageResult{}
	if req.Type == constant.ProductTypeProduct {
		// 商品规格, 必填
		var packageNames []string
		var flavorNames []string
		var flavors []CheckProductFlavorParam
		for _, flavor := range req.Flavors {
			flavors = append(flavors, CheckProductFlavorParam{
				Uuid:         flavor.Uuid,
				Price:        flavor.Price,
				BarcodeValue: flavor.BarcodeValue,
				BomUuid:      flavor.BomUuid,
				IsDelete:     flavor.IsDelete,
			})
			if flavor.IsDelete {
				productFlavor, _ := productRepo.GetProductFlavor(
					productRepo.WhereUuid(flavor.Uuid),
					commonRepo.WhereBySoftDelete(),
					productRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
				)
				if productFlavor.Uuid == 0 {
					return nil, nil, errors.New("商品规格不存在")
				}
				productPackageGroupItems, _ := productPackageGroupRepo.GetProductPackageGroupItems(
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereByProductBomUuid(flavor.BomUuid),
					productPackageGroupRepo.WithProductPackageGroup(commonRepo.WhereBySoftDelete()),
					productPackageGroupRepo.WithProductPackageGroupProduct(commonRepo.WhereBySoftDelete()),
					productPackageGroupRepo.WithProductPackageGroupProductMultiLanguageName(commonRepo.WhereBySoftDelete()),
				)
				for _, item := range productPackageGroupItems {
					if item.ProductPackageGroup != nil && item.ProductPackageGroup.ProductPackage != nil {
						if !slices.Contains(flavorNames, productFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage())) {
							flavorNames = append(flavorNames, productFlavor.MultiLanguageName.GetNameByLang(ctx.GetLanguage()))
						}
						if !slices.Contains(packageNames, item.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage())) {
							packageNames = append(packageNames, item.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage()))
						}
					}
				}
			}
		}
		if len(flavorNames) > 0 {
			return &product_resp.ProductEditResp{
				List: packageNames,
			}, []string{strings.Join(flavorNames, "、")}, errors.New("规格“%s”已关联如下套餐，暂时无法删除，请先修改套餐")
		}
		result, err := productCheckSrv.CheckProductFlavor(db, flavors)
		if err != nil {
			return nil, nil, err
		}
		flavorListResult = *result
		flavorListResult.Status = req.Status
		flavorListResult.StockNum = 99999999
		// 商品属性, 可选
		if len(req.Attributes) > 0 {
			var attributes []CheckProductAttributeGroupParam
			for _, attribute := range req.Attributes {
				var attributeParams []CheckProductAttributeParam
				for _, attribute := range attribute.Attributes {
					attributeParams = append(attributeParams, CheckProductAttributeParam{
						Uuid:              attribute.Uuid,
						IsDefaultSelected: attribute.IsDefaultSelected,
						IsDelete:          attribute.IsDelete,
					})
				}
				attributes = append(attributes, CheckProductAttributeGroupParam{
					Uuid:         attribute.Uuid,
					IsMust:       attribute.IsMust,
					IsOpenInput:  attribute.IsOpenInput,
					MaxSelection: attribute.MaxSelection,
					Attributes:   attributeParams,
					IsDelete:     attribute.IsDelete,
				})
			}
			result, err := productCheckSrv.CheckProductAttribute(db, attributes)
			if err != nil {
				return nil, nil, err
			}
			attributeListResult = result
		}
		// 商品加料, 可选
		if len(req.Sauce.Sauces) > 0 {
			var sauceListParam []CheckProductSauceItemParam
			for _, sauceReq := range req.Sauce.Sauces {
				sauceListParam = append(sauceListParam, CheckProductSauceItemParam{
					Uuid:              sauceReq.Uuid,
					IsDefaultSelected: sauceReq.IsDefaultSelected,
					BomUuid:           sauceReq.BomUuid,
					IsDelete:          sauceReq.IsDelete,
				})
			}
			result, err := productCheckSrv.CheckProductSauce(db, CheckProductSauceParam{
				IsMust:       req.Sauce.IsMust,
				IsOpenInput:  req.Sauce.IsOpenInput,
				MaxSelection: req.Sauce.MaxSelection,
				Sauces:       sauceListParam,
			})
			if err != nil {
				return nil, nil, err
			}
			sauceListResult = *result
			sauceListResult.Status = req.Status
		}
	} else {
		var groups []CheckProductPackageGroupParam
		for _, group := range req.Package.Groups {
			var products []CheckProductPackageGroupProductParam
			for _, product := range group.Products {
				products = append(products, CheckProductPackageGroupProductParam{
					Uuid:     product.Uuid,
					BomUuid:  product.BomUuid,
					Num:      product.Num,
					Sort:     product.Sort,
					IsDelete: product.IsDelete,
				})
			}
			groups = append(groups, CheckProductPackageGroupParam{
				Uuid:       group.Uuid,
				LocaleName: group.LocaleName,
				Products:   products,
				IsDelete:   group.IsDelete,
			})
		}
		result, err := productCheckSrv.CheckProductPackage(ctx, db, CheckProductPackageParam{
			Price:  req.Package.Price,
			Groups: groups,
		})
		if err != nil {
			return nil, nil, err
		}
		bom, err := productBomRepo.GetProductBom(
			commonRepo.WhereByProductPackageUuid(req.Uuid),
			commonRepo.WhereBySoftDelete(),
		)
		if err != nil {
			return nil, nil, err
		}
		if bom.ID == 0 {
			return nil, nil, errors.New("商品不存在")
		}
		packageResult = *result
		flavorListResult = CheckProductFlavorResult{
			MinPrice: packageResult.Price,
			MaxPrice: packageResult.Price,
			StockNum: packageResult.StockNum,
			Status:   req.Status,
			Flavors: []CheckProductFlavorItemResult{
				{
					BomUuid: bom.Uuid,
					Name:    req.LocaleName.ToJson(),
					Price:   packageResult.Price,
				},
			},
			IsPackage: true,
		}
	}
	// 商品税类
	if err := productCheckSrv.CheckProductTax(ctx, db, CheckProductTaxParam{
		DineUuid:    req.Tax.DineUuid,
		TakeoutUuid: req.Tax.TakeoutUuid,
	}); err != nil {
		return nil, nil, err
	}
	// 商品状态
	if err := productCheckSrv.CheckProductStatus(req.Status); err != nil {
		return nil, nil, err
	}
	// 商品图片
	if req.ImageFileUuid != 0 {
		if err := productCheckSrv.CheckProductImage(ctx, db, req.ImageFileUuid); err != nil {
			return nil, nil, err
		}
	}
	// 商品计价方式
	if err := productCheckSrv.CheckProductNumType(req.NumType); err != nil {
		return nil, nil, err
	}
	// 商品库存计算方式
	if err := productCheckSrv.CheckProductDeductStockType(req.DeductStockType); err != nil {
		return nil, nil, err
	}
	// 商品显示设置
	if err := productCheckSrv.CheckProductShow(CheckProductShowParam{
		IsShowCashier:   req.Show.IsShowCashier,
		IsShowTablet:    req.Show.IsShowTablet,
		IsShowKitchen:   req.Show.IsShowKitchen,
		IsShowAssistant: req.Show.IsShowAssistant,
		IsShowH5:        req.Show.IsShowH5,
		IsShowDelivery:  req.Show.IsShowDelivery,
	}); err != nil {
		return nil, nil, err
	}
	// 商品会员折扣
	if err := productCheckSrv.CheckProductMemberDiscount(req.Discount.IsEnableMemberDiscount); err != nil {
		return nil, nil, err
	}
	// 商品整单折扣
	if err := productCheckSrv.CheckProductOverallDiscount(req.Discount.IsEnableOverallDiscount); err != nil {
		return nil, nil, err
	}

	// 编辑商品
	err := db.Transaction(func(tx *gorm.DB) error {

		// 编辑商品包
		productPackageRes, err := s.EditProductPackage(ctx, tx, req, flavorListResult.MinPrice)
		if err != nil {
			return err
		}
		productPackageUuid := productPackageRes.Uuid
		erpCode := productPackageRes.ErpCode
		// 保存商品bom
		err = s.SaveProductPackageBom(ctx, tx, productPackageUuid, req.UnitUuid, erpCode, flavorListResult, sauceListResult)
		if err != nil {
			return err
		}
		// 商品属性
		err = s.SaveProductPackageAttribute(tx, attributeListResult, productPackageUuid)
		if err != nil {
			return err
		}

		if req.Type == constant.ProductTypePackage {
			// 套餐商品组
			err = s.SaveProductPackageGroup(tx, packageResult.Groups, productPackageUuid)
			if err != nil {
				return err
			}
		} else {
			if req.Status == 0 {
				// 套餐商品组
				productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)
				productPackageGroupItems, err := productPackageGroupRepo.GetProductPackageGroupItems(
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereByRelatedUuid(productPackageUuid),
					productPackageGroupRepo.WithProductPackageGroup(
						commonRepo.WhereBySoftDelete(),
					),
					productPackageGroupRepo.WithProductPackageGroupProduct(
						commonRepo.WhereBySoftDelete(),
					),
				)
				if err != nil {
					return errors.WithMessage(err, "获取商品套餐组商品失败")
				}
				for _, item := range productPackageGroupItems {
					if item.ProductPackageGroup != nil && item.ProductPackageGroup.ProductPackage != nil {
						err = tx.Model(&model.ProductPackage{}).Select("status").Where("uuid = ?", item.ProductPackageGroup.ProductPackage.Uuid).Updates(map[string]any{
							"status": 0,
						}).Error
						if err != nil {
							return errors.WithMessage(err, "修改商品套餐状态失败")
						}
						err = tx.Model(&model.ProductBom{}).Select("status").Where("product_package_uuid = ?", item.ProductPackageGroup.ProductPackage.Uuid).Updates(map[string]any{
							"status": 0,
						}).Error
						if err != nil {
							return errors.WithMessage(err, "修改商品套餐组商品状态失败")
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("添加商品失败", zap.Any("func", "AddProductShop"), zap.Any("params", req), zap.Error(err))
		return nil, nil, errors.WithMessage(err, "添加商品失败")
	}

	return nil, nil, nil
}

// EditProductPackage 编辑商品包
func (s *productSrv) EditProductPackage(ctx context.Context, tx *gorm.DB, req req.ProductShopEditReq, price float64) (*AddProductPackageRes, error) {
	commonRepo := repository.NewCommonRepo()
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
	productPackageRepo := repository.NewProductPackageRepo(tx)

	productPackage, err := productPackageRepo.GetProductPackage(
		commonRepo.WhereByUuid(req.Uuid),
		commonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "商品不存在")
	}
	// 保存多语言名称
	err = multiLanguageNameRepo.UpdateMultiLanguageName(productPackage.MultiLanguageNameUuid, model.MultiLanguageName{
		ZhName:   req.LocaleName.ZH,
		ThName:   req.LocaleName.TH,
		EnName:   req.LocaleName.EN,
		ZhTwName: req.LocaleName.ZHTW,
		JaName:   req.LocaleName.JA,
		KoName:   req.LocaleName.KO,
		MyName:   req.LocaleName.MY,
		TrName:   req.LocaleName.TR,
		SvName:   req.LocaleName.SV,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "保存多语言名称失败")
	}
	err = productPackageRepo.UpdateProductPackage(map[string]any{
		"name":                  req.LocaleName.ToJson(),
		"image_file_uuid":       req.ImageFileUuid,
		"deduct_stock_type":     req.DeductStockType,
		"num_type":              req.NumType,
		"unit_uuid":             req.UnitUuid,
		"dine_tax_uuid":         req.Tax.DineUuid,
		"category_uuid":         req.CategoryUuid,
		"takeout_tax_uuid":      req.Tax.TakeoutUuid,
		"status":                req.Status,
		"is_show_cashier":       req.Show.IsShowCashier,
		"is_show_tablet":        req.Show.IsShowTablet,
		"is_show_kitchen":       req.Show.IsShowKitchen,
		"is_show_assistant":     req.Show.IsShowAssistant,
		"is_show_h5":            req.Show.IsShowH5,
		"is_show_delivery":      req.Show.IsShowDelivery,
		"price":                 price,
		"open_discount":         req.Discount.IsEnableMemberDiscount,
		"open_overall_discount": req.Discount.IsEnableOverallDiscount,
	}, commonRepo.WhereByUuid(productPackage.Uuid))
	if err != nil {
		return nil, errors.WithMessage(err, "保存商品包失败")
	}

	return &AddProductPackageRes{
		Uuid:    req.Uuid,
		ErpCode: productPackage.ErpCode,
	}, nil
}

type AddProductPackageRes struct {
	Uuid    uint64 `json:"uuid"`
	ErpCode string `json:"erp_code"`
}

// AddProductPackage 添加商品包
func (s *productSrv) AddProductPackage(ctx context.Context, tx *gorm.DB, request req.ProductShopAddReq, price float64) (*AddProductPackageRes, error) {
	commonRepo := repository.NewCommonRepo()
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
	productRepo := repository.NewProductRepo(tx)
	productPackageRepo := repository.NewProductPackageRepo(tx)

	// 保存多语言名称
	multiLanguageName := model.NewMultiLanguageName(request.LocaleName.ToJson())
	multiLanguageNameUuid, err := multiLanguageNameRepo.CreateMultiLanguageName(*multiLanguageName)
	if err != nil {
		return nil, errors.WithMessage(err, "保存多语言名称失败")
	}
	// 保存商品包
	maxSort, err := productRepo.GetProductShopMaxSort(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByCategoryUuid(request.CategoryUuid),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品最大排序失败")
	}
	uuid, _ := utils.GetID()
	sort := maxSort + 1

	// 同步商品到erp
	erpCode := ""
	// 只有商品需要同步到erp时创建item模版，套餐在bom中同步
	if request.Type == constant.ProductTypeProduct {
		if ctx.GetCompany().IsOpenErp() {
			multiLanguageName := model.NewMultiLanguageName(request.LocaleName.ToJson())
			enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
			if err != nil {
				return nil, errors.WithMessage(err, "翻译失败")
			}
			productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(request.UnitUuid))
			if errGetUnit != nil {
				return nil, errors.WithMessage(errGetUnit, "获取商品单位失败")
			}

			erpSrv := erp.NewIErpSrv(s.dbm)
			itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
				ItemName: enName,
				StockUom: productUnit.ErpnextUom,
			})
			if errErp != nil {
				return nil, errors.WithMessage(errErp, "同步商品到erp失败")
			}
			erpCode = itemInfo.ItemCode
		}
	}
	productPackage := &model.ProductPackage{
		BaseModel: model.BaseModel{
			Uuid:       uuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  request.LocaleName.ToJson(),
		ErpCode:               erpCode,
		MultiLanguageNameUuid: multiLanguageNameUuid,
		ImageFileUuid:         request.ImageFileUuid,
		DeductStockType:       uint(request.DeductStockType),
		NumType:               uint(request.NumType),
		UnitUuid:              request.UnitUuid,
		DineTaxUuid:           request.Tax.DineUuid,
		CategoryUuid:          request.CategoryUuid,
		TakeoutTaxUuid:        request.Tax.TakeoutUuid,
		Status:                uint(request.Status),
		IsShowCashier:         uint(request.Show.IsShowCashier),
		IsShowTablet:          uint(request.Show.IsShowTablet),
		IsShowKitchen:         uint(request.Show.IsShowKitchen),
		IsShowAssistant:       uint(request.Show.IsShowAssistant),
		IsShowH5:              uint(request.Show.IsShowH5),
		IsShowDelivery:        uint(request.Show.IsShowDelivery),
		Sort:                  uint(sort),
		Price:                 price,
		ProductType:           uint(request.Type),
		OpenDiscount:          uint(request.Discount.IsEnableMemberDiscount),
		OpenOverallDiscount:   uint(request.Discount.IsEnableOverallDiscount),
	}
	err = productPackageRepo.CreateProductPackage(productPackage)
	if err != nil {
		return nil, errors.WithMessage(err, "保存商品包失败")
	}

	return &AddProductPackageRes{
		Uuid:    uuid,
		ErpCode: erpCode,
	}, nil
}

// SaveProductPackageBom 添加商品bom
func (s *productSrv) SaveProductPackageBom(ctx context.Context, tx *gorm.DB, productPackageUuid uint64, uintUuid uint64, templateItemCode string, flavorListResult CheckProductFlavorResult, sauceResult CheckProductSauceResult) error {
	commonRepo := repository.NewCommonRepo()
	productPackageRepo := repository.NewProductPackageRepo(tx)
	productBomRepo := repository.NewProductBomRepo(tx)
	warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
	warehouseMonthlyFormRepo := repository.NewWarehouseMonthlyFormRepo(tx)
	setting, err := s.settingSrv.GetCompanySetting(ctx)

	// 商品规格
	for _, flavor := range flavorListResult.Flavors {
		if flavor.IsDelete {
			err := productBomRepo.UpdateProductBom(map[string]any{
				"delete_time": time.Now().Unix(),
			}, commonRepo.WhereByUuid(flavor.BomUuid))
			if err != nil {
				return errors.WithMessage(err, "删除商品bom失败")
			}
			if setting.SaleStock == 1 {
				// 删除出库
				outFormUuid, _ := utils.GetID()
				warehouseOutForm := model.WarehouseOutForm{
					BaseModel: model.BaseModel{
						Uuid: outFormUuid,
					},
					FormNo:       warehouseFormRepo.GenerateWarehouseOutFormNo(setting.Timezone),
					Scene:        constant.WarehouseOutFormSceneDelete,
					Status:       constant.WarehouseOutFormStatusSuccess,
					OperatorUuid: ctx.GetStaffUuid(),
				}
				err = warehouseFormRepo.CreateWarehouseOutFormRecord(warehouseOutForm)
				if err != nil {
					return errors.WithMessage(err, "保存出库单失败")
				}
				warehouseOutFormItem := model.WarehouseOutFormItem{
					Num:                  flavorListResult.StockNum,
					Scene:                constant.WarehouseOutFormSceneDelete,
					Status:               1,
					ReduceStock:          constant.WarehouseOutFormItemReduceStockSuccess,
					WarehouseOutFormUuid: outFormUuid,
					ProductBomUuid:       flavor.BomUuid,
				}
				err = warehouseFormRepo.CreateWarehouseOutFormItemRecord(warehouseOutFormItem)
				if err != nil {
					return errors.WithMessage(err, "保存出库单明细失败")
				}
			}
		} else {
			if flavor.BomUuid == 0 {
				erpCode := ""
				if flavorListResult.IsPackage {
					// 同步套餐到erp
					if ctx.GetCompany().IsOpenErp() {
						multiLanguageName := model.NewMultiLanguageName(flavor.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(uintUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						stockUom := productUnit.ErpnextUom
						params := req.PackageAddErpReq{
							ItemName: enName,
							StockUom: stockUom,
						}
						itemInfo, errErp := erpSrv.AddPackage(ctx, params)
						if errErp != nil {
							return errors.WithMessage(errErp)
						}
						erpCode = itemInfo.ItemCode
					}
				} else {
					// 同步商品到erp
					if ctx.GetCompany().IsOpenErp() {
						multiLanguageName := model.NewMultiLanguageName(flavor.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(uintUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						stockUom := productUnit.ErpnextUom
						params := req.ProductAddErpReq{
							ItemName:          enName,
							StockUom:          stockUom,
							TemplateItemCode:  templateItemCode,
							ItemSpecification: enName,
						}
						itemInfo, errErp := erpSrv.AddProduct(ctx, params)
						if errErp != nil {
							return errors.WithMessage(errErp)
						}
						erpCode = itemInfo.ItemCode
					}
				}

				flavorUuid, _ := utils.GetID()
				_, err := productBomRepo.CreateProductBom(model.ProductBom{
					BaseModel: model.BaseModel{
						Uuid: flavorUuid,
					},
					Price:              flavor.Price,
					Name:               flavor.Name,
					ErpCode:            erpCode,
					ProductFlavorUuid:  flavor.Uuid,
					ProductPackageUuid: productPackageUuid,
					StockNum:           flavorListResult.StockNum,
					BarcodeValue:       flavor.BarcodeValue,
					Status:             flavorListResult.Status,
					IsOpenStock:        1,
				})
				if err != nil {
					return errors.WithMessage(err, "保存商品bom失败")
				}
				// 开启库存管理
				if setting.SaleStock == 1 {
					// 添加入库
					warehouseForm := model.WarehouseForm{
						FormNo:         warehouseFormRepo.GenerateWarehouseFormNo(setting.Timezone),
						Scene:          constant.WarehouseFormSceneAddStock,
						Num:            int(flavorListResult.StockNum),
						ProductBomUuid: flavorUuid,
						OperatorUuid:   ctx.GetStaffUuid(),
					}
					err = warehouseFormRepo.CreateWarehouseFormRecord(warehouseForm)
					if err != nil {
						return errors.WithMessage(err, "保存入库单失败")
					}
					// 添加月初库存记录
					warehouseMonthlyProductBomForm := model.WarehouseMonthlyProductBomForm{
						Year:           utils.SetTimezone(setting.Timezone).Now().Year(),
						Month:          int(utils.SetTimezone(setting.Timezone).Now().Month()),
						Scene:          constant.WarehouseMonthlyFormSceneStart,
						ProductBomUuid: flavorUuid,
						Stock:          float64(flavorListResult.StockNum),
					}
					err = warehouseMonthlyFormRepo.CreateWarehouseMonthlyProductBomForm(warehouseMonthlyProductBomForm)
					if err != nil {
						return errors.WithMessage(err, "保存月初库存记录失败")
					}
				}
			} else {
				err = productBomRepo.UpdateProductBom(map[string]any{
					"price":                flavor.Price,
					"name":                 flavor.Name,
					"product_flavor_uuid":  flavor.Uuid,
					"product_package_uuid": productPackageUuid,
					"barcode_value":        flavor.BarcodeValue,
					"status":               flavorListResult.Status,
					"is_open_stock":        1,
				}, commonRepo.WhereByUuid(flavor.BomUuid))
				if err != nil {
					return errors.WithMessage(err, "更新商品bom失败")
				}

				// 同步商品到erp
				if ctx.GetCompany().IsOpenErp() {
					if flavorListResult.IsPackage {
						// 同步套餐到erp
						multiLanguageName := model.NewMultiLanguageName(flavor.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(uintUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						productBom, errGetBom := productBomRepo.GetProductBom(commonRepo.WhereByUuid(flavor.BomUuid))
						if errGetBom != nil {
							return errors.WithMessage(errGetBom, "获取商品bom失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						_, errErp := erpSrv.AddPackage(ctx, req.PackageAddErpReq{
							ItemName: enName,
							StockUom: productUnit.ErpnextUom,
							ItemCode: productBom.ErpCode,
						})
						if errErp != nil {
							return errors.WithMessage(errErp, "同步套餐到erp失败")
						}
					} else {
						// 同步商品到erp
						multiLanguageName := model.NewMultiLanguageName(flavor.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(uintUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						productBom, errGetBom := productBomRepo.GetProductBom(commonRepo.WhereByUuid(flavor.BomUuid))
						if errGetBom != nil {
							return errors.WithMessage(errGetBom, "获取商品bom失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
							ItemName:          enName,
							StockUom:          productUnit.ErpnextUom,
							ItemCode:          productBom.ErpCode,
							TemplateItemCode:  templateItemCode,
							ItemSpecification: enName,
						})
						if errErp != nil {
							return errors.WithMessage(errErp)
						}
					}
				}

				if setting.SaleStock == 1 {
					bom, err := productBomRepo.GetProductBom(commonRepo.WhereByUuid(flavor.BomUuid))
					if err != nil {
						return errors.WithMessage(err, "获取商品bom失败")
					}
					diffStockNum := math.Abs(decimal.NewFromFloat(flavorListResult.StockNum).Sub(decimal.NewFromFloat(bom.StockNum)).InexactFloat64())
					if diffStockNum > 0 {
						// 调整入库
						if flavorListResult.StockNum > bom.StockNum {
							warehouseForm := model.WarehouseForm{
								FormNo:         warehouseFormRepo.GenerateWarehouseFormNo(setting.Timezone),
								Scene:          constant.WarehouseFormSceneAddStock,
								Num:            int(diffStockNum),
								ProductBomUuid: flavor.BomUuid,
								OperatorUuid:   ctx.GetStaffUuid(),
							}
							err = warehouseFormRepo.CreateWarehouseFormRecord(warehouseForm)
							if err != nil {
								return errors.WithMessage(err, "保存入库单失败")
							}
						}
						// 调整出库
						if flavorListResult.StockNum < bom.StockNum {
							outFormUuid, _ := utils.GetID()
							warehouseOutForm := model.WarehouseOutForm{
								BaseModel: model.BaseModel{
									Uuid: outFormUuid,
								},
								FormNo:       warehouseFormRepo.GenerateWarehouseOutFormNo(setting.Timezone),
								Scene:        constant.WarehouseOutFormSceneAdjust,
								Status:       constant.WarehouseOutFormStatusSuccess,
								OperatorUuid: ctx.GetStaffUuid(),
							}
							err = warehouseFormRepo.CreateWarehouseOutFormRecord(warehouseOutForm)
							if err != nil {
								return errors.WithMessage(err, "保存出库单失败")
							}
							warehouseOutFormItem := model.WarehouseOutFormItem{
								Num:                  diffStockNum,
								Scene:                constant.WarehouseOutFormSceneAdjust,
								Status:               1,
								ReduceStock:          constant.WarehouseOutFormItemReduceStockSuccess,
								WarehouseOutFormUuid: outFormUuid,
								ProductBomUuid:       flavor.BomUuid,
							}
							err = warehouseFormRepo.CreateWarehouseOutFormItemRecord(warehouseOutFormItem)
							if err != nil {
								return errors.WithMessage(err, "保存出库单明细失败")
							}
						}
					}
				}
			}
		}
	}
	// 商品小料
	for _, sauce := range sauceResult.Sauces {
		if sauce.IsDelete {
			err := productBomRepo.UpdateProductBom(map[string]any{
				"delete_time": time.Now().Unix(),
			}, commonRepo.WhereByUuid(sauce.BomUuid))
			if err != nil {
				return errors.WithMessage(err, "删除商品bom失败")
			}
		} else {
			if sauce.BomUuid == 0 {
				_, err := productBomRepo.CreateProductBom(model.ProductBom{
					Price:              sauce.Price,
					Name:               sauce.Name,
					ProductSauceUuid:   sauce.Uuid,
					ProductPackageUuid: productPackageUuid,
					StockNum:           99999999,
					Status:             flavorListResult.Status,
					IsOpenStock:        1,
					IsDefaultSelect:    sauce.IsDefaultSelected,
				})
				if err != nil {
					return errors.WithMessage(err, "保存商品bom失败")
				}
			} else {
				err := productBomRepo.UpdateProductBom(map[string]any{
					"price":                sauce.Price,
					"name":                 sauce.Name,
					"product_sauce_uuid":   sauce.Uuid,
					"product_package_uuid": productPackageUuid,
					"stock_num":            99999999,
					"status":               flavorListResult.Status,
					"is_open_stock":        1,
					"is_default_select":    sauce.IsDefaultSelected,
				}, commonRepo.WhereByUuid(sauce.BomUuid))
				if err != nil {
					return errors.WithMessage(err, "更新商品bom失败")
				}
			}
		}
	}
	// 更新商品包
	err = productPackageRepo.UpdateProductPackage(map[string]any{
		"sauce_required":      sauceResult.IsMust,
		"sauce_max_selection": sauceResult.MaxSelection,
	}, commonRepo.WhereByUuid(productPackageUuid))
	if err != nil {
		return errors.WithMessage(err, "更新商品包失败")
	}

	return nil
}

// SaveProductPackageAttribute 添加商品属性
func (s *productSrv) SaveProductPackageAttribute(tx *gorm.DB, attributeGroupList []CheckProductAttributeGroupParam, productPackageUuid uint64) error {
	commonRepo := repository.NewCommonRepo()
	productPackageAttributeGroupRepo := repository.NewProductPackageAttributeGroupRepo(tx)
	productPackageAttributeRepo := repository.NewProductPackageAttributeRepo(tx)

	for _, attributeGroup := range attributeGroupList {
		// 删除该属性组，则删除该组下所有的属性值
		if attributeGroup.IsDelete {
			productPackageAttributeGroup, err := productPackageAttributeGroupRepo.GetProductPackageAttributeGroup(
				commonRepo.WhereByProductPackageUuid(productPackageUuid),
				commonRepo.WhereByProductAttributeGroupUuid(attributeGroup.Uuid),
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return errors.WithMessage(err, "获取商品关联属性组失败")
			}
			err = productPackageAttributeGroupRepo.DeleteProductPackageAttributeGroup(commonRepo.WhereByUuid(productPackageAttributeGroup.Uuid))
			if err != nil {
				return errors.WithMessage(err, "删除商品关联属性组失败")
			}
			err = productPackageAttributeRepo.DeleteProductPackageAttribute(commonRepo.WhereByProductPackageAttributeGroupUuid(productPackageAttributeGroup.Uuid))
			if err != nil {
				return errors.WithMessage(err, "删除商品关联属性值失败")
			}
		} else {
			productPackageAttributeGroup, _ := productPackageAttributeGroupRepo.GetProductPackageAttributeGroup(
				commonRepo.WhereByProductPackageUuid(productPackageUuid),
				commonRepo.WhereByProductAttributeGroupUuid(attributeGroup.Uuid),
				commonRepo.WhereBySoftDelete(),
			)
			isAdd := false
			if productPackageAttributeGroup == nil || productPackageAttributeGroup.Uuid == 0 {
				isAdd = true
			}
			if isAdd {
				uuid, _ := utils.GetID()
				// 新增商品包关联属性组
				productPackageAttributeGroup = &model.ProductPackageAttributeGroup{
					BaseModel: model.BaseModel{
						Uuid:       uuid,
						CreateTime: time.Now().Unix(),
						UpdateTime: time.Now().Unix(),
					},
					IsMust:                    uint(attributeGroup.IsMust),
					MaxSelection:              uint(attributeGroup.MaxSelection),
					ProductPackageUuid:        productPackageUuid,
					ProductAttributeGroupUuid: attributeGroup.Uuid,
				}
				err := productPackageAttributeGroupRepo.CreateProductPackageAttributeGroups([]model.ProductPackageAttributeGroup{*productPackageAttributeGroup})
				if err != nil {
					return errors.WithMessage(err, "保存商品包关联属性组失败")
				}
				// 新增商品包关联属性值
				productPackageAttributeList := make([]model.ProductPackageAttribute, 0)
				for _, attribute := range attributeGroup.Attributes {
					productPackageAttributeList = append(productPackageAttributeList, model.ProductPackageAttribute{
						ProductPackageAttributeGroupUuid: uuid,
						AttributeUuid:                    attribute.Uuid,
						IsDefaultSelected:                uint(attribute.IsDefaultSelected),
					})
				}
				err = productPackageAttributeRepo.CreateProductPackageAttributes(productPackageAttributeList)
				if err != nil {
					return errors.WithMessage(err, "保存商品包关联属性值失败")
				}
			} else {
				// 更新商品包关联属性组
				err := productPackageAttributeGroupRepo.UpdateProductPackageAttributeGroup(map[string]any{
					"is_must":       attributeGroup.IsMust,
					"max_selection": attributeGroup.MaxSelection,
				}, commonRepo.WhereByUuid(productPackageAttributeGroup.Uuid))
				if err != nil {
					return errors.WithMessage(err, "更新商品包关联属性组失败")
				}
				for _, attribute := range attributeGroup.Attributes {
					if attribute.IsDelete {
						err := productPackageAttributeRepo.DeleteProductPackageAttribute(
							commonRepo.WhereByProductPackageAttributeGroupUuid(productPackageAttributeGroup.Uuid),
							commonRepo.WhereByAttributeUuid(attribute.Uuid))
						if err != nil {
							return errors.WithMessage(err, "删除商品包关联属性值失败")
						}
					} else {
						productPackageAttribute, _ := productPackageAttributeRepo.GetProductPackageAttribute(
							commonRepo.WhereByProductPackageAttributeGroupUuid(productPackageAttributeGroup.Uuid),
							commonRepo.WhereByAttributeUuid(attribute.Uuid),
							commonRepo.WhereBySoftDelete(),
						)
						if productPackageAttribute.Uuid == 0 {
							uuid, _ := utils.GetID()
							productPackageAttribute = &model.ProductPackageAttribute{
								BaseModel: model.BaseModel{
									Uuid:       uuid,
									CreateTime: time.Now().Unix(),
									UpdateTime: time.Now().Unix(),
								},
								ProductPackageAttributeGroupUuid: productPackageAttributeGroup.Uuid,
								AttributeUuid:                    attribute.Uuid,
								IsDefaultSelected:                uint(attribute.IsDefaultSelected),
							}
							err := productPackageAttributeRepo.CreateProductPackageAttributes([]model.ProductPackageAttribute{*productPackageAttribute})
							if err != nil {
								return errors.WithMessage(err, "保存商品包关联属性值失败")
							}
						} else {
							err := productPackageAttributeRepo.UpdateProductPackageAttribute(map[string]any{
								"is_default_selected": attribute.IsDefaultSelected,
							}, commonRepo.WhereByUuid(productPackageAttribute.Uuid))
							if err != nil {
								return errors.WithMessage(err, "更新商品包关联属性值失败")
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (s *productSrv) SaveProductPackageGroup(tx *gorm.DB, groupList []CheckProductPackageGroupResult, productPackageUuid uint64) error {
	commonRepo := repository.NewCommonRepo()
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
	productBomRepo := repository.NewProductBomRepo(tx)

	for _, group := range groupList {
		if group.IsDelete {
			err := productPackageGroupRepo.DeleteProductPackageGroup(
				commonRepo.WhereByUuid(group.Uuid),
			)
			if err != nil {
				return errors.WithMessage(err, "删除套餐组失败")
			}
			err = productPackageGroupRepo.DeleteProductPackageGroupItem(
				commonRepo.WhereByProductPackageGroupUuid(group.Uuid),
			)
			if err != nil {
				return errors.WithMessage(err, "删除套餐组商品失败")
			}
		} else {
			if group.Uuid == 0 {
				// 保存多语言名称
				multiLanguageNameUuid, err := multiLanguageNameRepo.CreateMultiLanguageName(model.MultiLanguageName{
					ZhName:   group.LocaleName.ZH,
					ThName:   group.LocaleName.TH,
					EnName:   group.LocaleName.EN,
					ZhTwName: group.LocaleName.ZHTW,
					JaName:   group.LocaleName.JA,
					KoName:   group.LocaleName.KO,
					MyName:   group.LocaleName.MY,
					TrName:   group.LocaleName.TR,
					SvName:   group.LocaleName.SV,
				})
				if err != nil {
					return errors.WithMessage(err, "保存多语言名称失败")
				}
				groupUuid, _ := utils.GetID()
				err = productPackageGroupRepo.CreateProductPackageGroup(&model.ProductPackageGroup{
					BaseModel:             model.BaseModel{Uuid: groupUuid},
					Name:                  group.LocaleName.ToJson(),
					MultiLanguageNameUuid: multiLanguageNameUuid,
					ProductPackageUuid:    productPackageUuid,
				})
				if err != nil {
					return errors.WithMessage(err, "保存套餐组失败")
				}
				for _, item := range group.Products {
					bom, err := productBomRepo.GetProductBom(
						commonRepo.WhereByUuid(item.BomUuid),
						commonRepo.WhereBySoftDelete(),
					)
					if err != nil || bom.ID == 0 {
						return errors.WithMessage(err, "获取商品bom失败")
					}
					itemUuid, _ := utils.GetID()
					productPackageGroupRepo.CreateProductPackageGroupItem(&model.ProductPackageGroupItem{
						BaseModel:               model.BaseModel{Uuid: itemUuid},
						ProductPackageGroupUuid: groupUuid,
						RelatedUuid:             bom.ProductPackageUuid,
						ProductBomUuid:          item.BomUuid,
						Num:                     float64(item.Num),
						Sort:                    item.Sort,
					})
				}
			} else {
				curGroup, err := productPackageGroupRepo.GetProductPackageGroup(
					commonRepo.WhereByUuid(group.Uuid),
					commonRepo.WhereBySoftDelete(),
				)
				if err != nil || curGroup.ID == 0 {
					return errors.WithMessage(err, "获取套餐组失败")
				}
				// 保存多语言名称
				err = multiLanguageNameRepo.UpdateMultiLanguageName(curGroup.MultiLanguageNameUuid, model.MultiLanguageName{
					ZhName:   group.LocaleName.ZH,
					ThName:   group.LocaleName.TH,
					EnName:   group.LocaleName.EN,
					ZhTwName: group.LocaleName.ZHTW,
					JaName:   group.LocaleName.JA,
					KoName:   group.LocaleName.KO,
					MyName:   group.LocaleName.MY,
					TrName:   group.LocaleName.TR,
					SvName:   group.LocaleName.SV,
				})
				if err != nil {
					return errors.WithMessage(err, "保存多语言名称失败")
				}
				err = productPackageGroupRepo.UpdateProductPackageGroup(map[string]any{
					"name": group.LocaleName.ToJson(),
				}, commonRepo.WhereByUuid(group.Uuid))
				if err != nil {
					return errors.WithMessage(err, "更新套餐组失败")
				}
				for _, item := range group.Products {
					if item.IsDelete {
						err := productPackageGroupRepo.DeleteProductPackageGroupItem(
							commonRepo.WhereByUuid(item.Uuid),
						)
						if err != nil {
							return errors.WithMessage(err, "删除套餐组商品失败")
						}
					} else {
						bom, err := productBomRepo.GetProductBom(
							commonRepo.WhereByUuid(item.BomUuid),
							commonRepo.WhereBySoftDelete(),
						)
						if err != nil || bom.ID == 0 {
							return errors.WithMessage(err, "获取商品bom失败")
						}
						if item.Uuid == 0 {
							itemUuid, _ := utils.GetID()
							err := productPackageGroupRepo.CreateProductPackageGroupItem(&model.ProductPackageGroupItem{
								BaseModel:               model.BaseModel{Uuid: itemUuid},
								ProductPackageGroupUuid: curGroup.Uuid,
								RelatedUuid:             bom.ProductPackageUuid,
								ProductBomUuid:          item.BomUuid,
								Num:                     float64(item.Num),
								Sort:                    item.Sort,
							})
							if err != nil {
								return errors.WithMessage(err, "保存套餐组商品失败")
							}
						} else {
							err := productPackageGroupRepo.UpdateProductPackageGroupItem(map[string]any{
								"related_uuid":     bom.ProductPackageUuid,
								"product_bom_uuid": item.BomUuid,
								"num":              item.Num,
								"sort":             item.Sort,
							}, commonRepo.WhereByUuid(item.Uuid))
							if err != nil {
								return errors.WithMessage(err, "更新套餐组商品失败")
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// ProductTaxList 获取商品税类列表
func (s *productSrv) ProductTaxList(ctx context.Context) product_resp.ProductTaxListResp {
	db := s.dbm.GetDB(ctx.GetDbId())
	taxRepo := repository.NewTaxRepo(db)

	list := make([]product_resp.ProductTaxItemResp, 0)
	taxes, err := taxRepo.GetTaxCategoryList()
	if err != nil {
		return product_resp.ProductTaxListResp{
			List: list,
		}
	}
	for _, tax := range taxes {
		list = append(list, product_resp.ProductTaxItemResp{
			Uuid: tax.Uuid,
			Name: tax.Name,
			Rate: tax.TaxRate,
		})
	}

	return product_resp.ProductTaxListResp{
		List: list,
	}
}

// ProductShopChangePrice 商品改价
func (s *productSrv) ProductShopChangePrice(ctx context.Context, req req.ProductShopChangePriceReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	err := db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)
		productPackageRepo := repository.NewProductPackageRepo(tx)
		productBomRepo := repository.NewProductBomRepo(tx)

		productPackage, err := productRepo.GetProduct(
			commonRepo.WhereByUuid(req.Uuid),
			commonRepo.WhereBySoftDelete(),
		)
		if err != nil || productPackage.ID == 0 {
			return errors.WithMessage(err, "商品不存在")
		}

		price := 0.0
		for _, item := range req.Prices {
			if !productRepo.CheckPrice(item.Price, 0, 100000000, 2) {
				return errors.New("商品价格范围错误")
			}
			productBom, err := productBomRepo.GetProductBom(
				commonRepo.WhereByUuid(item.Uuid),
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil || productBom.ID == 0 {
				return errors.WithMessage(err, "商品Bom不存在")
			}

			err = productBomRepo.UpdateProductBom(map[string]any{
				"price": item.Price,
			}, commonRepo.WhereByUuid(item.Uuid))
			if err != nil {
				return errors.WithMessage(err, "商品BOM改价失败")
			}
			if price == 0 {
				price = item.Price
			} else {
				price = utils.IfFloat64(item.Price <= price, item.Price, price)
			}
		}

		err = productPackageRepo.UpdateProductPackage(map[string]any{
			"price": price,
		}, commonRepo.WhereByUuid(req.Uuid))
		if err != nil {
			return errors.WithMessage(err, "商品改价失败")
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "商品改价失败")
	}

	return nil
}
