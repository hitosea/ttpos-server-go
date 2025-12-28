package printer

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/modules/printer/template"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
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
	PrintingDishes(printType int, saleBillUuid uint64, saleOrderUuid uint64, products printer_model.Products) bool
	PrintingStatementOrder(printType int, saleBill *model.SaleBill, saleOrderUuid uint64, firstExecution int, payMethodUuid uint64) (*resp.PrinterData, error)
	PrintingInvoice(saleBill *model.SaleBill, saleOrderUuid uint64, firstExecution int) (*resp.PrinterData, error)
	PrintingRechargeOrder(order model.MemberRechargeOrder, firstExecution int) (*resp.PrinterData, error)
	PrintingHandoverOrder(log *model.StaffShiftLog, businessData *business_data_resp.BusinessDataAll, firstExecution int, openMoneybox bool, deviceSnId ...string) (*resp.PrinterData, error)
	PrintingBusinessData(businessData *template.PrintingBusinessData, startTime int64, endTime int64, firstExecution int) (*resp.PrinterData, error)
	PrintingTakeoutOrder(memberSaleOrder *model.MemberSaleOrder, saleBill *model.SaleBill, saleOrderUuid uint64) (*resp.PrinterData, error)
	PrintingPlatformTakeoutReceipt(order *takeoutModel.TakeoutOrder, receiptType string, firstExecution int) (*resp.PrinterData, error)
	DeleteProductPrinterListCache()     // 删除商品打印机列表缓存
	SetFinishedTime(finishedTime int64) // 设置完成时间
	GetFinishedTime() int64             // 获取完成时间
	SetPrinterWidth(printerWidth int)   // 设置打印机宽度
	GetPrinterWidth() int               // 获取打印机宽度
	Is58mmPrinter() bool                // 是否58mm打印机
}

type PrinterRepoImpl struct {
	ctx             context.Context
	dbm             *database.DBManager
	cache           cache.Cache
	setting         *setting.Srv
	storeSetting    respSetting.Store
	printerSetting  respSetting.Printer
	currencySetting respSetting.Currency
	printMethod     int
	Lang            string // 可选语言参数
	finishedTime    int64  // 完成时间
	printerWidth    int    // 打印机宽度mm
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

// 获取打印机模板
func (p *PrinterRepoImpl) GetPrinterTemplate(id uint64) (int, uint64, string) {
	// 获取打印机模板
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())
	printerTemplateRepo, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(id)
	if err != nil {
		logger.Logger.Error("获取打印机模板失败", zap.Error(err))
		return 1, 0, ""
	}
	return printerTemplateRepo.Template, printerTemplateRepo.TmpUuid, printerTemplateRepo.TmpData
}

// 获取打印机模板详情
func (p *PrinterRepoImpl) GetPrinterTemplateInfo(id uint64) model.PrinterTemplate {
	// 获取打印机模板
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())
	printerTemplateRepo, err := repository.NewPrinterTemplateRepo(db).GetPrinterTemplateInfo(id)
	if err != nil {
		logger.Logger.Error("获取打印机模板失败", zap.Error(err))
		return model.PrinterTemplate{}
	}
	return printerTemplateRepo
}

// 获取商品打印机列表
func (p *PrinterRepoImpl) getProductPrinterList(widthPrintMode int) ([]model.ProductPrinter, error) {
	// 构建缓存键
	cacheKey := constant.GetRedisKeyProductPrinterListV2(p.ctx.GetCompanyUuid(), widthPrintMode)

	// 尝试从缓存获取
	if cachedData, found := p.cache.Get(cacheKey); found {
		// 尝试反序列化缓存数据
		var printers []model.ProductPrinter
		cachedBytes, ok := cachedData.(string)
		if ok {
			// 解压数据
			decompressed, err := decompress(cachedBytes)
			if err == nil {
				// 反序列化
				err := json.Unmarshal([]byte(decompressed), &printers)
				if err == nil && len(printers) > 0 {
					return printers, nil
				}
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
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterRegions",
			Args: []any{
				repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
			},
		}),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterItems",
			Args: []any{
				repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
			},
		}),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterItems.Printer",
			Args: []any{
				repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
			},
		}),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterItems.Printer.PrinterType",
			Args: []any{
				repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
			},
		}),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPrinterProductItems",
			Args: []any{
				repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
			},
		}),
	)
	//
	if err != nil {
		logger.Logger.Error("获取商品打印机列表失败", zap.Error(err))
		return []model.ProductPrinter{}, err
	}

	// 查询成功，将结果存入缓存
	if len(printers) > 0 {
		// 对每个ProductPrinter的ProductPrinterItems进行去重
		for i := range printers {
			printers[i].ProductPrinterItems = removeDuplicateProductPrinterItems(printers[i].ProductPrinterItems)
		}
		// 序列化
		printersBytes, err := json.Marshal(printers)
		if err == nil {
			// 缓存1天
			compressed, err := compress(printersBytes)
			if err == nil {
				p.cache.Set(cacheKey, compressed, 24*time.Hour)
			}
		}
	}

	return printers, nil
}

// compress 压缩数据
func compress(data []byte) (string, error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression) // 或使用gzip.DefaultCompression
	if err != nil {
		return "", err
	}
	_, err = zw.Write([]byte(data))
	if err != nil {
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded, nil
}

// decompress 解压数据
func decompress(data string) (string, error) {
	// 解码Base64
	compressed, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	// 解压数据
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return "", err
	}
	defer zr.Close()
	//
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// 获取打印机打印方式
func (p *PrinterRepoImpl) SetPrinterMethod(printMethod int, isKitchen ...bool) int {
	if printMethod != 0 {
		p.printMethod = printMethod
		if len(isKitchen) > 0 && isKitchen[0] {
			p.printerSetting.KitchenPrintMethod = strconv.Itoa(printMethod)
		} else {
			p.printerSetting.PrintMethod = strconv.Itoa(printMethod)
		}
	}
	return p.GetPrinterMethod(isKitchen...)
}

// 获取打印机打印方式
func (p *PrinterRepoImpl) GetPrinterMethod(isKitchen ...bool) int {
	if p.printMethod != 0 {
		return p.printMethod
	}
	printMethod := printerConst.PrinterLogPrintMethodText
	if len(isKitchen) > 0 && isKitchen[0] {
		if p.printerSetting.KitchenPrintMethod == "2" {
			printMethod = printerConst.PrinterLogPrintMethodImage
		}
	} else {
		if p.printerSetting.PrintMethod == "2" {
			printMethod = printerConst.PrinterLogPrintMethodImage
		}
	}
	return printMethod
}

// 获取打印机打印方式
func (p *PrinterRepoImpl) IsImagePrinterMethod(isKitchen ...bool) bool {
	return p.GetPrinterMethod(isKitchen...) == printerConst.PrinterLogPrintMethodImage
}

// 获取打印机宽度
func (p *PrinterRepoImpl) Is58mmPrinter() bool {
	return p.GetPrinterWidth() == 58
}

// 设置完成时间
func (p *PrinterRepoImpl) SetFinishedTime(finishedTime int64) {
	p.finishedTime = finishedTime
}

// 获取完成时间
func (p *PrinterRepoImpl) GetFinishedTime() int64 {
	return p.finishedTime
}

// 设置打印机宽度
func (p *PrinterRepoImpl) SetPrinterWidth(printerWidth int) {
	p.printerWidth = printerWidth
}

// 获取打印机宽度
func (p *PrinterRepoImpl) GetPrinterWidth() int {
	return p.printerWidth
}

// removeDuplicateProductPrinterItems 去除重复的ProductPrinterItems
// 根据PrinterUuid进行去重，保留第一个出现的记录
func removeDuplicateProductPrinterItems(items []*model.ProductPrinterItem) []*model.ProductPrinterItem {
	if len(items) <= 1 {
		return items
	}
	// 使用map记录已经出现过的PrinterUuid
	seen := make(map[uint64]bool)
	result := make([]*model.ProductPrinterItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		// 如果这个PrinterUuid还没有出现过，则添加到结果中
		if !seen[item.PrinterUuid] {
			seen[item.PrinterUuid] = true
			result = append(result, item)
		}
	}
	return result
}

// 删除商品打印机列表缓存
func (p *PrinterRepoImpl) DeleteProductPrinterListCache() {
	if p.cache == nil {
		return
	}
	companyUuid := p.ctx.GetCompanyUuid()
	if companyUuid == 0 {
		return
	}
	p.cache.Del(constant.GetRedisKeyProductPrinterListV2(companyUuid, 0))
	p.cache.Del(constant.GetRedisKeyProductPrinterListV2(companyUuid, 1))
	p.cache.Del(constant.GetRedisKeyProductPrinterListV2(companyUuid, -1))
	p.cache.Del(constant.GetRedisKeyProductPrinterListV2(companyUuid, -2))
}
