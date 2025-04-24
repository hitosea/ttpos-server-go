package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

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
