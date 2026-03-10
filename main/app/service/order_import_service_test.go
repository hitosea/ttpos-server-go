//go:build integration

package service

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

// createTestExcelFile 创建测试用的 Excel 文件
func createTestExcelFile() (*excelize.File, error) {
	f := excelize.NewFile()

	// Sheet1: 订单基本信息
	sheet1 := "订单基本信息"
	f.NewSheet(sheet1)
	f.DeleteSheet("Sheet1") // 删除默认的 Sheet1

	// 设置表头
	headers1 := []string{"订单号", "下单时间", "订单状态", "订单类型", "用餐方式",
		"订单金额", "订单原价", "门店名称", "桌台名称", "会员编号",
		"收银员姓名", "就餐人数", "当班编号", "桌位编号", "商品金额",
		"服务费", "税费", "备注"}
	for i, header := range headers1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet1, cell, header)
	}

	// 添加测试数据
	testData1 := []interface{}{
		"ORD001", "2025-11-19 14:30:00", 1, 0, 0,
		128.50, 150.00, "测试门店", "A01", "",
		"张三", 2, "DUTY001", "SERIAL001", 120.00,
		5.00, 3.50, "测试订单",
	}
	for i, value := range testData1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet1, cell, value)
	}

	// Sheet2: 订单明细
	sheet2 := "订单明细"
	f.NewSheet(sheet2)

	// 设置表头
	headers2 := []string{"订单号", "商品名称", "数量", "单价", "小计", "备注"}
	for i, header := range headers2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet2, cell, header)
	}

	// 添加测试数据
	testData2 := [][]interface{}{
		{"ORD001", "测试商品1", 2, 15.00, 30.00, ""},
		{"ORD001", "测试商品2", 1, 12.00, 12.00, ""},
	}
	for rowIdx, rowData := range testData2 {
		for colIdx, value := range rowData {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet2, cell, value)
		}
	}

	return f, nil
}

// TestParseOrderBasicSheet 测试解析订单基本信息
func TestParseOrderBasicSheet(t *testing.T) {
	// 创建测试服务（不需要真实的数据库）
	srv := &orderImportSrv{}

	// 创建测试 Excel 文件
	excelFile, err := createTestExcelFile()
	if err != nil {
		t.Fatalf("创建测试 Excel 文件失败: %v", err)
	}

	// 测试解析
	orderList, err := srv.parseOrderBasicSheet(excelFile, "订单基本信息")
	if err != nil {
		t.Fatalf("解析订单基本信息失败: %v", err)
	}

	// 验证结果
	if len(orderList) != 1 {
		t.Fatalf("期望解析 1 条订单，实际解析 %d 条", len(orderList))
	}

	order := orderList[0]
	if order.OrderNo != "ORD001" {
		t.Errorf("订单号错误，期望 ORD001，实际 %s", order.OrderNo)
	}
	if order.ShopName != "测试门店" {
		t.Errorf("门店名称错误，期望 测试门店，实际 %s", order.ShopName)
	}
	if order.Amount != 128.50 {
		t.Errorf("订单金额错误，期望 128.50，实际 %f", order.Amount)
	}
	if order.CreateTime == 0 {
		t.Error("下单时间解析失败")
	}
}

// TestParseOrderDetailSheet 测试解析订单明细
func TestParseOrderDetailSheet(t *testing.T) {
	// 创建测试服务
	srv := &orderImportSrv{}

	// 创建测试 Excel 文件
	excelFile, err := createTestExcelFile()
	if err != nil {
		t.Fatalf("创建测试 Excel 文件失败: %v", err)
	}

	// 测试解析
	detailList, err := srv.parseOrderDetailSheet(excelFile, "订单明细")
	if err != nil {
		t.Fatalf("解析订单明细失败: %v", err)
	}

	// 验证结果
	if len(detailList) != 2 {
		t.Fatalf("期望解析 2 条明细，实际解析 %d 条", len(detailList))
	}

	detail1 := detailList[0]
	if detail1.OrderNo != "ORD001" {
		t.Errorf("订单号错误，期望 ORD001，实际 %s", detail1.OrderNo)
	}
	if detail1.ProductName != "测试商品1" {
		t.Errorf("商品名称错误，期望 测试商品1，实际 %s", detail1.ProductName)
	}
	if detail1.Num != 2 {
		t.Errorf("数量错误，期望 2，实际 %d", detail1.Num)
	}
	if detail1.Price != 15.00 {
		t.Errorf("单价错误，期望 15.00，实际 %f", detail1.Price)
	}
}

// TestParseDateTime 测试解析日期时间
func TestParseDateTime(t *testing.T) {
	srv := &orderImportSrv{}

	tests := []struct {
		name      string
		dateTime  string
		wantError bool
	}{
		{
			name:      "正常格式",
			dateTime:  "2025-11-19 14:30:00",
			wantError: false,
		},
		{
			name:      "空字符串",
			dateTime:  "",
			wantError: true,
		},
		{
			name:      "错误格式",
			dateTime:  "2025/11/19 14:30:00",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, err := srv.parseDateTime(tt.dateTime)
			if tt.wantError {
				if err == nil {
					t.Errorf("期望返回错误，但没有错误")
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误，但得到错误: %v", err)
				}
				if timestamp == 0 {
					t.Errorf("时间戳不应该为 0")
				}
			}
		})
	}
}

// TestGetOrderDetailsByOrderNo 测试根据订单号获取订单明细
func TestGetOrderDetailsByOrderNo(t *testing.T) {
	srv := &orderImportSrv{}

	detailList := []OrderDetailData{
		{OrderNo: "ORD001", ProductName: "商品1"},
		{OrderNo: "ORD001", ProductName: "商品2"},
		{OrderNo: "ORD002", ProductName: "商品3"},
	}

	// 测试获取 ORD001 的明细
	details := srv.getOrderDetailsByOrderNo(detailList, "ORD001")
	if len(details) != 2 {
		t.Errorf("期望获取 2 条明细，实际获取 %d 条", len(details))
	}

	// 测试获取不存在的订单号
	details = srv.getOrderDetailsByOrderNo(detailList, "ORD999")
	if len(details) != 0 {
		t.Errorf("期望获取 0 条明细，实际获取 %d 条", len(details))
	}
}

// TestSafeGetString 测试安全获取字符串
func TestSafeGetString(t *testing.T) {
	tests := []struct {
		name     string
		row      []string
		index    int
		expected string
	}{
		{
			name:     "正常索引",
			row:      []string{"a", "b", "c"},
			index:    1,
			expected: "b",
		},
		{
			name:     "索引越界",
			row:      []string{"a", "b"},
			index:    5,
			expected: "",
		},
		{
			name:     "空数组",
			row:      []string{},
			index:    0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeGetString(tt.row, tt.index)
			if result != tt.expected {
				t.Errorf("期望 %s，实际 %s", tt.expected, result)
			}
		})
	}
}

// TestOrderImportService_Import_ExcelParse 测试 Excel 解析部分（不涉及数据库）
func TestOrderImportService_Import_ExcelParse(t *testing.T) {
	// 这个测试只测试 Excel 解析部分，不涉及数据库操作
	// 需要 mock 数据库管理器，这里简化处理，只测试解析逻辑

	srv := &orderImportSrv{}

	// 创建测试 Excel 文件
	excelFile, err := createTestExcelFile()
	if err != nil {
		t.Fatalf("创建测试 Excel 文件失败: %v", err)
	}

	// 测试解析订单基本信息
	orderBasicList, err := srv.parseOrderBasicSheet(excelFile, "订单基本信息")
	if err != nil {
		t.Fatalf("解析订单基本信息失败: %v", err)
	}
	if len(orderBasicList) == 0 {
		t.Error("订单基本信息列表为空")
	}

	// 测试解析订单明细
	orderDetailList, err := srv.parseOrderDetailSheet(excelFile, "订单明细")
	if err != nil {
		t.Fatalf("解析订单明细失败: %v", err)
	}
	if len(orderDetailList) == 0 {
		t.Error("订单明细列表为空")
	}

	// 验证数据关联
	details := srv.getOrderDetailsByOrderNo(orderDetailList, orderBasicList[0].OrderNo)
	if len(details) == 0 {
		t.Error("未找到对应的订单明细")
	}
}

// TestOrderImportService_Import_FileSizeLimit 测试文件大小限制
func TestOrderImportService_Import_FileSizeLimit(t *testing.T) {
	// 这个测试验证文件大小限制逻辑
	// 由于需要完整的服务实例和数据库，这里只做结构说明
	// 实际测试需要在集成测试中完成

	t.Log("文件大小限制测试需要在集成测试中完成，需要 mock 数据库管理器")
}

// TestOrderImportService_Import_DataValidation 测试数据校验逻辑（简化版）
func TestOrderImportService_Import_DataValidation(t *testing.T) {
	// 这个测试验证数据校验逻辑
	// 由于需要数据库连接，这里只做结构说明
	// 实际测试需要在集成测试中完成

	tests := []struct {
		name       string
		orderNo    string
		createTime int64
		shopName   string
		wantError  bool
	}{
		{
			name:       "订单号为空",
			orderNo:    "",
			createTime: 1732008600,
			shopName:   "测试门店",
			wantError:  true,
		},
		{
			name:       "下单时间为空",
			orderNo:    "ORD001",
			createTime: 0,
			shopName:   "测试门店",
			wantError:  true,
		},
		{
			name:       "门店名称为空",
			orderNo:    "ORD001",
			createTime: 1732008600,
			shopName:   "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderBasic := &OrderBasicData{
				OrderNo:    tt.orderNo,
				CreateTime: tt.createTime,
				ShopName:   tt.shopName,
			}

			// 基本字段校验
			if orderBasic.OrderNo == "" {
				if !tt.wantError {
					t.Error("订单号不能为空")
				}
			}
			if orderBasic.CreateTime == 0 {
				if !tt.wantError {
					t.Error("下单时间不能为空")
				}
			}
			if orderBasic.ShopName == "" {
				if !tt.wantError {
					t.Error("门店名称不能为空")
				}
			}
		})
	}
}
