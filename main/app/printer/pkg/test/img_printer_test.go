package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
	"ttpos-server-go/app/printer/pkg"
)

// TestComplexImgTemplate 测试复杂的JSON模板功能
func TestComplexImgTemplate(t *testing.T) {
	// 创建复杂的测试模板
	templateJSON, err := os.ReadFile("../template_json/dishes_complete_order_tmp.json")
	if err != nil {
		t.Fatalf("读取 tmp.json 文件失败: %v", err)
	}

	// 将 templateJSON 转换为字符串
	templateJSONStr := string(templateJSON)

	// 从JSON文件读取测试数据
	testDataBytes, err := os.ReadFile("../template_json/dishes_data.json")
	if err != nil {
		t.Fatalf("读取测试数据文件失败: %v", err)
	}

	var testData map[string]interface{}
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		t.Fatalf("解析测试数据JSON失败: %v", err)
	}

	// 创建解析器
	parser, err := pkg.NewImgTemplateParser(pkg.ImgBaseData{
		Language:             "zh",
		CurrencyUnit:         "$",
		CurrencyUnitPosition: 1,
	}, templateJSONStr, testData)
	if err != nil {
		t.Fatalf("创建复杂模板解析器失败: %v", err)
	}

	// 验证模板
	err = parser.ValidateTemplate()
	if err != nil {
		t.Fatalf("复杂模板验证失败: %v", err)
	}

	// 解析模板
	img, err := parser.Parse()
	if err != nil {
		t.Fatalf("解析复杂模板失败: %v", err)
	}

	// 保存测试图片
	img.Save("./tmp/printer/complex_template_test.png", false, 0)

	fmt.Println("复杂模板测试完成，图片已保存到: ./tmp/printer/complex_template_test.png")

	img.Cleanup()
	//
	// 不退出
	time.Sleep(1000000 * time.Minute)
}
