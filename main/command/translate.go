package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"

	"github.com/spf13/cobra"
)

var URL = "http://103.63.139.229:8088/api/translate"

func init() {
	rootCommand.AddCommand(translateCmd)
}

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "开始翻译",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
	},
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("-------translate start-------\n")
		//
		// websocket.PushClient(1, "*", "*", websocket.UPDATE_PRODUCT, "")
		execute()
		//
		fmt.Printf("-------translate end-------\n")
	},
}

// 执行命令
func execute() {

	languageList := i18n.GetLanguageList()
	fmt.Println(languageList)

	// 指定要扫描的目录
	dir := "./"

	// 创建一个map来存储提取的中文
	chineseTexts := make(map[string]string)

	// 遍历目录下的所有文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.go文件
		if !info.IsDir() &&
			strings.HasSuffix(path, ".go") &&
			!strings.HasSuffix(path, "docs.go") &&
			!strings.Contains(path, "trans") &&
			!strings.Contains(path, "old_model") &&
			!strings.Contains(path, "command") &&
			!strings.Contains(path, "model") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// 使用正则表达式匹配中文字符和相关内容
			re := regexp.MustCompile(`('|")((auto\.)?(([a-zA-Z0-9\/-]+)?[\p{Han}]+.*?))('|")`)
			matches := re.FindAllStringSubmatch(string(content), -1)

			// 将匹配到的中文添加到map中
			for _, match := range matches {
				chineseTexts[match[2]] = ""
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		return
	}

	// 读取 zh.json 文件
	zhContent, err := ioutil.ReadFile("./i18n/languages/zh.json")
	if err != nil {
		fmt.Printf("Error reading zh.json: %v\n", err)
		return
	}

	// 解析 zh.json 内容
	var zhData map[string]string
	err = json.Unmarshal(zhContent, &zhData)
	if err != nil {
		fmt.Printf("Error parsing zh.json: %v\n", err)
		return
	}

	// 过滤掉已存在的 key
	filteredTexts := make(map[string]string)
	for text := range chineseTexts {
		if _, exists := zhData[text]; !exists {
			filteredTexts[text] = ""
		}
	}

	// 更新 chineseTexts
	chineseTexts = filteredTexts

	// 打印提取的中文数量
	fmt.Println(len(chineseTexts))

	// 分组处理
	groupSize := 5
	textGroup := make([]string, 0, groupSize)
	for text := range chineseTexts {
		textGroup = append(textGroup, text)
		if len(textGroup) == groupSize {
			processGroup(textGroup)
			textGroup = textGroup[:0]
		}
	}
	if len(textGroup) > 0 {
		processGroup(textGroup)
	}
}

func processGroup(texts []string) {
	data := map[string][]map[string]string{"data": {}}
	for _, text := range texts {
		data["data"] = append(data["data"], map[string]string{
			"lang":    "zh",
			"content": text,
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
		return
	}

	resp, err := http.Post(URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error requesting Baidu API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if result["code"].(float64) == 200 {
		data := result["data"].(string)
		var translations []map[string]string
		err = json.Unmarshal([]byte(data), &translations)
		if err != nil {
			fmt.Printf("Error parsing translations: %v\n", err)
			return
		}

		for _, lang := range i18n.GetLanguageList() {
			filename := fmt.Sprintf("./i18n/languages/%s.json", lang)
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Printf("Error reading %s.json: %v\n", lang, err)
				continue
			}

			var langData map[string]string
			err = json.Unmarshal(content, &langData)
			if err != nil {
				fmt.Printf("Error parsing %s.json: %v\n", lang, err)
				continue
			}

			// UpdateBalance existing keys and collect new entries
			newEntries := make(map[string]string)
			for _, trans := range translations {
				key := trans["key"]
				langKey := lang
				if lang == "zhtw" {
					langKey = "zh-TW"
				}
				if val, ok := trans[langKey]; ok {
					if lang == "zh" {
						newEntries[val] = val
					} else {
						newEntries[key] = val
					}
				}
			}

			for k, v := range newEntries {
				langData[k] = v
			}

			updatedContent, err := json.MarshalIndent(langData, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling updated %s data: %v\n", lang, err)
				continue
			}

			err = ioutil.WriteFile(filename, updatedContent, 0644)
			if err != nil {
				fmt.Printf("Error writing updated %s.json: %v\n", lang, err)
				continue
			}
		}
	} else {
		fmt.Printf("API request failed with code: %v\n", result["code"])
	}

	//
	for _, text := range texts {
		fmt.Printf("成功翻译 %s \n", text)
	}
}
