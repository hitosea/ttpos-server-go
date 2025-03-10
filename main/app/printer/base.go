package printer

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// PPrinterRepo 打印
type PPrinterRepo interface {
	PrintingDishes(printType int, saleBillUuid uint64, products Products) bool
}

type PrinterRepoImpl struct {
	Ctx            context.Context
	dbm            *database.DBManager
	cache          cache.Cache
	setting        *setting.Srv
	storeSetting   respSetting.Store
	printerSetting respSetting.Printer
}

func NewPrinterRepo(ctx context.Context) PPrinterRepo {
	dbm := database.GetDBManager(config.DatabaseConf{})
	//
	setting := setting.NewSrvImpl(dbm, cache.Global)
	// 获取门店设置
	storeSetting, err := setting.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店设置失败", zap.Error(err))
		return nil
	}
	// 获取打印机设置
	printerSetting, err := setting.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
		return nil
	}
	//
	return &PrinterRepoImpl{
		Ctx:            ctx,
		dbm:            dbm,
		cache:          cache.Global,
		setting:        setting,
		storeSetting:   storeSetting,
		printerSetting: printerSetting,
	}
}

// Products 送厨商品列表
type Products []OrderProduct
type OrderProduct struct {
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	TotalNum       uint               `json:"total_num"`        // 总数量
}

// ProductPrinter 商品打印机
type ProductPrinter struct {
	Uuid   uint64 `json:"uuid"`   // 商品打印机uuid
	Name   string `json:"name"`   // 商品打印机名称
	Status int    `json:"status"` // 商品打印机状态
}

// 获取商品打印机列表
func (p *PrinterRepoImpl) getProductPrinterList() ([]model.ProductPrinter, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("PRODUCT_PRINTER_LIST:%d", p.Ctx.GetCompanyUuid())

	// 尝试从缓存获取
	if cachedData, found := p.cache.Get(cacheKey); found {
		// 尝试反序列化缓存数据
		var printers []model.ProductPrinter
		cachedBytes, ok := cachedData.([]byte)
		if ok {
			err := json.Unmarshal(cachedBytes, &printers)
			if err == nil && len(printers) > 0 {
				return printers, nil
			}
		}
		// 反序列化失败或数据为空，删除无效缓存
		p.cache.Del(cacheKey)
	}

	// 缓存未命中，从数据库查询
	db := p.dbm.GetDB(p.Ctx.GetCompanyUuid())
	// 创建商品打印机仓库
	productPrinterRepo := repository.NewProductPrinterRepo(db)
	// 获取商品打印机列表
	printers, err := productPrinterRepo.GetProductPrinters(
		productPrinterRepo.WhereStatus(constant.ProductPrinterStatusOpen),
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterRegions",
		}),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterItems.Printer.PrinterType",
		}),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterProductItems",
		}),
	)
	//
	if err != nil {
		logger.Logger.Error("获取商品打印机列表失败", zap.Error(err))
		return []model.ProductPrinter{}, err
	}

	// 查询成功，将结果存入缓存
	if len(printers) > 0 {
		printersBytes, err := json.Marshal(printers)
		if err == nil {
			// 缓存1天
			p.cache.Set(cacheKey, printersBytes, 24*time.Hour)
		}
	}

	return printers, nil
}
