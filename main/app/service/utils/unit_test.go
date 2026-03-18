//go:build integration

package utils

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestProductMustPlanCheck_CheckResult(t *testing.T) {
	tests := []struct {
		name   string
		rule   Rule
		check  Check
		expect ProductMustPlanCheckResult
	}{
		{
			name: "Test case 1: All required products are selected",
			rule: Rule{
				EachPersonProductPlan: map[uint64]map[uint64]uint{
					1: {101: 2, 102: 3},
				},
				EachOrderProductPlan: map[uint64]uint{
					2: 5,
				},
			},
			check: Check{
				PerProduct: map[uint64]map[uint64]uint{
					1: {101: 2, 102: 3},
				},
				CombinationProduct: map[uint64]uint{
					2: 5,
				},
			},
			expect: ProductMustPlanCheckResult{
				PerProduct:         map[uint64]map[uint64]uint{},
				CombinationProduct: map[uint64]uint{},
			},
		},
		{
			name: "Test case 2: Some required products are not selected",
			rule: Rule{
				EachPersonProductPlan: map[uint64]map[uint64]uint{
					1: {101: 2, 102: 3},
				},
				EachOrderProductPlan: map[uint64]uint{
					2: 5,
				},
			},
			check: Check{
				PerProduct: map[uint64]map[uint64]uint{
					1: {101: 1, 102: 2},
				},
				CombinationProduct: map[uint64]uint{
					2: 3,
				},
			},
			expect: ProductMustPlanCheckResult{
				PerProduct: map[uint64]map[uint64]uint{
					1: {101: 1, 102: 1},
				},
				CombinationProduct: map[uint64]uint{
					2: 2,
				},
			},
		},
		{
			name: "Test case 3: No products are selected",
			rule: Rule{
				EachPersonProductPlan: map[uint64]map[uint64]uint{
					1: {101: 2, 102: 3},
				},
				EachOrderProductPlan: map[uint64]uint{
					2: 5,
				},
			},
			check: Check{
				PerProduct:         map[uint64]map[uint64]uint{},
				CombinationProduct: map[uint64]uint{},
			},
			expect: ProductMustPlanCheckResult{
				PerProduct: map[uint64]map[uint64]uint{
					1: {101: 2, 102: 3},
				},
				CombinationProduct: map[uint64]uint{
					2: 5,
				},
			},
		},
		{
			name: "Test case 4: Extra products are selected",
			rule: Rule{
				EachPersonProductPlan: map[uint64]map[uint64]uint{
					1: {101: 2, 102: 3},
				},
				EachOrderProductPlan: map[uint64]uint{
					2: 5,
				},
			},
			check: Check{
				PerProduct: map[uint64]map[uint64]uint{
					1: {101: 3, 102: 4},
				},
				CombinationProduct: map[uint64]uint{
					2: 6,
				},
			},
			expect: ProductMustPlanCheckResult{
				PerProduct:         map[uint64]map[uint64]uint{},
				CombinationProduct: map[uint64]uint{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := ProductMustPlanCheck{
				Rule:  tt.rule,
				Check: tt.check,
			}
			result := checker.CheckResult()

			db := NewGormDB()
			tips, err := Tips(db, result)
			if err != nil {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
			jsonData, _ := json.MarshalIndent(tips, "", "    ")
			fmt.Println(fmt.Sprintf("tips: %s", string(jsonData)))
			if !reflect.DeepEqual(result, &tt.expect) {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func NewGormDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/ttpos?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

// NewSaleBillSetting1 不收取服务费、不收取消费税
func NewSaleBillSetting1() model.SaleBillSetting {
	setting := model.SaleBillSetting{}
	setting.ServiceFeeType = constant.SaleBillSettingServiceFeeTypeNone // 不收取服务费
	setting.TaxFeeType = constant.SaleBillSettingTaxFeeTypeNone         // 关闭消费税
	return setting
}

// NewSaleBillSetting2 按比例收取服务费、商品未含税
func NewSaleBillSetting2(serviceFeeValue float64) model.SaleBillSetting {
	setting := NewSaleBillSetting1()
	setting.ServiceFeeType = constant.SaleBillSettingServiceFeeTypePercentTax // 按比例收取服务费
	setting.ServiceFeeValue = serviceFeeValue
	setting.TaxFeeType = constant.SaleBillSettingTaxFeeTypePercent // 商品未含税
	setting.ServiceApply = 1                                       // 收取服务费
	return setting
}

// 测试商品计算
func TestProductCalc(t *testing.T) {
	setting := NewSaleBillSetting2(0.07)
	NewProduct := func(
		saucePrice []float64, // 小料价格
		flavorPrice float64, // 规格价格
		taxRate float64, // 商品税率
		memberDiscountRate float64, // 会员折扣率
		memberCardDiscountRate float64, // 会员卡折扣率
		customDiscountRate float64, // 自定义折扣率
	) *model.SaleOrderProduct {
		saleOrderProduct := &model.SaleOrderProduct{
			TaxRate:                taxRate,
			FlavorPrice:            flavorPrice,
			MemberDiscountRate:     memberDiscountRate,
			MemberCardDiscountRate: memberCardDiscountRate,
			CustomDiscountRate:     customDiscountRate,
			OpenMemberDiscount:     constant.ProductMemberDiscountOn,
			SaleOrderProductBoms: []*model.SaleOrderProductBom{
				{
					IsFlavorBom: 1,
					Price:       flavorPrice, // 商品规格价格
				},
			},
		}
		for _, price := range saucePrice {
			saleOrderProduct.SaleOrderProductBoms = append(saleOrderProduct.SaleOrderProductBoms, &model.SaleOrderProductBom{
				IsFlavorBom: 0,
				Price:       price,
			})
		}
		return saleOrderProduct
	}

	tests := []struct {
		name    string
		product *model.SaleOrderProduct
		setting model.SaleBillSetting
		expect  model.SaleOrderProductCalc
	}{
		{
			name:    "test 1",
			product: NewProduct([]float64{2.3, 3, 10} /*小料价格*/, 10.27 /*规格价格*/, 0.10 /*商品税率*/, 0.88 /*会员折扣率*/, 0.85 /*会员卡折扣率*/, 0.85 /*自定义折扣率*/),
			setting: setting,
			expect: model.SaleOrderProductCalc{
				SaucePrice:        15.3,
				ProductPrice:      25.57,
				SalePrice:         25.57,
				SalePriceNoTax:    25.57,
				Price:             16.26,
				MemberDiscountFee: 6.44,
				CustomDiscountFee: 2.87,
				DiscountFee:       9.31,
				TaxFee:            1.63,
				ServiceFee:        1.14,
				ServiceTaxFee:     0.11,
				TotalPrice:        19.14,
				OriginTotalPrice:  30.1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.product.CalcSaleOrderProduct(test.setting)
			if !reflect.DeepEqual(result, test.expect) {
				t.Errorf(`
				expected %+v, 
				got      %+v`, test.expect, result)
			}
		})
	}
}

type LogEntry struct {
	Level   string `json:"level"`
	Time    string `json:"time"`
	Caller  string `json:"caller"`
	Message string `json:"msg"`
}

func (l *LogEntry) GetDuration() (float64, string) {
	re := regexp.MustCompile(`\[(\d+\.\d+)(ms|s)\]`)
	matches := re.FindStringSubmatch(l.Message)
	if len(matches) < 2 {
		return 0, "error"
	}
	durationStr := matches[1]
	duration, _ := strconv.ParseFloat(durationStr, 64)

	unit := "s"
	if strings.Contains(matches[0], "ms") {
		unit = "ms"
	}
	return duration, unit
}

// 获取指定目录下所有以.sql.log结尾的文件
func GetFiles(dirPath string) ([]string, error) {
	var sqlLogFiles []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql.log") {
			sqlLogFiles = append(sqlLogFiles, path)
		}
		return nil
	})
	return sqlLogFiles, err
}

func TestLogEntryGetDuration(t *testing.T) {
	files, err2 := GetFiles("/Users/xiezhihuan/dev/ttpos-server-go/main/log")
	if err2 != nil {
		panic(err2)
	}
	fmt.Println("files:", files)
	list, err := ParseMulFile(files)
	if err != nil {
		panic(err)
	}

	distribution := make(map[string]int)
	for _, entry := range list {
		key := fmt.Sprintf("%.1f%s", entry.Duration, entry.Unit)
		distribution[key]++
	}

	// 将分布情况按数量由高到低排序
	type kv struct {
		Key   string
		Value int
	}

	var sortedDistribution []kv
	for k, v := range distribution {
		sortedDistribution = append(sortedDistribution, kv{k, v})
	}

	sort.Slice(sortedDistribution, func(i, j int) bool {
		return sortedDistribution[i].Value > sortedDistribution[j].Value
	})

	for _, kv := range sortedDistribution {
		fmt.Println("分布情况:", kv.Key, kv.Value)
	}
}

func ParseMulFile(paths []string) ([]Duration, error) {
	list := make([]Duration, 0)
	for _, path := range paths {
		durations, err := Parse(path)
		if err != nil {
			return nil, err
		}
		list = append(list, durations...)
	}
	return list, nil
}

type Duration struct {
	Duration float64
	Unit     string
}

func Parse(path string) ([]Duration, error) {
	list := make([]Duration, 0)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		logEntry := parseLogEntry(line)
		duration, unit := logEntry.GetDuration()
		list = append(list, Duration{Duration: duration, Unit: unit})
		//fmt.Println("duration:", duration, "unit:", unit)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func parseLogEntry(lineJson string) LogEntry {
	var logEntry LogEntry
	err := json.Unmarshal([]byte(lineJson), &logEntry)
	if err != nil {
		fmt.Println(lineJson)
	}
	return logEntry
}

func TestInsert(t *testing.T) {
	Insert()
}

func Insert() {
	InitializeSonyFlakeId()
	// 数据库连接信息
	dsn := "root:cfeb18fa768c2d5f@tcp(127.0.0.1:13306)/shop4477708931072000?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 插入语句
	insertStmt := `
	INSERT INTO ttpos_product_package (
		 uuid, name, multi_language_name_uuid, image_name, image_file_uuid, 
		deduct_stock_type, unit_uuid, dine_tax_uuid, category_uuid, special_category_uuid, 
		takeout_tax_uuid, printer_tag_uuid, supplier_uuid, status, is_show_cashier, 
		is_show_tablet, is_show_kitchen, is_show_assistant, is_show_h5, sort, limit_num, 
		sauce_required, sauce_max_selection, %describe%, open_discount, actual_sale_num, 
		create_time, update_time, delete_time
	) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	insertStmt = strings.ReplaceAll(insertStmt, "%", "`")

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// 准备插入
	stmt, err := tx.Prepare(insertStmt)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// 插入数据
	for i := 0; i < 1000*10*10*10; i++ {
		uuid, _ := utils.GetID()

		_, err := stmt.Exec(
			uuid, "{\"zh\":\"姜汁汽水\",\"en\":\"Ginger Ale\"}", 3661596362539041+i, "", 76+i,
			1, 2, 2, 14, 0, 2, 0, 1724054114+i, 1, 1, 1, 1, 1, 1,
			19+i, 0, 0, 0, "", 1, 140.0000+float64(i), 1739077400, 1745663007, 0,
		)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully inserted 10,000 records.")
}

var SonyFlakeIdOnce sync.Once

func InitializeSonyFlakeId() {
	SonyFlakeIdOnce.Do(func() {
		utils.InitSonyFlakeId()
	})
}
