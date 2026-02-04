package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/spf13/cobra"
)

var URL = "https://aitrans.ttpos.com/translate"

var newLanguage string
var num int
var forceTrans bool

// skipTexts 不需要翻译的文案列表（日文、繁体中文等非简体中文文案）
var skipTexts = map[string]struct{}{
	"明細表":    {},
	"收款小記":   {},
	"時間帯":    {},
	"訂單金額":   {},
	"名稱":     {},
	"総時間":    {},
	"退款金額":   {},
	"點餐-外賣":  {},
	"桌台訂單金額": {},
	"営業日":    {},
	"總耗時":    {},
	"消費桌數":   {},
	"點餐訂單金額": {},
	"営業入金統計": {},
	"時段營業統計": {},
	"支払方法":   {},
	"菜品出品明細": {},
	"全額返金回数": {},
	"最小注文金額": {},
	"最大訂單金額": {},
	"的可出库数量不足。\\n\\n请联系发货门店": {},
	"注文金額":       {},
	"訂單金額人均":     {},
	"一件当たり注文金額":  {},
	"用戶分析統計":     {},
	"伝菜総時間":      {},
	"整單退款筆數":     {},
	"発送店舗却下":     {},
	"時間帯別売上統計":   {},
	"綜合營運統計":     {},
	"分類":         {},
	"平均客単価":      {},
	"利用テーブル数":    {},
	"實付金額":       {},
	"完成數量":       {},
	"最小訂單金額":     {},
	"料理の提供詳細":    {},
	"實付金額人均":     {},
	"一人あたり":      {},
	"店內點餐":       {},
	"最長出品時長":     {},
	"製作総時間":      {},
	"一人当たり支払い金額": {},
	"門店通過":       {},
	"門店駁回":       {},
	"下單時間":       {},
	"製作總耗時":      {},
	"支付金額":       {},
	"支払い金額":      {},
	"門店匯總統計":     {},
	"全額返金金額":     {},
	"次數":         {},
	"原価販売額":      {},
	"平均出品時長":     {},
	"平均訂單金額":     {},
	"重新發起":       {},
	"伝菜完了時刻":     {},
	"料理の提供明細":    {},
	"外送訂單金額":     {},
	"整單退款金額":     {},
	"最短の製造時間":    {},
	"現金レシート":     {},
	"收貨門店通過":     {},
	"時段":         {},
	"門店名稱":       {},
	"一件当たり支払い金額": {},
	"最短出品時長":     {},
	"商品名稱":       {},
	"配達":         {},
	"店舗承認":       {},
	"實付金額單均":     {},
	"製作完成時間":     {},
	"完了時刻":       {},
	"部分退款金額":     {},
	"所有訂單數":      {},
	"傳菜總耗時":      {},
	"発注時刻":       {},
	"物品 %s 的可出库数量不足。\\n\\n请联系发货门店":      {},
	"物品 %s 的可出库数量不足。\\n\\n请补充库存":        {},
	"物品 %s 未找到。\\n\\n提交后将移除该物品，是否继续提交？": {},
	"物品 %s 的可调拨数量不足。\\n\\n请更换出库仓库":      {},
}

func init() {
	rootCommand.AddCommand(translateCmd)
	translateCmd.Flags().StringVar(&newLanguage, "new-language", "", "新语言文件名")
	translateCmd.Flags().IntVar(&num, "num", 10, "每次关键字翻译数量")
	translateCmd.Flags().BoolVar(&forceTrans, "force-trans", false, "强制翻译")
}

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "开始翻译",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		if newLanguage != "" && !slices.Contains(i18n.GetLanguageList(), newLanguage) {
			log.Fatalf("language-name [%s] is not in language list: %v", newLanguage, i18n.GetLanguageList())
		}
	},
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("-------translate start-------\n")
		//
		execute()
		//
		fmt.Printf("-------translate end-------\n")
	},
}

// 执行命令
func execute() {

	languageList := i18n.GetLanguageList()
	fmt.Println("languageList: ", languageList)

	// 获取中文数据
	zhData := getLangData("zh")

	var chineseTexts []string
	if newLanguage == "" { // 增量更新翻译
		// 指定要扫描的目录
		dir := "./"
		// 提取的中文
		chineseTextMap := make(map[string]struct{})
		// 遍历目录下的所有文件
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// 只处理.go文件
			if !info.IsDir() &&
				strings.HasSuffix(path, ".go") &&
				!strings.HasSuffix(path, "docs.go") &&
				!strings.Contains(path, "trans/") &&
				!strings.Contains(path, "_test.go") &&
				!strings.Contains(path, "old_model") &&
				!strings.Contains(path, "command") &&
				!strings.Contains(path, "request_logger") &&
				!strings.Contains(path, "marketing_activity.go") &&
				!strings.Contains(path, "bucket.go") &&
				!strings.Contains(path, "country.go") &&
				!strings.Contains(path, "model") {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				// 使用正则表达式匹配中文字符和相关内容，包括以%s开头且后面跟汉字的情况
				re := regexp.MustCompile(`('|")((auto\.)?(([a-zA-Z0-9\/-]+)?[\p{Han}]+.*?|%s[\p{Han}]+.*?))('|")`)
				matches := re.FindAllStringSubmatch(string(content), -1)

				// 提取中文
				for _, match := range matches {
					chineseTextMap[match[2]] = struct{}{}
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error scanning directory: %v\n", err)
			return
		}
		// 过滤掉已存在的 key
		var filteredTexts []string
		for text := range chineseTextMap {
			if _, exists := zhData[text]; !exists {
				if strings.Contains(text, "\\n") && !strings.Contains(text, "\\n\\n") {
					continue
				}
				filteredTexts = append(filteredTexts, text)
			}
		}
		// 除了新的文案，还要判断各个语言文件中缺少中文的哪些文案
		missingTexts := checkText(maputil.Keys(zhData))
		// 更新 chineseTexts
		chineseTexts = append(filteredTexts, missingTexts...)
		chineseTexts = slice.Unique(chineseTexts)

		// 删除其他语言中多语的文案
		defer handleDuplicateText()

	} else { // 新添加语言翻译
		chineseTexts = maputil.Keys(zhData)
	}

	// 过滤掉不需要翻译的文案
	chineseTexts = filterSkipTexts(chineseTexts)

	// 打印提取的中文数量
	fmt.Println(utils.ToJsonString(chineseTexts))
	fmt.Println(len(chineseTexts))

	// 分组处理
	chunks := slice.Chunk(chineseTexts, num)
	var count = 0
	for _, chunk := range chunks {
		processGroup(chunk)
		fmt.Printf("process: %.2f%%\n", float64(len(chunk)+count)/float64(len(chineseTexts))*100)
		count += len(chunk)
	}
}

// 处理分组
func processGroup(texts []string) {
	type TextDataItem struct {
		Lang    string `json:"lang"`
		Content string `json:"content"`
	}
	data := make(map[string]any)
	var textData = []TextDataItem{}
	for _, text := range texts {
		textData = append(textData, TextDataItem{
			Lang:    "zh",
			Content: text,
		})
	}
	data["data"] = textData
	if newLanguage != "" {
		data["trans"] = []string{newLanguage}
	}

	if forceTrans {
		data["force_trans"] = true
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
		return
	}

	resp, err := http.Post(URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error requesting ai translate api: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			os.WriteFile("translate_response.log", body, os.ModePerm)
			execute()
		}
	}()

	var result map[string]any
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if result["code"].(float64) == 200 {
		// 直接使用data数组，不需要转换为字符串
		translations := result["data"].([]any)
		for _, lang := range i18n.GetLanguageList() {
			if newLanguage != "" {
				if lang != newLanguage {
					continue
				}
				newLangFile := fmt.Sprintf("./i18n/languages/%s.json", lang)
				if !utils.IsFileExist(newLangFile) {
					utils.CreateFile(newLangFile, []byte("{}"), 0644)
				}
			}
			langData := getLangData(lang)
			// 更新现有键并收集新条目
			newEntries := make(map[string]string)
			for _, trans := range translations {
				transMap := trans.(map[string]any)
				key := transMap["key"].(string)
				langKey := lang
				if lang == "zhtw" {
					langKey = "zh-TW"
				}
				if val, ok := transMap[langKey]; ok {
					if lang == "zh" {
						newEntries[val.(string)] = val.(string)
					} else {
						newEntries[key] = val.(string)
					}
				}
			}
			for k, v := range newEntries {
				if langData[k] == "" {
					langData[k] = v
				}
			}
			// 写入语言文件
			writeContent(lang, langData)
		}
	} else {
		fmt.Printf("API request failed with code: %v\n", result["code"])
	}

	//
	for _, text := range texts {
		fmt.Printf("成功翻译 %s \n", text)
	}
}

// 处理多语的文案
func handleDuplicateText() {
	// 解析中文
	zhData := getLangData("zh")
	writeContent("zh", zhData)
	// 遍历其他语言，删除多语的文案
	for _, lang := range i18n.GetLanguageList() {
		if lang == "zh" {
			continue
		}
		// 解析语言文件
		langData := getLangData(lang)
		for k := range langData {
			if _, ok := zhData[k]; !ok {
				delete(langData, k)
			}
		}
		// 写入语言文件
		writeContent(lang, langData)
	}
}

// 过滤掉不需要翻译的文案
func filterSkipTexts(texts []string) []string {
	var result []string
	for _, text := range texts {
		if _, skip := skipTexts[text]; !skip {
			result = append(result, text)
		}
	}
	return result
}

// 其他语言中缺少中文的文案
func checkText(chineseTexts []string) []string {
	var missingTexts []string
	for _, lang := range i18n.GetLanguageList() {
		langData := getLangData(lang)
		var langDataMap = make(map[string]struct{})
		for k := range langData {
			if _, ok := langDataMap[k]; !ok {
				langDataMap[k] = struct{}{}
			}
		}

		for _, text := range chineseTexts {
			if _, ok := langDataMap[text]; !ok {
				missingTexts = append(missingTexts, text)
			}
		}
	}
	return missingTexts
}

// 解析语言文件，返回map[string]string
func getLangData(lang string) map[string]string {
	filename := fmt.Sprintf("./i18n/languages/%s.json", lang)
	if !utils.IsFileExist(filename) {
		log.Fatalf("file [%s] not exists\n", filename)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading %s.json: %v\n", lang, err)
	}
	var langData map[string]string
	err = json.Unmarshal(content, &langData)
	if err != nil {
		log.Fatalf("Error parsing %s.json: %v\n", lang, err)
	}
	return langData
}

// 写入语言文件
func writeContent(lang string, content map[string]string) {
	// 更新内容
	updatedContent, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling updated %s data: %v\n", lang, err)
	}
	// 写入文件
	err = os.WriteFile(fmt.Sprintf("./i18n/languages/%s.json", lang), updatedContent, 0644)
	if err != nil {
		log.Fatalf("Error writing updated %s.json: %v\n", lang, err)
	}
}
