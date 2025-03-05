package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

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
