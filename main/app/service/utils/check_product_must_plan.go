package utils

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

// Rule 桌台的必点商品规则
type Rule struct {
	// 全选必点方案
	// product_must_plan_uuid => product_package_uuid => num  这个全选必点方案1里每个商品要求的商品数量
	//                        => product_package_uuid => num
	//                        => product_package_uuid => num
	// product_must_plan_uuid => product_package_uuid => num  这个全选必点方案2里每个商品要求的商品数量
	//                        => product_package_uuid => num
	//                        => product_package_uuid => num
	EachPersonProductPlan map[uint64]map[uint64]uint
	// 任选必点方案
	// product_must_plan_uuid => num  这个任选方案已经点的商品数量
	EachOrderProductPlan map[uint64]uint
}

// Check 桌台在各方案的选购情况
type Check struct {
	// product_must_plan_uuid => product_package_uuid => num  这个全选必点方案里每个商品已经点的数量
	//                        => product_package_uuid => num
	//                        => product_package_uuid => num
	// product_must_plan_uuid => product_package_uuid => num  这个全选必点方案里每个商品已经点的数量
	//                        => product_package_uuid => num
	//                        => product_package_uuid => num
	// 判断以上哪个全选方案的哪些商品没有点够数量
	PerProduct map[uint64]map[uint64]uint
	// product_must_plan_uuid => num  这个任选方案已经点的商品数量
	// product_must_plan_uuid => num  这个任选方案已经点的商品数量
	// product_must_plan_uuid => num  这个任选方案已经点的商品数量
	// 判断以上哪个任选方案的商品数量没有点够
	CombinationProduct map[uint64]uint
}

type ProductMustPlanCheck struct {
	Rule  Rule  // 桌台的必点商品规则
	Check Check //桌台在各方案的选购情况
}

type ProductMustPlanCheckResult struct {
	PerProduct         map[uint64]map[uint64]uint // 哪些全选方案不通过，并给出不通过的商品及还需选购的数量
	CombinationProduct map[uint64]uint            // 哪些任选方案不通过，并给出不通过的任选方案及还需选购的数量
}

func (result *ProductMustPlanCheckResult) IsPassed() bool {
	return len(result.PerProduct) == 0 && len(result.CombinationProduct) == 0
}

func (result *ProductMustPlanCheckResult) HumanizedTips() string {
	type Tips struct {
		PerProductPlan []struct {
			ProductMustPlanUuid uint64                `json:"product_must_plan_uuid"`
			ProductMustPlan     model.ProductMustPlan `json:"product_must_plan"` // 全选必点方案信息
			ProductList         []struct {
				ProductUuid    uint64               `json:"product_uuid"`
				ProductPackage model.ProductPackage `json:"product_package"`
				Num            uint64               `json:"num"` // 还需选购的数量
			} `json:"product_list"`
		}
		CombinationProductPlan []struct {
			ProductMustPlanUuid uint64 `json:"product_must_plan_uuid"`
			//ProductMustPlan     model.ProductMustPlan `json:"product_must_plan"` // 任选必点方案信息
			RequiredNum uint64 `json:"required_num"`
		}
	}
	return ""
}

func (p *ProductMustPlanCheck) CheckResult() *ProductMustPlanCheckResult {
	result := &ProductMustPlanCheckResult{
		PerProduct:         make(map[uint64]map[uint64]uint),
		CombinationProduct: make(map[uint64]uint),
	}

	// 检查全选必点方案
	for planUuid, productMap := range p.Rule.EachPersonProductPlan {
		for productUuid, requiredNum := range productMap {
			if checkedNum, exists := p.Check.PerProduct[planUuid][productUuid]; !exists || checkedNum < requiredNum {
				if _, exists := result.PerProduct[planUuid]; !exists {
					result.PerProduct[planUuid] = make(map[uint64]uint)
				}
				result.PerProduct[planUuid][productUuid] = requiredNum - checkedNum
			}
		}
	}

	// 检查任选必点方案
	for planUuid, requiredNum := range p.Rule.EachOrderProductPlan {
		if checkedNum, exists := p.Check.CombinationProduct[planUuid]; !exists || checkedNum < requiredNum {
			result.CombinationProduct[planUuid] = requiredNum - checkedNum
		}
	}

	return result
}

type PerProductPlan struct {
	ProductMustPlanUuid uint64                `json:"product_must_plan_uuid"`
	ProductMustPlan     model.ProductMustPlan `json:"product_must_plan"` // 全选必点方案信息
	ProductPackageList  []ProductPackage      `json:"product_list"`
}

type ProductPackage struct {
	ProductPackageUuid uint64               `json:"product_package_uuid"`
	ProductPackage     model.ProductPackage `json:"product_package"`
	Num                uint64               `json:"num"` // 还需选购的数量
}

type CombinationProductPlan struct {
	ProductMustPlanUuid uint64                `json:"product_must_plan_uuid"`
	ProductMustPlan     model.ProductMustPlan `json:"product_must_plan"` // 任选必点方案信息
	RequiredNum         uint64                `json:"required_num"`
}

type ProductMustPlanCheckResultTips struct {
	PerProductPlanList         []PerProductPlan         `json:"per_product_plan_list"`
	CombinationProductPlanList []CombinationProductPlan `json:"combination_product_plan_list"`
}

func (t *ProductMustPlanCheckResultTips) IsPass() bool {
	return len(t.PerProductPlanList) == 0 && len(t.CombinationProductPlanList) == 0
}

func Tips(db *gorm.DB, result *ProductMustPlanCheckResult) (*ProductMustPlanCheckResultTips, error) {
	// 获取涉及到的ProductMustPlan的UUID列表
	var planUuids []uint64
	for planUuid := range result.PerProduct {
		planUuids = append(planUuids, planUuid)
	}
	for planUuid := range result.CombinationProduct {
		planUuids = append(planUuids, planUuid)
	}

	// 使用SQL IN查询获取ProductMustPlan数据
	productMustPlans, err := repository.NewProductMustPlanRepo(db).GetProductMustPlanListByUuids(planUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 将查询结果存入map中
	productMustPlanMap := make(map[uint64]model.ProductMustPlan)
	for _, plan := range productMustPlans {
		productMustPlanMap[plan.Uuid] = plan
	}

	// 获取涉及到的ProductPackage的UUID列表
	var productUuids []uint64
	for _, productMap := range result.PerProduct {
		for productUuid := range productMap {
			productUuids = append(productUuids, productUuid)
		}
	}

	// 使用SQL IN查询获取ProductPackage数据
	productPackages, err := repository.NewProductRepo(db).GetProductPackageListByUuids(productUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 将查询结果存入map中
	productPackageMap := make(map[uint64]model.ProductPackage)
	for _, pkg := range productPackages {
		productPackageMap[pkg.Uuid] = pkg
	}

	tips := &ProductMustPlanCheckResultTips{
		PerProductPlanList:         []PerProductPlan{},
		CombinationProductPlanList: []CombinationProductPlan{},
	}

	for planUuid, productMap := range result.PerProduct {
		perProductPlan := PerProductPlan{
			ProductMustPlanUuid: planUuid,
			ProductPackageList:  []ProductPackage{},
		}
		// 从map中获取ProductMustPlan
		perProductPlan.ProductMustPlan = productMustPlanMap[planUuid]

		for productUuid, num := range productMap {
			productPackage := ProductPackage{
				ProductPackageUuid: productUuid,
				Num:                uint64(num),
			}
			// 从map中获取ProductPackage数据
			productPackage.ProductPackage = productPackageMap[productUuid]

			perProductPlan.ProductPackageList = append(perProductPlan.ProductPackageList, productPackage)
		}
		tips.PerProductPlanList = append(tips.PerProductPlanList, perProductPlan)
	}

	for planUuid, requiredNum := range result.CombinationProduct {
		combinationProductPlan := CombinationProductPlan{
			ProductMustPlanUuid: planUuid,
			RequiredNum:         uint64(requiredNum),
		}
		// 从map中获取ProductMustPlan
		combinationProductPlan.ProductMustPlan = productMustPlanMap[planUuid]

		tips.CombinationProductPlanList = append(tips.CombinationProductPlanList, combinationProductPlan)
	}

	return tips, nil
}
