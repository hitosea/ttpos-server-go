package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IMustPlanSrv 定义必点方案服务接口
type IMustPlanSrv interface {
	GetProductMustPlans(ctx context.Context, deskUuid uint64) ([]*model.ProductMustPlan, error)                                                                                                          // 获取桌台的必选方案model列表
	GetDeskMustPlanProductPackageList(ctx context.Context, deskUuid uint64) ([]*model.ProductPackage, error)                                                                                             // 获取桌台的必点ProductPackage列表
	GetInstantMustPlanList(ctx context.Context, db *gorm.DB, shopCartMustProductInfo ro.MustPlanProductInfo, options ...func(option *CheckOption)) ([]resp.InstantProductMustPlan, error)                // 获取点餐的必点方案列表
	GetDeskMustPlanList(ctx context.Context, mealNum uint, shopCartMustProductInfo ro.MustPlanProductInfo, deskUuid uint64, options ...func(option *CheckOption)) ([]resp.InstantProductMustPlan, error) // 获取桌台的各个必点方案的点购情况，如是否满足必点、还差多少个
	GetMustPlanUuidByProductPackage(ctx context.Context, saleBillUuid, productPackageUuid uint64, deskUuid uint64) (uint64, error)                                                                       // 通过商品包获取必点方案uuid
}

// mustPlanSrv 必点方案服务结构体
type mustPlanSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewMustPlanSrv 创建新的必点方案服务
func NewMustPlanSrv(dbm *database.DBManager) IMustPlanSrv {
	return NewMustPlanSrvImpl(dbm)
}

// NewMustPlanSrvImpl 创建新的必点方案服务实现
func NewMustPlanSrvImpl(dbm *database.DBManager) IMustPlanSrv {
	return &mustPlanSrv{
		dbm: dbm,
	}
}

type CheckOption struct {
	Scene                    uint                    // 检查场景. 1-下单场景 2-结账场景
	SaleBillUuid             uint64                  // 桌台自助餐场景下，将必点商品价格记为0元
	SelectedMustPlanProducts *ro.MustPlanProductInfo // 桌台已经选择的必点商品。使用场景仅用于平板加购并送厨时，将新加购的商品构建为该对象
}

const (
	CheckSceneCooking  = 1 // 下单场景
	CheckSceneCheckout = 2 // 结账场景
)

func WithCheckSceneCooking() func(option *CheckOption) {
	return func(option *CheckOption) {
		option.Scene = CheckSceneCooking
	}
}

func WithCheckSceneCheckout() func(option *CheckOption) {
	return func(option *CheckOption) {
		option.Scene = CheckSceneCheckout
	}
}

func WithSaleBillUuid(saleBillUuid uint64) func(option *CheckOption) {
	return func(option *CheckOption) {
		option.SaleBillUuid = saleBillUuid
	}
}

func WithSelectedMustPlanProductsCheckOption(selectedMustPlanProducts *ro.MustPlanProductInfo) func(option *CheckOption) {
	return func(option *CheckOption) {
		option.SelectedMustPlanProducts = selectedMustPlanProducts
	}
}

// GetInstantMustPlanList 获取点餐的必点方案列表，用于获取各个必点方案的点餐情况
func (s *mustPlanSrv) GetInstantMustPlanList(ctx context.Context, db *gorm.DB, shopCartMustProductInfo ro.MustPlanProductInfo, options ...func(option *CheckOption)) ([]resp.InstantProductMustPlan, error) {
	option := &CheckOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	mustPlanList := make([]resp.InstantProductMustPlan, 0)

	// 获取必选方案信息
	productMustPlans, err := repository.NewProductMustPlanRepo(db).GetProductMustPlanListAllInfos(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 构建必点方案响应列表
	for _, plan := range productMustPlans {
		if !plan.IsInstantMustPlan() {
			// 忽略桌台必点方案
			continue
		}
		if plan.IsDelete() {
			continue
		}

		if option.Scene != 0 {
			if option.Scene == CheckSceneCheckout {
				// 结账场景下，如果“结账时检查必点商品”为关闭时，则不检查必点商品
				if plan.AutoCheckout == 0 {
					continue
				}
			}
			if option.Scene == CheckSceneCooking {
				// 结账场景下，如果“下单时检查必点商品”为关闭时，则不检查必点商品
				if plan.AutoCheck == 0 {
					continue
				}
			}
		}

		productPackageList := s.getInstantMustPlanProductList(ctx, plan)
		// 如果列表为空，跳过不显示
		if len(productPackageList) == 0 {
			continue
		}
		// 获取购物车中已经点了多个这个必点方案的商品
		selectedNum := float64(0)
		productPackageMap, ok := shopCartMustProductInfo[plan.Uuid]
		if ok {
			for i, productPackage := range productPackageList {
				productPackageUuid := productPackage.Product.Uuid
				num := productPackageMap[productPackageUuid]

				// 设置每个商品的已选数量
				productPackageList[i].SelectedNum = num
				if productPackageList[i].SelectedNum >= productPackageList[i].MustNum {
					productPackageList[i].NeedNum = 0
				} else {
					productPackageList[i].NeedNum = productPackageList[i].MustNum - productPackageList[i].SelectedNum
				}
				// 如果必点方案是“每单必点1份”且“固定商品”
				if plan.GetMustType() == constant.ProductMustPlanMustTypeEachOrder && plan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
					selectedNum += num
				}
				// 如果必点方案是 “每单必点1份” 且 “任选商品”
				if plan.GetMustType() == constant.ProductMustPlanMustTypeEachOrder && plan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
					// “已选择x份” 为选了多少个不同的商品
					if num > 0 {
						selectedNum += 1
					}
				}
			}
		}

		// 获取购物车中各个必点商品已经点了多少个，还差多少个
		mustMap := make(map[uint64]uint) // product_package_uuid => num 每个必点商品还差多少个
		if ok {
			for _, productPackage := range productPackageList {
				productPackageUuid := productPackage.Product.Uuid
				num := productPackageMap[productPackageUuid] // 该商品已点xx个
				mustNum := productPackage.MustNum            // 该商品要求点的数量
				result := int(mustNum) - int(num)            // uint类型，不能用result < 0来判断. uint不会小于0，负数是个很大的正数
				if result > 0 {
					mustMap[productPackageUuid] = uint(result)
				} else {
					mustMap[productPackageUuid] = 0
				}
			}
		}
		// 如果必点方案是可选商品，NeedNum的取值要么是1 要么是0
		// 当selectedNum>0时，NeedNum为0
		needNum := float64(0)
		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
			if selectedNum > 0 {
				needNum = 0
			} else {
				needNum = 1
			}
		}
		// 如果必点方案是固定商品时，NeedNum的取值为“必选弹框”列表中商品还差数量之和
		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
			for _, product := range productPackageList {
				needNum += product.NeedNum

			}
		}

		mustPlan := resp.InstantProductMustPlan{
			Uuid:         plan.Uuid,
			Name:         plan.Name,
			MustType:     plan.GetMustType(),
			MustRule:     plan.GetMustRule(),
			CanChangeNum: plan.IsCustomerCanChange(),
			MealNum:      1,           // 点餐必点方案的MealNum为1
			SelectedNum:  selectedNum, // 已选择xx份
			NeedNum:      needNum,     // 还差xx份。不应该算上自动加购商品的数量
			Products:     resp.ProductPackageList{List: productPackageList},
		}
		mustPlanList = append(mustPlanList, mustPlan)
	}

	return mustPlanList, nil
}

// 获取点餐的必点方案的商品列表。用于加购商品时判断该商品是不是这些点餐列表里的商品
func (s *mustPlanSrv) getInstantMustPlanProductList(ctx context.Context, mustPlan *model.ProductMustPlan) []resp.InstantMustPlanProductStat {
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	// 不是点餐的必单方案
	if !mustPlan.IsInstantMustPlan() {
		return []resp.InstantMustPlanProductStat{}
	}
	if mustPlan.IsDelete() {
		return []resp.InstantMustPlanProductStat{}
	}
	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
	stockNumFn := func(productBomUuid uint64) float64 {
		// 使用 ProductInventoryAppService 查询商品BOM库存
		inventory, err := appService.GetProductInventory(ctx, productBomUuid)
		if err != nil {
			// 查询失败时记录日志并返回0
			ctx.Log().Warn("查询商品BOM库存失败", zap.Uint64("productBomUuid", productBomUuid), zap.Error(err))
			return 0
		}
		return inventory
	}

	productList := make([]resp.InstantMustPlanProductStat, 0)
	// 点餐的必点方案都是“每单必点”，没有“每人必点”类型的
	// 如果是固定商品
	if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
		for _, planItem := range mustPlan.ProductMustPlanItems {
			productPackage := planItem.GetProductInfo(baseUrl, stockNumFn)
			if productPackage == nil {
				continue
			}

			canAutoAdd, productAutoAddReq := planItem.ProductPackage.IsNoSelectProduct()
			isAutoAdd := mustPlan.IsAutoCart() && canAutoAdd
			var autoAddReq resp.ProductAutoAddReq
			if isAutoAdd {
				autoAddReq = *productAutoAddReq
			}

			productStat := resp.InstantMustPlanProductStat{
				Product:           *productPackage,
				IsAutoAdd:         isAutoAdd, // 必点方案勾选自动加购且商品是无选择商品
				ProductAutoAddReq: autoAddReq,
				SelectedNum:       0, // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
				MustNum:           1, // 该商品要求必点数量
				NeedNum:           1, // 该商品还需点点数量。展示给前端之前判断购物车内是否已经点了该商品，减已点数量
			}
			productList = append(productList, productStat)
		}
	}
	// 如果是可选商品
	if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
		for _, planItem := range mustPlan.ProductMustPlanItems {
			productPackage := planItem.GetProductInfo(baseUrl, stockNumFn)
			if productPackage == nil {
				continue
			}
			productStat := resp.InstantMustPlanProductStat{
				Product:     *productPackage,
				IsAutoAdd:   false, // 可选商品的必点方案都不自动加购
				SelectedNum: 0,     // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
				MustNum:     0,
				NeedNum:     0,
			}
			productList = append(productList, productStat)
		}
	}

	return productList
}

// GetProductMustPlans 获取桌台的必选方案model列表
func (s *mustPlanSrv) GetProductMustPlans(ctx context.Context, deskUuid uint64) ([]*model.ProductMustPlan, error) {
	db := ctx.GetDB()
	// 获取全部必选方案
	productMustPlanList, err := repository.NewProductMustPlanRepo(db).GetProductMustPlanListDeskInfos(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 过滤出属于该桌台的必选方案
	productMustPlans := make([]*model.ProductMustPlan, 0) // 桌台的必选方案列表
	for index, mustPlan := range productMustPlanList {
		for _, planRegion := range mustPlan.ProductMustPlanRegions {
			for _, desk := range planRegion.DeskRegion.Desks {
				if desk.Uuid == deskUuid {
					productMustPlans = append(productMustPlans, productMustPlanList[index])
				}
			}
		}
	}
	return productMustPlans, nil
}

// GetDeskMustPlanProductPackageList 获取桌台的必点ProductPackage列表
func (s *mustPlanSrv) GetDeskMustPlanProductPackageList(ctx context.Context, deskUuid uint64) ([]*model.ProductPackage, error) {
	productMustPlans, err := s.GetProductMustPlans(ctx, deskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	productPackageList := make([]*model.ProductPackage, 0)
	for _, mustPlan := range productMustPlans {
		for _, planItem := range mustPlan.ProductMustPlanItems {
			productPackageList = append(productPackageList, &planItem.ProductPackage)
		}
	}
	return productPackageList, nil
}

// GetDeskMustPlanList 获取桌台的各个必点方案的点购情况，如是否满足必点、还差多少个
func (s *mustPlanSrv) GetDeskMustPlanList(ctx context.Context, mealNum uint, shopCartMustProductInfo ro.MustPlanProductInfo, deskUuid uint64, options ...func(option *CheckOption)) ([]resp.InstantProductMustPlan, error) {
	option := &CheckOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	if option.SelectedMustPlanProducts != nil {
		// 将新加购的商品与已经加购的商品合并
		// MustPlanUuid => ProductPackageUuid => num
		for mustPlanUuid, productPackageMap := range *option.SelectedMustPlanProducts {
			for productPackageUuid, num := range productPackageMap {
				if _, ok := shopCartMustProductInfo[mustPlanUuid]; !ok {
					shopCartMustProductInfo[mustPlanUuid] = make(map[uint64]float64)
				}
				shopCartMustProductInfo[mustPlanUuid][productPackageUuid] += num
			}
		}
	}

	// 获取该桌台的必选方案
	productMustPlanList, err := s.GetProductMustPlans(ctx, deskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	mustPlanList := make([]resp.InstantProductMustPlan, 0)

	// 获取自助餐商品列表
	buffetProductMap := make(map[uint64]bool) // product_package_uuid => true
	if option.SaleBillUuid != 0 {
		saleBill, err := repository.NewSaleBillRepo(ctx.GetDB()).GetSaleBillBuffetProductList(option.SaleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		buffetProducts := saleBill.GetBuffetProductList()
		for _, buffetProduct := range buffetProducts.List {
			buffetProductMap[buffetProduct.Uuid] = true
		}
	}

	// 构建必点方案响应列表
	for _, plan := range productMustPlanList {
		if plan.IsDelete() {
			continue
		}

		if option.Scene != 0 {
			if option.Scene == CheckSceneCheckout {
				// 结账场景下，如果“结账时检查必点商品”为关闭时，则不检查必点商品
				if plan.AutoCheckout == 0 {
					continue
				}
			}
			if option.Scene == CheckSceneCooking {
				// 结账场景下，如果“下单时检查必点商品”为关闭时，则不检查必点商品
				if plan.AutoCheck == 0 {
					continue
				}
			}
		}

		productPackages := s.getIDeskMustPlanProductList(ctx, plan, mealNum)
		// 如果列表为空，跳过不检查
		if len(productPackages) == 0 {
			continue
		}

		// 获取购物车中已经点了多个这个必点方案的商品
		selectedNum := float64(0)
		productPackageMap, ok := shopCartMustProductInfo[plan.Uuid]
		if ok {
			for _, productPackage := range productPackages {
				productPackageUuid := productPackage.Product.Uuid
				num := productPackageMap[productPackageUuid]
				// 设置每个商品的已选数量
				selectedNum += num
			}
		}
		// 获取购物车中各个必点商品已经点了多少个，还差多少个
		mustMap := make(map[uint64]float64) // product_package_uuid => num 每个必点商品还差多少个
		for i, productPackage := range productPackages {
			productPackageUuid := productPackage.Product.Uuid
			num := productPackageMap[productPackageUuid] // 该商品已点xx个
			mustNum := productPackage.MustNum            // 该商品要求点的数量
			result := mustNum - num
			// 如果已点数量大于要求点数量，则该商品还差0个. 注意：uint类型，不能用result < 0来判断
			if mustNum < num {
				result = 0
			}
			mustMap[productPackageUuid] = result
			// 设置每个商品的已选数量
			productPackages[i].SelectedNum = num
			if productPackages[i].SelectedNum >= productPackages[i].MustNum {
				productPackages[i].NeedNum = 0
			} else {
				productPackages[i].NeedNum = productPackages[i].MustNum - productPackages[i].SelectedNum
			}

			// 如果商品是自助餐商品，则价格为0元
			if option.SaleBillUuid != 0 && buffetProductMap[productPackageUuid] {
				productPackages[i].Product.Price = 0 // 自助餐商品价格为0元
				// 自助餐商品的所有规格价格为0元
				for j := range productPackages[i].Product.Flavors.List {
					flavor := &productPackages[i].Product.Flavors.List[j]
					flavor.Price = 0
				}
			}
		}

		// 如果必点方案是可选商品
		// 当selectedNum>=桌台人数时，NeedNum为0
		needNum := float64(0)
		// 可选商品、每人必点
		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAny && plan.GetMustType() == constant.ProductMustPlanMustTypeEachPerson {
			if selectedNum >= float64(mealNum) {
				needNum = 0
			} else {
				needNum = float64(mealNum) - selectedNum
			}
		}
		// 可选商品、每单必点
		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAny && plan.GetMustType() == constant.ProductMustPlanMustTypeEachOrder {
			if selectedNum >= 1 {
				needNum = 0
			} else {
				needNum = 1
			}
		}
		// 如果必点方案是固定商品时，NeedNum的取值为“必选弹框”列表中商品还差数量之和
		if plan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
			for _, num := range mustMap {
				needNum += num
			}
		}

		mustPlan := resp.InstantProductMustPlan{
			Uuid:         plan.Uuid,
			Name:         plan.Name,
			MustType:     plan.GetMustType(),
			MustRule:     plan.GetMustRule(),
			CanChangeNum: plan.IsCustomerCanChange(),
			MealNum:      mealNum,
			SelectedNum:  selectedNum, // 已选择xx份
			NeedNum:      needNum,     // 还差xx份。应该算上自动加购商品的数量
			Products:     resp.ProductPackageList{List: productPackages},
		}
		mustPlanList = append(mustPlanList, mustPlan)
	}

	return mustPlanList, nil
}

// 获取桌台的各个必点方案的要求，如某个商品要求选购多少个
func (s *mustPlanSrv) getIDeskMustPlanProductList(ctx context.Context, mustPlan *model.ProductMustPlan, personNum uint) []resp.InstantMustPlanProductStat {
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	// 不是桌台的必单方案
	if !mustPlan.IsDeskMustPlan() {
		return []resp.InstantMustPlanProductStat{}
	}
	if mustPlan.IsDelete() {
		return []resp.InstantMustPlanProductStat{}
	}
	productList := make([]resp.InstantMustPlanProductStat, 0)

	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
	stockNumFn := func(productBomUuid uint64) float64 {
		// 使用 ProductInventoryAppService 查询商品BOM库存
		inventory, err := appService.GetProductInventory(ctx, productBomUuid)
		if err != nil {
			// 查询失败时记录日志并返回0
			ctx.Log().Warn("查询商品BOM库存失败", zap.Uint64("productBomUuid", productBomUuid), zap.Error(err))
			return 0
		}
		return inventory
	}

	// 如果是"每单必点"、固定商品
	if mustPlan.GetMustType() == constant.ProductMustPlanMustTypeEachOrder {
		if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
			for _, planItem := range mustPlan.ProductMustPlanItems {

				productPackage := planItem.GetProductInfo(baseUrl, stockNumFn)
				if productPackage == nil {
					continue
				}

				canAutoAdd, productAutoAddReq := planItem.ProductPackage.IsNoSelectProduct()
				isAutoAdd := mustPlan.IsAutoCart() && canAutoAdd
				var autoAddReq resp.ProductAutoAddReq
				if isAutoAdd {
					autoAddReq = *productAutoAddReq
				}

				productStat := resp.InstantMustPlanProductStat{
					Product:           *productPackage,
					IsAutoAdd:         isAutoAdd, // 必点方案勾选自动加购且商品是无选择商品
					ProductAutoAddReq: autoAddReq,
					SelectedNum:       0, // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
					MustNum:           1, // 该商品要求必点数量
					NeedNum:           1, // 该商品还需点点数量。展示给前端之前判断购物车内是否已经点了该商品，减已点数量
				}
				productList = append(productList, productStat)
			}
		}
		// 如果是"每单必点"、可选商品
		if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
			for _, planItem := range mustPlan.ProductMustPlanItems {
				productPackage := planItem.GetProductInfo(baseUrl, stockNumFn)
				if productPackage == nil {
					continue
				}
				productStat := resp.InstantMustPlanProductStat{
					Product:     *productPackage,
					IsAutoAdd:   false, // 可选商品的必点方案都不自动加购
					SelectedNum: 0,     // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
					MustNum:     0,
					NeedNum:     0,
				}
				productList = append(productList, productStat)
			}
		}
	}

	// "每人必点"
	if mustPlan.GetMustType() == constant.ProductMustPlanMustTypeEachPerson {
		// 如果是"每人必点"、固定商品
		if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAll {
			for _, planItem := range mustPlan.ProductMustPlanItems {
				productPackage := planItem.GetProductInfo(baseUrl, stockNumFn)
				if productPackage == nil {
					continue
				}

				canAutoAdd, productAutoAddReq := planItem.ProductPackage.IsNoSelectProduct()
				isAutoAdd := mustPlan.IsAutoCart() && canAutoAdd
				var autoAddReq resp.ProductAutoAddReq
				if isAutoAdd {
					autoAddReq = *productAutoAddReq
				}

				productStat := resp.InstantMustPlanProductStat{
					Product:           *productPackage,
					IsAutoAdd:         isAutoAdd, // 必点方案勾选自动加购且商品是无选择商品
					ProductAutoAddReq: autoAddReq,
					SelectedNum:       0,                      // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
					MustNum:           1 * float64(personNum), // 该商品要求必点数量
					NeedNum:           1 * float64(personNum), // 该商品还需点点数量。展示给前端之前判断购物车内是否已经点了该商品，减已点数量
				}
				productList = append(productList, productStat)
			}
		}
		// 如果是"每人必点"、可选商品
		if mustPlan.GetMustRule() == constant.ProductMustPlanMustRuleAny {
			for _, planItem := range mustPlan.ProductMustPlanItems {
				productPackage := planItem.GetProductInfo(baseUrl, stockNumFn)
				if productPackage == nil {
					continue
				}
				productStat := resp.InstantMustPlanProductStat{
					Product:     *productPackage,
					IsAutoAdd:   false, // 可选商品的必点方案都不自动加购
					SelectedNum: 0,     // 该商品已经点的数量。展示给前端之前判断购物车内是否已经点了该商品，加上已点数量
					MustNum:     0,
					NeedNum:     0,
				}
				productList = append(productList, productStat)
			}
		}
	}

	return productList
}

// GetMustPlanUuidByProductPackage 通过productPackageUuid判断该商品包是哪个必点方案的
func (s *mustPlanSrv) GetMustPlanUuidByProductPackage(ctx context.Context, saleBillUuid, productPackageUuid uint64, deskUuid uint64) (uint64, error) {
	// 用product_package_uuid在ttpos_product_must_plan_item表中查询
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	productPlanItems, err := repository.NewProductMustPlanItemRepo(db).GetProductMustPlanItemByPackageUuid(productPackageUuid)
	if err != nil {
		ctx.Log().Debug("加购商品时判断商品是不是必点商品时，查询必点商品失败", zap.Error(err))
		return 0, errors.WithMessage(err)
	}
	// 该商品不是必点商品
	if len(productPlanItems) == 0 {
		return 0, nil
	}
	// 该商品是必点商品且只属于一个必点方案
	if len(productPlanItems) == 1 {
		mustPlanUuid := productPlanItems[0].ProductMustPlan.Uuid
		ctx.Log().Debug("加购商品时判断商品是不是必点商品", zap.Any("商品包productPackageUuid", productPackageUuid), zap.Any("必点方案mustPlanUuid", mustPlanUuid))
		return mustPlanUuid, nil
	}
	// 该商品是必点商品且属于多个必点方案。选择第一个未满足必点要求的方案

	// 判断这些方案是不是该销售账单的必点方案

	// 获取某个方案的必点情况
	var plans []resp.InstantProductMustPlan

	// 查询到购物车信息
	shopCart, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	if deskUuid != 0 {
		deskMustPlanList, err := s.GetDeskMustPlanList(ctx, shopCart.SaleBill.MealNum, shopCart.GetMustPlanProductInfo(), deskUuid)
		if err != nil {
			ctx.Log().Debug("加购商品时判断商品是不是必点商品时，获取桌台必点方案失败", zap.Error(err))
			return 0, errors.WithMessage(err, "获取桌台必点方案失败")
		}
		plans = deskMustPlanList
	} else {
		instantMustPlanList, err := s.GetInstantMustPlanList(ctx, db, shopCart.GetMustPlanProductInfo())
		if err != nil {
			ctx.Log().Debug("加购商品时判断商品是不是必点商品时，获取点餐必点方案失败", zap.Error(err))
			return 0, errors.WithMessage(err, "获取点餐必点方案失败")
		}
		plans = instantMustPlanList
	}
	deskPlanUuidMap := make(map[string]resp.InstantProductMustPlan)

	for i, plan := range plans {
		deskPlanUuidMap[plan.Name] = plans[i]
	}

	productPlanItemList := make([]model.ProductMustPlanItem, 0)
	for i, item := range productPlanItems {
		name := item.ProductMustPlan.Name
		if _, exit := deskPlanUuidMap[name]; exit {
			productPlanItemList = append(productPlanItemList, productPlanItems[i])
		}
	}

	if len(productPlanItemList) == 1 {
		mustPlanUuid := productPlanItemList[0].ProductMustPlan.Uuid
		ctx.Log().Debug("加购商品时判断商品是不是必点商品", zap.Any("商品包productPackageUuid", productPackageUuid), zap.Any("必点方案mustPlanUuid", mustPlanUuid))
		return mustPlanUuid, nil
	}

	for _, mustPlanItem := range productPlanItemList {
		name := mustPlanItem.ProductMustPlan.Name
		plan := deskPlanUuidMap[name]
		if plan.NeedNum > 0 {
			return mustPlanItem.ProductMustPlanUuid, nil
		}
	}

	ctx.Log().Debug("获取商品的必点方案可能异常")
	return 0, nil
}
