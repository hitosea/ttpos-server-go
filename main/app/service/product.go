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
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 导入状态常量
const (
	ImportStatusStart      = "start"      // 开始导入
	ImportStatusProcessing = "processing" // 正在处理
	ImportStatusFinish     = "finish"     // 导入完成
	ImportStatusError      = "error"      // 导入失败
)

// ImportProgressData 导入进度数据结构
type ImportProgressData struct {
	Time     int64               `json:"time"`     // 时间戳
	Status   string              `json:"status"`   // 状态：start|processing|finish|error
	Progress int                 `json:"progress"` // 进度百分比 (0-100)
	Total    int                 `json:"total"`    // 总数量
	Current  int                 `json:"current"`  // 当前处理数量
	Success  int                 `json:"success"`  // 成功数量
	Failed   int                 `json:"failed"`   // 失败数量
	Error    string              `json:"error"`    // 总体错误信息
	Errors   []ImportErrorDetail `json:"errors"`   // 详细错误列表
}

// ImportErrorDetail 导入错误详情
type ImportErrorDetail struct {
	Row     int    `json:"row"`     // 行号
	Message string `json:"message"` // 错误信息
}

// pushImportProgress 推送导入进度到前端
func (s *productSrv) pushImportProgress(companyUuid uint64, deviceSn string, data ImportProgressData) {
	data.Time = time.Now().Unix()
	go websocket.PushClient(companyUuid, websocket.SourceShop, deviceSn, websocket.IMPORT_PRODUCT, map[string]any{
		"time":     data.Time,
		"status":   data.Status,
		"progress": data.Progress,
		"total":    data.Total,
		"current":  data.Current,
		"success":  data.Success,
		"failed":   data.Failed,
		"error":    data.Error,
		"errors":   data.Errors,
	})
}

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
	UpdateProductFlavorErp(ctx context.Context, tx *gorm.DB) error                                                      // 更新商品规格到erp

	// 导入商品
	ImportProductList(ctx context.Context, req req.ProductImportListReq) (product_resp.ProductImportResp, error) // 导入商品列表
	ImportProduct(ctx context.Context, req req.ProductImportReq) error                                           // 导入商品

	GetProductSingleList(ctx context.Context, req req.ProductSingleListReq) (*product_resp.ProductSingleListResp, error) // 获取单规格商品列表

	GetProductShopList(ctx context.Context, req req.ProductShopListReq) (*product_resp.ProductShopListResp, error)                 // 获取商品列表（商家端）
	SortProductShopList(ctx context.Context, req req.SortProductShopListReq) error                                                 // 排序商品列表
	GetProductDetail(ctx context.Context, req req.ProductDetailReq) (*product_resp.ProductDetailResp, error)                       // 获取商品详情
	ProductShopStatus(ctx context.Context, req req.ProductShopStatusReq) error                                                     // 修改商品状态
	ProductTaxList(ctx context.Context) product_resp.ProductTaxListResp                                                            // 获取商品税类列表
	AddProductShop(ctx context.Context, req req.ProductShopAddReq) error                                                           // 添加商品
	EditProductShop(ctx context.Context, req req.ProductShopEditReq) (*product_resp.ProductEditResp, []string, error)              // 编辑商品
	DeleteProductShop(ctx context.Context, req req.ProductShopDeleteReq) (*product_resp.ProductDeleteResp, error)                  // 删除商品
	ProductShopChangePrice(ctx context.Context, req req.ProductShopChangePriceReq) error                                           // 商品改价
	AddProductPackage(ctx context.Context, tx *gorm.DB, req req.ProductShopAddReq, price float64) (*AddProductPackageRes, error)   // 添加商品包
	EditProductPackage(ctx context.Context, tx *gorm.DB, req req.ProductShopEditReq, price float64) (*AddProductPackageRes, error) // 编辑商品包
	SaveProductPackageBom(ctx context.Context, tx *gorm.DB, params SaveProductPackageBomParams) error                              // 保存商品bom
	SaveProductPackageAttribute(tx *gorm.DB, param []CheckProductAttributeGroupParam, productPackageUuid uint64) error             // 保存商品属性
	SaveProductPackageGroup(tx *gorm.DB, groupList []CheckProductPackageGroupResult, productPackageUuid uint64) error              // 保存商品套餐组

	// 分批类型管理
	GetBatchTagList(ctx context.Context, req req.BatchTagListReq) (*product_resp.BatchTagList, error) // 获取分批类型列表
	GetBatchTag(ctx context.Context, req req.BatchTagReq) (*product_resp.BatchTagDetail, error)       // 获取分批类型详情
	AddBatchTag(ctx context.Context, req req.BatchTagAddReq) error                                    // 添加分批类型
	EditBatchTag(ctx context.Context, req req.BatchTagEditReq) error                                  // 编辑分批类型
	DeleteBatchTag(ctx context.Context, req req.BatchTagDeleteReq) error                              // 删除分批类型
	SortBatchTag(ctx context.Context, req req.BatchTagSortReq) error                                  // 排序分批类型
	GetBatchTagColorUsage(ctx context.Context) (*product_resp.BatchTagColorUsageList, error)          // 获取色块被选择情况
	SaveBatchProduct(ctx context.Context, req req.SaveBatchProductReq) error                          // 保存分批商品

	SyncProductShopCategory(ctx context.Context) error   // 同步产品分类
	SyncProductTax(ctx context.Context) error            // 同步商品税类
	SyncUnit(ctx context.Context) error                  // 获取总部最新单位数据
	SyncProductFlavor(ctx context.Context) error         // 同步商品规格
	SyncSauce(ctx context.Context) error                 // 获取总部最新加料数据
	SyncAttributeGroup(ctx context.Context) error        // 获取总部最新属性组数据
	SyncProduct(ctx context.Context) error               // 同步商品
	SyncProductStockByBomCard(ctx context.Context) error // 计算所有关联成本卡的商品的库存
	SyncProductPackageImage(ctx context.Context) error   // 同步商品包图片
}

type productSrv struct {
	dbm          *database.DBManager // 数据库管理器
	localeSrv    ILocaleSrv          // 多语言名称服务
	settingSrv   setting.ISrv        // 设置服务
	cache        cache.Cache         // 缓存
	translateSrv ITranslateSrv       // 翻译服务
	systemLock   lock.Lock           // 系统锁
}

func NewProductSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, cache cache.Cache, translateSrv ITranslateSrv) IProductSrv {
	return NewProductSrvImpl(dbm, localeSrv, settingSrv, cache, translateSrv)
}

func NewProductSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, cache cache.Cache, translateSrv ITranslateSrv) IProductSrv {
	return &productSrv{
		dbm:          dbm,
		localeSrv:    localeSrv,
		settingSrv:   settingSrv,
		cache:        cache,
		translateSrv: translateSrv,
		systemLock:   lock.NewSystemLock(),
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

			packageItem := product_resp.Product{
				Uuid:       product.Uuid,
				Image:      image,
				LocaleName: product.MultiLanguageName.GetNames(),
				Unit:       unit,
				Price: func() float64 {
					if len(flavors) > 0 {
						return flavors[0].Price
					}
					return 999999.5
				}(),
				LimitNum:            product.LimitNum,
				CategoryUuid:        product.CategoryUuid,
				SpecialCategoryUuid: product.SpecialCategoryUuid,
				FirstCategoryUuid:   product.ProductCategory.GetFirstCategoryUuid(),
				Describe:            product.Describe,
				IsShowKitchen:       product.IsShowKitchen,
				ProductType:         product.ProductType,
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
			}
			if len(flavors) > 0 {
				packageItem.Price = flavors[0].Price
				list = append(list, packageItem)
			} else {
				logger.Logger.Error("套餐没有规格，无法显示价格", zap.String("name", product.MultiLanguageName.GetNames().EN), zap.Any("product", product))
			}
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
						Uuid:        child.Uuid,
						Name:        child.MultiLanguageName.GetNameByLang(language),
						Code:        child.Code,
						ParentUuid:  child.ParentUuid,
						IsSpecial:   child.IsSpecial == 1,
						Sort:        child.Sort,
						Status:      child.Status,
						IsEditable:  isEditable(ctx, child.HeadquarterUuid),
						CategoryKey: child.CategoryKey,
						Children: product_resp.ProductShopCategoryListResp{
							List: make([]product_resp.ProductShopCategory, 0),
						},
					})
				}
			}
			list = append(list, product_resp.ProductShopCategory{
				Uuid:        category.Uuid,
				Name:        category.MultiLanguageName.GetNameByLang(language),
				Code:        category.Code,
				ParentUuid:  category.ParentUuid,
				IsSpecial:   category.IsSpecial == 1,
				Sort:        category.Sort,
				Status:      category.Status,
				IsEditable:  isEditable(ctx, category.HeadquarterUuid),
				CategoryKey: category.CategoryKey,
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
		Code:         productCategory.Code,
		IsEditable:   isEditable(ctx, productCategory.HeadquarterUuid),
		CategoryKey:  productCategory.CategoryKey,
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
	// 大写编码
	addReq.Code = strings.ToUpper(addReq.Code)
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

	// 检查分类编码是否已存在
	if addReq.Code != "" {
		if exist := productRepo.CheckProductCategoryCodeExist(addReq.Code, 0); exist {
			return errors.New("分类编码已存在")
		}
	}

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
			Code:                  addReq.Code,
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
	// 大写编码
	editReq.Code = strings.ToUpper(editReq.Code)
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
	changeCode := false // 是否修改了分类编码
	if editReq.Code != productCategory.Code {
		changeCode = true
	}
	if changeCode {
		if editReq.Code != "" {
			if exist := productRepo.CheckProductCategoryCodeExist(editReq.Code, editReq.Uuid); exist {
				return errors.New("分类编码已存在")
			}
		}
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
			"code":        editReq.Code,
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
		if ctx.GetCompany().IsOpenErp() {
			if changeCode {
				// 获取关联的商品。更新商品（某规格）的分类编码、套餐的分类编码
				productRepo := repository.NewProductRepo(tx)
				products, err := productRepo.GetProducts(
					commonRepo.WhereBySoftDelete(),
					productRepo.WhereCategoryUuid(editReq.Uuid),
					repository.CommonRepo.Preload(
						repository.WithPreload{
							Query: "ProductBoms",
						},
					),
				)
				if err != nil {
					return errors.WithMessage(err, "获取商品失败")
				}
				if len(products) > 0 {
					erpSrv := erp.NewIErpSrv(s.dbm)
					for _, product := range products {
						for _, productBom := range product.ProductBoms {
							productMultiLanguageName := model.NewMultiLanguageName(product.Name)
							productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
							if err != nil {
								return errors.WithMessage(err, "翻译失败")
							}

							multiLanguageName := model.NewMultiLanguageName(productBom.Name)
							enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
							if err != nil {
								return errors.WithMessage(err, "翻译失败")
							}
							productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(product.UnitUuid))
							if errGetUnit != nil {
								return errors.WithMessage(errGetUnit, "获取商品单位失败")
							}

							localeName := language.JsonToLocaleResponse(editReq.LocaleName.ToJson())
							classification, err := s.getEnName(ctx, *localeName)
							if err != nil {
								return errors.WithMessage(err, "翻译失败")
							}
							if productBom.IsFlavor() {
								itemName := fmt.Sprintf("%s-%s", productEnName, enName)
								if _, err := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
									ItemName:          itemName,
									StockUom:          productUnit.ErpnextUom,
									ItemCode:          productBom.ErpCode,
									TemplateItemCode:  product.ErpCode,
									ItemSpecification: enName,
									Classification:    classification,
									ClassificationCode: func() string {
										if editReq.Code != "" {
											return editReq.Code
										}
										return " " // 如果分类编码为空，则设置为空
									}(),
								}); err != nil {
									return errors.WithMessage(err, "更新分类编码到erp失败")
								}
							}
							if productBom.IsPackageFlavor() {
								itemName := fmt.Sprintf("%s-%s", productEnName, enName)
								if _, err := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
									ItemName:          itemName,
									StockUom:          productUnit.ErpnextUom,
									ItemCode:          productBom.ErpCode,
									TemplateItemCode:  product.ErpCode,
									ItemSpecification: enName,
									Classification:    classification,
									ClassificationCode: func() string {
										if editReq.Code != "" {
											return editReq.Code
										}
										return " " // 如果分类编码为空，则设置为空
									}(),
								}); err != nil {
									return errors.WithMessage(err, "更新分类编码到erp失败")
								}
							}
						}
					}
				}
			}
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
func (s *productSrv) DeleteProductShop(ctx context.Context, request req.ProductShopDeleteReq) (*product_resp.ProductDeleteResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(db)
	setting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return nil, errors.New("获取公司设置失败")
	}

	product, err := productRepo.GetProduct(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(request.Uuid),
		productRepo.WithProductBoms(
			commonRepo.WhereBySoftDelete(),
		),
		productRepo.WithProductUnit(),
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "ProductBoms.ProductSauce",
			},
			repository.WithPreload{
				Query: "ProductPackageGroups",
			},
		),
	)
	if product.Uuid == 0 || err != nil {
		return nil, errors.New("商品不存在")
	}

	productPackageGroupItems, err := productPackageGroupRepo.GetProductPackageGroupItems(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByRelatedUuid(request.Uuid),
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
		productBomRepo := repository.NewProductBomRepo(tx) // 避免事务失效
		// 如果删除的是套餐
		if product.IsPackage() {
			// 删除ttpos_product_package_group_item、ttpos_product_package_group表的记录
			for _, productPackageGroup := range product.ProductPackageGroups {
				err := tx.Model(&model.ProductPackageGroupItem{}).Where("product_package_group_uuid = ?", productPackageGroup.Uuid).Updates(map[string]any{
					"delete_time": time.Now().Unix(),
				}).Error
				if err != nil {
					return err
				}
			}
			err = tx.Model(&model.ProductPackageGroup{}).Where("product_package_uuid = ?", request.Uuid).Updates(map[string]any{
				"delete_time": time.Now().Unix(),
			}).Error
			if err != nil {
				return err
			}
		}

		// 删除商品包
		err := tx.Model(&model.ProductPackage{}).Where("uuid = ?", request.Uuid).Updates(map[string]any{
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
			commonRepo.WhereByProductPackageUuid(request.Uuid),
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

		// 将商品模板和规格禁售
		if ctx.GetCompany().IsOpenErp() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			if !product.IsPackage() {
				if err := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
					ItemCode:   product.ErpCode,
					NotForSale: true,
					Disabled:   product.Status == constant.ProductStatusOffSale,
				}); err != nil {
					return errors.WithMessage(err, "设置商品模板禁售失败")
				}
				for _, productBom := range product.ProductBoms {
					if productBom.IsFlavor() {
						if err := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
							ItemCode:     productBom.ErpCode,
							NotForSale:   true,
							InternalCode: productBom.InternalCode,
							Disabled:     product.Status == constant.ProductStatusOffSale,
						}); err != nil {
							return errors.WithMessage(err, "设置商品规格禁售失败")
						}
					}
				}
			} else {
				items := []req.DeleteProductErpItemReq{}
				multiLanguageName := model.NewMultiLanguageName(product.Name)
				enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
				if err != nil {
					return errors.WithMessage(err, "翻译失败")
				}
				erpCode := product.GetPackageProductBom().ErpCode
				items = append(items, req.DeleteProductErpItemReq{
					ItemCode: erpCode,
					ItemName: enName,
					StockUom: product.ProductUnit.ErpnextUom,
				})
				if err := erpSrv.DeleteProduct(ctx, req.DeleteProductErpReq{
					Items: items,
				}); err != nil {
					return errors.WithMessage(err, "删除商品到erp失败")
				}
			}
		}
		return nil
	})

	if err != nil {
		logger.Logger.Error("删除商品失败", zap.Any("func", "DeleteProductShop"), zap.Any("params", request), zap.Error(err))
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
	if productCategory.CategoryKey == "all" {
		return errors.New("不可删除的分类")
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

// SyncProductShopCategory 同步产品分类
func (s *productSrv) SyncProductShopCategory(ctx context.Context) error {
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	if !company.IsOpenErp() {
		return errors.New("公司未授权erp")
	}
	if !companySetting.IsSubShop() {
		return nil
	}
	commonRepo := repository.NewCommonRepo()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	productRepo := repository.NewProductRepo(headquarterDB)
	categories, err := productRepo.GetProductCategoryList(
		commonRepo.WhereByCategoryKey(""),
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithSort("ASC"),
		commonRepo.WhereByHeadquarterUuid(0),
		productRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
	)
	if err != nil {
		return errors.WithMessage(err, "获取总部产品分类失败")
	}
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	err = subShopDB.Transaction(func(tx *gorm.DB) error {
		productRepo = repository.NewProductRepo(tx)
		categoryRepo := repository.NewProductCategoryRepo(tx)
		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
		for _, category := range categories {
			subShopCategory, err := categoryRepo.GetProductCategory(
				commonRepo.WhereByUuid(category.Uuid),
				commonRepo.WhereByHeadquarterUuid(companySetting.HeadquarterUuid),
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil || subShopCategory.Uuid == 0 {
				time := time.Now().Unix()
				multiLanguageName := model.MultiLanguageName{
					BaseModel: model.BaseModel{
						Uuid:       category.MultiLanguageName.Uuid,
						CreateTime: time,
						UpdateTime: time,
					},
					EnName:   category.MultiLanguageName.EnName,
					ZhName:   category.MultiLanguageName.ZhName,
					ZhTwName: category.MultiLanguageName.ZhTwName,
					ThName:   category.MultiLanguageName.ThName,
					MyName:   category.MultiLanguageName.MyName,
					JaName:   category.MultiLanguageName.JaName,
					KoName:   category.MultiLanguageName.KoName,
					TrName:   category.MultiLanguageName.TrName,
					SvName:   category.MultiLanguageName.SvName,
				}
				multiLanguageNameRepo.CreateMultiLanguageName(multiLanguageName)
				maxSort, err := productRepo.GetProductCategoryMaxSort(
					commonRepo.WhereBySoftDelete(),
					productRepo.WhereParentUuid(category.ParentUuid),
					productRepo.WhereByIsSpecial(utils.IfUint(category.IsSpecial == 1, 1, 0)),
				)
				if err != nil {
					return errors.WithMessage(err, "获取一级分类最大排序失败")
				}
				sort := uint(maxSort + 1)
				_, err = categoryRepo.CreateProductCategory(model.ProductCategory{
					BaseModel: model.BaseModel{
						Uuid:       category.Uuid,
						CreateTime: category.CreateTime,
						UpdateTime: category.UpdateTime,
					},
					Name:                  category.Name,
					MultiLanguageNameUuid: category.MultiLanguageName.Uuid,
					Status:                category.Status,
					ParentUuid:            category.ParentUuid,
					IsSpecial:             category.IsSpecial,
					Sort:                  sort,
					Code:                  category.Code,
					HeadquarterUuid:       companySetting.HeadquarterUuid,
				})
				if err != nil {
					return errors.WithMessage(err, "创建分类失败")
				}
			} else {
				changeCode := false // 是否修改了分类编码
				if category.Code != subShopCategory.Code {
					changeCode = true
				}
				if changeCode {
					if category.Code != "" {
						if exist := productRepo.CheckProductCategoryCodeExist(category.Code, category.Uuid); exist {
							return errors.New("分类编码已存在")
						}
					}
				}
				err = categoryRepo.UpdateProductCategory(subShopCategory.ID, model.ProductCategory{
					BaseModel: model.BaseModel{
						UpdateTime: category.UpdateTime,
					},
					Name:                  category.Name,
					MultiLanguageNameUuid: subShopCategory.MultiLanguageNameUuid,
					Status:                category.Status,
					ParentUuid:            category.ParentUuid,
					IsSpecial:             category.IsSpecial,
					Sort:                  category.Sort,
					Code:                  category.Code,
				})
				if err != nil {
					return errors.WithMessage(err, "更新分类失败")
				}
				if ctx.GetCompany().IsOpenErp() {
					if changeCode {
						// 获取关联的商品。更新商品（某规格）的分类编码、套餐的分类编码
						productRepo := repository.NewProductRepo(tx)
						products, err := productRepo.GetProducts(
							commonRepo.WhereBySoftDelete(),
							productRepo.WhereCategoryUuid(category.Uuid),
							commonRepo.WhereByHeadquarterUuid(0), // 只查询子店自己的商品
							productRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
						)
						if err != nil {
							return errors.WithMessage(err, "获取商品失败")
						}
						if len(products) > 0 {
							erpSrv := erp.NewIErpSrv(s.dbm)
							for _, product := range products {
								for _, productBom := range product.ProductBoms {
									productMultiLanguageName := model.NewMultiLanguageName(product.Name)
									productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
									if err != nil {
										return errors.WithMessage(err, "翻译失败")
									}

									multiLanguageName := model.NewMultiLanguageName(productBom.Name)
									enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
									if err != nil {
										return errors.WithMessage(err, "翻译失败")
									}
									productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(product.UnitUuid))
									if errGetUnit != nil {
										return errors.WithMessage(errGetUnit, "获取商品单位失败")
									}

									localeName := language.JsonToLocaleResponse(category.Name)
									classification, err := s.getEnName(ctx, *localeName)
									if err != nil {
										return errors.WithMessage(err, "翻译失败")
									}
									if productBom.IsFlavor() {
										itemName := fmt.Sprintf("%s-%s", productEnName, enName)
										if _, err := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
											ItemName:          itemName,
											StockUom:          productUnit.ErpnextUom,
											ItemCode:          productBom.ErpCode,
											TemplateItemCode:  product.ErpCode,
											ItemSpecification: enName,
											Classification:    classification,
											ClassificationCode: func() string {
												if category.Code != "" {
													return category.Code
												}
												return " " // 如果分类编码为空，则设置为空
											}(),
										}); err != nil {
											return errors.WithMessage(err, "更新分类编码到erp失败")
										}
									}
									if productBom.IsPackageFlavor() {
										itemName := fmt.Sprintf("%s-%s", productEnName, enName)
										if _, err := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
											ItemName:          itemName,
											StockUom:          productUnit.ErpnextUom,
											ItemCode:          productBom.ErpCode,
											TemplateItemCode:  product.ErpCode,
											ItemSpecification: enName,
											Classification:    classification,
											ClassificationCode: func() string {
												if category.Code != "" {
													return category.Code
												}
												return " " // 如果分类编码为空，则设置为空
											}(),
										}); err != nil {
											return errors.WithMessage(err, "更新分类编码到erp失败")
										}
									}
								}
							}
						}
					}
				}
			}
		}
		// 删除普通分类缓存
		tag := cryptor.Md5String(fmt.Sprintf("category%d%d%d", ctx.GetCompanyUuid(), 1, 1))
		err = cache.NewTaggedCache(s.cache).TagClear(tag)
		if err != nil {
			return err
		}
		// 删除特殊分类缓存
		tag = cryptor.Md5String(fmt.Sprintf("category%d%d%d", ctx.GetCompanyUuid(), 0, 1))
		err = cache.NewTaggedCache(s.cache).TagClear(tag)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "同步商品分类失败")
	}
	return nil
}

// SyncProductTax 同步商品税类
func (s *productSrv) SyncProductTax(ctx context.Context) error {
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	if !company.IsOpenErp() {
		return errors.New("公司未授权erp")
	}
	if !companySetting.IsSubShop() {
		return nil
	}
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	taxRepo := repository.NewTaxRepo(headquarterDB)
	taxes, err := taxRepo.GetTaxCategoryList()
	if err != nil {
		return errors.WithMessage(err, "获取总部商品税类失败")
	}
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	err = subShopDB.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		taxRepo = repository.NewTaxRepo(tx)
		for _, tax := range taxes {
			subShopTax, _ := taxRepo.GetTaxCategory(
				commonRepo.WhereByUuid(tax.Uuid),
				commonRepo.WhereByHeadquarterUuid(companySetting.HeadquarterUuid),
			)
			if subShopTax.Uuid == 0 {
				err = taxRepo.CreateTax(model.Tax{
					BaseModel: model.BaseModel{
						Uuid:       tax.Uuid,
						CreateTime: tax.CreateTime,
						UpdateTime: tax.UpdateTime,
					},
					Name:            tax.Name,
					TaxRate:         tax.TaxRate,
					HeadquarterUuid: companySetting.HeadquarterUuid,
				})
				if err != nil {
					return errors.WithMessage(err, "创建商品税类失败")
				}
			} else {
				taxRepo.UpdataTax(
					map[string]any{
						"name":        tax.Name,
						"tax_rate":    tax.TaxRate,
						"delete_time": constant.NotDeleted,
					},
					commonRepo.WhereByUuid(subShopTax.Uuid),
				)
			}
		}
		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "同步商品税类失败")
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
			IsEditable:          isEditable(ctx, unit.HeadquarterUuid),
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
		IsEditable: isEditable(ctx, unit.HeadquarterUuid),
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
	db.Model(&model.ProductUnit{}).Scopes(repository.NotDeleted, repository.ExcludeHeadquarter).Select("ifnull(max(sort), 0)").Scan(&maxSort)

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

		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		// 开启了ERP，并且是TTPOS站点，同步到ERPNext
		if company.IsOpenErp() {
			enName, err := s.getEnName(ctx, addReq.LocaleName)
			if err != nil {
				return errors.WithMessage(errors.New("翻译失败"), err.Error())
			}
			erpUom := company.Name + "-" + enName
			err = erp.NewIErpSrv(s.dbm).SaveUom(ctx.GetContext(), req.SaveUomReq{
				SiteCode:    companySetting.ErpnextSiteCode,
				CompanyAbbr: companySetting.ErpnextCompanyAbbr,
				Branch:      companySetting.ErpnextBranchName,
				UomName:     erpUom,
				AliasName:   enName,
			})
			if err != nil {
				return errors.WithMessage(errors.New("同步单位到erp失败"), err.Error())
			}
			err = tx.Model(&model.ProductUnit{}).Where("uuid = ?", productUnit.Uuid).Update("erpnext_uom", erpUom).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存erp单位失败"), err.Error())
			}

			// 修改商品的单位UUID
			for _, productPackageUuid := range addReq.ProductPackageUuids {
				// 同步更新商品到erp
				if ctx.GetCompany().IsOpenErp() {
					productPackageRepo := repository.NewProductPackageRepo(tx)
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
						productMultiLanguageName := model.NewMultiLanguageName(productPackage.Name)
						productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						multiLanguageName := model.NewMultiLanguageName(productBom.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						itemName := fmt.Sprintf("%s-%s", productEnName, enName)
						_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
							ItemName: itemName,
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
		return nil
	})
	return err
}

func (s *productSrv) getEnName(ctx context.Context, locale dto.LocaleResponse) (string, error) {
	return GetEnName(ctx, s.settingSrv, locale)
}

func GetEnName(ctx context.Context, settingSrv setting.ISrv, locale dto.LocaleResponse) (string, error) {
	enName := locale.EN
	if enName != "" {
		return enName, nil
	}
	storeSetting, err := settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return "", errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
	}
	if len(storeSetting.Language) == 0 {
		return "", errors.New("门店未设置默认语言")
	}
	defaultLanguage := storeSetting.Language[0].Name
	res, err := utils.NewTranslateClient().Translate(ctx.GetContext(), []utils.TranslateItem{
		{
			Lang:    defaultLanguage,
			Content: locale.GetLocale(defaultLanguage),
		},
	})
	if err != nil {
		return "", errors.WithMessage(errors.New("翻译失败"), err.Error())
	}
	return res.Data[0].En, nil
}

func GetMultiLanguageName(ctx context.Context, enName string) (*dto.LocaleResponse, error) {
	companySetting := ctx.GetCompanySetting()
	defaultLanguage := companySetting.GetDefaultLanguage()
	res, err := utils.NewTranslateClient().Translate(ctx.GetContext(), []utils.TranslateItem{
		{
			Lang:    defaultLanguage,
			Content: enName,
		},
	})
	if err != nil {
		return &dto.LocaleResponse{
			EN:   enName,
			ZH:   enName,
			TH:   enName,
			ZHTW: enName,
			JA:   enName,
			KO:   enName,
			MY:   enName,
			TR:   enName,
			SV:   enName,
		}, nil
		// return nil, errors.WithMessage(errors.New("翻译失败"), err.Error())
	}
	return &dto.LocaleResponse{
		EN:   res.Data[0].En,
		ZH:   res.Data[0].Zh,
		TH:   res.Data[0].Th,
		ZHTW: res.Data[0].ZhTw,
		JA:   res.Data[0].Ja,
		KO:   res.Data[0].Ko,
		MY:   res.Data[0].My,
		TR:   res.Data[0].Tr,
		SV:   res.Data[0].Sv,
	}, nil
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

	if !isEditable(ctx, productUnit.HeadquarterUuid) {
		return errors.New("单位不可编辑")
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
								productMultiLanguageName := model.NewMultiLanguageName(productPackage.Name)
								productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								multiLanguageName := model.NewMultiLanguageName(productBom.Name)
								enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								erpSrv := erp.NewIErpSrv(s.dbm)
								itemName := fmt.Sprintf("%s-%s", productEnName, enName)
								_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
									ItemName: itemName,
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

	if err == nil {
		// 单位 - 将多语言名称uuid从待翻译集合中删除
		s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), productUnit.MultiLanguageNameUuid)
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
	if !isEditable(ctx, productUnit.HeadquarterUuid) {
		return errors.New("单位不可删除")
	}
	// 是否关联商品
	if len(productUnit.ProductPackages) > 0 {
		return errors.New("该单位下存在商品，不允许删除")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		// 开启了ERP，并且是TTPOS站点，同步到ERPNext
		if company.IsOpenErp() && productUnit.ErpnextUom != "" {
			err = erp.NewIErpSrv(s.dbm).DeleteUom(ctx.GetContext(), req.DeleteUomReq{
				SiteCode:    companySetting.ErpnextSiteCode,
				CompanyAbbr: companySetting.ErpnextCompanyAbbr,
				Branch:      companySetting.ErpnextBranchName,
				UomName:     productUnit.ErpnextUom,
			})
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return errors.WithMessage(errors.New("删除单位失败"), err.Error())
			}
		}
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
			productBomCardName, _ = repository.NewProductBomCardRepo(db).GetProductBomCardName(productSauce.ProductBomCardUuid)
		}
		productSauceListResp = append(productSauceListResp, product_resp.ProductSauceItem{
			Uuid:                productSauce.Uuid,
			Name:                productSauce.MultiLanguageName.GetNameByLang(language),
			Price:               productSauce.Price,
			Sort:                productSauce.Sort,
			ProductPackageCount: productSauce.ProductPackageCount,
			ProductBomCardUuid:  productSauce.ProductBomCardUuid,
			ProductBomCardName:  productBomCardName,
			IsEditable:          isEditable(ctx, productSauce.HeadquarterUuid),
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
		IsEditable: isEditable(ctx, productSauce.HeadquarterUuid),
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

	if len(addReq.ProductPackageUuids) > 0 {
		productRepo := repository.NewProductRepo(db)
		// 检查商品是否存在
		productPackages, err := productRepo.GetProductPackageListByUuids(addReq.ProductPackageUuids)
		if err != nil {
			return errors.WithMessage(errors.New("商品不存在"), err.Error())
		}
		if len(productPackages) != len(addReq.ProductPackageUuids) {
			return errors.New("商品不存在")
		}
	}

	// 获取当前最大的排序值
	var maxSort int
	db.Model(&model.ProductSauce{}).Scopes(repository.NotDeleted, repository.ExcludeHeadquarter).Select("ifnull(max(sort), 0)").Scan(&maxSort)

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

		// 同步新增加料到erp
		// v2.7 加料暂时不需要处理（本任务不从ERP同步属性跟加料以及TTPOS添加修改之后不同步到ERP，子店同步时，从ttpos的总部获取数据）
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

// isEditable 是否可编辑
func isEditable(_ context.Context, headquarterUuid uint64) bool {
	return headquarterUuid == 0
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
	if !isEditable(ctx, productSauce.HeadquarterUuid) {
		return errors.New("加料不可编辑")
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
	if err == nil {
		// 加料 - 将多语言名称uuid从待翻译集合中删除
		s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), productSauce.MultiLanguageNameUuid)
	}
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
	if !isEditable(ctx, productSauce.HeadquarterUuid) {
		return errors.New("加料不可删除")
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

		// 同步删除加料到erp
		if ctx.GetCompany().IsOpenErp() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
				ItemCode:   productSauce.ErpCode,
				NotForSale: true,
			})
			if errErp != nil {
				return errors.WithMessage(errErp, "同步删除加料到erp失败")
			}
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
				IsEditable: isEditable(ctx, productAttribute.HeadquarterUuid),
			})
		}
		productAttributeGroupListResp = append(productAttributeGroupListResp, product_resp.ProductAttributeGroupItem{
			Uuid:          productAttributeGroup.Uuid,
			Name:          productAttributeGroup.MultiLanguageName.GetNameByLang(language),
			AttributeName: strings.Join(attributeNames, "、"),
			Attributes:    attributeList,
			Sort:          productAttributeGroup.Sort,
			IsEditable:    isEditable(ctx, productAttributeGroup.HeadquarterUuid),
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
		IsEditable: isEditable(ctx, productAttributeGroup.HeadquarterUuid),
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
			Sort:       productAttribute.Sort,
			IsEditable: isEditable(ctx, productAttribute.HeadquarterUuid),
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
			HeadquarterUuid:     productFlavor.HeadquarterUuid,
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
		HeadquarterUuid: productFlavor.HeadquarterUuid,
	}, nil
}

// AddProductAttributeGroup 添加商品属性组
func (s *productSrv) AddProductAttributeGroup(ctx context.Context, addReq req.ProductAttributeGroupAddReq) error {
	// companySetting := ctx.GetCompanySetting()
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
	db.Model(&model.ProductAttributeGroup{}).Scopes(repository.NotDeleted, repository.ExcludeHeadquarter).Select("ifnull(max(sort), 0)").Scan(&maxSort)

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

	// company := ctx.GetCompany()
	// 开启了ERP，并且是TTPOS站点，同步到ERPNext
	// v2.7 本任务不从ERP同步属性跟加料以及TTPOS添加修改之后不同步到ERP，子店同步时，从ttpos的总部获取数据
	// if company.IsOpenErp() {
	// 	attributeValueList := []req.SaveAttributeValueReq{}
	// 	uuidErpValueNameMap := make(map[uint64]string)
	// 	for attributeUuid, locale := range uuidLocaleMap {
	// 		enValueName, err := s.getEnName(ctx, locale)
	// 		if err != nil {
	// 			return errors.WithMessage(errors.New("翻译失败"), err.Error())
	// 		}
	// 		erpValueName := company.Name + "-" + enValueName
	// 		uuidErpValueNameMap[attributeUuid] = erpValueName
	// 		attributeValueList = append(attributeValueList, req.SaveAttributeValueReq{
	// 			AttributeValue: erpValueName,
	// 			Abbr:           enValueName,
	// 		})
	// 	}
	// 	enGroupName, err := s.getEnName(ctx, addReq.LocaleName)
	// 	if err != nil {
	// 		return errors.WithMessage(errors.New("翻译失败"), err.Error())
	// 	}
	// 	erpGroupName := company.Name + "-" + enGroupName
	// 	err = erp.NewIErpSrv(s.dbm).SaveAttribute(ctx.GetContext(), req.SaveAttributeReq{
	// 		SiteCode:           companySetting.ErpnextSiteCode,
	// 		CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
	// 		Branch:             companySetting.ErpnextBranchName,
	// 		AttributeName:      erpGroupName,
	// 		AliasName:          enGroupName,
	// 		AttributeValueList: attributeValueList,
	// 	})
	// 	if err != nil {
	// 		return errors.WithMessage(errors.New("同步属性组到erp失败"), err.Error())
	// 	}
	// 	// 保存属性组 erpnext_attribute_group_name
	// 	err = db.Model(&model.ProductAttributeGroup{}).Where("uuid = ?", productAttributeGroup.Uuid).Update("erpnext_attribute_group_name", erpGroupName).Error
	// 	if err != nil {
	// 		return errors.WithMessage(errors.New("保存erp属性组失败"), err.Error())
	// 	}
	// 	// 保存属性值 erpnext_attribute_value
	// 	for attributeUuid, erpValueName := range uuidErpValueNameMap {
	// 		err = db.Model(&model.ProductAttribute{}).Where("uuid = ?", attributeUuid).Update("erpnext_attribute_value", erpValueName).Error
	// 		if err != nil {
	// 			return errors.WithMessage(errors.New("保存erp属性值失败"), err.Error())
	// 		}
	// 	}
	// }

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

	company := ctx.GetCompany()

	err = db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)
		productFlavorRepo := repository.NewProductFlavorRepo(tx)
		warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
		warehouseMonthlyFormRepo := repository.NewWarehouseMonthlyFormRepo(tx)
		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
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
		multiLanguageNameUuid, err := multiLanguageNameRepo.CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return errors.WithMessage(err, "保存多语言名称失败")
		}
		// 保存产品规格
		productFlavor := model.ProductFlavor{
			Name:                  addReq.LocaleName.ToJson(),
			MultiLanguageNameUuid: multiLanguageNameUuid,
			Sort:                  sort,
		}
		err = productFlavorRepo.CreateProductFlavor(productFlavor)
		if err != nil {
			return errors.WithMessage(err, "保存规格失败")
		}
		// 保存规格名称到erp
		if company.IsOpenErp() {
			err = s.UpdateProductFlavorErp(ctx, tx)
			if err != nil {
				return errors.WithMessage(err, "更新规格到erp失败")
			}
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
					commonRepo.Preload(
						repository.WithPreload{
							Query: "ProductCategory",
						},
					),
				)
				if productPackage.ID == 0 || err != nil {
					return errors.WithMessage(err, "商品不存在")
				}

				// 同步新增商品到erp
				erpCode := ""
				if company.IsOpenErp() {
					productMultiLanguageName := model.NewMultiLanguageName(productPackage.Name)
					productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					multiLanguageName := model.NewMultiLanguageName(addReq.LocaleName.ToJson())
					enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
					if errGetUnit != nil {
						return errors.WithMessage(errGetUnit, "获取商品单位失败")
					}

					localeName := language.JsonToLocaleResponse(productPackage.ProductCategory.Name)
					classification, err := s.getEnName(ctx, *localeName)
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					erpSrv := erp.NewIErpSrv(s.dbm)
					itemName := fmt.Sprintf("%s-%s", productEnName, enName)
					itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
						ItemName:           itemName,
						StockUom:           productUnit.ErpnextUom,
						TemplateItemCode:   productPackage.ErpCode,
						ItemSpecification:  enName,
						Classification:     classification,
						ClassificationCode: productPackage.ProductCategory.Code,
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
					StockNum:           99999999,
				})

				setting, err := s.settingSrv.GetCompanySetting(ctx)
				// 开启库存管理
				if setting.SaleStock == 1 {
					// 添加入库
					warehouseForm := model.WarehouseForm{
						FormNo:         warehouseFormRepo.GenerateWarehouseFormNo(setting.Timezone),
						Scene:          constant.WarehouseFormSceneAddStock,
						Num:            99999999,
						ProductBomUuid: uuid,
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
						ProductBomUuid: uuid,
						Stock:          99999999,
					}
					err = warehouseMonthlyFormRepo.CreateWarehouseMonthlyProductBomForm(warehouseMonthlyProductBomForm)
					if err != nil {
						return errors.WithMessage(err, "保存月初库存记录失败")
					}
				}
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

// UpdateProductFlavorErp 更新商品规格到erp
func (s *productSrv) UpdateProductFlavorErp(ctx context.Context, tx *gorm.DB) error {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(tx)
	productFlavorRepo := repository.NewProductFlavorRepo(tx)
	flavorList, err := productRepo.GetProductFlavorList([]repository.DBOption{
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByHeadquarterUuid(0),
		productRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
	}...)
	if err != nil {
		return errors.WithMessage(err, "获取规格列表失败")
	}
	companySetting := ctx.GetCompanySetting()
	groupName := fmt.Sprintf("%s-Specifications", companySetting.ErpnextCompanyAbbr)
	maxErpnextValueNo, err := productRepo.GetProductFlavorMaxErpnextValueNo(
		commonRepo.WhereByHeadquarterUuid(0),
	)
	if err != nil {
		return errors.WithMessage(err, "获取最大erpnext规格值编号失败")
	}
	valueList := make([]req.SaveErpFlavorValueReq, 0, len(flavorList))
	for _, flavor := range flavorList {
		enName, err := s.getEnName(ctx, flavor.MultiLanguageName.GetNames())
		if err != nil {
			return errors.WithMessage(err, "翻译失败")
		}
		if flavor.ErpnextValueNo == 0 {
			maxErpnextValueNo += 1
			valueName := fmt.Sprintf("%s-%s-%s", companySetting.ErpnextCompanyAbbr, enName, fmt.Sprintf("%04d", maxErpnextValueNo))
			err = productFlavorRepo.UpdateProductFlavor(map[string]any{
				"erpnext_group_name": groupName,
				"erpnext_value_name": valueName,
				"erpnext_alias_name": enName,
				"erpnext_value_no":   maxErpnextValueNo,
			}, commonRepo.WhereByUuid(flavor.Uuid))
			if err != nil {
				return errors.WithMessage(err, "更新erpnext规格值编号失败")
			}
			valueList = append(valueList, req.SaveErpFlavorValueReq{
				ValueName:      valueName,
				ValueAliasName: enName,
			})
		} else {
			valueList = append(valueList, req.SaveErpFlavorValueReq{
				ValueName:      flavor.ErpnextValueName,
				ValueAliasName: flavor.ErpnextAliasName,
			})
		}
	}

	erpSrv := erp.NewIErpSrv(s.dbm)
	err = erpSrv.SaveFlavor(ctx.GetContext(), req.SaveErpFlavorReq{
		SiteCode:       companySetting.ErpnextSiteCode,
		GroupName:      groupName,
		GroupAliasName: groupName,
		// Branch:         companySetting.ErpnextBranchName,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		ValueList:   valueList,
	})
	if err != nil {
		return errors.WithMessage(err, "同步规格到erp失败")
	}

	return nil
}

// EditProductAttributeGroup 编辑商品属性组
func (s *productSrv) EditProductAttributeGroup(ctx context.Context, editReq req.ProductAttributeGroupEditReq) error {
	// companySetting := ctx.GetCompanySetting()
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

	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := repository.NewProductRepo(db)

	var manualTranslatedUuids []uint64
	// 检查属性组是否存在
	attributeGroup, err := productRepo.GetProductAttributeGroup(
		productRepo.WhereUuid(editReq.Uuid),
		productRepo.WithProductAttributes(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("属性组不存在"), err.Error())
	}
	manualTranslatedUuids = append(manualTranslatedUuids, attributeGroup.MultiLanguageNameUuid)
	if !isEditable(ctx, attributeGroup.HeadquarterUuid) {
		return errors.New("属性组不可编辑")
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
		manualTranslatedUuids = append(manualTranslatedUuids, attribute.MultiLanguageNameUuid)
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
				err = tx.Model(&model.ProductAttribute{}).Where("uuid = ?", productAttribute.Uuid).Updates(map[string]any{
					"name": productAttribute.LocaleName.ToJson(),
					"sort": productAttribute.Sort,
				}).Error
				if err != nil {
					return errors.WithMessage(errors.New("保存属性值失败"), err.Error())
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
	} else {
		// 属性（组） - 将多语言名称uuid从待翻译集合中删除
		s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), manualTranslatedUuids...)
	}

	// company := ctx.GetCompany()
	// 开启了ERP，并且是TTPOS站点，同步到ERPNext
	// v2.7 本任务不从ERP同步属性跟加料以及TTPOS添加修改之后不同步到ERP，子店同步时，从ttpos的总部获取数据
	// if company.IsOpenErp() {
	// 	var valueList []req.SaveAttributeValueReq
	// 	uuidErpValueNameMap := make(map[uint64]string)
	// 	// 构建 valueList
	// 	for uuid, locale := range uuidLocaleMap {
	// 		enValueName, err := s.getEnName(ctx, locale)
	// 		if err != nil {
	// 			return errors.WithMessage(errors.New("翻译失败"), err.Error())
	// 		}
	// 		var erpValueName string
	// 		if v, ok := uuidAttributeMap[uuid]; !ok || v.ErpnextAttributeValue == "" { // 新增属性值，或erpnext_attribute_value为空
	// 			erpValueName = company.Name + "-" + enValueName
	// 			uuidErpValueNameMap[uuid] = erpValueName
	// 		} else { // 已存在属性值，且erpnext_attribute_value不为空，则使用旧值
	// 			erpValueName = v.ErpnextAttributeValue
	// 		}
	// 		valueList = append(valueList, req.SaveAttributeValueReq{
	// 			AttributeValue: erpValueName,
	// 			Abbr:           enValueName, // 总是获取最新别名
	// 		})
	// 	}
	// 	// 新属性组别名
	// 	enGroupName, err := s.getEnName(ctx, editReq.LocaleName)
	// 	if err != nil {
	// 		return errors.WithMessage(errors.New("翻译失败"), err.Error())
	// 	}
	// 	err = erp.NewIErpSrv(s.dbm).SaveAttribute(ctx.GetContext(), req.SaveAttributeReq{
	// 		SiteCode:           companySetting.ErpnextSiteCode,
	// 		CompanyAbbr:        companySetting.ErpnextCompanyAbbr,
	// 		Branch:             companySetting.ErpnextBranchName,
	// 		AttributeName:      attributeGroup.ErpnextAttributeGroupName,
	// 		AliasName:          enGroupName,
	// 		AttributeValueList: valueList,
	// 	})
	// 	if err != nil {
	// 		return errors.WithMessage(errors.New("同步属性值到erp失败"), err.Error())
	// 	}
	// 	for uuid, erpValueName := range uuidErpValueNameMap {
	// 		err = db.Model(&model.ProductAttribute{}).Where("uuid = ?", uuid).Update("erpnext_attribute_value", erpValueName).Error
	// 		if err != nil {
	// 			return errors.WithMessage(errors.New("保存erp属性值失败"), err.Error())
	// 		}
	// 	}
	// }

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

	lang := ctx.GetLanguage()
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
					if !slices.Contains(productNames, productPackage.MultiLanguageName.GetNameByLang(lang)) {
						productNames = append(productNames, productPackage.MultiLanguageName.GetNameByLang(lang))
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
						if !slices.Contains(packageNames, packageItem.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(lang)) {
							packageNames = append(packageNames, packageItem.ProductPackageGroup.ProductPackage.MultiLanguageName.GetNameByLang(lang))
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

	var manualTranslatedUuid uint64
	err := db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)
		warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
		warehouseMonthlyFormRepo := repository.NewWarehouseMonthlyFormRepo(tx)

		flavor, err := productRepo.GetProductFlavor(
			productRepo.WhereUuid(editReq.Uuid),
			commonRepo.WhereBySoftDelete(),
		)
		if err != nil || flavor.ID == 0 {
			return errors.WithMessage(err, "获取商品规格详情失败")
		}
		manualTranslatedUuid = flavor.MultiLanguageNameUuid
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

		company := ctx.GetCompany()
		companySetting, err := s.settingSrv.GetCompanySetting(ctx)
		if err != nil {
			return errors.WithMessage(err, "获取公司设置失败")
		}
		if company.IsOpenErp() {
			err = s.UpdateProductFlavorErp(ctx, tx)
			if err != nil {
				return errors.WithMessage(err, "更新规格到erp失败")
			}
		}

		// 更新关联商品包
		for _, item := range editReq.List {
			if item.IsDelete {
				productBom, _ := productRepo.GetProductBom(
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereByUuid(item.BomUuid),
					productRepo.WithProductPackage(commonRepo.WhereBySoftDelete()),
					productRepo.WithProductPackageProductUnit(commonRepo.WhereBySoftDelete()),
				)
				if productBom.ID == 0 {
					return errors.New("商品bom不存在")
				}
				// 删除商品BOM
				err := tx.Model(&model.ProductBom{}).Where("uuid = ?", productBom.Uuid).Updates(map[string]any{
					"delete_time": time.Now().Unix(),
				}).Error
				if err != nil {
					return err
				}
				// 删除商品规格的item
				if company.IsOpenErp() {
					languageName := model.NewMultiLanguageName(productBom.Name)
					enName, err := s.getEnName(ctx, languageName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					erpSrv := erp.NewIErpSrv(s.dbm)
					if err := erpSrv.DeleteProduct(ctx, req.DeleteProductErpReq{
						Items: []req.DeleteProductErpItemReq{
							{
								ItemCode: productBom.ErpCode,
								ItemName: enName,
								StockUom: productBom.ProductPackage.ProductUnit.ErpnextUom,
							},
						},
					}); err != nil {
						return errors.WithMessage(err, "删除商品规格的item失败")
					}
				}
				if companySetting.SaleStock == 1 {
					// 删除出库
					outFormUuid, _ := utils.GetID()
					warehouseForm := model.WarehouseOutForm{
						BaseModel: model.BaseModel{
							Uuid: outFormUuid,
						},
						FormNo:       warehouseFormRepo.GenerateWarehouseOutFormNo(companySetting.Timezone),
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
					if company.IsOpenErp() {
						productMultiLanguageName := model.NewMultiLanguageName(productPackage.Name)
						productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						multiLanguageName := model.NewMultiLanguageName(editReq.LocaleName.ToJson())
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}

						localeName := language.JsonToLocaleResponse(productPackage.ProductCategory.Name)
						classification, err := s.getEnName(ctx, *localeName)
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}

						erpSrv := erp.NewIErpSrv(s.dbm)
						itemName := fmt.Sprintf("%s-%s", productEnName, enName)
						itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
							ItemName:           itemName,
							StockUom:           productUnit.ErpnextUom,
							TemplateItemCode:   productPackage.ErpCode,
							ItemSpecification:  enName,
							Classification:     classification,
							ClassificationCode: productPackage.ProductCategory.Code,
						})
						if errErp != nil {
							return errors.WithMessage(errErp, "同步商品到erp失败")
						}
						erpCode = itemInfo.ItemCode
					}
					// 新增商品BOM
					uuid, _ := utils.GetID()
					err := tx.Create(&model.ProductBom{
						BaseModel: model.BaseModel{
							Uuid: uuid,
						},
						Price:              item.Price,
						Name:               editReq.LocaleName.ToJson(),
						ErpCode:            erpCode,
						Status:             int(productPackage.Status),
						ProductFlavorUuid:  flavor.Uuid,
						ProductPackageUuid: item.Uuid,
						StockNum:           99999999,
					}).Error
					if err != nil {
						return err
					}
					// 开启库存管理
					if companySetting.SaleStock == 1 {
						// 添加入库
						warehouseForm := model.WarehouseForm{
							FormNo:         warehouseFormRepo.GenerateWarehouseFormNo(companySetting.Timezone),
							Scene:          constant.WarehouseFormSceneAddStock,
							Num:            99999999,
							ProductBomUuid: uuid,
							OperatorUuid:   ctx.GetStaffUuid(),
						}
						err = warehouseFormRepo.CreateWarehouseFormRecord(warehouseForm)
						if err != nil {
							return errors.WithMessage(err, "保存入库单失败")
						}
						// 添加月初库存记录
						warehouseMonthlyProductBomForm := model.WarehouseMonthlyProductBomForm{
							Year:           utils.SetTimezone(companySetting.Timezone).Now().Year(),
							Month:          int(utils.SetTimezone(companySetting.Timezone).Now().Month()),
							Scene:          constant.WarehouseMonthlyFormSceneStart,
							ProductBomUuid: uuid,
							Stock:          99999999,
						}
						err = warehouseMonthlyFormRepo.CreateWarehouseMonthlyProductBomForm(warehouseMonthlyProductBomForm)
						if err != nil {
							return errors.WithMessage(err, "保存月初库存记录失败")
						}
					}
				} else {
					// 同步更新商品到erp
					if company.IsOpenErp() {
						productMultiLanguageName := model.NewMultiLanguageName(productPackage.Name)
						productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
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
						itemName := fmt.Sprintf("%s-%s", productEnName, enName)
						_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
							ItemName:          itemName,
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
	} else {
		// 规格 - 将多语言名称uuid从待翻译集合中删除
		s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), manualTranslatedUuid)
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
	if !isEditable(ctx, productAttribute.HeadquarterUuid) {
		return errors.New("属性值不可删除")
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
	if !isEditable(ctx, productAttributeGroup.HeadquarterUuid) {
		return errors.New("属性组不可删除")
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
	if productFlavor.HeadquarterUuid != 0 {
		return errors.New("无法删除总部商品规格")
	}
	// 判断商品规格是否关联了商品
	productBomCount, _ := productRepo.GetProductBomCount(
		commonRepo.WhereByProductFlavorUuid(productFlavor.Uuid),
		commonRepo.WhereBySoftDelete(),
	)
	if productBomCount > 0 {
		return errors.New("该规格已经关联了商品，不可删除")
	}

	// 获取公司
	company := ctx.GetCompany()

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

		// 同步规格到erp
		if company.IsOpenErp() {
			err = s.UpdateProductFlavorErp(ctx, tx)
			if err != nil {
				return errors.WithMessage(err, "更新规格到erp失败")
			}
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

// SyncProductFlavor 同步商品规格
func (s *productSrv) SyncProductFlavor(ctx context.Context) error {
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	if !company.IsOpenErp() {
		return errors.New("公司未授权erp")
	}

	// 获取erp规格列表
	erpFlavorList, err := erp.NewIErpSrv(s.dbm).GetFlavorList(ctx.GetContext(), req.GetErpFlavorListReq{
		SiteCode: companySetting.ErpnextSiteCode,
		// Branch:      companySetting.ErpnextBranchName,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
	})
	if err != nil {
		return errors.WithMessage(err, "获取erp规格列表失败")
	}

	var multiLanguageNameUuids []uint64
	var saveProductFlavorUuids []uint64

	db := s.dbm.GetDB(ctx.GetDbId())
	err = db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)
		productFlavorRepo := repository.NewProductFlavorRepo(tx)
		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)

		// 同步erp规格到本地
		maxSort, _ := productRepo.GetProductFlavorMaxSort(
			commonRepo.WhereByHeadquarterUuid(0),
		)
		for _, erpFlavor := range erpFlavorList.List {
			// 仅同步本店abbr-Specifications规格
			if erpFlavor.AttributeName != companySetting.ErpnextCompanyAbbr+"-Specifications" {
				continue
			}
			for _, erpFlavorValue := range erpFlavor.AttributeValueList {
				existsFlavor, _ := productFlavorRepo.GetProductFlavor(
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereByErpnextValueName(erpFlavorValue.AttributeValue),
				)
				if existsFlavor.Uuid == 0 {
					// 新增
					maxSort += 1
					mutilLanguageName := model.MultiLanguageName{
						EnName:   erpFlavorValue.Abbr,
						ZhName:   erpFlavorValue.Abbr,
						ZhTwName: erpFlavorValue.Abbr,
						ThName:   erpFlavorValue.Abbr,
						MyName:   erpFlavorValue.Abbr,
						JaName:   erpFlavorValue.Abbr,
						KoName:   erpFlavorValue.Abbr,
						TrName:   erpFlavorValue.Abbr,
						SvName:   erpFlavorValue.Abbr,
					}
					multiLanguageNameUuid, err := multiLanguageNameRepo.CreateMultiLanguageName(mutilLanguageName)
					if err != nil {
						return errors.WithMessage(err, "创建多语言名称失败")
					}
					multiLanguageNameUuids = append(multiLanguageNameUuids, multiLanguageNameUuid)
					// CFG-20 milliliter-0011，根据- 分割，取最后一个元素
					noArr := strings.Split(erpFlavorValue.AttributeValue, "-")
					no := ""
					if len(noArr) > 0 {
						no = noArr[len(noArr)-1] // 取最后一个元素
					}
					// 将no前面的0去掉
					no = strings.TrimLeft(no, "0")
					noInt, err := strconv.Atoi(no)
					if err != nil {
						// 转换no失败，跳过
						continue
					}
					newProductFlavor := model.ProductFlavor{
						Name:                  mutilLanguageName.ToJson(),
						MultiLanguageNameUuid: multiLanguageNameUuid,
						Sort:                  int(maxSort),
						ErpnextGroupName:      erpFlavor.AttributeName,
						ErpnextValueName:      erpFlavorValue.AttributeValue,
						ErpnextAliasName:      erpFlavorValue.Abbr,
						ErpnextValueNo:        noInt,
					}
					err = productFlavorRepo.CreateProductFlavor(newProductFlavor)
					if err != nil {
						return errors.WithMessage(err, "新增规格到本地失败")
					}
					saveProductFlavorUuids = append(saveProductFlavorUuids, newProductFlavor.Uuid)
				} else {
					// 更新
					err = productFlavorRepo.UpdateProductFlavor(map[string]any{
						"erpnext_group_name": erpFlavor.AttributeName,
						"erpnext_alias_name": erpFlavorValue.Abbr,
						"delete_time":        constant.NotDeleted,
					}, commonRepo.WhereByUuid(existsFlavor.Uuid))
					if err != nil {
						return errors.WithMessage(err, "更新规格到本地失败")
					}
					saveProductFlavorUuids = append(saveProductFlavorUuids, existsFlavor.Uuid)
				}
			}
		}

		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "同步erp规格到本地失败")
	}

	// 同步总部规格到子店
	if companySetting.IsSubShop() {
		headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(headquarterDb)

		headquarterFlavorList, err := productRepo.GetProductFlavorList(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByHeadquarterUuid(0),
			productRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
		)
		if err != nil {
			return errors.WithMessage(err, "获取总部规格列表失败")
		}

		delFlavorUuids := make([]uint64, 0)
		delMultiLanguageNameUuids := make([]uint64, 0)
		newFlavorList := make([]model.ProductFlavor, 0)
		newMultiLanguageNameList := make([]model.MultiLanguageName, 0)
		for _, flavor := range headquarterFlavorList {
			time := time.Now().Unix()
			newFlavorList = append(newFlavorList, model.ProductFlavor{
				BaseModel:             model.BaseModel{Uuid: flavor.Uuid, CreateTime: time, UpdateTime: time},
				Name:                  flavor.Name,
				MultiLanguageNameUuid: flavor.MultiLanguageName.Uuid,
				Sort:                  flavor.Sort,
				HeadquarterUuid:       companySetting.HeadquarterUuid,
				ErpnextGroupName:      flavor.ErpnextGroupName,
				ErpnextValueName:      flavor.ErpnextValueName,
				ErpnextAliasName:      flavor.ErpnextAliasName,
				ErpnextValueNo:        flavor.ErpnextValueNo,
			})
			newMultiLanguageNameList = append(newMultiLanguageNameList, model.MultiLanguageName{
				BaseModel: model.BaseModel{Uuid: flavor.MultiLanguageName.Uuid, CreateTime: time, UpdateTime: time},
				EnName:    flavor.MultiLanguageName.EnName,
				ZhName:    flavor.MultiLanguageName.ZhName,
				ZhTwName:  flavor.MultiLanguageName.ZhTwName,
				ThName:    flavor.MultiLanguageName.ThName,
				MyName:    flavor.MultiLanguageName.MyName,
				JaName:    flavor.MultiLanguageName.JaName,
				KoName:    flavor.MultiLanguageName.KoName,
				TrName:    flavor.MultiLanguageName.TrName,
				SvName:    flavor.MultiLanguageName.SvName,
			})
			delFlavorUuids = append(delFlavorUuids, flavor.Uuid)
			delMultiLanguageNameUuids = append(delMultiLanguageNameUuids, flavor.MultiLanguageName.Uuid)
		}

		db := s.dbm.GetDB(ctx.GetDbId())
		subFlavorList, err := productRepo.GetProductFlavorList(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByHeadquarterUuid(companySetting.HeadquarterUuid),
			productRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
		)
		if err != nil {
			return errors.WithMessage(err, "获取子店规格列表失败")
		}
		for _, flavor := range subFlavorList {
			delFlavorUuids = append(delFlavorUuids, flavor.Uuid)
			delMultiLanguageNameUuids = append(delMultiLanguageNameUuids, flavor.MultiLanguageName.Uuid)
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			productFlavorRepo := repository.NewProductFlavorRepo(tx)
			multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)

			if len(delFlavorUuids) > 0 {
				err = productFlavorRepo.DestroyProductFlavor(commonRepo.WhereInUuids(delFlavorUuids))
				if err != nil {
					return errors.WithMessage(err, "销毁子店规格失败")
				}
			}
			if len(delMultiLanguageNameUuids) > 0 {
				err = multiLanguageNameRepo.DestroyMultiLanguageName(commonRepo.WhereInUuids(delMultiLanguageNameUuids))
				if err != nil {
					return errors.WithMessage(err, "销毁子店多语言名称失败")
				}
			}
			if len(newFlavorList) > 0 {
				err = productFlavorRepo.CreateProductFlavorList(newFlavorList)
				if err != nil {
					return errors.WithMessage(err, "创建子店规格失败")
				}
			}
			if len(newMultiLanguageNameList) > 0 {
				err = multiLanguageNameRepo.CreateMultiLanguageNameList(newMultiLanguageNameList)
				if err != nil {
					return errors.WithMessage(err, "创建子店多语言名称失败")
				}
			}

			return nil
		})
		if err != nil {
			return errors.WithMessage(err, "同步总部规格到子店失败")
		}
	}

	// 添加多语言uuid到待翻译集合中
	if len(multiLanguageNameUuids) > 0 {
		if err := s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), multiLanguageNameUuids...); err != nil {
			logger.Logger.Error("同步基础规格添加多语言uuid到待翻译集合中失败", zap.Error(err))
		}
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
	companyUuid := ctx.GetCompanyUuid()
	deviceSn := ctx.GetDeviceSn()
	db := s.dbm.GetDB(ctx.GetDbId())
	language := ctx.GetLanguage()
	// 生成锁的key
	lockKey := fmt.Sprintf("%d_v02_import_product", companyUuid)

	// 用信道锁 禁止并发导入 - 按公司UUID加锁确保同一公司的商品导入操作不会并发执行
	if !s.systemLock.TryLockUuidString(lockKey) {
		return nil
	}

	// 验证条形码是否重复
	duplicateRows := reqs.GetBarcodeDuplicateRows()
	if len(duplicateRows) > 0 {
		s.systemLock.UnlockUuidString(lockKey)
		return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(duplicateRows[0]) + "]: " + i18n.Translate(language, "商品条码不能重复"))
	}

	// 预验证阶段 - 检查商品名称和条形码是否已存在
	for _, item := range reqs.List {
		// 验证是否已经存在
		productNameIsExist := repository.NewProductRepo(db).CheckMultiLanguageNameExist(item.LocaleName)
		if !productNameIsExist.IsNull() {
			s.systemLock.UnlockUuidString(lockKey)
			return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "商品名称已存在"))
		}
		// 验证条形码存在性检查
		if repository.NewProductRepo(db).CheckBarcodeExist(item.Barcode, 0) {
			s.systemLock.UnlockUuidString(lockKey)
			return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "商品条码已存在"))
		}
	}

	// 多规格合并
	lists := make(map[string]req.ProductShopAddReq)
	for _, item := range reqs.List {
		md5Key := item.LocaleName.GetMd5()

		// 创建新的规格项
		newFlavor := req.ProductShopAddFlavorReq{
			Uuid:         item.SkuUuid,
			Price:        item.ProductPrice,
			BarcodeValue: item.Barcode,
		}

		if _, exists := lists[md5Key]; !exists {
			// 如果商品不存在，创建新的商品记录
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
				NumType:         utils.IfInt(item.NumType == 1, 0, 1),
				DeductStockType: utils.IfInt(item.DeductStockType == 2, 0, 1),
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
				Row:     item.Row,
				Flavors: []req.ProductShopAddFlavorReq{newFlavor}, // 直接添加新规格
			}
		} else {
			// 如果商品已存在，直接添加新规格到现有规格列表
			temp := lists[md5Key]
			temp.Flavors = append(temp.Flavors, newFlavor)
			lists[md5Key] = temp
		}
	}

	// 异步导入
	go func() {
		defer s.systemLock.UnlockUuidString(lockKey)

		totalCount := len(lists)
		progressData := ImportProgressData{
			Time:    time.Now().Unix(),
			Status:  ImportStatusStart,
			Total:   totalCount,
			Current: 0,
			Success: 0,
			Failed:  0,
			Errors:  make([]ImportErrorDetail, 0),
		}

		// 实际导入阶段
		time.Sleep(300 * time.Millisecond)
		s.pushImportProgress(companyUuid, deviceSn, progressData)

		// 重置进度数据
		progressData.Current = 0
		progressData.Success = 0
		progressData.Failed = 0
		progressData.Errors = make([]ImportErrorDetail, 0) // 重置错误列表，只保留导入阶段的错误
		progressData.Status = ImportStatusProcessing

		totalItems := len(lists)
		currentIndex := 0

		for _, item := range lists {
			currentIndex++
			progressData.Current = currentIndex
			progressData.Progress = 30 + int(float64(currentIndex)/float64(totalItems)*70) // 导入占70%进度，从30%开始

			err := s.AddProductShop(ctx, item)
			if err != nil {
				progressData.Failed++
				progressData.Errors = append(progressData.Errors, ImportErrorDetail{
					Row:     item.Row,
					Message: err.Error(),
				})
				logger.Logger.Error("导入商品失败",
					zap.Int("row", item.Row),
					zap.Error(err),
					zap.Uint64("companyUuid", companyUuid))
			} else {
				progressData.Success++
			}

			// 推送导入进度
			if currentIndex%5 == 0 || currentIndex == totalItems { // 每5条或最后一条推送进度
				s.pushImportProgress(companyUuid, deviceSn, progressData)
			}
		}

		// 推送最终结果
		progressData.Progress = 100
		if progressData.Failed > 0 {
			progressData.Status = ImportStatusError
			progressData.Error = fmt.Sprintf("导入完成，成功%d条，失败%d条", progressData.Success, progressData.Failed)
		} else {
			progressData.Status = ImportStatusFinish
			progressData.Error = fmt.Sprintf("导入成功，共处理%d条商品", progressData.Success)
		}

		// 延迟500毫秒
		time.Sleep(500 * time.Millisecond)
		progressData.Time = time.Now().Unix()
		s.pushImportProgress(companyUuid, deviceSn, progressData)
	}()

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
					productBomCardName, _ = repository.NewProductBomCardRepo(db).GetProductBomCardName(productBom.ProductBomCardUuid)
				}
				productItem := product_resp.ProductSingleListItemResp{
					Uuid:               productBom.Uuid,
					Name:               productPackage.MultiLanguageName.GetNames(),
					FlavorName:         productBom.ProductFlavor.MultiLanguageName.GetNames(),
					InternalCode:       productBom.InternalCode,
					CategoryUuid:       productPackage.CategoryUuid,
					ProductBomCardUuid: productBom.ProductBomCardUuid,
					ProductBomCardName: productBomCardName,
					HeadquarterUuid:    productPackage.HeadquarterUuid,
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
						Uuid:         productBom.Uuid,
						LocaleName:   productBom.ProductFlavor.MultiLanguageName.GetNames(),
						Price:        productBom.Price,
						InternalCode: productBom.InternalCode,
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

	productPrinterList, err := repository.NewProductPrinterRepo(db).GetProductPrintersByProductPackageUuid(productPackage.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品打印机列表失败")
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
		ProductPrinters: product_resp.ProductPrinterList{
			List: func() []product_resp.ProductPrinter {
				printers := make([]product_resp.ProductPrinter, 0, len(productPrinterList))
				for _, productPrinter := range productPrinterList {
					printers = append(printers, product_resp.ProductPrinter{
						Uuid:   productPrinter.Uuid,
						Name:   productPrinter.Name,
						Status: productPrinter.Status,
					})
				}
				return printers
			}(),
		},
		HeadquarterUuid: productPackage.HeadquarterUuid,
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
		productRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
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
		err = tx.Model(&model.ProductBom{}).Select("status").Where("product_package_uuid = ?", productPackage.Uuid).Updates(map[string]any{
			"status": req.Status,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "修改商品规格状态失败")
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
		company := ctx.GetCompany()
		if company.IsOpenErp() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			if !productPackage.IsPackage() {
				// 禁用商品模板
				err = erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
					ItemCode: productPackage.ErpCode,
					Disabled: *req.Status == 0,
				})
				if err != nil {
					return errors.WithMessage(err, "同步商品模板到erp失败")
				}
			}
			for _, productBom := range productPackage.ProductBoms {
				if productBom.IsFlavor() || productBom.IsPackageFlavor() {
					err = erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
						ItemCode: productBom.ErpCode,
						Disabled: *req.Status == 0,
					})
					if err != nil {
						return errors.WithMessage(err, "同步商品bom到erp失败")
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
	// 检查商品规格内部编码
	for idx, flavor := range req.Flavors {
		if flavor.InternalCode != "" {
			// 大写编码
			internalCode := strings.ToUpper(strings.TrimSpace(flavor.InternalCode))
			req.Flavors[idx].InternalCode = internalCode
			if repository.NewProductRepo(db).CheckProductFlavorInternalCodeExist(internalCode, flavor.Uuid) {
				return errors.New("内部编码已存在")
			}
		}
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
				InternalCode: flavor.InternalCode,
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
		if req.Package.InternalCode != "" {
			if repository.NewProductRepo(db).CheckProductFlavorInternalCodeExist(req.Package.InternalCode, 0) {
				return errors.New("内部编码已存在")
			}
		}
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
					Name:         req.LocaleName.ToJson(),
					Price:        packageResult.Price,
					InternalCode: req.Package.InternalCode,
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
		err = s.SaveProductPackageBom(ctx, tx, SaveProductPackageBomParams{
			ProductPackageUuid: productPackageUuid,
			UnitUuid:           req.UnitUuid,
			TemplateItemCode:   erpCode,
			FlavorListResult:   flavorListResult,
			SauceResult:        sauceListResult,
			CategoryUuid:       req.CategoryUuid,
		})
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
		} else if ctx.Version(context.GTE, "2.7.0") {
			// 检查商品打印机
			if err := productCheckSrv.CheckProductPrinter(ctx, db, productPackageUuid, req.ProductPrinterUuids); err != nil {
				return errors.WithMessage(err)
			}
			// 新增商品包关联打印机
			err = repository.NewProductPrinterRepo(tx).CreateProductPackagePrinter(productPackageUuid, req.ProductPrinterUuids)
			if err != nil {
				return errors.WithMessage(err, "保存商品包关联打印机失败")
			}
			// 删除商品打印机列表缓存
			printer.NewPrinterRepo(ctx).DeleteProductPrinterListCache()
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
	// 检查商品规格内部编码
	for idx, flavor := range req.Flavors {
		if flavor.InternalCode != "" {
			// 大写编码
			internalCode := strings.ToUpper(flavor.InternalCode)
			req.Flavors[idx].InternalCode = internalCode
			if repository.NewProductRepo(db).CheckProductFlavorInternalCodeExist(flavor.InternalCode, flavor.BomUuid) {
				return nil, nil, errors.New("内部编码已存在")
			}
		}
	}

	// 商品专用检查
	flavorListResult := CheckProductFlavorResult{}
	sauceListResult := CheckProductSauceResult{}
	attributeListResult := []CheckProductAttributeGroupParam{}
	packageResult := CheckProductPackageResult{}
	if req.Type == constant.ProductTypeProduct { // 编辑商品
		// 商品规格, 必填
		var packageNames []string
		var flavorNames []string
		var flavors []CheckProductFlavorParam
		for _, flavor := range req.Flavors { // 商品规格
			flavors = append(flavors, CheckProductFlavorParam{
				Uuid:         flavor.Uuid,
				Price:        flavor.Price,
				BarcodeValue: flavor.BarcodeValue,
				InternalCode: flavor.InternalCode,
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
	} else { // 编辑套餐
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
					BomUuid:      bom.Uuid,
					Name:         req.LocaleName.ToJson(),
					InternalCode: req.Package.InternalCode,
					Price:        packageResult.Price,
					BarcodeValue: req.Flavors[0].BarcodeValue,
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
		err = s.SaveProductPackageBom(ctx, tx, SaveProductPackageBomParams{
			ProductPackageUuid: productPackageUuid,
			UnitUuid:           req.UnitUuid,
			TemplateItemCode:   erpCode,
			FlavorListResult:   flavorListResult,
			SauceResult:        sauceListResult,
			CategoryUuid:       req.CategoryUuid,
		})
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

			// 新增商品包关联打印机
			if ctx.Version(context.GTE, "2.7.0") {
				// 商品打印机
				if err := productCheckSrv.CheckProductPrinter(ctx, db, productPackageUuid, req.ProductPrinterUuids); err != nil {
					return errors.WithMessage(err)
				}
				// 新增商品包关联打印机
				err = repository.NewProductPrinterRepo(tx).CreateProductPackagePrinter(productPackageUuid, req.ProductPrinterUuids)
				if err != nil {
					return errors.WithMessage(err, "保存商品包关联打印机失败")
				}
				// 删除商品打印机列表缓存
				printer.NewPrinterRepo(ctx).DeleteProductPrinterListCache()
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
	// 处理外送端,如果外送端未开启, 套餐商品或者小数计价,则不显示外送端
	isShowDelivery := uint(req.Show.IsShowDelivery)
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsOpenRider() || req.Type == constant.ProductTypePackage || req.NumType == constant.ProductNumTypeDecimal {
		isShowDelivery = 0
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
		"is_show_delivery":      isShowDelivery,
		"price":                 price,
		"open_discount":         req.Discount.IsEnableMemberDiscount,
		"open_overall_discount": req.Discount.IsEnableOverallDiscount,
	}, commonRepo.WhereByUuid(productPackage.Uuid))
	if err != nil {
		return nil, errors.WithMessage(err, "保存商品包失败")
	}
	// 商品 - 将多语言名称uuid从待翻译集合中删除
	s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), productPackage.MultiLanguageNameUuid) // 商品

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
			productCategory, errGetCategory := repository.NewProductCategoryRepo(tx).GetProductCategory(
				commonRepo.WhereByUuid(request.CategoryUuid),
				repository.NewProductCategoryRepo(tx).WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
			)
			if errGetCategory != nil {
				return nil, errors.WithMessage(errGetCategory, "获取商品分类失败")
			}
			productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(request.UnitUuid))
			if errGetUnit != nil {
				return nil, errors.WithMessage(errGetUnit, "获取商品单位失败")
			}

			flavorUuids := make([]uint64, 0, len(request.Flavors))
			for _, v := range request.Flavors {
				flavorUuids = append(flavorUuids, v.Uuid)
			}
			flavorList, err := repository.NewProductFlavorRepo(tx).GetProductFlavorList(flavorUuids...)
			if err != nil {
				return nil, errors.WithMessage(err, "获取商品规格失败")
			}
			var flavors []req.Flavor
			for _, v := range flavorList {
				flavors = append(flavors, req.Flavor{
					Name:  v.ErpnextGroupName,
					Value: v.ErpnextValueName,
				})
			}
			classification := productCategory.MultiLanguageName.GetNames()
			getEnClassification, err := s.getEnName(ctx, classification)
			if err != nil {
				return nil, errors.WithMessage(err, "翻译失败")
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
				ItemName:           enName,
				StockUom:           productUnit.ErpnextUom,
				Classification:     getEnClassification,
				ClassificationCode: productCategory.Code,
				Flavors:            flavors,
			})
			if errErp != nil {
				return nil, errors.WithMessage(errErp, "同步商品到erp失败")
			}
			erpCode = itemInfo.ItemCode
		}
	}
	// 处理外送端,如果外送端未开启, 套餐商品或者小数计价,则不显示外送端
	isShowDelivery := uint(request.Show.IsShowDelivery)
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsOpenRider() || request.Type == constant.ProductTypePackage || request.NumType == constant.ProductNumTypeDecimal {
		isShowDelivery = 0
	}
	openOverallDiscount := uint(0)
	if request.Discount.IsEnableOverallDiscount == 1 {
		openOverallDiscount = 1
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
		IsShowDelivery:        isShowDelivery,
		Sort:                  uint(sort),
		Price:                 price,
		ProductType:           uint(request.Type),
		OpenDiscount:          uint(request.Discount.IsEnableMemberDiscount),
		OpenOverallDiscount:   &openOverallDiscount,
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

type SaveProductPackageBomParams struct {
	ProductPackageUuid uint64                   // 商品包UUID
	UnitUuid           uint64                   // 单位UUID
	TemplateItemCode   string                   // 模板商品编码
	FlavorListResult   CheckProductFlavorResult // 商品规格列表
	SauceResult        CheckProductSauceResult  // 商品加料列表
	CategoryUuid       uint64                   // 分类UUID
}

// SaveProductPackageBom 添加商品bom
func (s *productSrv) SaveProductPackageBom(ctx context.Context, tx *gorm.DB, params SaveProductPackageBomParams) error {
	commonRepo := repository.NewCommonRepo()
	productPackageRepo := repository.NewProductPackageRepo(tx)
	productBomRepo := repository.NewProductBomRepo(tx)
	warehouseFormRepo := repository.NewWarehouseFormRepo(tx)
	warehouseMonthlyFormRepo := repository.NewWarehouseMonthlyFormRepo(tx)
	setting, err := s.settingSrv.GetCompanySetting(ctx)

	// flavorListResult.Flavors 根据IsDelete排序，如果IsDelete=true，则排在前面，优先删除，然后再添加
	sort.Slice(params.FlavorListResult.Flavors, func(i, j int) bool {
		return params.FlavorListResult.Flavors[i].IsDelete && !params.FlavorListResult.Flavors[j].IsDelete
	})

	// 商品规格
	for _, flavor := range params.FlavorListResult.Flavors {
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
					Num:                  params.FlavorListResult.StockNum,
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
			// 删除商品规格的item
			if ctx.GetCompany().IsOpenErp() {
				// 获取商品bom信息
				productBom, err := productBomRepo.GetProductBom(commonRepo.WhereByUuid(flavor.BomUuid))
				if err != nil {
					return errors.WithMessage(err, "获取商品bom失败")
				}
				erpSrv := erp.NewIErpSrv(s.dbm)
				if errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
					ItemCode:   productBom.ErpCode,
					NotForSale: true,
				}); errErp != nil {
					return errors.WithMessage(errErp, "设置商品规格禁售失败")
				}
			}
		} else {
			if flavor.BomUuid == 0 {
				var isAdd bool
				var flavorUuid uint64
				erpCode := ""
				if params.FlavorListResult.IsPackage {
					isAdd = true
					// 同步套餐到erp
					if ctx.GetCompany().IsOpenErp() {
						// 获取商品包信息
						productPackage, err := productPackageRepo.GetProductPackage(
							commonRepo.WhereByUuid(params.ProductPackageUuid),
							commonRepo.Preload(
								repository.WithPreload{
									Query: "ProductUnit",
								},
								repository.WithPreload{
									Query: "ProductCategory.MultiLanguageName",
								},
							),
						)
						if err != nil {
							return errors.WithMessage(err, "获取商品包失败")
						}

						multiLanguageName := model.NewMultiLanguageName(flavor.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(params.UnitUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						stockUom := productUnit.ErpnextUom
						classification := productPackage.ProductCategory.MultiLanguageName.GetNames()
						getEnClassification, err := s.getEnName(ctx, classification)
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}

						params := req.PackageAddErpReq{
							ItemName:           enName,
							StockUom:           stockUom,
							InternalCode:       params.FlavorListResult.Flavors[0].InternalCode,
							Classification:     getEnClassification,
							ClassificationCode: productPackage.ProductCategory.Code,
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
						erpSrv := erp.NewIErpSrv(s.dbm)
						bom, err := productBomRepo.GetProductBom(
							commonRepo.WhereByProductPackageUuid(params.ProductPackageUuid),
							commonRepo.WhereByProductFlavorUuid(flavor.Uuid),
						)
						if err != nil || bom.Uuid == 0 {
							isAdd = true
						}
						if isAdd {
							// 获取商品包信息
							itemInfo, errErp := erpSrv.AddProductBom(ctx, req.ProductBomAddErpReq{
								VariantsOf:   params.TemplateItemCode,
								InternalCode: flavor.InternalCode,
								Flavors: []req.Flavor{
									{
										Name:  flavor.ErpnextGroupName,
										Value: flavor.ErpnextValueName,
									},
								},
							})
							if errErp != nil {
								return errors.WithMessage(errErp, "同步商品bom到erp失败")
							}
							erpCode = itemInfo.ItemCode
						} else {
							errErp := erpSrv.UpdateProduct(ctx, erp.UpdateProductReq{
								ItemCode:     bom.ErpCode,
								InternalCode: flavor.InternalCode,
								Disabled:     bom.Status == constant.ProductStatusOffSale,
								Attributes: []erp.UpdateProductFlavor{
									{
										Name:  flavor.ErpnextGroupName,
										Value: flavor.ErpnextValueName,
									},
								},
							})
							if errErp != nil {
								return errors.WithMessage(errErp, "更新商品bom到erp失败")
							}
							erpCode = bom.ErpCode
							flavorUuid = bom.Uuid
						}

					}
				}

				if isAdd {
					flavorUuid, _ = utils.GetID()
					_, err := productBomRepo.CreateProductBom(model.ProductBom{
						BaseModel: model.BaseModel{
							Uuid: flavorUuid,
						},
						Price:              flavor.Price,
						Name:               flavor.Name,
						ErpCode:            erpCode,
						ProductFlavorUuid:  flavor.Uuid,
						ProductPackageUuid: params.ProductPackageUuid,
						StockNum:           params.FlavorListResult.StockNum,
						BarcodeValue:       flavor.BarcodeValue,
						InternalCode:       flavor.InternalCode,
						Status:             params.FlavorListResult.Status,
						IsOpenStock:        1,
					})
					if err != nil {
						return errors.WithMessage(err, "保存商品bom失败")
					}
				} else {
					err = productBomRepo.UpdateProductBom(map[string]any{
						"price":                flavor.Price,
						"name":                 flavor.Name,
						"erp_code":             erpCode,
						"product_flavor_uuid":  flavor.Uuid,
						"product_package_uuid": params.ProductPackageUuid,
						"stock_num":            params.FlavorListResult.StockNum,
						"barcode_value":        flavor.BarcodeValue,
						"internal_code":        flavor.InternalCode,
						"status":               params.FlavorListResult.Status,
						"is_open_stock":        1,
						"delete_time":          0,
					}, commonRepo.WhereByUuid(flavorUuid))
					if err != nil {
						return errors.WithMessage(err, "更新商品bom失败")
					}
				}
				// 开启库存管理
				if setting.SaleStock == 1 {
					// 添加入库
					warehouseForm := model.WarehouseForm{
						FormNo:         warehouseFormRepo.GenerateWarehouseFormNo(setting.Timezone),
						Scene:          constant.WarehouseFormSceneAddStock,
						Num:            int(params.FlavorListResult.StockNum),
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
						Stock:          float64(params.FlavorListResult.StockNum),
					}
					err = warehouseMonthlyFormRepo.CreateWarehouseMonthlyProductBomForm(warehouseMonthlyProductBomForm)
					if err != nil {
						return errors.WithMessage(err, "保存月初库存记录失败")
					}
				}
			} else {
				updateData := map[string]any{
					"price":                flavor.Price,
					"name":                 flavor.Name,
					"product_flavor_uuid":  flavor.Uuid,
					"product_package_uuid": params.ProductPackageUuid,
					"barcode_value":        flavor.BarcodeValue,
					"internal_code":        flavor.InternalCode,
					"status":               params.FlavorListResult.Status,
					"is_open_stock":        1,
				}

				err = productBomRepo.UpdateProductBom(updateData, commonRepo.WhereByUuid(flavor.BomUuid))
				if err != nil {
					return errors.WithMessage(err, "更新商品bom失败")
				}

				// 同步商品到erp
				if ctx.GetCompany().IsOpenErp() {
					if params.FlavorListResult.IsPackage {
						// 同步套餐到erp
						multiLanguageName := model.NewMultiLanguageName(flavor.Name)
						enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(params.UnitUuid))
						if errGetUnit != nil {
							return errors.WithMessage(errGetUnit, "获取商品单位失败")
						}
						productBom, errGetBom := productBomRepo.GetProductBom(
							commonRepo.WhereByUuid(flavor.BomUuid),
							commonRepo.Preload(
								repository.WithPreload{
									Query: "ProductPackage.ProductCategory.MultiLanguageName",
								},
							),
						)
						if errGetBom != nil {
							return errors.WithMessage(errGetBom, "获取商品bom失败")
						}
						erpSrv := erp.NewIErpSrv(s.dbm)
						classification := productBom.ProductPackage.ProductCategory.MultiLanguageName.GetNames()
						getEnClassification, err := s.getEnName(ctx, classification)
						if err != nil {
							return errors.WithMessage(err, "翻译失败")
						}
						_, errErp := erpSrv.AddPackage(ctx, req.PackageAddErpReq{
							ItemName: enName,
							StockUom: productUnit.ErpnextUom,
							ItemCode: productBom.ErpCode,
							InternalCode: func() string {
								if flavor.InternalCode != "" {
									return flavor.InternalCode
								}
								return " " // 内部编码为空时，传空格给ErpNext
							}(),
							Classification:     getEnClassification,
							ClassificationCode: productBom.ProductPackage.ProductCategory.Code,
						})
						if errErp != nil {
							return errors.WithMessage(errErp, "同步套餐到erp失败")
						}
					}
				}
				if setting.SaleStock == 1 {
					bom, err := productBomRepo.GetProductBom(commonRepo.WhereByUuid(flavor.BomUuid))
					if err != nil {
						return errors.WithMessage(err, "获取商品bom失败")
					}
					diffStockNum := math.Abs(decimal.NewFromFloat(params.FlavorListResult.StockNum).Sub(decimal.NewFromFloat(bom.StockNum)).InexactFloat64())
					if diffStockNum > 0 {
						// 调整入库
						if params.FlavorListResult.StockNum > bom.StockNum {
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
						if params.FlavorListResult.StockNum < bom.StockNum {
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
	for _, sauce := range params.SauceResult.Sauces {
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
					ProductPackageUuid: params.ProductPackageUuid,
					StockNum:           99999999,
					Status:             params.FlavorListResult.Status,
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
					"product_package_uuid": params.ProductPackageUuid,
					"stock_num":            99999999,
					"status":               params.FlavorListResult.Status,
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
		"sauce_required":      params.SauceResult.IsMust,
		"sauce_max_selection": params.SauceResult.MaxSelection,
	}, commonRepo.WhereByUuid(params.ProductPackageUuid))
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

// SyncUnit 同步单位，暂不考虑erp禁用的情况；如果总部取消给某个子店查看某个单位，如何处理，暂不处理
func (s *productSrv) SyncUnit(ctx context.Context) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}

	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	// 子店获取erp单位列表
	erp := erp.NewIErpSrv(s.dbm)
	uomList, err := erp.GetUomList(ctx.GetContext(), req.GetUomListReq{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	})
	if err != nil {
		return errors.WithMessage(err, "获取单位列表失败")
	}
	var uomNames []string
	for _, uom := range uomList.List {
		uomNames = append(uomNames, uom.UomName)
	}

	// 子店获取总部单位
	var headquarter model.CompanySetting
	var headquarterUnits []model.ProductUnit
	if companySetting.IsSubShop() {
		err := s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("uuid = ?", companySetting.HeadquarterUuid).Scopes(repository.NotDeleted).Debug().First(&headquarter).Error
		if err != nil || headquarter.Uuid == 0 {
			return errors.WithMessage(errors.New("获取总部公司失败"))
		}
		s.dbm.GetDB(headquarter.Uuid).Model(&model.ProductUnit{}).Preload("MultiLanguageName").Find(&headquarterUnits)
	}

	// 子店ttpos已有单位 和 要标记删除的单位
	var units []model.ProductUnit
	var deletingUnitUuids []uint64
	unitMap := make(map[string]model.ProductUnit)
	db.Model(&model.ProductUnit{}).Scopes(repository.ExcludeHeadquarter).Where("erpnext_uom != ''").Find(&units)
	for _, unit := range units {
		if !slices.Contains(uomNames, unit.ErpnextUom) {
			deletingUnitUuids = append(deletingUnitUuids, unit.Uuid)
		}
		unitMap[unit.ErpnextUom] = unit
	}

	// 子店ttpos单位最大的排序
	var unitSort int
	db.Model(&model.ProductUnit{}).Scopes(repository.NotDeleted, repository.ExcludeHeadquarter).Select("ifnull(MAX(sort), 0)").Scan(&unitSort)
	unitSort++

	// 要恢复的单位Uuid
	var recoveringUnitUuids []uint64
	// 要翻译的多语言Uuid
	var multiLanguageNameUuids []uint64
	// 已存在的多语言Uuid
	var existsMultiLanguageUuids []uint64
	db.Model(&model.MultiLanguageName{}).Pluck("uuid", &existsMultiLanguageUuids)

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除不在erp单位列表中的单位
		if len(deletingUnitUuids) > 0 {
			err := tx.Model(&model.ProductUnit{}).Where("uuid IN (?)", deletingUnitUuids).Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(errors.New("erp已删除单位，标记删除ttpos单位失败"), err.Error())
			}
		}
		// 要添加的单位
		var insertingProductUnits []model.ProductUnit
		for _, uom := range uomList.List {
			translateName := uom.UomName
			if uom.AliasName != "" {
				translateName = uom.AliasName
			}
			multiLanguageName := model.MultiLanguageName{
				EnName:   translateName,
				ZhName:   translateName,
				ThName:   translateName,
				MyName:   translateName,
				JaName:   translateName,
				KoName:   translateName,
				TrName:   translateName,
				SvName:   translateName,
				ZhTwName: translateName,
			}
			if unit, ok := unitMap[uom.UomName]; !ok {
				// 不存在则新建
				err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
				if err != nil {
					return err
				}
				insertingProductUnits = append(insertingProductUnits, model.ProductUnit{
					Name:                  multiLanguageName.ToJson(),
					MultiLanguageNameUuid: multiLanguageName.Uuid,
					Sort:                  unitSort,
					ErpnextUom:            uom.UomName,
				})
				multiLanguageNameUuids = append(multiLanguageNameUuids, multiLanguageName.Uuid)
				unitSort++
			} else if unit.DeleteTime > 0 { // 但是被标记为删除，需要恢复为被删除
				recoveringUnitUuids = append(recoveringUnitUuids, unit.Uuid)
			}
		}
		// 同步总部ttpos单位
		if len(headquarterUnits) > 0 {
			// 删除仓库
			tx.Where("headquarter_uuid > 0").Delete(&model.ProductUnit{})

			var insertingMultiLanguageNames []model.MultiLanguageName
			for _, headquarterUnit := range headquarterUnits {
				if headquarterUnit.MultiLanguageName.Uuid != 0 && !slices.Contains(existsMultiLanguageUuids, headquarterUnit.MultiLanguageName.Uuid) {
					insertingMultiLanguageNames = append(insertingMultiLanguageNames, model.MultiLanguageName{
						BaseModel: model.BaseModel{
							Uuid:       headquarterUnit.MultiLanguageName.Uuid,
							CreateTime: headquarterUnit.MultiLanguageName.CreateTime,
							UpdateTime: headquarterUnit.MultiLanguageName.UpdateTime,
							DeleteTime: headquarterUnit.MultiLanguageName.DeleteTime,
						},
						EnName:   headquarterUnit.MultiLanguageName.EnName,
						ZhName:   headquarterUnit.MultiLanguageName.ZhName,
						ThName:   headquarterUnit.MultiLanguageName.ThName,
						MyName:   headquarterUnit.MultiLanguageName.MyName,
						JaName:   headquarterUnit.MultiLanguageName.JaName,
						KoName:   headquarterUnit.MultiLanguageName.KoName,
						TrName:   headquarterUnit.MultiLanguageName.TrName,
						SvName:   headquarterUnit.MultiLanguageName.SvName,
						ZhTwName: headquarterUnit.MultiLanguageName.ZhTwName,
					})
				}
				insertingProductUnits = append(insertingProductUnits, model.ProductUnit{
					BaseModel: model.BaseModel{
						Uuid:       headquarterUnit.Uuid,
						CreateTime: headquarterUnit.CreateTime,
						UpdateTime: headquarterUnit.UpdateTime,
						DeleteTime: headquarterUnit.DeleteTime,
					},
					Name:                  headquarterUnit.MultiLanguageName.ToJson(),
					MultiLanguageNameUuid: headquarterUnit.MultiLanguageName.Uuid,
					Sort:                  headquarterUnit.Sort,
					ErpnextUom:            headquarterUnit.ErpnextUom,
					HeadquarterUuid:       headquarter.Uuid,
				})
			}
			if len(insertingMultiLanguageNames) > 0 {
				err := tx.Model(&model.MultiLanguageName{}).Create(&insertingMultiLanguageNames).Error
				if err != nil {
					return err
				}
			}
		}
		if len(insertingProductUnits) > 0 {
			err := tx.Model(&model.ProductUnit{}).Create(&insertingProductUnits).Error
			if err != nil {
				return err
			}
		}
		// 恢复为未删除
		if len(recoveringUnitUuids) > 0 {
			err := tx.Model(&model.ProductUnit{}).Where("uuid IN (?)", recoveringUnitUuids).Update("delete_time", 0).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if len(multiLanguageNameUuids) > 0 {
		if err := s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), multiLanguageNameUuids...); err != nil {
			logger.Logger.Error("单位同步添加多语言uuid到待翻译集合中失败", zap.Error(err), zap.Any("multiLanguageNameUuids", multiLanguageNameUuids))
		}
	}
	return err
}

// SyncSauce 同步加料
func (s *productSrv) SyncSauce(ctx context.Context) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}
	// v2.7 加料暂时不需要处理（本任务不从ERP同步属性跟加料以及TTPOS添加修改之后不同步到ERP，子店同步时，从ttpos的总部获取数据）
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsSubShop() {
		return nil
	}
	var headquarter model.CompanySetting
	err := s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("uuid = ?", companySetting.HeadquarterUuid).Scopes(repository.NotDeleted).Debug().First(&headquarter).Error
	if err != nil || headquarter.Uuid == 0 {
		return errors.WithMessage(errors.New("获取总部公司失败"))
	}
	var headquarterSauces []model.ProductSauce
	s.dbm.GetDB(headquarter.Uuid).Model(&model.ProductSauce{}).Preload("MultiLanguageName").Find(&headquarterSauces)

	if len(headquarterSauces) > 0 {
		err := s.dbm.GetDB(companySetting.CompanyUuid).Transaction(func(tx *gorm.DB) error {
			// 删除多语言
			tx.Where("uuid IN (?)", tx.Model(&model.ProductSauce{}).Where("headquarter_uuid > 0").Select("multi_language_name_uuid")).Delete(&model.MultiLanguageName{})
			// 删除加料
			tx.Where("headquarter_uuid > 0").Delete(&model.ProductSauce{})

			var insertingMultiLanguageNames []model.MultiLanguageName
			var insertingProductSauce []model.ProductSauce
			for _, headquarterSauce := range headquarterSauces {
				if headquarterSauce.MultiLanguageName.Uuid == 0 {
					continue
				}
				insertingMultiLanguageNames = append(insertingMultiLanguageNames, model.MultiLanguageName{
					BaseModel: model.BaseModel{
						Uuid:       headquarterSauce.MultiLanguageName.Uuid,
						CreateTime: headquarterSauce.MultiLanguageName.CreateTime,
						UpdateTime: headquarterSauce.MultiLanguageName.UpdateTime,
						DeleteTime: headquarterSauce.MultiLanguageName.DeleteTime,
					},
					EnName:   headquarterSauce.MultiLanguageName.EnName,
					ZhName:   headquarterSauce.MultiLanguageName.ZhName,
					ThName:   headquarterSauce.MultiLanguageName.ThName,
					MyName:   headquarterSauce.MultiLanguageName.MyName,
					JaName:   headquarterSauce.MultiLanguageName.JaName,
					KoName:   headquarterSauce.MultiLanguageName.KoName,
					TrName:   headquarterSauce.MultiLanguageName.TrName,
					SvName:   headquarterSauce.MultiLanguageName.SvName,
					ZhTwName: headquarterSauce.MultiLanguageName.ZhTwName,
				})
				insertingProductSauce = append(insertingProductSauce, model.ProductSauce{
					BaseModel: model.BaseModel{
						Uuid:       headquarterSauce.Uuid,
						CreateTime: headquarterSauce.CreateTime,
						UpdateTime: headquarterSauce.UpdateTime,
						DeleteTime: headquarterSauce.DeleteTime,
					},
					Name:                  headquarterSauce.MultiLanguageName.ToJson(),
					Price:                 headquarterSauce.Price,
					MultiLanguageNameUuid: headquarterSauce.MultiLanguageName.Uuid,
					Sort:                  headquarterSauce.Sort,
					ProductBomCardUuid:    headquarterSauce.ProductBomCardUuid, // 同步的成本卡Uuid
					ErpCode:               headquarterSauce.ErpCode,
					HeadquarterUuid:       headquarter.Uuid,
				})
			}
			if len(insertingMultiLanguageNames) > 0 {
				err := tx.Model(&model.MultiLanguageName{}).Create(&insertingMultiLanguageNames).Error
				if err != nil {
					return err
				}
			}
			if len(insertingProductSauce) > 0 {
				err := tx.Model(&model.ProductSauce{}).Create(&insertingProductSauce).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return errors.WithMessage(errors.New("同步总部加料失败"), err.Error())
		}
	}

	return err

}

// SyncAttributeGroup 同步属性组、属性
func (s *productSrv) SyncAttributeGroup(ctx context.Context) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}
	companySetting := ctx.GetCompanySetting()
	// v2.7 本任务不从ERP同步属性跟加料以及TTPOS添加修改之后不同步到ERP，子店同步时，从ttpos的总部获取数据
	if !companySetting.IsSubShop() {
		return nil
	}
	var headquarter model.CompanySetting
	err := s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("uuid = ?", companySetting.HeadquarterUuid).Scopes(repository.NotDeleted).Debug().First(&headquarter).Error
	if err != nil || headquarter.Uuid == 0 {
		return errors.WithMessage(errors.New("获取总部公司失败"))
	}

	var headquarterAttributeGroups []model.ProductAttributeGroup
	s.dbm.GetDB(headquarter.Uuid).Model(&model.ProductAttributeGroup{}).Preload("MultiLanguageName").Preload("ProductAttributes").Preload("ProductAttributes.MultiLanguageName").Find(&headquarterAttributeGroups)

	var existsMultiLanguageUuids []uint64
	s.dbm.GetDB(companySetting.CompanyUuid).Model(&model.MultiLanguageName{}).Pluck("uuid", &existsMultiLanguageUuids)
	if len(headquarterAttributeGroups) > 0 {
		err := s.dbm.GetDB(companySetting.CompanyUuid).Transaction(func(tx *gorm.DB) error {
			// 删除属性值
			tx.Where("attribute_group_uuid IN (?)", tx.Model(&model.ProductAttributeGroup{}).Where("headquarter_uuid > 0").Select("uuid")).Delete(&model.ProductAttribute{})
			// 删除属性组
			tx.Where("headquarter_uuid > 0").Delete(&model.ProductAttributeGroup{})
			var insertingMultiLanguageNames []model.MultiLanguageName
			var insertingProductAttributeGroups []model.ProductAttributeGroup
			var insertingProductAttributes []model.ProductAttribute
			for _, headquarterAttributeGroup := range headquarterAttributeGroups {
				if headquarterAttributeGroup.MultiLanguageName.Uuid != 0 && !slices.Contains(existsMultiLanguageUuids, headquarterAttributeGroup.MultiLanguageName.Uuid) {
					insertingMultiLanguageNames = append(insertingMultiLanguageNames, model.MultiLanguageName{
						BaseModel: model.BaseModel{
							Uuid:       headquarterAttributeGroup.MultiLanguageName.Uuid,
							CreateTime: headquarterAttributeGroup.MultiLanguageName.CreateTime,
							UpdateTime: headquarterAttributeGroup.MultiLanguageName.UpdateTime,
							DeleteTime: headquarterAttributeGroup.MultiLanguageName.DeleteTime,
						},
						EnName:   headquarterAttributeGroup.MultiLanguageName.EnName,
						ZhName:   headquarterAttributeGroup.MultiLanguageName.ZhName,
						ThName:   headquarterAttributeGroup.MultiLanguageName.ThName,
						MyName:   headquarterAttributeGroup.MultiLanguageName.MyName,
						JaName:   headquarterAttributeGroup.MultiLanguageName.JaName,
						KoName:   headquarterAttributeGroup.MultiLanguageName.KoName,
						TrName:   headquarterAttributeGroup.MultiLanguageName.TrName,
						SvName:   headquarterAttributeGroup.MultiLanguageName.SvName,
						ZhTwName: headquarterAttributeGroup.MultiLanguageName.ZhTwName,
					})
				}
				insertingProductAttributeGroups = append(insertingProductAttributeGroups, model.ProductAttributeGroup{
					BaseModel: model.BaseModel{
						Uuid:       headquarterAttributeGroup.Uuid,
						CreateTime: headquarterAttributeGroup.CreateTime,
						UpdateTime: headquarterAttributeGroup.UpdateTime,
						DeleteTime: headquarterAttributeGroup.DeleteTime,
					},
					Name:                      headquarterAttributeGroup.MultiLanguageName.ToJson(),
					MultiLanguageNameUuid:     headquarterAttributeGroup.MultiLanguageName.Uuid,
					Sort:                      headquarterAttributeGroup.Sort,
					ErpnextAttributeGroupName: headquarterAttributeGroup.ErpnextAttributeGroupName,
					HeadquarterUuid:           headquarter.Uuid,
				})
				for _, headquarterProductAttribute := range headquarterAttributeGroup.ProductAttributes {
					if headquarterProductAttribute.MultiLanguageName.Uuid != 0 && !slices.Contains(existsMultiLanguageUuids, headquarterProductAttribute.MultiLanguageName.Uuid) {
						insertingMultiLanguageNames = append(insertingMultiLanguageNames, model.MultiLanguageName{
							BaseModel: model.BaseModel{
								Uuid:       headquarterProductAttribute.MultiLanguageName.Uuid,
								CreateTime: headquarterProductAttribute.MultiLanguageName.CreateTime,
								UpdateTime: headquarterProductAttribute.MultiLanguageName.UpdateTime,
								DeleteTime: headquarterProductAttribute.MultiLanguageName.DeleteTime,
							},
							EnName:   headquarterProductAttribute.MultiLanguageName.EnName,
							ZhName:   headquarterProductAttribute.MultiLanguageName.ZhName,
							ThName:   headquarterProductAttribute.MultiLanguageName.ThName,
							MyName:   headquarterProductAttribute.MultiLanguageName.MyName,
							JaName:   headquarterProductAttribute.MultiLanguageName.JaName,
							KoName:   headquarterProductAttribute.MultiLanguageName.KoName,
							TrName:   headquarterProductAttribute.MultiLanguageName.TrName,
							SvName:   headquarterProductAttribute.MultiLanguageName.SvName,
							ZhTwName: headquarterProductAttribute.MultiLanguageName.ZhTwName,
						})
					}
					insertingProductAttributes = append(insertingProductAttributes, model.ProductAttribute{
						BaseModel: model.BaseModel{
							Uuid:       headquarterProductAttribute.Uuid,
							CreateTime: headquarterProductAttribute.CreateTime,
							UpdateTime: headquarterProductAttribute.UpdateTime,
							DeleteTime: headquarterProductAttribute.DeleteTime,
						},
						Name:                  headquarterProductAttribute.MultiLanguageName.ToJson(),
						MultiLanguageNameUuid: headquarterProductAttribute.MultiLanguageName.Uuid,
						AttributeGroupUuid:    headquarterProductAttribute.AttributeGroupUuid,
						Sort:                  headquarterProductAttribute.Sort,
						ErpnextAttributeValue: headquarterProductAttribute.ErpnextAttributeValue,
						HeadquarterUuid:       headquarter.Uuid,
					})
				}
			}
			if len(insertingMultiLanguageNames) > 0 {
				err := tx.Model(&model.MultiLanguageName{}).Create(&insertingMultiLanguageNames).Error
				if err != nil {
					return err
				}
			}
			if len(insertingProductAttributeGroups) > 0 {
				err := tx.Model(&model.ProductAttributeGroup{}).Create(&insertingProductAttributeGroups).Error
				if err != nil {
					return err
				}
			}
			if len(insertingProductAttributes) > 0 {
				err := tx.Model(&model.ProductAttribute{}).Create(&insertingProductAttributes).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return errors.WithMessage(errors.New("同步总部属性组失败"), err.Error())
		}
	}

	return nil
}

// SyncProduct 同步商品
func (s *productSrv) SyncProduct(ctx context.Context) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}
	companySetting := ctx.GetCompanySetting()

	// 同步erp商品
	erpSrv := erp.NewIErpSrv(s.dbm)
	erpProducts, err := erpSrv.GetProductList(ctx, erp.GetErpProductListReq{
		SiteCode:        companySetting.ErpnextSiteCode,
		Branch:          companySetting.ErpnextBranchName,
		CompanyAbbr:     companySetting.ErpnextCompanyAbbr,
		ContainDisabled: true,
	})
	if err != nil {
		return errors.WithMessage(err, "获取erp商品列表失败")
	}

	var multiLanguageNameUuids []uint64
	db := ctx.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		productRepo := repository.NewProductRepo(tx)
		productPackageRepo := repository.NewProductPackageRepo(tx)
		productUnitRepo := repository.NewProductUnitRepo(tx)
		productCategoryRepo := repository.NewProductCategoryRepo(tx)
		productFlavorRepo := repository.NewProductFlavorRepo(tx)
		productBomRepo := repository.NewProductBomRepo(tx)
		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
		maxSort, _ := productRepo.GetProductShopMaxSort(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByHeadquarterUuid(0),
		)
		for _, erpProduct := range erpProducts.ItemList {
			if !strings.HasPrefix(erpProduct.ItemCode, "SP") {
				continue
			}
			if erpProduct.VariantOf == "" {
				// 商品包
				unit, err := productUnitRepo.GetProductUnit(
					productUnitRepo.WhereByErpnextUom(erpProduct.StockUom),
				)
				if err != nil || unit.Uuid == 0 {
					logger.Logger.Error("商品单位不存在", zap.String("erpnext_uom", erpProduct.StockUom), zap.Error(err))
					continue
				}
				var categoryUuid uint64
				category, err := productCategoryRepo.GetProductCategory(
					commonRepo.WhereByCategoryKey(""),
					commonRepo.WhereBySoftDelete(),
					commonRepo.WhereByCode(erpProduct.ClassificationCode),
				)
				if err == nil && category.Uuid > 0 {
					categoryUuid = category.Uuid
				}
				existsProductPackage, err := productPackageRepo.GetProductPackage(
					commonRepo.WhereByProductPackageErpCode(erpProduct.ItemCode),
					commonRepo.WhereByHeadquarterUuid(0),
					productPackageRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
				)
				status := utils.IfUint(erpProduct.Disabled, constant.ProductStatusOffSale, constant.ProductStatusOnSale)
				if err != nil || existsProductPackage.Uuid == 0 {
					enName := erpProduct.ItemName
					lang := model.MultiLanguageName{
						EnName:   enName,
						ZhName:   enName,
						ZhTwName: enName,
						ThName:   enName,
						MyName:   enName,
						JaName:   enName,
						KoName:   enName,
						TrName:   enName,
						SvName:   enName,
					}
					multiLanguageNameUuid, err := multiLanguageNameRepo.CreateMultiLanguageName(lang)
					if err != nil {
						logger.Logger.Error("创建多语言名称失败", zap.String("item_code", erpProduct.ItemCode), zap.String("item_name", erpProduct.ItemName), zap.Error(err))
						continue
					}
					multiLanguageNameUuids = append(multiLanguageNameUuids, multiLanguageNameUuid)
					maxSort += 1
					var deleteTime int64 = 0
					if erpProduct.NotForSale {
						deleteTime = time.Now().Unix()
					}
					openOverallDiscount := uint(1)
					productPackage := model.ProductPackage{
						BaseModel:             model.BaseModel{DeleteTime: deleteTime},
						Name:                  lang.ToJson(),
						ErpCode:               erpProduct.ItemCode,
						MultiLanguageNameUuid: multiLanguageNameUuid,
						NumType:               constant.ProductNumTypeDecimal,
						UnitUuid:              unit.Uuid,
						CategoryUuid:          categoryUuid,
						Status:                status,
						IsShowCashier:         1,
						IsShowTablet:          1,
						IsShowKitchen:         1,
						IsShowAssistant:       1,
						IsShowH5:              1,
						Sort:                  uint(maxSort),
						OpenDiscount:          1,
						OpenOverallDiscount:   &openOverallDiscount,
					}
					err = productPackageRepo.CreateProductPackage(&productPackage)
					if err != nil {
						logger.Logger.Error("创建商品包失败", zap.String("item_code", erpProduct.ItemCode), zap.Any("productPackage", productPackage), zap.Error(err))
						continue
					}
				} else {
					updateData := map[string]any{
						"unit_uuid": unit.Uuid,
						"status":    status,
					}
					if erpProduct.NotForSale && existsProductPackage.DeleteTime == 0 {
						updateData["delete_time"] = time.Now().Unix()
					} else if !erpProduct.NotForSale {
						updateData["delete_time"] = 0
					}
					err = productPackageRepo.UpdateProductPackage(updateData, commonRepo.WhereByUuid(existsProductPackage.Uuid))
					if err != nil {
						logger.Logger.Error("更新商品包失败", zap.String("item_code", erpProduct.ItemCode), zap.Any("updateData", updateData), zap.Error(err))
						continue
					}
				}
			} else {
				// 商品规格
				productPackage, err := productPackageRepo.GetProductPackage(
					commonRepo.WhereByProductPackageErpCode(erpProduct.VariantOf),
					commonRepo.WhereByHeadquarterUuid(0),
					productPackageRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
				)
				if err != nil || productPackage.Uuid == 0 {
					logger.Logger.Error("商品包不存在", zap.String("item_code", erpProduct.ItemCode), zap.Error(err))
					continue
				}
				for _, attribute := range erpProduct.Attributes {
					flavor, err := productFlavorRepo.GetProductFlavor(
						commonRepo.WhereBySoftDelete(),
						commonRepo.WhereByErpnextValueName(attribute.AttributeValue),
						productRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
					)
					if err != nil || flavor.Uuid == 0 {
						logger.Logger.Error("商品规格不存在", zap.String("item_code", erpProduct.ItemCode), zap.String("erpnext_value_name", attribute.AttributeValue), zap.Error(err))
						continue
					}
					existsProductBom, err := productBomRepo.GetProductBom(
						commonRepo.WhereByProductPackageErpCode(erpProduct.ItemCode),
						commonRepo.WhereByProductPackageUuid(productPackage.Uuid),
						commonRepo.WhereByProductFlavorUuid(flavor.Uuid),
					)
					status := utils.IfInt(erpProduct.Disabled, constant.ProductStatusOffSale, constant.ProductStatusOnSale)
					if err != nil || existsProductBom.Uuid == 0 {
						var deleteTime int64 = 0
						if erpProduct.NotForSale {
							deleteTime = time.Now().Unix()
						}
						productBom := model.ProductBom{
							BaseModel:          model.BaseModel{DeleteTime: deleteTime},
							Name:               flavor.Name,
							ErpCode:            erpProduct.ItemCode,
							StockNum:           erpProduct.OpeningStock,
							BarcodeValue:       erpProduct.Barcode,
							InternalCode:       erpProduct.InternalCode,
							Status:             status,
							ProductFlavorUuid:  flavor.Uuid,
							ProductPackageUuid: productPackage.Uuid,
						}
						_, err = productBomRepo.CreateProductBom(productBom)
						if err != nil {
							logger.Logger.Error("创建商品bom失败", zap.String("item_code", erpProduct.ItemCode), zap.Any("productBom", productBom), zap.Error(err))
							continue
						}
					} else {
						updateData := map[string]any{
							"barcode_value": erpProduct.Barcode,
							"status":        status,
						}
						if erpProduct.NotForSale && existsProductBom.DeleteTime == 0 {
							updateData["delete_time"] = time.Now().Unix()
						} else if !erpProduct.NotForSale {
							updateData["delete_time"] = 0
						}
						err = productBomRepo.UpdateProductBom(updateData, commonRepo.WhereByUuid(existsProductBom.Uuid))
						if err != nil {
							logger.Logger.Error("更新商品bom失败", zap.String("item_code", erpProduct.ItemCode), zap.Any("updateData", updateData), zap.Error(err))
							continue
						}
					}

				}
			}
		}
		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "同步erp商品到本地失败", err.Error())
	}

	// 同步总店商品到子店
	if companySetting.IsSubShop() {
		commonRepo := repository.NewCommonRepo()
		headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
		productPackageRepo := repository.NewProductPackageRepo(headquarterDb)
		headProductPackageList, err := productPackageRepo.GetProductPackageList(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByHeadquarterUuid(0),
			productPackageRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
			productPackageRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
			productPackageRepo.WithProductPackageAttributeGroups(commonRepo.WhereBySoftDelete()),
			productPackageRepo.WithProductPackageAttributeGroupAttributes(commonRepo.WhereBySoftDelete()),
			productPackageRepo.WithProductPackageGroups(commonRepo.WhereBySoftDelete()),
			productPackageRepo.WithProductPackageGroupItems(commonRepo.WhereBySoftDelete()),
			productPackageRepo.WithProductPackageGroupMultiLanguageName(commonRepo.WhereBySoftDelete()),
		)
		if err != nil {
			return errors.WithMessage(err, "获取总部商品包列表失败")
		}
		// 需要保存的商品包、多语言名称、商品bom、商品包属性组、商品包属性、商品套餐组、商品套餐商品
		delMultiLanguageNameUuid := make([]uint64, 0)
		newProductPackageList := make([]model.ProductPackage, 0)
		newMultiLanguageNameList := make([]model.MultiLanguageName, 0)
		newProductBomList := make([]model.ProductBom, 0)
		newProductPackageAttributeGroupList := make([]model.ProductPackageAttributeGroup, 0)
		newProductPackageAttributeList := make([]model.ProductPackageAttribute, 0)
		newProductPackageGroupList := make([]model.ProductPackageGroup, 0)
		newProductPackageGroupItemList := make([]model.ProductPackageGroupItem, 0)
		for _, productPackage := range headProductPackageList {
			time := time.Now().Unix()
			newProductPackageList = append(newProductPackageList, model.ProductPackage{
				BaseModel:             model.BaseModel{Uuid: productPackage.Uuid, CreateTime: time, UpdateTime: time},
				Name:                  productPackage.MultiLanguageName.ToJson(),
				ErpCode:               productPackage.ErpCode,
				MultiLanguageNameUuid: productPackage.MultiLanguageNameUuid,
				ImageName:             productPackage.ImageName,
				ImageFileUuid:         productPackage.ImageFileUuid,
				DeductStockType:       productPackage.DeductStockType,
				NumType:               productPackage.NumType,
				UnitUuid:              productPackage.UnitUuid,
				DineTaxUuid:           productPackage.DineTaxUuid,
				CategoryUuid:          productPackage.CategoryUuid,
				TakeoutTaxUuid:        productPackage.TakeoutTaxUuid,
				SpecialCategoryUuid:   productPackage.SpecialCategoryUuid,
				PrinterTagUuid:        productPackage.PrinterTagUuid,
				SupplierUuid:          productPackage.SupplierUuid,
				Status:                productPackage.Status,
				IsShowCashier:         productPackage.IsShowCashier,
				IsShowTablet:          productPackage.IsShowTablet,
				IsShowKitchen:         productPackage.IsShowKitchen,
				IsShowAssistant:       productPackage.IsShowAssistant,
				IsShowH5:              productPackage.IsShowH5,
				IsShowDelivery:        productPackage.IsShowDelivery,
				Sort:                  productPackage.Sort,
				LimitNum:              productPackage.LimitNum,
				Describe:              productPackage.Describe,
				ActualSaleNum:         productPackage.ActualSaleNum,
				Price:                 productPackage.Price,
				ProductType:           productPackage.ProductType,
				SauceRequired:         productPackage.SauceRequired,
				SauceMaxSelection:     productPackage.SauceMaxSelection,
				OpenDiscount:          productPackage.OpenDiscount,
				OpenOverallDiscount:   productPackage.OpenOverallDiscount,
				IsBatch:               productPackage.IsBatch,
				HeadquarterUuid:       companySetting.HeadquarterUuid,
			})
			newMultiLanguageNameList = append(newMultiLanguageNameList, model.MultiLanguageName{
				BaseModel: model.BaseModel{Uuid: productPackage.MultiLanguageName.Uuid, CreateTime: time, UpdateTime: time},
				EnName:    productPackage.MultiLanguageName.EnName,
				ZhName:    productPackage.MultiLanguageName.ZhName,
				ZhTwName:  productPackage.MultiLanguageName.ZhTwName,
				ThName:    productPackage.MultiLanguageName.ThName,
				MyName:    productPackage.MultiLanguageName.MyName,
				JaName:    productPackage.MultiLanguageName.JaName,
				KoName:    productPackage.MultiLanguageName.KoName,
				TrName:    productPackage.MultiLanguageName.TrName,
				SvName:    productPackage.MultiLanguageName.SvName,
			})
			delMultiLanguageNameUuid = append(delMultiLanguageNameUuid, productPackage.MultiLanguageName.Uuid)
			for _, productBom := range productPackage.ProductBoms {
				newProductBomList = append(newProductBomList, model.ProductBom{
					BaseModel:          model.BaseModel{Uuid: productBom.Uuid, CreateTime: time, UpdateTime: time},
					PurchasePrice:      productBom.PurchasePrice,
					Price:              productBom.Price,
					Name:               productBom.Name,
					ErpCode:            productBom.ErpCode,
					StockNum:           productBom.StockNum,
					IsOpenStock:        productBom.IsOpenStock,
					BarcodeValue:       productBom.BarcodeValue,
					InternalCode:       productBom.InternalCode,
					IsDefaultSelect:    productBom.IsDefaultSelect,
					Status:             productBom.Status,
					IsSoldOut:          productBom.IsSoldOut,
					ActualSaleNum:      productBom.ActualSaleNum,
					ProductFlavorUuid:  productBom.ProductFlavorUuid,
					ProductSauceUuid:   productBom.ProductSauceUuid,
					ProductPackageUuid: productBom.ProductPackageUuid,
					ProductBomCardUuid: productBom.ProductBomCardUuid,
				})
			}
			for _, productPackageAttributeGroup := range productPackage.ProductPackageAttributeGroups {
				newProductPackageAttributeGroupList = append(newProductPackageAttributeGroupList, model.ProductPackageAttributeGroup{
					BaseModel:                 model.BaseModel{Uuid: productPackageAttributeGroup.Uuid, CreateTime: time, UpdateTime: time},
					IsMust:                    productPackageAttributeGroup.IsMust,
					MaxSelection:              productPackageAttributeGroup.MaxSelection,
					ProductPackageUuid:        productPackageAttributeGroup.ProductPackageUuid,
					ProductAttributeGroupUuid: productPackageAttributeGroup.ProductAttributeGroupUuid,
				})
				for _, productPackageAttribute := range productPackageAttributeGroup.ProductPackageAttributes {
					newProductPackageAttributeList = append(newProductPackageAttributeList, model.ProductPackageAttribute{
						BaseModel:                        model.BaseModel{Uuid: productPackageAttribute.Uuid, CreateTime: time, UpdateTime: time},
						ProductPackageAttributeGroupUuid: productPackageAttribute.ProductPackageAttributeGroupUuid,
						AttributeUuid:                    productPackageAttribute.AttributeUuid,
						IsDefaultSelected:                productPackageAttribute.IsDefaultSelected,
					})
				}
			}
			for _, productPackageGroup := range productPackage.ProductPackageGroups {
				newProductPackageGroupList = append(newProductPackageGroupList, model.ProductPackageGroup{
					BaseModel:             model.BaseModel{Uuid: productPackageGroup.Uuid, CreateTime: time, UpdateTime: time},
					Name:                  productPackageGroup.MultiLanguageName.ToJson(),
					ProductPackageUuid:    productPackageGroup.ProductPackageUuid,
					MultiLanguageNameUuid: productPackageGroup.MultiLanguageName.Uuid,
				})
				for _, productPackageGroupItem := range productPackageGroup.ProductPackageGroupItems {
					newProductPackageGroupItemList = append(newProductPackageGroupItemList, model.ProductPackageGroupItem{
						BaseModel:               model.BaseModel{Uuid: productPackageGroupItem.Uuid, CreateTime: time, UpdateTime: time},
						ProductPackageGroupUuid: productPackageGroupItem.ProductPackageGroupUuid,
						RelatedUuid:             productPackageGroupItem.RelatedUuid,
						ProductBomUuid:          productPackageGroupItem.ProductBomUuid,
						Num:                     productPackageGroupItem.Num,
						Sort:                    productPackageGroupItem.Sort,
					})
				}
				newMultiLanguageNameList = append(newMultiLanguageNameList, model.MultiLanguageName{
					BaseModel: model.BaseModel{Uuid: productPackageGroup.MultiLanguageName.Uuid, CreateTime: time, UpdateTime: time},
					EnName:    productPackageGroup.MultiLanguageName.EnName,
					ZhName:    productPackageGroup.MultiLanguageName.ZhName,
					ZhTwName:  productPackageGroup.MultiLanguageName.ZhTwName,
					ThName:    productPackageGroup.MultiLanguageName.ThName,
					MyName:    productPackageGroup.MultiLanguageName.MyName,
					JaName:    productPackageGroup.MultiLanguageName.JaName,
					KoName:    productPackageGroup.MultiLanguageName.KoName,
					TrName:    productPackageGroup.MultiLanguageName.TrName,
					SvName:    productPackageGroup.MultiLanguageName.SvName,
				})
				delMultiLanguageNameUuid = append(delMultiLanguageNameUuid, productPackageGroup.MultiLanguageName.Uuid)
			}
		}
		db := s.dbm.GetDB(companySetting.CompanyUuid)
		err = db.Transaction(func(tx *gorm.DB) error {
			productPackageRepo := repository.NewProductPackageRepo(tx)
			productBomRepo := repository.NewProductBomRepo(tx)
			productPackageAttributeGroupRepo := repository.NewProductPackageAttributeGroupRepo(tx)
			productPackageAttributeRepo := repository.NewProductPackageAttributeRepo(tx)
			productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)
			multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)

			// 需要保存的商品包、多语言名称、商品bom、商品包属性组、商品包属性、商品套餐组、商品套餐商品
			delProductPackageUuid := make([]uint64, 0)
			delProductBomUuid := make([]uint64, 0)
			delProductPackageAttributeGroupUuid := make([]uint64, 0)
			delProductPackageAttributeUuid := make([]uint64, 0)
			delProductPackageGroupUuid := make([]uint64, 0)
			delProductPackageGroupItemUuid := make([]uint64, 0)

			subProductPackageList, err := productPackageRepo.GetProductPackageList(
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByHeadquarterUuid(companySetting.HeadquarterUuid),
				productPackageRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
				productPackageRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
				productPackageRepo.WithProductPackageAttributeGroups(commonRepo.WhereBySoftDelete()),
				productPackageRepo.WithProductPackageAttributeGroupAttributes(commonRepo.WhereBySoftDelete()),
				productPackageRepo.WithProductPackageGroups(commonRepo.WhereBySoftDelete()),
				productPackageRepo.WithProductPackageGroupItems(commonRepo.WhereBySoftDelete()),
				productPackageRepo.WithProductPackageGroupMultiLanguageName(commonRepo.WhereBySoftDelete()),
			)
			if err != nil {
				return errors.WithMessage(err, "获取子店商品列表失败")
			}
			for _, productPackage := range subProductPackageList {
				delProductPackageUuid = append(delProductPackageUuid, productPackage.Uuid)
				delMultiLanguageNameUuid = append(delMultiLanguageNameUuid, productPackage.MultiLanguageNameUuid)
				for _, productBom := range productPackage.ProductBoms {
					delProductBomUuid = append(delProductBomUuid, productBom.Uuid)
				}
				for _, productPackageAttributeGroup := range productPackage.ProductPackageAttributeGroups {
					delProductPackageAttributeGroupUuid = append(delProductPackageAttributeGroupUuid, productPackageAttributeGroup.Uuid)
					for _, productPackageAttribute := range productPackageAttributeGroup.ProductPackageAttributes {
						delProductPackageAttributeUuid = append(delProductPackageAttributeUuid, productPackageAttribute.Uuid)
					}
				}
				for _, productPackageGroup := range productPackage.ProductPackageGroups {
					delProductPackageGroupUuid = append(delProductPackageGroupUuid, productPackageGroup.Uuid)
					for _, productPackageGroupItem := range productPackageGroup.ProductPackageGroupItems {
						delProductPackageGroupItemUuid = append(delProductPackageGroupItemUuid, productPackageGroupItem.Uuid)
					}
					delMultiLanguageNameUuid = append(delMultiLanguageNameUuid, productPackageGroup.MultiLanguageNameUuid)
				}
			}

			if len(delProductPackageUuid) > 0 {
				err = productPackageRepo.DestroyProductPackage(commonRepo.WhereInUuids(delProductPackageUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁商品包失败")
				}
			}
			if len(delMultiLanguageNameUuid) > 0 {
				err = multiLanguageNameRepo.DestroyMultiLanguageName(commonRepo.WhereInUuids(delMultiLanguageNameUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁多语言名称失败")
				}
			}
			if len(delProductBomUuid) > 0 {
				err = productBomRepo.DestroyProductBom(commonRepo.WhereInUuids(delProductBomUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁商品bom失败")
				}
			}
			if len(delProductPackageAttributeGroupUuid) > 0 {
				err = productPackageAttributeGroupRepo.DestroyProductPackageAttributeGroup(commonRepo.WhereInUuids(delProductPackageAttributeGroupUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁商品包属性组失败")
				}
			}
			if len(delProductPackageAttributeUuid) > 0 {
				err = productPackageAttributeRepo.DestroyProductPackageAttribute(commonRepo.WhereInUuids(delProductPackageAttributeUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁商品包属性失败")
				}
			}
			if len(delProductPackageGroupUuid) > 0 {
				err = productPackageGroupRepo.DestroyProductPackageGroup(commonRepo.WhereInUuids(delProductPackageGroupUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁商品包组失败")
				}
			}
			if len(delProductPackageGroupItemUuid) > 0 {
				err = productPackageGroupRepo.DestroyProductPackageGroupItem(commonRepo.WhereInUuids(delProductPackageGroupItemUuid))
				if err != nil {
					return errors.WithMessage(err, "销毁商品包组商品失败")
				}
			}

			if len(newProductPackageList) > 0 {
				err = productPackageRepo.CreateProductPackages(newProductPackageList)
				if err != nil {
					return errors.WithMessage(err, "创建商品包失败")
				}
			}
			if len(newProductBomList) > 0 {
				for _, v := range newProductBomList {
					_, err = productBomRepo.CreateProductBom(v)
					if err != nil {
						logger.Logger.Error("创建商品bom失败", zap.Error(err))
						// return errors.WithMessage(err, "创建商品bom失败")
					}
				}
			}
			if len(newProductPackageAttributeGroupList) > 0 {
				for _, v := range newProductPackageAttributeGroupList {
					err = productPackageAttributeGroupRepo.CreateProductPackageAttributeGroups([]model.ProductPackageAttributeGroup{v})
					if err != nil {
						logger.Logger.Error("创建商品包属性组失败", zap.Error(err))
						// return errors.WithMessage(err, "创建商品包属性组失败")
					}
				}
			}
			if len(newProductPackageAttributeList) > 0 {
				for _, v := range newProductPackageAttributeList {
					err = productPackageAttributeRepo.CreateProductPackageAttributes([]model.ProductPackageAttribute{v})
					if err != nil {
						logger.Logger.Error("创建商品包属性失败", zap.Error(err))
						// return errors.WithMessage(err, "创建商品包属性失败")
					}
				}
			}
			if len(newProductPackageGroupList) > 0 {
				for _, v := range newProductPackageGroupList {
					err = productPackageGroupRepo.CreateProductPackageGroups([]model.ProductPackageGroup{v})
					if err != nil {
						logger.Logger.Error("创建商品包组失败", zap.Error(err))
						// return errors.WithMessage(err, "创建商品包组失败")
					}
				}
			}
			if len(newProductPackageGroupItemList) > 0 {
				for _, v := range newProductPackageGroupItemList {
					err = productPackageGroupRepo.CreateProductPackageGroupItems([]model.ProductPackageGroupItem{v})
					if err != nil {
						logger.Logger.Error("创建商品包组商品失败", zap.Error(err))
						// return errors.WithMessage(err, "创建商品包组商品失败")
					}
				}
			}
			if len(newMultiLanguageNameList) > 0 {
				for _, v := range newMultiLanguageNameList {
					_, err = multiLanguageNameRepo.CreateMultiLanguageName(v)
					if err != nil {
						logger.Logger.Error("创建多语言名称失败", zap.Error(err))
						// return errors.WithMessage(err, "创建多语言名称失败")
					}
				}
			}
			return nil
		})
		if err != nil {
			return errors.WithMessage(err, "同步总店商品到子店失败")
		}
	}

	if len(multiLanguageNameUuids) > 0 {
		err = s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), multiLanguageNameUuids...)
		if err != nil {
			logger.Logger.Error("同步erp商品添加多语言uuid到待翻译集合中失败", zap.Error(err))
		}
	}

	return nil
}

// GetProductBatchTypeList 获取分批类型列表
func (s *productSrv) GetBatchTagList(ctx context.Context, req req.BatchTagListReq) (*product_resp.BatchTagList, error) {
	batchTagRepo := repository.NewBatchTagRepo(s.dbm.GetDB(ctx.GetDbId()))
	batchTags, err := batchTagRepo.GetBatchTagList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取分批类型列表失败")
	}

	// 转换为响应格式
	list := make([]product_resp.BatchTag, len(batchTags))
	for i, batchType := range batchTags {
		list[i] = product_resp.BatchTag{
			Uuid:       batchType.Uuid,
			LocaleName: batchType.MultiLanguageName.GetNames(),
			Color:      batchType.Color,
			Sort:       batchType.Sort,
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Sort < list[j].Sort
	})

	return &product_resp.BatchTagList{
		List: list,
	}, nil
}

// GetBatchTag 获取分批类型详情
func (s *productSrv) GetBatchTag(ctx context.Context, req req.BatchTagReq) (*product_resp.BatchTagDetail, error) {
	batchTagRepo := repository.NewBatchTagRepo(s.dbm.GetDB(ctx.GetDbId()))
	batchTag, err := batchTagRepo.GetBatchTagInfo(req.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取分批类型详情失败")
	}

	return &product_resp.BatchTagDetail{
		Uuid:       batchTag.Uuid,
		LocaleName: batchTag.MultiLanguageName.GetNames(),
		Color:      batchTag.Color,
		Sort:       batchTag.Sort,
	}, nil
}

// AddProductBatchType 添加分批类型
func (s *productSrv) AddBatchTag(ctx context.Context, req req.BatchTagAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		batchTagRepo := repository.NewBatchTagRepo(tx)
		// 检查颜色是否已被使用
		if batchTagRepo.CheckColorExists(req.Color, 0) {
			return errors.New("该颜色已被其他分批类型使用")
		}

		// 获取下一个排序值
		maxSort, err := batchTagRepo.GetMaxSort()
		if err != nil {
			return errors.WithMessage(err, "获取当前最大的排序值失败")
		}
		nextSort := maxSort + 1

		// 创建多语言名称
		multiLanguageName := model.MultiLanguageName{}
		multiLanguageName.InitByLocaleResponse(req.LocaleName)
		multiLanguageNameUuid, err := repository.NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}

		// 创建分批类型
		batchTag := model.BatchTag{
			Name:                  req.LocaleName.ToJson(),
			MultiLanguageNameUuid: multiLanguageNameUuid,
			Color:                 req.Color,
			Sort:                  nextSort,
		}

		err = batchTagRepo.CreateBatchTag(batchTag)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err, "添加分批类型失败")
	}

	return nil
}

// EditProductBatchType 编辑分批类型
func (s *productSrv) EditBatchTag(ctx context.Context, req req.BatchTagEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		batchTagRepo := repository.NewBatchTagRepo(tx)
		batchTag, err := batchTagRepo.GetBatchTagInfo(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "获取分批类型详情失败")
		}

		// 检查颜色是否已被其他分批类型使用（排除自己）
		if batchTagRepo.CheckColorExists(req.Color, req.Uuid) {
			return errors.New("该颜色已被其他分批类型使用")
		}

		// 更新多语言名称
		batchTag.MultiLanguageName.InitByLocaleResponse(req.LocaleName)
		repository.NewMultiLanguageNameRepo(tx).UpdateMultiLanguageName(batchTag.MultiLanguageNameUuid, *batchTag.MultiLanguageName)

		// 更新分批类型
		batchTag.Color = req.Color
		batchTag.Name = req.LocaleName.ToJson()
		err = batchTagRepo.UpdateBatchTag(*batchTag)
		if err != nil {
			return errors.WithMessage(err, "更新分批类型失败")
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err, "编辑分批类型失败")
	}

	return nil
}

// DeleteProductBatchType 删除分批类型
func (s *productSrv) DeleteBatchTag(ctx context.Context, req req.BatchTagDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	batchTagRepo := repository.NewBatchTagRepo(db)
	batchTag, err := batchTagRepo.GetBatchTag(repository.CommonRepo.WhereByUuid(req.Uuid), repository.CommonRepo.WhereBySoftDelete())
	if err != nil {
		return errors.WithMessage(err, "获取分批类型详情失败")
	}

	// 查询分批类型是否被使用. sale_bill中是否被使用、正在进行中的订单中是否有商品正在使用
	// 查询所有正在进行中的订单
	var saleBillUuids []uint64
	db.Model(&model.SaleBill{}).Where("status = ?", constant.SaleBillStatusPending).Where("delete_time = ?", 0).Where("batch_tag_uuid <> ?", 0).Select("uuid").Scan(&saleBillUuids)
	var count int64
	db.Model(&model.SaleOrderProduct{}).Where("sale_bill_uuid in (?) AND batch_tag_uuid = ?", saleBillUuids, req.Uuid).Where("delete_time = ?", 0).Count(&count)
	if count > 0 {
		return errors.WithMessage(errors.New("该分批类型正在使用，无法删除"), "分批类型正在被使用")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		batchTagRepo := repository.NewBatchTagRepo(tx)
		// 删除分批类型
		err := batchTagRepo.DeleteBatchTag(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "删除分批类型失败")
		}
		// 删除多语言名称
		err = repository.NewMultiLanguageNameRepo(tx).DeleteMultiLanguageName(batchTag.MultiLanguageNameUuid)
		if err != nil {
			return errors.WithMessage(err, "删除名称多语言失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err, "删除分批类型失败")
	}

	return nil
}

// SortProductBatchType 排序分批类型
func (s *productSrv) SortBatchTag(ctx context.Context, req req.BatchTagSortReq) error {
	sorts := make(map[uint64]int)
	for _, item := range req.List {
		sorts[item.Uuid] = item.Sort
	}

	productRepo := repository.NewProductRepo(s.dbm.GetDB(ctx.GetDbId()))
	err := productRepo.BatchUpdateSort(&model.BatchTag{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序分批类型失败"), err.Error())
	}

	return nil
}

// GetBatchTagColorUsage 获取色块被选择情况
func (s *productSrv) GetBatchTagColorUsage(ctx context.Context) (*product_resp.BatchTagColorUsageList, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	batchTagRepo := repository.NewBatchTagRepo(db)
	batchTags, err := batchTagRepo.GetBatchTagList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取分批类型列表失败")
	}

	// #FF585B
	// #FC0169
	// #FF9900
	// #BC3BBB
	// #7A55D4
	// #97B92D
	// #006E5E
	// #C18000
	// #8C5A3F
	settingRepo := repository.NewSettingRepo(s.dbm.GetDB(ctx.GetDbId()))
	settingData := settingRepo.GetByKey(constant.SettingBatchColor)
	if settingData.Key == "" {
		logger.Logger.Error("商家设置不存在或获取失败", zap.Uint64("companyUuid", ctx.GetDbId()))
		return nil, errors.WithMessage(errors.New("商家设置不存在或获取失败"), "商家设置不存在或获取失败")
	}
	// settingData.Values是json格式，需要转换为[]string。 json数组格式示例：["#FF585B", "#FC0169", "#FF9900", "#BC3BBB", "#7A55D4", "#97B92D", "#006E5E", "#C18000", "#8C5A3F"]
	colors := make([]string, 0)
	err = json.Unmarshal([]byte(settingData.Values), &colors)
	if err != nil {
		return nil, errors.WithMessage(err, "转换商家设置失败")
	} else if len(colors) == 0 {
		return nil, errors.WithMessage(errors.New("没有设置分批类型颜色"), "没有设置分批类型颜色")
	}

	// 将batchTags转换为map
	batchTagMap := make(map[string]*model.BatchTag)
	for _, batchTag := range batchTags {
		batchTagMap[batchTag.Color] = batchTag
	}

	colorUsageMap := make([]product_resp.BatchTagColorUsage, len(colors))
	for i, color := range colors {
		batchTag, ok := batchTagMap[color]
		if ok {
			colorUsageMap[i] = product_resp.BatchTagColorUsage{
				Color:        color,
				IsUsed:       true,
				UsedBy:       batchTag.Name,
				BatchTagUuid: batchTag.Uuid,
			}
		} else {
			colorUsageMap[i] = product_resp.BatchTagColorUsage{
				Color:        color,
				IsUsed:       false,
				UsedBy:       "",
				BatchTagUuid: 0,
			}
		}
	}

	return &product_resp.BatchTagColorUsageList{
		List: colorUsageMap,
	}, nil
}

// SaveBatchProduct 保存分批商品
func (s *productSrv) SaveBatchProduct(ctx context.Context, req req.SaveBatchProductReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productPackageRepo := repository.NewProductPackageRepo(db)
	// 查询出所选商品中还没有是分批商品的商品包
	productPackages, err := productPackageRepo.GetProductPackageListByUuidsAndIsBatch(req.Uuids, 0)
	if err != nil {
		return errors.WithMessage(err, "查询商品包失败")
	}
	// 查询所有已经是分批商品的商品包
	productPackagesBatch, err := productPackageRepo.GetProductPackageListByIsBatch(1)
	if err != nil {
		return errors.WithMessage(err, "查询商品包失败")
	}
	batchUuids := make(map[uint64]bool) // 已经是分批商品的商品包uuid列表
	for _, productPackage := range productPackagesBatch {
		batchUuids[productPackage.Uuid] = true
	}
	requestUuids := make(map[uint64]bool) // 请求的商品包uuid列表
	for _, uuid := range req.Uuids {
		requestUuids[uuid] = true
	}
	uuids := make([]uint64, 0) // 需要变为分批商品的。 不是分批商品 --> 分批商品
	for _, productPackage := range productPackages {
		uuids = append(uuids, productPackage.Uuid)
	}
	// 需要变成非分批商品的。 分批商品 --> 非分批商品.
	nonBatchUuids := make([]uint64, 0)
	for _, productPackage := range productPackagesBatch {
		if !requestUuids[productPackage.Uuid] {
			nonBatchUuids = append(nonBatchUuids, productPackage.Uuid)
		}
	}

	// 将这些商品包设置为分批商品
	if len(uuids) > 0 {
		err = productPackageRepo.SetProductPackageBatch(uuids, 1)
		if err != nil {
			return errors.WithMessage(err, "保存分批商品失败")
		}
	}
	// 将这些商品包设置为非分批商品
	if len(nonBatchUuids) > 0 {
		err = productPackageRepo.SetProductPackageBatch(nonBatchUuids, 0)
		if err != nil {
			return errors.WithMessage(err, "保存非分批商品失败")
		}
	}
	return nil
}

func (s *productSrv) SyncProductStockByBomCard(ctx context.Context) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	productBomRepo := repository.NewProductBomRepo(db)
	productBomList, err := productBomRepo.GetProductBomsByHasCard()
	if err != nil {
		return errors.WithMessage(err, "获取有成本卡商品列表失败")
	}
	err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		for _, productBom := range productBomList {
			productBomCard := productBom.ProductBomCard
			if productBomCard == nil {
				logger.Logger.Error("商品规格没有成本卡无法重新计算商品库存", zap.Uint64("productBomUuid", productBom.Uuid), zap.Any("productBom", productBom))
				continue
			}
			expectedProductionNum := productBomCard.CalculateExpectedProductionNum()
			if err := repository.NewProductBomRepo(tx).UpdateProductBomCard(productBom.Uuid, productBomCard.Uuid, expectedProductionNum); err != nil {
				return errors.WithMessage(err, "更新商品库存失败")
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "计算商品库存失败")
	}
	return nil
}

// SyncProductPackageImage 同步商品包图片
func (s *productSrv) SyncProductPackageImage(ctx context.Context) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsSubShop() {
		return nil
	}
	var files []model.File
	var fileGroups []model.FileGroup
	headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
	fileUuidQuery := headquarterDb.Model(&model.ProductPackage{}).Where("image_file_uuid > 0").Select("image_file_uuid")
	err := headquarterDb.Model(&model.File{}).Where("uuid in (?)", fileUuidQuery).Find(&files).Error
	if err != nil {
		return errors.WithMessage(errors.New("查询文件失败"), err.Error())
	}
	fileGroupUuidQuery := headquarterDb.Model(&model.File{}).Where("uuid in (?)", fileUuidQuery).Where("group_uuid > 0").Select("group_uuid")
	err = headquarterDb.Model(&model.FileGroup{}).Where("uuid in (?)", fileGroupUuidQuery).Find(&fileGroups).Error
	if err != nil {
		return errors.WithMessage(err, "查询文件分组失败")
	}
	var newFiles []model.File
	var newFileGroups []model.FileGroup
	for _, file := range files {
		newFiles = append(newFiles, model.File{
			BaseModel: model.BaseModel{
				Uuid:       file.Uuid,
				CreateTime: file.CreateTime,
				UpdateTime: file.UpdateTime,
				DeleteTime: file.DeleteTime,
			},
			Storage:         file.Storage,
			GroupUuid:       file.GroupUuid,
			HeadquarterUuid: companySetting.HeadquarterUuid,
			FileUrl:         file.FileUrl,
			SaveName:        file.SaveName,
			FileName:        file.FileName,
			FileSize:        file.FileSize,
			FileType:        file.FileType,
			RealName:        file.RealName,
			UrlParam:        file.UrlParam,
			IndexFileName:   file.IndexFileName,
			Extension:       file.Extension,
			IsUser:          file.IsUser,
			IsRecycle:       file.IsRecycle,
		})
	}
	for _, fileGroup := range fileGroups {
		newFileGroups = append(newFileGroups, model.FileGroup{
			BaseModel: model.BaseModel{
				Uuid:       fileGroup.Uuid,
				CreateTime: fileGroup.CreateTime,
				UpdateTime: fileGroup.UpdateTime,
				DeleteTime: fileGroup.DeleteTime,
			},
			GroupType:       fileGroup.GroupType,
			GroupName:       fileGroup.GroupName,
			Sort:            fileGroup.Sort,
			HeadquarterUuid: companySetting.HeadquarterUuid,
		})
	}
	// 删除后迁移
	err = s.dbm.GetDB(companySetting.CompanyUuid).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).Delete(&model.File{}).Error; err != nil {
			return errors.WithMessage(errors.New("删除总部文件失败"), err.Error())
		}
		if err := tx.Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).Delete(&model.FileGroup{}).Error; err != nil {
			return errors.WithMessage(errors.New("删除总部文件分组失败"), err.Error())
		}
		if len(newFiles) > 0 {
			if err := tx.Create(&newFiles).Error; err != nil {
				return errors.WithMessage(errors.New("同步总部文件失败"), err.Error())
			}
		}
		if len(newFileGroups) > 0 {
			if err := tx.Create(&newFileGroups).Error; err != nil {
				return errors.WithMessage(errors.New("同步总部分组失败"), err.Error())
			}
		}
		return nil
	})
	return err
}
