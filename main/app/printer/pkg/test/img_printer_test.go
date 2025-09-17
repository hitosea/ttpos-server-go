package test

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
	"ttpos-server-go/app/printer/pkg"
)

// TestComplexImgTemplate 测试复杂的JSON模板功能
func TestComplexImgTemplate(t *testing.T) {
	// 创建复杂的测试模板
	templateJSON, err := ioutil.ReadFile("../template_json/statement_order.json")
	if err != nil {
		t.Fatalf("读取 tmp.json 文件失败: %v", err)
	}

	// 将 templateJSON 转换为字符串
	templateJSONStr := string(templateJSON)

	// 创建复杂测试数据
	testData := map[string]interface{}{
		"brand_name": "TTPOS",
		"store": map[string]interface{}{
			"name":               "重庆高老九火锅店",
			"address":            "北京市朝阳区建国路88号",
			"phone":              "010-12345678",
			"logo":               "../../../../tmp/printer/shop/logo/79d5d080453715c8d1b0fa73801ecca0.png",
			"company":            "重庆高老九火锅店",
			"company_addr":       "北京市朝阳区建国路88号",
			"company_phone":      "010-12345678",
			"company_tax_number": "1234567890",
			"cashier_sn":         "231231123123123122154",
			"printer_sn":         "12345667890",
		},
		"order": map[string]interface{}{
			"serial_no":    "取单号: A08 (4人)",
			"order_no":     "202501559871231",
			"remark":       "这是桌台的备注这是桌台的备注这是桌台的备注这是桌台的备注这是桌台的备注这是桌台的备注",
			"cashier_name": "张三",
			"time":         "2024-12-20 15:30:25",
			"buffets": []map[string]interface{}{
				{"name": "我是自助餐1", "text": "2124*121", "subtotal": "1112312.00", "info": "大人"},
				{"name": "我是自助餐2", "text": "24*12", "subtotal": 88.00, "info": "小孩"},
			},
			"delays": []map[string]interface{}{
				{"name": "我是加钟1", "text": "24*12", "subtotal": 138.00},
			},
			"products": []map[string]interface{}{
				{"name": "北京烤鸭", "text": "24*12", "subtotal": 138.00, "info": "份"},
				{"name": "清蒸鲈鱼", "text": "24*12", "subtotal": 88.00, "info": "个"},
				{
					"name":           "麻婆豆腐",
					"text":           "24*12",
					"subtotal":       72.00,
					"info":           "碗",
					"is_gift":        1,
					"is_package":     1,
					"is_sub_product": 0,
				},
				{
					"name":           "麻婆豆腐",
					"text":           "24*12",
					"subtotal":       72.00,
					"info":           "碗",
					"is_gift":        1,
					"is_package":     1,
					"is_sub_product": 1,
				},
			},
			"product_num":               10,
			"product_amount":            1000.00,
			"service_fee":               100.00,
			"tax_rate":                  7,
			"tax_fee":                   100.00,
			"is_contain_tax":            0,
			"discount_fee":              30.00,
			"discount_rate":             1,
			"member_discount_fee":       100.00,
			"member_discount_rate":      10,
			"member_card_discount_rate": 10,
			"member_points_discount":    10,
			"coupon_exchange_amount":    100.00,
			"check_out_zero_fee":        10.00,
			"return_amount":             17.00,
			"payment_commission_fee":    10.00,
			"free_amount":               120.00,
			"actual_receive_price":      1000.00,
			"payment_methods": []map[string]interface{}{
				{"name": "支付方式", "text": "微信支付"},
				{"name": "实际金额", "text": "100.00$"},
				{"name": "支付方式", "text": "现金支付"},
				{"name": "实际金额", "text": "100.00$"},
				{"name": "找零", "text": "100.00$"},
			},
			"percentage_lists": []map[string]interface{}{
				{"name": "VAT (7%)", "text": "100.00 (9.09)"},
				{"name": "VAT (7%)", "text": "100.00 (9.09)"},
			},
			"is_free":                  1,
			"member_remaining_balance": 100.00,
			"member_points":            100.00,
			"payment_name":             "微信",
			"payment_qrcode":           "http://nginx/tmp/printer/shop/logo/79d5d080453715c8d1b0fa73801ecca0.png",
		},
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
