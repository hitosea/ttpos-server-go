package printer

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/printer/template"
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
	PrintingDishes(printType int, saleBillUuid uint64, products printer_model.Products) bool
	PrintingStatementOrder(printType int, saleBill *model.SaleBill, saleOrderUuid uint64, FirstExecution int, payMethodUuid uint64) (*resp.PrinterData, error)
	PrintingInvoice(saleBill *model.SaleBill, saleOrderUuid uint64, firstExecution int) (*resp.PrinterData, error)
	PrintingRechargeOrder(order model.MemberRechargeOrder, FirstExecution int) (*resp.PrinterData, error)
	PrintingHandoverOrder(log *model.StaffShiftLog, businessData *business_data_resp.BusinessDataAll, FirstExecution int, openMoneybox bool, deviceSnId ...string) (*resp.PrinterData, error)
	PrintingBusinessData(businessData *template.PrintingBusinessData, startTime int64, endTime int64, deviceSnId ...string) (*resp.PrinterData, error)
}

type PrinterRepoImpl struct {
	ctx             context.Context
	dbm             *database.DBManager
	cache           cache.Cache
	setting         *setting.Srv
	storeSetting    respSetting.Store
	printerSetting  respSetting.Printer
	currencySetting respSetting.Currency
	Lang            string // 可选语言参数
}

func NewPrinterRepo(ctx context.Context, langs ...string) PPrinterRepo {
	dbm := database.GetDBManager(config.DatabaseConf{})
	//
	setting := setting.NewSrvImpl(dbm, cache.Global)
	// 获取门店设置
	storeSetting, err := setting.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店设置失败", zap.Error(err))
		fmt.Println("获取门店设置失败", zap.Error(err))
	}
	// 获取打印机设置
	printerSetting, err := setting.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
		fmt.Println("获取打印机设置失败", zap.Error(err))
	}
	// 获取货币设置
	currencySetting, err := setting.GetCurrencySetting(ctx)
	if err != nil {
		logger.Logger.Error("获取货币设置失败", zap.Error(err))
		fmt.Println("获取货币设置失败", zap.Error(err))
	}

	// 创建打印机实例
	printerRepo := &PrinterRepoImpl{
		ctx:             ctx,
		dbm:             dbm,
		cache:           cache.Global,
		setting:         setting,
		storeSetting:    storeSetting,
		printerSetting:  printerSetting,
		currencySetting: currencySetting,
	}

	// 设置语言参数
	if len(langs) > 0 {
		printerRepo.Lang = langs[0]
	} else {
		// 使用打印机设置中的默认语言
		printerRepo.Lang = printerSetting.DefaultLanguage
	}

	return printerRepo
}

// 获取商品打印机列表
func (p *PrinterRepoImpl) GetPrinterTemplate(id uint64) int {
	// 获取打印机模板
	printerTemplateRepo, err := repository.NewPrinterTemplateRepo(p.dbm.GetDB(p.ctx.GetCompanyUuid())).GetPrinterTemplateInfo(id)
	if err != nil {
		logger.Logger.Error("获取打印机模板失败", zap.Error(err))
		return 1
	}
	return printerTemplateRepo.Template
}

// 获取商品打印机列表
func (p *PrinterRepoImpl) getProductPrinterList(widthPrintMode int) ([]model.ProductPrinter, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("PRODUCT_PRINTER_LIST:%d:%d", p.ctx.GetCompanyUuid(), widthPrintMode)

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
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())
	// 创建商品打印机仓库
	productPrinterRepo := repository.NewProductPrinterRepo(db)
	// 获取商品打印机列表
	printers, err := productPrinterRepo.GetProductPrinters(
		productPrinterRepo.WhereStatus(constant.ProductPrinterStatusOpen),
		productPrinterRepo.WidthPrintMode(widthPrintMode),
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
