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
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IProductSrv 定义产品服务接口
type IProductSrv interface {
	GetProductList(ctx context.Context, req req.ProductListReq) (product_resp.ProductListWithPaginationResp, error)               // 获取产品列表
	GetProductCategoryList(dbId uint64) (product_resp.ProductCategoryListResp, error)                                             // 获取产品类别列表
	GetProductUnitList(ctx context.Context, req req.ProductUnitListReq) (product_resp.ProductUnitListResp, error)                 // 获取产品单位列表
	GetProductUnit(ctx context.Context, req req.ProductUnitReq) (product_resp.ProductUnitDetail, error)                           // 获取产品单位详情
	GetProductRecommendList(ctx context.Context, req req.ProductRecommendListReq) (*product_resp.ProductRecommendListResp, error) // 获取产品推荐列表
	SearchProducts(ctx context.Context, req req.ProductSearchReq) (*product_resp.ProductSearchResp, error)                        // 搜索商品

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
		tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName)
		// 保存产品单位
		productUnit := model.ProductUnit{
			Sort:                  maxSort + 1,
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Name:                  addReq.LocaleName.ToJson(),
		}
		tx.Model(&model.ProductUnit{}).Create(&productUnit)

		// 修改商品包的单位UUID
		for _, productPackageUuid := range addReq.ProductPackageUuids {
			tx.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			})
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
		tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productUnit.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    editUnitReq.LocaleName.ZH,
			"th_name":    editUnitReq.LocaleName.TH,
			"en_name":    editUnitReq.LocaleName.EN,
			"zh_tw_name": editUnitReq.LocaleName.ZHTW,
			"ja_name":    editUnitReq.LocaleName.JA,
			"ko_name":    editUnitReq.LocaleName.KO,
			"my_name":    editUnitReq.LocaleName.MY,
			"tr_name":    editUnitReq.LocaleName.TR,
			"sv_name":    editUnitReq.LocaleName.SV,
		})
		// 修改产品单位
		tx.Model(&model.ProductUnit{}).Where("uuid = ?", editUnitReq.Uuid).Updates(map[string]any{
			"name": editUnitReq.LocaleName.ToJson(),
		})
		// 修改商品包的单位UUID, 如果商品包的单位UUID是当前单位，则修改为0
		tx.Model(&model.ProductPackage{}).Where("unit_uuid = ?", editUnitReq.Uuid).Updates(map[string]any{
			"unit_uuid": 0,
		})
		// 修改商品包的单位UUID
		if len(editUnitReq.ProductPackageUuids) > 0 {
			// 修改商品包的单位UUID
			tx.Model(&model.ProductPackage{}).Where("uuid in (?)", editUnitReq.ProductPackageUuids).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			})
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
		tx.Model(&model.ProductUnit{}).Where("uuid = ?", deleteUnitReq.Uuid).Update("delete_time", time.Now().Unix())
		// 删除多语言名称
		tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productUnit.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix())
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
			tx.Model(&model.ProductUnit{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			})
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
	// 检查商品包是否存在
	productPackages, err := productRepo.GetProductPackageListByUuids(addReq.ProductPackageUuids)
	if err != nil {
		return errors.WithMessage(err, "商品包不存在")
	}
	if len(productPackages) != len(addReq.ProductPackageUuids) {
		return errors.New("商品包不存在")
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
		tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName)
		// 保存商品加料
		productSauce := model.ProductSauce{
			Sort:                  maxSort + 1,
			Price:                 addReq.Price,
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Name:                  addReq.LocaleName.ToJson(),
		}
		tx.Model(&model.ProductSauce{}).Create(&productSauce)

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
		repository.NewProductBomRepo(tx).CreateProductBoms(boms)
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

	// 检查商品包是否存在
	productPackages, err := productRepo.GetProductPackageListByUuids(editReq.ProductPackageUuids)
	if err != nil {
		return errors.WithMessage(err, "商品包不存在")
	}
	if len(productPackages) != len(editReq.ProductPackageUuids) {
		return errors.New("商品包不存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		name := editReq.LocaleName.ToJson()
		// 修改多语言名称
		tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productSauce.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    editReq.LocaleName.ZH,
			"th_name":    editReq.LocaleName.TH,
			"en_name":    editReq.LocaleName.EN,
			"zh_tw_name": editReq.LocaleName.ZHTW,
			"ja_name":    editReq.LocaleName.JA,
			"ko_name":    editReq.LocaleName.KO,
			"my_name":    editReq.LocaleName.MY,
			"tr_name":    editReq.LocaleName.TR,
			"sv_name":    editReq.LocaleName.SV,
		})
		// 修改商品加料
		tx.Model(&model.ProductSauce{}).Where("uuid = ?", editReq.Uuid).Updates(map[string]any{
			"name":  name,
			"price": editReq.Price,
		})
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
			bomRepo.CreateProductBoms(newBoms)
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
		tx.Model(&model.ProductSauce{}).Where("uuid = ?", deleteReq.Uuid).Update("delete_time", time.Now().Unix())
		// 删除多语言名称
		tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", productSauce.MultiLanguageNameUuid).Update("delete_time", time.Now().Unix())
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
			tx.Model(&model.ProductSauce{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
				"sort": item.Sort,
			})
		}
		return nil
	})
	return err
}
