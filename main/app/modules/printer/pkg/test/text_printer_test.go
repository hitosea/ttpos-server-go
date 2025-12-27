package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer/pkg"
)

// TestTextPrinterWithSendHelper 测试文本打印机与发送助手
func TestTextPrinterWithSendHelper(t *testing.T) {
	// 创建复杂的测试模板
	templateJSON, err := os.ReadFile("../template_json/dishes_one_dish_one_order_tmp.json")
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

	// 打印测试数据
	var testData map[string]interface{}
	if err := json.Unmarshal(testDataBytes, &testData); err != nil {
		t.Fatalf("解析测试数据JSON失败: %v", err)
	}

	// 创建文本解析器
	parser, err := pkg.NewTextTemplateParser(pkg.TextBaseData{
		Language:             "zh",
		CurrencyUnit:         "元",
		CurrencyUnitPosition: 1, // 1表示在数字后面
		// DotsPerLine:          384,
		// AppID:                "test_app",
		// AppKey:               "test_key",
	}, templateJSONStr, testData)
	if err != nil {
		t.Fatalf("创建文本解析器失败: %v", err)
	}

	// 解析模板生成打印内容
	printContent, err := parser.Parse(false)
	if err != nil {
		t.Fatalf("解析模板失败: %v", err)
	}

	// 打印内容到控制台（用于调试）
	fmt.Printf("生成的打印内容:\n%s\n", printContent)
	fmt.Printf("打印内容长度: %d 字符\n", len(printContent))

	// 测试发送到打印机（需要配置打印机信息）
	t.Run("SendToPrinter", func(t *testing.T) {
		// 注意：这里需要真实的打印机配置才能测试
		// 如果没有配置打印机，可以跳过这个测试
		printerConfig := model.PrinterConfigJson{
			IP:      "192.168.100.235",                  // 替换为实际的打印机IP
			SN:      "N439254810352",                    // 替换为实际的打印机SN
			APP_ID:  "d0a273417b0f415895ef1adc1831fa14", // 替换为实际的APP_ID
			APP_KEY: "58d19a6e080a400daae071dba3779629", // 替换为实际的APP_KEY
		}

		// 测试商米云打印
		// err := pkg.PrintSunmiTicket(printerConfig, printContent)
		// if err != nil {
		// 	fmt.Printf("商米云打印测试失败（这是正常的，因为没有真实配置）: %v \n", err)
		// } else {
		// 	fmt.Printf("商米云打印测试成功 \n")
		// }

		// 测试局域网打印
		err = pkg.PrintTicket(printerConfig.IP, "9100", printContent, constant.Yes) // 0表示文本打印
		if err != nil {
			fmt.Printf("局域网打印测试失败（这是正常的，因为没有真实配置）: %v \n", err)
		} else {
			fmt.Printf("局域网打印测试成功 \n")
		}

		// 如果没有配置打印机，可以跳过这个测试
		// printerConfig = model.PrinterConfigJson{
		// 	IP:      "192.168.100.241",                  // 替换为实际的打印机IP
		// 	SN:      "N439254810352",                    // 替换为实际的打印机SN
		// 	APP_ID:  "d0a273417b0f415895ef1adc1831fa14", // 替换为实际的APP_ID
		// 	APP_KEY: "58d19a6e080a400daae071dba3779629", // 替换为实际的APP_KEY
		// }
		// // 测试局域网打印
		// err = pkg.PrintTicket(printerConfig.IP, "9100", printContent, constant.Yes) // 0表示文本打印
		// if err != nil {
		// 	fmt.Printf("局域网打印测试失败（这是正常的，因为没有真实配置）: %v \n", err)
		// } else {
		// 	fmt.Printf("局域网打印测试成功 \n")
		// }
	})

	// 不退出
	time.Sleep(1000000 * time.Minute)
}
