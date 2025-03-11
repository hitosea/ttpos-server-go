package printer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"unicode"

	"go.uber.org/zap"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// PPrinterRepo 打印
type PPrinterRepo interface {
	PrintingDishes(printType int, saleBillUuid uint64, products printer_model.Products) bool
}

type PrinterRepoImpl struct {
	ctx             context.Context
	dbm             *database.DBManager
	cache           cache.Cache
	setting         *setting.Srv
	storeSetting    respSetting.Store
	printerSetting  respSetting.Printer
	currencySetting respSetting.Currency
}

// hex2bin 将十六进制字符串转换为二进制数据
// 类似于PHP中的同名函数
func hex2bin(hexStr string) string {
	// 移除可能存在的空格
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	// 解码十六进制字符串
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Printf("解析十六进制字符串出错: %v\n", err)
		return ""
	}

	// 返回解码后的字符串
	return string(decoded)
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
	// 获取货币设置
	currencySetting, err := setting.GetCurrencySetting(ctx)
	if err != nil {
		logger.Logger.Error("获取货币设置失败", zap.Error(err))
		return nil
	}
	//
	return &PrinterRepoImpl{
		ctx:             ctx,
		dbm:             dbm,
		cache:           cache.Global,
		setting:         setting,
		storeSetting:    storeSetting,
		printerSetting:  printerSetting,
		currencySetting: currencySetting,
	}
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
func (p *PrinterRepoImpl) getProductPrinterList() ([]model.ProductPrinter, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("PRODUCT_PRINTER_LIST:%d", p.ctx.GetCompanyUuid())

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

// ExecutePrinting 连接打印机并发送打印内容
// 支持多种语言（泰语、韩语、中文等）的字符编码处理
func (p *PrinterRepoImpl) ExecutePrinting(printerIP string, content string) error {
	// 连接打印机（TCP连接，端口9100是标准打印机端口）
	conn, err := net.DialTimeout("tcp", printerIP+":9100", 3*time.Second)
	if err != nil {
		return fmt.Errorf("连接打印机出错: %v", err)
	}
	defer conn.Close()
	// 转换十六进制字符串为二进制数据
	content = hex2bin(content)
	// 替换特殊字符
	content = strings.ReplaceAll(content, "ー", "-")
	// 使用正则表达式分割文本，保留泰语、韩语和泰铢符号
	thaiHangulRegex := regexp.MustCompile(`([\p{Thai}\p{Hangul}฿]+)`)
	segments := thaiHangulRegex.Split(content, -1)
	matches := thaiHangulRegex.FindAllString(content, -1)
	// 合并分割结果和匹配结果，按原始顺序
	allSegments := make([]string, 0)
	matchIndex := 0
	for _, seg := range segments {
		if seg != "" {
			allSegments = append(allSegments, seg)
		}
		if matchIndex < len(matches) {
			allSegments = append(allSegments, matches[matchIndex])
			matchIndex++
		}
	}
	// 处理每个文本段落，根据语言类型选择对应的编码
	for _, segment := range allSegments {
		if segment == "" {
			continue
		}
		// 检查是否包含泰语字符或泰铢符号
		isThai := containsThai(segment) || strings.Contains(segment, "฿")
		isKorean := containsKorean(segment)
		//
		if isThai {
			// 泰语处理
			// 切换到泰语字符集
			_, err = conn.Write([]byte{0x1C, 0x2E})
			if err != nil {
				return err
			}
			// 转换为CP874编码（泰语字符集）
			encoded, err := encodeTo(segment, "cp874")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		} else if isKorean {
			// 韩语处理
			// 切换到韩语字符集
			_, err = conn.Write([]byte{0x1C, 0x26})
			if err != nil {
				return err
			}
			// 转换为CP949编码（韩语字符集）
			encoded, err := encodeTo(segment, "cp949")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		} else {
			// 其他语言（默认使用中文GBK编码）
			// 切换到中文字符集
			_, err = conn.Write([]byte{0x1C, 0x26})
			if err != nil {
				return err
			}
			// 转换为GBK编码（中文字符集）
			encoded, err := encodeTo(segment, "gbk")
			if err != nil {
				return err
			}
			conn.Write(encoded)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// 检查字符串是否包含泰语字符
func containsThai(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Thai, r) {
			return true
		}
	}
	return false
}

// 检查字符串是否包含韩语字符
func containsKorean(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

// 将字符串转换为指定编码
func encodeTo(s string, encoding string) ([]byte, error) {
	switch encoding {
	case "cp874":
		enc := charmap.Windows874.NewEncoder()
		return enc.Bytes([]byte(s))
	case "cp949":
		enc := korean.EUCKR.NewEncoder()
		return enc.Bytes([]byte(s))
	case "gbk":
		enc := simplifiedchinese.GBK.NewEncoder()
		return enc.Bytes([]byte(s))
	default:
		return []byte(s), nil
	}
}
