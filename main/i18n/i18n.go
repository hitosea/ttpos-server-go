package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var dict = make(map[string]map[string]string)

const AcceptLanguage = "Accept-Language"
const (
	LanguageZH   = "zh"   // 简体中文
	LanguageZHTW = "zhtw" // 繁体中文
	LanguageEN   = "en"   // 英文
	LanguageJA   = "ja"   // 日文
	LanguageKO   = "ko"   // 韩文
	LanguageMY   = "my"   // 缅甸语
	LanguageTH   = "th"   // 泰语
	LanguageTR   = "tr"   // 土耳其语
	LanguageDE   = "de"   // 德语
	LanguageSV   = "sv"   // 瑞典语
)

// GetLanguageList 获取语言列表
func GetLanguageList() []string {
	return []string{LanguageZH, LanguageZHTW, LanguageEN, LanguageJA, LanguageKO, LanguageMY, LanguageTH, LanguageTR, LanguageDE, LanguageSV}
}

// GetAcceptLanguage 获取语言
func GetAcceptLanguage(c *gin.Context) string {
	language := c.GetHeader(AcceptLanguage)
	if !slices.Contains([]string{LanguageZH, LanguageZHTW, LanguageEN, LanguageJA, LanguageKO, LanguageMY, LanguageTH, LanguageTR, LanguageDE, LanguageSV}, language) {
		language = LanguageEN
	}
	return language
}

func Init() {
	var (
		dir     = "./i18n/languages"
		entries []os.DirEntry
		err     error
	)
	// 读取目录下的文件
	if entries, err = os.ReadDir(dir); err != nil {
		log.Fatal("读取语言包文件失败：", err.Error())
	}
	// 遍历文件
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".json") {
			continue
		}
		parts := strings.SplitN(filename, ".", 2)
		language := parts[0]
		if !slices.Contains([]string{LanguageZH, LanguageZHTW, LanguageEN, LanguageJA, LanguageKO, LanguageMY, LanguageTH, LanguageTR, LanguageDE, LanguageSV}, language) {
			continue
		}
		var content []byte
		if content, err = os.ReadFile(dir + "/" + filename); err != nil {
			log.Println("读取语言包文件失败2：", err.Error())
			continue
		}
		m := make(map[string]string)
		if err = json.Unmarshal(content, &m); err != nil {
			log.Println("读取语言包文件失败3：", err.Error())
			continue
		}
		dict[language] = m
	}
}

func Exists(language, key string) bool {
	_, exists := dict[language][key]
	return exists
}

func toAnySlice(strSlice []string) []any {
	anySlice := make([]any, len(strSlice))
	for i, v := range strSlice {
		anySlice[i] = v
	}
	return anySlice
}

// Translate 翻译，限定替换参数类型为字符串（如果有）
func Translate(language, key string, replace ...string) string {
	if msg, exists := dict[language][key]; exists {
		if len(replace) == 0 {
			return msg
		}
		args := toAnySlice(replace)
		return fmt.Sprintf(msg, args...)
	} else {
		return key
	}
}
