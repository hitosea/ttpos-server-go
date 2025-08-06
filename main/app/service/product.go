package service

import (
	"encoding/json"
	"slices"
	"sort"
	"strconv"
	"strings"
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
)

// IProductSrv 定义产品服务接口
type IProductSrv interface {
	GetProductList(ctx context.Context, req req.ProductListReq) (product_resp.ProductListWithPaginationResp, error)               // 获取产品列表
	GetProductCategoryList(dbId uint64) (product_resp.ProductCategoryListResp, error)                                             // 获取产品类别列表
	GetProductRecommendList(ctx context.Context, req req.ProductRecommendListReq) (*product_resp.ProductRecommendListResp, error) // 获取产品推荐列表
	SearchProducts(ctx context.Context, req req.ProductSearchReq) (*product_resp.ProductSearchResp, error)                        // 搜索商品
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

		if product.ProductType == constant.Yes {
			packageGroupList := make([]product_resp.ProductPackageGroup, 0)
			for _, group := range product.ProductPackageGroups {
				productList := make([]product_resp.PackageProductDetail, 0)
				for _, item := range group.ProductPackageGroupItems {
					flavor := getFlavor(item.ProductBom)            // 商品规格
					attributeGroups := getAttributeGroups(&product) // 商品属性组
					productDetail := product_resp.PackageProductDetail{
						Detail: product_resp.Product{
							Uuid:       item.ProductBom.Uuid,
							LocaleName: item.ProductPackage.MultiLanguageName.GetNames(),
							Image:      image,
							Unit:       unit,
							Price:      0, // 商品价格，套餐内目前是0元
							Flavors: product_resp.ProductFlavorList{
								List: []product_resp.ProductFlavor{flavor},
							},
							AttributeGroups: product_resp.ProductAttributeGroupList{
								List: attributeGroups,
							},
							Describe: product.Describe,
						},
					}
					productDetail.CanEdit = productDetail.GetCanEdit() // 是否可以编辑
					productList = append(productList, productDetail)
				}

				packageGroup := product_resp.ProductPackageGroup{
					Uuid:       group.Uuid,
					LocaleName: group.MultiLanguageName.GetNames(),
					Products: product_resp.ProductList{
						List: productList,
					},
				}
				packageGroup.IsFull = packageGroup.GetIsFull() // 是否选满
				packageGroupList = append(packageGroupList, packageGroup)
			}
			list = append(list, product_resp.Product{
				Uuid:              product.Uuid,
				Image:             image,
				LocaleName:        product.MultiLanguageName.GetNames(),
				Unit:              unit,
				Price:             product.Price,
				LimitNum:          product.LimitNum,
				CategoryUuid:      product.CategoryUuid,
				FirstCategoryUuid: product.ProductCategory.GetFirstCategoryUuid(),
				Describe:          product.Describe,
				IsShowKitchen:     product.IsShowKitchen,
				ProductType:       product.ProductType,
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
