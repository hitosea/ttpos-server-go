package command

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/model"
	newsetting "ttpos-server-go/app/modules/setting"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
	oldsetting "ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	pkgctx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(verifySettingServiceCmd)
}

// 验证 Setting 服务命令
var verifySettingServiceCmd = &cobra.Command{
	Use:   "verify-setting-service",
	Short: "验证新旧 Setting 服务输出一致性",
	Long:  `验证重构后的 Setting 服务（DDD 模块）与旧服务的输出是否完全一致`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s", redColor, err, resetColor)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 初始化id生成器
		utils.InitIdGenerator()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runSettingServiceVerification()
	},
}

// MethodVerifyResult 方法验证结果
type MethodVerifyResult struct {
	MethodName string
	Passed     bool
	OldJSON    string
	NewJSON    string
	Error      string
}

// runSettingServiceVerification 运行 Setting 服务验证
func runSettingServiceVerification() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("Setting 服务兼容性验证")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// 第一步：创建测试商户
	fmt.Printf("%s 第一步：创建测试商户... %s\n", blueColor, resetColor)
	testCompanyUuid, err := createSettingTestCompany()
	if err != nil {
		fmt.Printf("%s 错误: 创建测试商户失败: %v %s\n", redColor, err, resetColor)
		return
	}
	defer cleanupSettingTestCompany(testCompanyUuid)

	// 第二步：创建测试数据库并初始化
	fmt.Printf("\n%s 第二步：创建测试数据库并初始化... %s\n", blueColor, resetColor)
	dbName := fmt.Sprintf("%s%d", constant.DBNamePrefix, testCompanyUuid)
	if err := createSettingTestDatabase(dbName); err != nil {
		fmt.Printf("%s 错误: 创建测试数据库失败: %v %s\n", redColor, err, resetColor)
		return
	}

	// 执行 shop_01.sql
	if err := executeSettingShop01SQL(dbName); err != nil {
		fmt.Printf("%s 错误: 执行 shop_01.sql 失败: %v %s\n", redColor, err, resetColor)
		return
	}
	fmt.Printf("%s  ✓ 测试数据库初始化成功 %s\n", greenColor, resetColor)

	// 第三步：初始化新旧服务
	fmt.Printf("\n%s 第三步：初始化新旧 Setting 服务... %s\n", blueColor, resetColor)
	dbm := database.GetDBManager(config.Database)
	cacheInstance := cache.Global

	oldSrv := oldsetting.NewSrv(dbm, cacheInstance)
	newSrv := newsetting.NewSrv(dbm, cacheInstance)

	fmt.Printf("%s  ✓ 服务初始化成功 %s\n", greenColor, resetColor)

	// 第四步：创建测试 Context
	// 创建一个模拟的 gin.Context 用于设置服务中的请求处理
	mockGinCtx := &gin.Context{}
	mockGinCtx.Request = &http.Request{
		Host: "localhost:8080",
		URL:  &url.URL{Host: "localhost:8080"},
	}

	ctx := pkgctx.NewContext(
		pkgctx.WithCompanyUuid(testCompanyUuid),
		pkgctx.WithGinContext(mockGinCtx),
	)
	// 设置数据库连接
	companyDB := dbm.GetDB(testCompanyUuid)
	ctx.SetDB(companyDB)

	// 第五步：逐个方法对比验证
	fmt.Printf("\n%s 第四步：开始方法对比验证... %s\n", blueColor, resetColor)
	fmt.Println("-" + strings.Repeat("-", 79))

	results := verifyAllMethods(ctx, oldSrv, newSrv)

	// 第六步：输出验证结果
	fmt.Printf("\n%s 第五步：验证结果汇总... %s\n", blueColor, resetColor)
	fmt.Println("=" + strings.Repeat("=", 79))

	passCount := 0
	failCount := 0
	for _, result := range results {
		if result.Passed {
			passCount++
			fmt.Printf("%s ✅ %s %s\n", greenColor, result.MethodName, resetColor)
		} else {
			failCount++
			fmt.Printf("%s ❌ %s %s\n", redColor, result.MethodName, resetColor)
			if result.Error != "" {
				fmt.Printf("    错误: %s\n", result.Error)
			} else {
				fmt.Printf("    差异详情:\n")
				printJSONDiff(result.OldJSON, result.NewJSON)
			}
		}
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Printf("总计: %d 个方法, %s%d 通过%s, %s%d 失败%s\n",
		len(results),
		greenColor, passCount, resetColor,
		redColor, failCount, resetColor)

	if failCount == 0 {
		fmt.Printf("\n%s ✅ 所有方法验证通过！新旧服务输出完全一致。%s\n", greenColor, resetColor)
	} else {
		fmt.Printf("\n%s ⚠️  存在 %d 个方法输出不一致，请检查差异。%s\n", redColor, failCount, resetColor)
	}
}

// verifyAllMethods 验证所有方法
func verifyAllMethods(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) []MethodVerifyResult {
	var results []MethodVerifyResult

	// 1. GetStoreSetting
	results = append(results, verifyGetStoreSetting(ctx, oldSrv, newSrv))

	// 2. GetStoreLanguageList
	results = append(results, verifyGetStoreLanguageList(ctx, oldSrv, newSrv))

	// 3. GetStoreLanguage
	results = append(results, verifyGetStoreLanguage(ctx, oldSrv, newSrv))

	// 4. GetBusinessSetting
	results = append(results, verifyGetBusinessSetting(ctx, oldSrv, newSrv))

	// 5. GetShopBusinessSetting
	results = append(results, verifyGetShopBusinessSetting(ctx, oldSrv, newSrv))

	// 6. GetCurrencySetting
	results = append(results, verifyGetCurrencySetting(ctx, oldSrv, newSrv))

	// 7. GetKioskSetting
	results = append(results, verifyGetKioskSetting(ctx, oldSrv, newSrv))

	// 8. GetPointsSetting
	results = append(results, verifyGetPointsSetting(ctx, oldSrv, newSrv))

	// 9. GetCloudBasicSetting
	results = append(results, verifyGetCloudBasicSetting(ctx, oldSrv, newSrv))

	// 10. GetServiceFeeSetting
	results = append(results, verifyGetServiceFeeSetting(ctx, oldSrv, newSrv))

	// 11. GetTaxRateSetting
	results = append(results, verifyGetTaxRateSetting(ctx, oldSrv, newSrv))

	// 12. GetCompanySetting
	results = append(results, verifyGetCompanySetting(ctx, oldSrv, newSrv))

	// 13. GetCashierLanguage
	results = append(results, verifyGetCashierLanguage(ctx, oldSrv, newSrv))

	// 14. GetCashierAd
	results = append(results, verifyGetCashierAd(ctx, oldSrv, newSrv))

	// 15. GetCashierBaseSetting
	results = append(results, verifyGetCashierBaseSetting(ctx, oldSrv, newSrv))

	// 16. GetAcceptOrderSetting
	results = append(results, verifyGetAcceptOrderSetting(ctx, oldSrv, newSrv))

	// 17. GetMenuQrcode
	results = append(results, verifyGetMenuQrcode(ctx, oldSrv, newSrv))

	// 18. GetPaymentMethodList
	results = append(results, verifyGetPaymentMethodList(ctx, oldSrv, newSrv))

	// 19. GetDataManageSetting
	results = append(results, verifyGetDataManageSetting(ctx, oldSrv, newSrv))

	// 20. SymbolPosition
	results = append(results, verifySymbolPosition(ctx, oldSrv, newSrv))

	// 21. VerifyPassword
	results = append(results, verifyVerifyPassword(ctx, oldSrv, newSrv))

	// 需要 languageList 的方法
	oldLanguageList, _ := oldSrv.GetStoreLanguageList(ctx)
	newLanguageList := convertToNewLanguageList(oldLanguageList)

	// 22. GetCashierSetting
	results = append(results, verifyGetCashierSetting(ctx, oldSrv, newSrv, oldLanguageList, newLanguageList))

	// 23. GetPrinterSetting
	results = append(results, verifyGetPrinterSetting(ctx, oldSrv, newSrv, oldLanguageList, newLanguageList))

	// 24. GetAssistantSetting
	results = append(results, verifyGetAssistantSetting(ctx, oldSrv, newSrv, oldLanguageList, newLanguageList))

	// 25. GetH5Setting
	results = append(results, verifyGetH5Setting(ctx, oldSrv, newSrv, oldLanguageList, newLanguageList))

	// 26. GetTabletSetting
	results = append(results, verifyGetTabletSetting(ctx, oldSrv, newSrv, oldLanguageList, newLanguageList))

	// 需要 companySetting 的方法
	oldCompanySetting, _ := oldSrv.GetCompanySetting(ctx)

	// 27. GetBuffetSetting
	results = append(results, verifyGetBuffetSetting(ctx, oldSrv, newSrv, oldCompanySetting))

	// 28. GetPaymentSetting
	results = append(results, verifyGetPaymentSetting(ctx, oldSrv, newSrv, oldCompanySetting))

	// 29. GetKitchenSetting
	results = append(results, verifyGetKitchenSetting(ctx, oldSrv, newSrv, oldCompanySetting, oldLanguageList, newLanguageList))

	// 30. GetPrinterInfo (需要先获取 printerSetting)
	results = append(results, verifyGetPrinterInfo(ctx, oldSrv, newSrv, oldLanguageList, newLanguageList))

	// 31. UpdateSetting
	results = append(results, verifyUpdateSetting(ctx, oldSrv, newSrv))

	// 32. VerifyAdvancedPassword
	results = append(results, verifyVerifyAdvancedPassword(ctx, oldSrv, newSrv))

	// 33. CheckUpdate
	results = append(results, verifyCheckUpdate(ctx, oldSrv, newSrv))

	// Edit 方法暂不验证（需要编写操作后验证效果的逻辑）

	return results
}

// convertToNewLanguageList 将旧的 LanguageItem 列表转换为新的
func convertToNewLanguageList(oldList []dto.LanguageItem) []valueobject.LanguageItem {
	newList := make([]valueobject.LanguageItem, 0, len(oldList))
	for _, item := range oldList {
		newList = append(newList, valueobject.LanguageItem{
			Name:  item.Name,
			Value: item.Value,
		})
	}
	return newList
}

// compareJSON 比较两个对象的 JSON 序列化结果
func compareJSON(old, new any) (bool, string, string) {
	oldJSON, err1 := json.Marshal(old)
	newJSON, err2 := json.Marshal(new)

	if err1 != nil || err2 != nil {
		return false, fmt.Sprintf("error: %v", err1), fmt.Sprintf("error: %v", err2)
	}

	if bytes.Equal(oldJSON, newJSON) {
		return true, string(oldJSON), string(newJSON)
	}

	// 找出具体差异：解析为 map 然后比较
	var oldMap, newMap map[string]interface{}
	json.Unmarshal(oldJSON, &oldMap)
	json.Unmarshal(newJSON, &newMap)

	diff := findMapDiff(oldMap, newMap, "")
	if diff != "" {
		return false, string(oldJSON), string(newJSON)
	}

	return false, string(oldJSON), string(newJSON)
}

// findMapDiff 找出两个 map 的差异
func findMapDiff(old, new map[string]interface{}, prefix string) string {
	var diffs []string

	// 找出 old 中有但 new 中没有的键
	for k := range old {
		if _, exists := new[k]; !exists {
			diffs = append(diffs, fmt.Sprintf("缺少字段 %s%s", prefix, k))
		}
	}

	// 找出 new 中有但 old 中没有的键
	for k := range new {
		if _, exists := old[k]; !exists {
			diffs = append(diffs, fmt.Sprintf("多余字段 %s%s", prefix, k))
		}
	}

	// 比较共同的键
	for k, oldVal := range old {
		if newVal, exists := new[k]; exists {
			oldJSON, _ := json.Marshal(oldVal)
			newJSON, _ := json.Marshal(newVal)
			if !bytes.Equal(oldJSON, newJSON) {
				// 对于数组，逐个比较元素
				if oldArr, ok := oldVal.([]interface{}); ok {
					if newArr, ok := newVal.([]interface{}); ok && len(oldArr) == len(newArr) {
						for i := range oldArr {
							if oldItemJSON, _ := json.Marshal(oldArr[i]); !bytes.Equal(oldItemJSON, func() []byte { newItemJSON, _ := json.Marshal(newArr[i]); return newItemJSON }()) {
								diffs = append(diffs, fmt.Sprintf("字段 %s%s[%d] 不同: old=%s, new=%s", prefix, k, i, string(oldItemJSON), func() string { newItemJSON, _ := json.Marshal(newArr[i]); return string(newItemJSON) }()))
								break
							}
						}
					} else {
						// 数组长度不同
						diffs = append(diffs, fmt.Sprintf("字段 %s%s 长度不同: old_len=%d, new_len=%d", prefix, k, len(oldArr), len(newArr)))
					}
				} else {
					// 普通字段，只显示前 100 字符的差异
					oldStr := string(oldJSON)
					newStr := string(newJSON)
					if len(oldStr) > 100 {
						oldStr = oldStr[:100] + "..."
					}
					if len(newStr) > 100 {
						newStr = newStr[:100] + "..."
					}
					diffs = append(diffs, fmt.Sprintf("字段 %s%s 不同: old=%s, new=%s", prefix, k, oldStr, newStr))
				}
			}
		}
	}

	return strings.Join(diffs, "\n")
}

// printJSONDiff 打印 JSON 差异
func printJSONDiff(oldJSON, newJSON string) {
	// 简化输出：只显示前 500 字符
	maxLen := 500
	if len(oldJSON) > maxLen {
		oldJSON = oldJSON[:maxLen] + "..."
	}
	if len(newJSON) > maxLen {
		newJSON = newJSON[:maxLen] + "..."
	}
	fmt.Printf("    旧服务: %s\n", oldJSON)
	fmt.Printf("    新服务: %s\n", newJSON)
}

// === 各方法的验证函数 ===

// safeVerify 安全执行验证，捕获 panic
func safeVerify(methodName string, fn func() MethodVerifyResult) (result MethodVerifyResult) {
	result = MethodVerifyResult{MethodName: methodName}
	defer func() {
		if r := recover(); r != nil {
			result.Error = fmt.Sprintf("panic: %v", r)
			result.Passed = false
		}
	}()
	return fn()
}

func verifyGetStoreSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetStoreSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetStoreSetting"}

		oldResult, oldErr := oldSrv.GetStoreSetting(ctx)
		newResult, newErr := newSrv.GetStoreSetting(ctx)

		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}

		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetStoreLanguageList(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetStoreLanguageList", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetStoreLanguageList"}
		oldResult, oldErr := oldSrv.GetStoreLanguageList(ctx)
		newResult, newErr := newSrv.GetStoreLanguageList(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetStoreLanguage(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetStoreLanguage", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetStoreLanguage"}
		oldResult, oldErr := oldSrv.GetStoreLanguage(ctx)
		newResult, newErr := newSrv.GetStoreLanguage(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetBusinessSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetBusinessSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetBusinessSetting"}
		oldResult, oldErr := oldSrv.GetBusinessSetting(ctx)
		newResult, newErr := newSrv.GetBusinessSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetShopBusinessSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetShopBusinessSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetShopBusinessSetting"}
		oldResult, oldErr := oldSrv.GetShopBusinessSetting(ctx)
		newResult, newErr := newSrv.GetShopBusinessSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetCurrencySetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetCurrencySetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCurrencySetting"}
		oldResult, oldErr := oldSrv.GetCurrencySetting(ctx)
		newResult, newErr := newSrv.GetCurrencySetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetKioskSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetKioskSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetKioskSetting"}
		oldResult, oldErr := oldSrv.GetKioskSetting(ctx)
		newResult, newErr := newSrv.GetKioskSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetPointsSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetPointsSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetPointsSetting"}
		oldResult, oldErr := oldSrv.GetPointsSetting(ctx)
		newResult, newErr := newSrv.GetPointsSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetCloudBasicSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetCloudBasicSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCloudBasicSetting"}
		oldResult, oldErr := oldSrv.GetCloudBasicSetting(ctx)
		newResult, newErr := newSrv.GetCloudBasicSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetServiceFeeSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetServiceFeeSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetServiceFeeSetting"}
		oldResult, oldErr := oldSrv.GetServiceFeeSetting(ctx)
		newResult, newErr := newSrv.GetServiceFeeSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetTaxRateSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetTaxRateSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetTaxRateSetting"}
		oldResult, oldErr := oldSrv.GetTaxRateSetting(ctx)
		newResult, newErr := newSrv.GetTaxRateSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetCompanySetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetCompanySetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCompanySetting"}
		oldResult, oldErr := oldSrv.GetCompanySetting(ctx)
		newResult, newErr := newSrv.GetCompanySetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetCashierLanguage(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetCashierLanguage", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCashierLanguage"}
		oldResult, oldErr := oldSrv.GetCashierLanguage(ctx)
		newResult, newErr := newSrv.GetCashierLanguage(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetCashierAd(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetCashierAd", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCashierAd"}
		oldResult, oldErr := oldSrv.GetCashierAd(ctx)
		newResult, newErr := newSrv.GetCashierAd(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetCashierBaseSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetCashierBaseSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCashierBaseSetting"}
		oldResult, oldErr := oldSrv.GetCashierBaseSetting(ctx)
		newResult, newErr := newSrv.GetCashierBaseSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetAcceptOrderSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetAcceptOrderSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetAcceptOrderSetting"}
		oldResult, oldErr := oldSrv.GetAcceptOrderSetting(ctx)
		newResult, newErr := newSrv.GetAcceptOrderSetting(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetMenuQrcode(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetMenuQrcode", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetMenuQrcode"}
		oldResult, oldErr := oldSrv.GetMenuQrcode(ctx)
		newResult, newErr := newSrv.GetMenuQrcode(ctx)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		// GetMenuQrcode 返回包含时间戳的 token，所以无法进行精确比较
		// 改为验证格式：应该包含 "/home?token="
		if strings.Contains(oldResult, "/home?token=") && strings.Contains(newResult, "/home?token=") {
			result.Passed = true
			result.OldJSON = "valid_token_format"
			result.NewJSON = "valid_token_format"
		} else {
			result.Passed = false
			result.OldJSON = oldResult
			result.NewJSON = newResult
		}
		return result
	})
}

func verifyGetPaymentMethodList(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetPaymentMethodList", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetPaymentMethodList"}
		oldResult := oldSrv.GetPaymentMethodList(ctx)
		newResult := newSrv.GetPaymentMethodList(ctx)
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetDataManageSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("GetDataManageSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetDataManageSetting"}
		oldResult := oldSrv.GetDataManageSetting(ctx)
		newResult := newSrv.GetDataManageSetting(ctx)
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifySymbolPosition(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("SymbolPosition", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "SymbolPosition"}
		testPrice := 123.45
		oldResult := oldSrv.SymbolPosition(ctx, testPrice)
		newResult := newSrv.SymbolPosition(ctx, testPrice)
		result.Passed = oldResult == newResult
		result.OldJSON = oldResult
		result.NewJSON = newResult
		return result
	})
}

func verifyVerifyPassword(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("VerifyPassword", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "VerifyPassword"}
		source := "cashier"
		typ := "manager"
		password := "123456"
		oldResult := oldSrv.VerifyPassword(ctx, source, typ, password)
		newResult := newSrv.VerifyPassword(ctx, source, typ, password)
		result.Passed = oldResult == newResult
		result.OldJSON = fmt.Sprintf("%v", oldResult)
		result.NewJSON = fmt.Sprintf("%v", newResult)
		return result
	})
}

func verifyGetCashierSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetCashierSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetCashierSetting"}
		oldResult, oldErr := oldSrv.GetCashierSetting(ctx, oldLanguageList)
		newResult, newErr := newSrv.GetCashierSetting(ctx, newLanguageList)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetPrinterSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetPrinterSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetPrinterSetting"}
		oldResult, oldErr := oldSrv.GetPrinterSetting(ctx, oldLanguageList)
		newResult, newErr := newSrv.GetPrinterSetting(ctx, newLanguageList)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetAssistantSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetAssistantSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetAssistantSetting"}
		oldResult, oldErr := oldSrv.GetAssistantSetting(ctx, oldLanguageList)
		newResult, newErr := newSrv.GetAssistantSetting(ctx, newLanguageList)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetH5Setting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetH5Setting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetH5Setting"}
		oldResult, oldErr := oldSrv.GetH5Setting(ctx, oldLanguageList)
		newResult, newErr := newSrv.GetH5Setting(ctx, newLanguageList)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetTabletSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetTabletSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetTabletSetting"}
		oldResult, oldErr := oldSrv.GetTabletSetting(ctx, oldLanguageList)
		newResult, newErr := newSrv.GetTabletSetting(ctx, newLanguageList)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetBuffetSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, companySetting model.CompanySetting) MethodVerifyResult {
	return safeVerify("GetBuffetSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetBuffetSetting"}
		oldResult, oldErr := oldSrv.GetBuffetSetting(ctx, companySetting)
		newResult, newErr := newSrv.GetBuffetSetting(ctx, companySetting)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetPaymentSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, companySetting model.CompanySetting) MethodVerifyResult {
	return safeVerify("GetPaymentSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetPaymentSetting"}
		oldResult, oldErr := oldSrv.GetPaymentSetting(ctx, companySetting)
		newResult, newErr := newSrv.GetPaymentSetting(ctx, companySetting)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetKitchenSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, companySetting model.CompanySetting, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetKitchenSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetKitchenSetting"}
		oldResult, oldErr := oldSrv.GetKitchenSetting(ctx, companySetting, oldLanguageList)
		newResult, newErr := newSrv.GetKitchenSetting(ctx, companySetting, newLanguageList)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyGetPrinterInfo(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv, oldLanguageList []dto.LanguageItem, newLanguageList []valueobject.LanguageItem) MethodVerifyResult {
	return safeVerify("GetPrinterInfo", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "GetPrinterInfo"}
		// 先获取 printerSetting
		oldPrinterSetting, oldErr := oldSrv.GetPrinterSetting(ctx, oldLanguageList)
		newPrinterSetting, newErr := newSrv.GetPrinterSetting(ctx, newLanguageList)
		if oldErr != nil || newErr != nil {
			result.Error = fmt.Sprintf("获取 PrinterSetting 失败: oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		// 使用测试设备ID
		deviceId := "test_device_001"
		oldResult, oldErr := oldSrv.GetPrinterInfo(ctx, oldPrinterSetting, deviceId)
		newResult, newErr := newSrv.GetPrinterInfo(ctx, newPrinterSetting, deviceId)
		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}
		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

func verifyUpdateSetting(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("UpdateSetting", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "UpdateSetting"}
		// 测试参数
		settingKey := "test_key"
		values := map[string]interface{}{"test": "value"}

		oldErr := oldSrv.UpdateSetting(ctx, settingKey, values)
		newErr := newSrv.UpdateSetting(ctx, settingKey, values)

		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}

		result.Passed = true
		result.OldJSON = "success"
		result.NewJSON = "success"
		return result
	})
}

func verifyVerifyAdvancedPassword(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("VerifyAdvancedPassword", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "VerifyAdvancedPassword"}
		// 测试参数 - 使用默认密码 666888
		password := "666888"

		oldErr := oldSrv.VerifyAdvancedPassword(ctx, password)
		newErr := newSrv.VerifyAdvancedPassword(ctx, password)

		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}

		result.Passed = true
		result.OldJSON = "success"
		result.NewJSON = "success"
		return result
	})
}

func verifyCheckUpdate(ctx pkgctx.Context, oldSrv oldsetting.ISrv, newSrv newsetting.ISrv) MethodVerifyResult {
	return safeVerify("CheckUpdate", func() MethodVerifyResult {
		result := MethodVerifyResult{MethodName: "CheckUpdate"}
		// 测试参数
		appType := 1
		brand := "test_brand"
		language := "zh"

		oldResult, oldErr := oldSrv.CheckUpdate(ctx, appType, brand, language)
		newResult, newErr := newSrv.CheckUpdate(ctx, appType, brand, language)

		if oldErr != nil || newErr != nil {
			// 如果两个服务都返回错误，比较错误消息是否相同
			if oldErr != nil && newErr != nil {
				oldErrMsg := fmt.Sprintf("%v", oldErr)
				newErrMsg := fmt.Sprintf("%v", newErr)
				if oldErrMsg == newErrMsg {
					result.Passed = true
					result.OldJSON = oldErrMsg
					result.NewJSON = newErrMsg
					return result
				}
			}
			result.Error = fmt.Sprintf("oldErr: %v, newErr: %v", oldErr, newErr)
			return result
		}

		result.Passed, result.OldJSON, result.NewJSON = compareJSON(oldResult, newResult)
		return result
	})
}

// === 测试环境相关函数 ===

// createSettingTestCompany 创建测试商户
func createSettingTestCompany() (uint64, error) {
	fmt.Printf("  正在创建测试商户...\n")

	// 连接 saas 数据库
	saasDB, err := database.NewMySQLConnection(config.Database, "saas")
	if err != nil {
		return 0, fmt.Errorf("连接 saas 数据库失败: %w", err)
	}

	// 生成测试商户 UUID
	testUuid, err := utils.GetID()
	if err != nil {
		return 0, fmt.Errorf("生成商户UUID失败: %w", err)
	}

	// 创建测试商户记录
	company := &model.Company{
		BaseModel: model.BaseModel{
			Uuid:       testUuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:          "测试商户_验证Setting服务",
		Logo:          "",
		ExpireTime:    0,
		AuthDay:       0,
		Status:        1,
		AuthStartTime: time.Now().Unix(),
		OldCompanyId:  0,
		IsEnableErp:   0,
		LastSyncTime:  0,
	}

	if err := saasDB.Create(company).Error; err != nil {
		return 0, fmt.Errorf("创建商户记录失败: %w", err)
	}

	fmt.Printf("%s  ✓ 测试商户创建成功，UUID: %d %s\n", greenColor, testUuid, resetColor)
	return testUuid, nil
}

// createSettingTestDatabase 创建测试数据库
func createSettingTestDatabase(dbName string) error {
	fmt.Printf("  正在创建测试数据库: %s...\n", dbName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", dbName))
	if err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	fmt.Printf("%s  ✓ 数据库创建成功 %s\n", greenColor, resetColor)

	return nil
}

// executeSettingShop01SQL 执行 shop_01.sql 和迁移文件
func executeSettingShop01SQL(dbName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	workDir, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("获取项目根目录失败: %w", err)
	}

	// 执行 shop_01.sql
	sqlFile := filepath.Join(workDir, "admin", "database", "seeds", "shop_01.sql")
	fmt.Printf("  执行 shop_01.sql...\n")
	if err := executeSQLFile(db, sqlFile); err != nil {
		return fmt.Errorf("执行 shop_01.sql 失败: %w", err)
	}

	// 执行 shop_init_data.sql（如果存在）
	initDataFile := filepath.Join(workDir, "admin", "database", "seeds", "shop_init_data.sql")
	if _, err := os.Stat(initDataFile); err == nil {
		fmt.Printf("  执行 shop_init_data.sql...\n")
		if err := executeSQLFile(db, initDataFile); err != nil {
			return fmt.Errorf("执行 shop_init_data.sql 失败: %w", err)
		}
	}

	// 执行迁移文件（可选，如果 PHP 不可用则跳过）
	fmt.Printf("  执行数据库迁移...\n")
	if err := runSettingMigrations(); err != nil {
		fmt.Printf("  ⚠️  跳过数据库迁移（%v）\n", err)
	} else {
		fmt.Printf("  ✓ 数据库迁移成功\n")
	}

	return nil
}

// runSettingMigrations 执行数据库迁移文件
func runSettingMigrations() error {
	workDir, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("获取项目根目录失败: %w", err)
	}
	adminDir := filepath.Join(workDir, "admin")

	// 检查 admin 目录是否存在
	if _, err := os.Stat(adminDir); os.IsNotExist(err) {
		return fmt.Errorf("admin 目录不存在: %s", adminDir)
	}

	// 检查 think 文件是否存在
	thinkFile := filepath.Join(adminDir, "think")
	if _, err := os.Stat(thinkFile); os.IsNotExist(err) {
		return fmt.Errorf("think 文件不存在: %s", thinkFile)
	}

	// 执行迁移命令: make php-migrate (需要在根目录执行，因为 php-migrate 定义在根目录的 Makefile 中)
	cmd := exec.Command("make", "php-migrate")
	cmd.Dir = workDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("迁移命令执行失败: %w\n输出: %s", err, string(output))
	}

	return nil
}

// cleanupSettingTestCompany 清理测试商户和数据库
func cleanupSettingTestCompany(companyUuid uint64) {
	fmt.Printf("\n%s 正在清理测试数据... %s\n", blueColor, resetColor)

	// 删除数据库
	dbName := fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port)

	db, err := sql.Open("mysql", dsn)
	if err == nil {
		_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		db.Close()
		fmt.Printf("%s  ✓ 测试数据库已删除 %s\n", greenColor, resetColor)
	}

	// 删除 saas 库中的商户记录
	saasDB, err := database.NewMySQLConnection(config.Database, "saas")
	if err == nil {
		_ = saasDB.Where("uuid = ?", companyUuid).Delete(&model.Company{})
		fmt.Printf("%s  ✓ 测试商户记录已删除 %s\n", greenColor, resetColor)
	}
}
