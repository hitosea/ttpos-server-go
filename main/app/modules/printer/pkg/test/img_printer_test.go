package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"ttpos-server-go/app/modules/printer/pkg"
)

// TestImgTemplateParser verifies that the image template parser renders without error
// using local fixture files. No network access required.
func TestImgTemplateParser(t *testing.T) {
	templateJSON, err := os.ReadFile("../template/takeout_customer_receipt_tmp.json")
	if err != nil {
		t.Skipf("fixture file not found, skipping: %v", err)
	}

	testDataBytes, err := os.ReadFile("../template/takeout_merchant_receipt_data.json")
	if err != nil {
		t.Skipf("fixture file not found, skipping: %v", err)
	}

	var testData map[string]any
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		t.Fatalf("解析测试数据JSON失败: %v", err)
	}

	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             "zh",
		CurrencyUnit:         "$",
		CurrencyUnitPosition: 1,
	}, string(templateJSON), testData)
	if err != nil {
		t.Fatalf("创建图片模板解析器失败: %v", err)
	}

	if err := parser.ValidateTemplate(); err != nil {
		t.Fatalf("模板验证失败: %v", err)
	}

	img, err := parser.Parse()
	if err != nil {
		t.Fatalf("解析模板失败: %v", err)
	}

	fmt.Printf("图片模板解析成功, 内容大小: %v\n", img)
}
