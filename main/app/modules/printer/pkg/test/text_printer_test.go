package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"ttpos-server-go/app/modules/printer/pkg"
)

// TestTextTemplateParser verifies that the text template parser produces non-empty output
// from the local fixture files. No network access required.
func TestTextTemplateParser(t *testing.T) {
	templateJSON, err := os.ReadFile("../template_json/dishes_one_dish_one_order_tmp.json")
	if err != nil {
		t.Skipf("fixture file not found, skipping: %v", err)
	}

	testDataBytes, err := os.ReadFile("../template_json/dishes_data.json")
	if err != nil {
		t.Skipf("fixture file not found, skipping: %v", err)
	}

	var testData map[string]any
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		t.Fatalf("解析测试数据JSON失败: %v", err)
	}

	parser, err := pkg.NewTextTemplateParser(pkg.TextBaseData{
		Language:             "zh",
		CurrencyUnit:         "元",
		CurrencyUnitPosition: 1,
	}, string(templateJSON), testData)
	if err != nil {
		t.Fatalf("创建文本解析器失败: %v", err)
	}

	printContent, err := parser.Parse(false)
	if err != nil {
		t.Fatalf("解析模板失败: %v", err)
	}

	if len(printContent) == 0 {
		t.Fatal("expected non-empty print content")
	}

	fmt.Printf("生成的打印内容长度: %d 字符\n", len(printContent))
}
