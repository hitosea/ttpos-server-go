package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PrintFormatMigration Print Format 迁移数据结构
type PrintFormatMigration struct {
	Name            string `json:"name"`
	DocType         string `json:"doc_type"`
	Standard        int    `json:"standard"`
	Module          string `json:"module"`
	HTML            string `json:"html"`
	CSS             string `json:"css"`
	PrintFormatType string `json:"print_format_type"`
	PrintFormatFor  string `json:"print_format_for"`
	CustomFormat    int    `json:"custom_format"`
}

func main() {
	// CSV 文件路径
	csvPath := "manifest/printformat/Wallace Print Format.csv"
	// 输出目录
	outputDir := "manifest/erp-migrate/v2.9/02_print_format"

	// 读取 CSV 文件
	file, err := os.Open(csvPath)
	if err != nil {
		fmt.Printf("打开 CSV 文件失败: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("读取 CSV 文件失败: %v\n", err)
		os.Exit(1)
	}

	if len(records) < 2 {
		fmt.Println("CSV 文件格式错误：至少需要表头和数据行")
		os.Exit(1)
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	// 解析数据行（跳过表头）
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 8 {
			fmt.Printf("跳过第 %d 行：字段数量不足\n", i+1)
			continue
		}

		// 解析数据
		name := strings.Trim(record[0], `"`)
		standardStr := strings.Trim(record[1], `"`)
		module := strings.Trim(record[2], `"`)
		docType := strings.Trim(record[3], `"`)
		printFormatFor := strings.Trim(record[4], `"`)
		css := strings.Trim(record[5], `"`)
		html := strings.Trim(record[6], `"`)
		customFormatStr := strings.Trim(record[7], `"`)

		// 转换 Standard 字段
		standard := 0
		if strings.ToLower(standardStr) == "yes" {
			standard = 1
		}

		// 转换 Custom Format 字段
		customFormat := 0
		if customFormatStr == "1" {
			customFormat = 1
		}

		// 确定 Print Format Type
		printFormatType := "Jinja"
		if strings.Contains(html, "{%") || strings.Contains(html, "{{") {
			printFormatType = "Jinja"
		} else {
			printFormatType = "Standard"
		}

		// 构建迁移数据结构
		migration := PrintFormatMigration{
			Name:            name,
			DocType:         docType,
			Standard:        standard,
			Module:          module,
			HTML:            html,
			CSS:             css,
			PrintFormatType: printFormatType,
			PrintFormatFor:  printFormatFor,
			CustomFormat:    customFormat,
		}

		// 生成文件名（将空格替换为下划线，转小写）
		fileName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
		fileName = strings.ReplaceAll(fileName, "-", "_")
		fileName = filepath.Join(outputDir, fileName+".json")

		// 转换为 JSON
		jsonData, err := json.MarshalIndent(migration, "", "  ")
		if err != nil {
			fmt.Printf("转换 JSON 失败 [%s]: %v\n", name, err)
			continue
		}

		// 写入文件
		if err := os.WriteFile(fileName, jsonData, 0644); err != nil {
			fmt.Printf("写入文件失败 [%s]: %v\n", fileName, err)
			continue
		}

		fmt.Printf("✅ 已生成: %s\n", fileName)
	}

	fmt.Println("\n✅ 所有 Print Format 迁移文档生成完成！")
}
