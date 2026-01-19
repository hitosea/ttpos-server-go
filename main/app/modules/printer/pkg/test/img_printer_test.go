package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
	"ttpos-server-go/app/modules/printer/pkg"
)

// TestComplexImgTemplate 测试复杂的JSON模板功能
func TestComplexImgTemplate(t *testing.T) {
	// 创建复杂的测试模板
	templateJSON, err := os.ReadFile("../template/takeout_customer_receipt_tmp.json")
	if err != nil {
		t.Fatalf("读取 tmp.json 文件失败: %v", err)
	}

	// 将 templateJSON 转换为字符串
	templateJSONStr := string(templateJSON)

	// 从JSON文件读取测试数据
	testDataBytes, err := os.ReadFile("../template/takeout_merchant_receipt_data.json")
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
	img.SegmentationHeight = 200000
	img.Save("./tmp/printer/complex_template_test.png", false, 0)

	fmt.Println("复杂模板测试完成，图片已保存到: ./tmp/printer/complex_template_test.png")

	// 测试发送到打印机（需要配置打印机信息）
	// printContent := img.Save("./tmp/printer/complex_template_test.png", false, 0)
	// t.Run("SendToPrinter", func(t *testing.T) {
	// 	// 注意：这里需要真实的打印机配置才能测试
	// 	// 如果没有配置打印机，可以跳过这个测试
	// 	printerConfig := model.PrinterConfigJson{
	// 		IP:      "192.168.100.235",                  // 替换为实际的打印机IP
	// 		SN:      "N439254810352",                    // 替换为实际的打印机SN
	// 		APP_ID:  "d0a273417b0f415895ef1adc1831fa14", // 替换为实际的APP_ID
	// 		APP_KEY: "58d19a6e080a400daae071dba3779629", // 替换为实际的APP_KEY
	// 	}

	// 	// 测试商米云打印
	// 	// err := pkg.PrintSunmiTicket(printerConfig, printContent)
	// 	// if err != nil {
	// 	// 	fmt.Printf("商米云打印测试失败（这是正常的，因为没有真实配置）: %v \n", err)
	// 	// } else {
	// 	// 	fmt.Printf("商米云打印测试成功 \n")
	// 	// }

	// 	// 测试局域网打印
	// 	err = pkg.PrintTicket(printerConfig.IP, "9100", printContent, constant.No) // 0表示文本打印
	// 	if err != nil {
	// 		fmt.Printf("局域网打印测试失败（这是正常的，因为没有真实配置）: %v \n", err)
	// 	} else {
	// 		fmt.Printf("局域网打印测试成功 \n")
	// 	}
	// })

	// img.Cleanup()
	//
	// 不退出
	time.Sleep(1000000 * time.Minute)
}
