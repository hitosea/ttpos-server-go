package service

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// GetProductPackageInventoryOption 商品包库存查询选项
type GetProductPackageInventoryOption struct {
	Strategy IProductPackageInventoryStrategy // 商品包库存计算策略（min或max），如果为nil则使用默认策略
}

// WithStrategy 设置商品包库存计算策略
func WithStrategy(strategy IProductPackageInventoryStrategy) func(option *GetProductPackageInventoryOption) {
	return func(option *GetProductPackageInventoryOption) {
		option.Strategy = strategy
	}
}

// IProductInventoryDomainService 商品库存领域服务接口
type IProductInventoryDomainService interface {
	// GetProductInventory 获取商品库存
	// productBomUuid: 商品BOM的UUID
	// 返回: 库存数量（float64），无限库存返回 99999999
	GetProductInventory(ctx context.Context, productBomUuid uint64) (float64, error)

	// GetProductInventoriesBatch 批量获取商品库存
	// productBomUuids: 商品BOM的UUID列表
	// 返回: map[UUID]库存数量，无限库存返回 99999999
	GetProductInventoriesBatch(ctx context.Context, productBomUuids []uint64) (map[uint64]float64, error)

	// GetProductPackageInventory 获取商品包库存
	// productPackageUuid: 商品包的UUID
	// opts: 可选参数，使用 WithStrategy 设置策略
	// 返回: 库存数量（float64），根据策略计算该商品包下所有BOM库存的最小值或最大值
	GetProductPackageInventory(ctx context.Context, productPackageUuid uint64, opts ...func(option *GetProductPackageInventoryOption)) (float64, error)

	// GetProductPackageInventoriesBatch 批量获取商品包库存
	// productPackageUuids: 商品包的UUID列表
	// opts: 可选参数，使用 WithStrategy 设置策略
	// 返回: map[商品包UUID]库存数量，根据策略计算每个商品包下所有BOM库存的最小值或最大值
	GetProductPackageInventoriesBatch(ctx context.Context, productPackageUuids []uint64, opts ...func(option *GetProductPackageInventoryOption)) (map[uint64]float64, error)
}

// productInventoryDomainService 商品库存领域服务实现
type productInventoryDomainService struct {
	productBomRepo                  repository.IProductBomRepository
	productPackageRepo              repository.IProductPackageRepository
	strategies                      map[string]IInventoryStrategy
	defaultPackageInventoryStrategy IProductPackageInventoryStrategy // 默认商品包库存策略（最小值）
}

// NewProductInventoryDomainService 创建商品库存领域服务
func NewProductInventoryDomainService(
	productBomRepo repository.IProductBomRepository,
	productPackageRepo repository.IProductPackageRepository,
) IProductInventoryDomainService {
	return NewProductInventoryDomainServiceWithStrategies(
		productBomRepo,
		productPackageRepo,
		NewBomCardProductInventoryStrategy(),
		NewFlavorMaterialsProductInventoryStrategy(),
		NewSauceBomCardProductInventoryStrategy(),
		NewSauceMaterialsProductInventoryStrategy(),
		NewSauceNonBomCardProductInventoryStrategy(),
		NewNonBomCardProductInventoryStrategy(),
		NewMaxProductPackageInventoryStrategy(), // 默认使用最大值策略
	)
}

// NewProductInventoryDomainServiceWithStrategies 创建商品库存领域服务（支持策略注入，用于测试）
func NewProductInventoryDomainServiceWithStrategies(
	productBomRepo repository.IProductBomRepository,
	productPackageRepo repository.IProductPackageRepository,
	bomCardStrategy IInventoryStrategy,
	flavorMaterialsStrategy IInventoryStrategy,
	sauceBomCardStrategy IInventoryStrategy,
	sauceMaterialsStrategy IInventoryStrategy,
	sauceNonBomCardStrategy IInventoryStrategy,
	nonBomCardStrategy IInventoryStrategy,
	packageInventoryStrategy IProductPackageInventoryStrategy,
) IProductInventoryDomainService {
	strategies := make(map[string]IInventoryStrategy)
	strategies["bom_card"] = bomCardStrategy
	strategies["flavor_materials"] = flavorMaterialsStrategy
	strategies["sauce_bom_card"] = sauceBomCardStrategy
	strategies["sauce_materials"] = sauceMaterialsStrategy
	strategies["sauce_non_bom_card"] = sauceNonBomCardStrategy
	strategies["non_bom_card"] = nonBomCardStrategy

	return &productInventoryDomainService{
		productBomRepo:                  productBomRepo,
		productPackageRepo:              productPackageRepo,
		strategies:                      strategies,
		defaultPackageInventoryStrategy: packageInventoryStrategy,
	}
}

// GetProductInventory 获取商品库存
func (s *productInventoryDomainService) GetProductInventory(
	ctx context.Context,
	productBomUuid uint64,
) (float64, error) {
	// 1. 查询商品BOM
	productBomInterface, err := s.productBomRepo.FindByUuid(ctx, productBomUuid)
	if err != nil {
		return 0, errors.WithMessage(err, "查询商品BOM失败")
	}
	if productBomInterface == nil {
		return 0, errors.New("商品不存在")
	}

	// 2. 类型断言为 *model.ProductBom
	productBom, ok := productBomInterface.(*model.ProductBom)
	if !ok {
		return 0, errors.New("商品BOM类型错误")
	}

	// 3. 判断商品类型，选择策略
	var strategy IInventoryStrategy
	if productBom.HasProductBomCard() {
		strategy = s.strategies["bom_card"]
	} else if len(productBom.FlavorMaterials) > 0 {
		strategy = s.strategies["flavor_materials"]
	} else if productBom.IsSauce() {
		// 小料：优先判断是否有成本卡，其次判断是否有关联材料，都没有则使用专门的小料策略
		if productBom.ProductSauce.HasProductBomCard() {
			strategy = s.strategies["sauce_bom_card"]
		} else if len(productBom.ProductSauce.SauceMaterials) > 0 {
			strategy = s.strategies["sauce_materials"]
		} else {
			// 小料既没有成本卡也没有关联材料，使用专门的小料策略（检查 ProductSauce.SauceMaterials）
			strategy = s.strategies["sauce_non_bom_card"]
		}
	} else {
		strategy = s.strategies["non_bom_card"]
	}

	// 4. 计算库存
	inventory, err := strategy.CalculateInventory(ctx, productBom)
	if err != nil {
		return 0, errors.WithMessage(err, "计算库存失败")
	}

	return inventory, nil
}

// GetProductInventoriesBatch 批量获取商品库存
func (s *productInventoryDomainService) GetProductInventoriesBatch(
	ctx context.Context,
	productBomUuids []uint64,
) (map[uint64]float64, error) {
	// 1. 参数校验
	if len(productBomUuids) == 0 {
		return make(map[uint64]float64), nil
	}

	// 2. 批量查询商品BOM
	productBomInterfaces, err := s.productBomRepo.FindByUuids(ctx, productBomUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品BOM失败")
	}

	// 3. 初始化结果map
	result := make(map[uint64]float64, len(productBomUuids))

	// 4. 遍历每个BOM，计算库存
	for _, bomInterface := range productBomInterfaces {
		productBom, ok := bomInterface.(*model.ProductBom)
		if !ok {
			// 记录错误但继续处理其他BOM
			logger.Logger.Warn("*model.ProductBom类型错误", zap.Uint64("company_uuid", ctx.GetCompanyUuid()), zap.Error(errors.New("商品BOM类型错误")))
			continue
		}

		// 5. 判断商品类型，选择策略
		var strategy IInventoryStrategy
		if productBom.HasProductBomCard() {
			strategy = s.strategies["bom_card"]
		} else if len(productBom.FlavorMaterials) > 0 {
			strategy = s.strategies["flavor_materials"]
		} else if productBom.IsSauce() {
			// 小料：优先判断是否有成本卡，其次判断是否有关联材料，都没有则使用专门的小料策略
			if productBom.ProductSauce.HasProductBomCard() {
				strategy = s.strategies["sauce_bom_card"]
			} else if len(productBom.ProductSauce.SauceMaterials) > 0 {
				strategy = s.strategies["sauce_materials"]
			} else {
				// 小料既没有成本卡也没有关联材料，使用专门的小料策略（检查 ProductSauce.SauceMaterials）
				strategy = s.strategies["sauce_non_bom_card"]
			}
		} else {
			strategy = s.strategies["non_bom_card"]
		}

		// 6. 计算库存
		inventory, err := strategy.CalculateInventory(ctx, productBom)
		if err != nil {
			// 记录错误但继续处理其他BOM
			logger.Logger.Warn("计算商品库存失败", zap.Uint64("company_uuid", ctx.GetCompanyUuid()), zap.Uint64("product_bom_uuid", productBom.Uuid), zap.Error(err))
			continue
		}

		result[productBom.Uuid] = inventory
	}

	return result, nil
}

// GetProductPackageInventory 获取商品包库存
// 根据策略计算该商品包下所有商品BOM库存的最小值或最大值
func (s *productInventoryDomainService) GetProductPackageInventory(
	ctx context.Context,
	productPackageUuid uint64,
	opts ...func(option *GetProductPackageInventoryOption),
) (float64, error) {
	// 解析选项
	option := &GetProductPackageInventoryOption{}
	for _, opt := range opts {
		opt(option)
	}

	// 1. 查询商品包信息，检查状态
	productPackage, err := s.productPackageRepo.FindByUuid(ctx, productPackageUuid)
	if err != nil {
		return 0, errors.WithMessage(err, "查询商品包失败")
	}
	if productPackage == nil {
		return 0, errors.New("商品包不存在")
	}

	// 2. 如果商品包已删除或下架，直接返回库存为0
	if productPackage.IsDelete() || productPackage.IsDown() {
		return 0, nil
	}

	// 3. 查询商品包下所有BOM
	productBomInterfaces, err := s.productBomRepo.FindByProductPackageUuid(ctx, productPackageUuid)
	if err != nil {
		return 0, errors.WithMessage(err, "查询商品包BOM列表失败")
	}

	if len(productBomInterfaces) == 0 {
		return 0, errors.New("商品包下没有BOM")
	}

	// 4. 收集所有BOM的UUID
	productBomUuids := make([]uint64, 0, len(productBomInterfaces))
	for _, bomInterface := range productBomInterfaces {
		productBom, ok := bomInterface.(*model.ProductBom)
		if !ok {
			// 记录错误但继续处理其他BOM
			continue
		}
		productBomUuids = append(productBomUuids, productBom.Uuid)
	}

	// 5. 批量查询库存
	inventoryMap, err := s.GetProductInventoriesBatch(ctx, productBomUuids)
	if err != nil {
		return 0, errors.WithMessage(err, "批量查询商品包BOM库存失败")
	}

	// 6. 收集库存值
	inventories := make([]float64, 0, len(inventoryMap))
	for _, uuid := range productBomUuids {
		if inventory, exists := inventoryMap[uuid]; exists {
			inventories = append(inventories, inventory)
		}
	}

	// 7. 如果没有有效的库存，返回错误
	if len(inventories) == 0 {
		return 0, errors.New("无法计算商品包库存：所有BOM库存查询失败")
	}

	// 8. 使用策略计算商品包库存
	// 如果未提供策略，使用默认策略
	strategy := option.Strategy
	if strategy == nil {
		strategy = s.defaultPackageInventoryStrategy
	}

	return strategy.CalculatePackageInventory(ctx, inventories)
}

// GetProductPackageInventoriesBatch 批量获取商品包库存
// 根据策略计算每个商品包下所有商品BOM库存的最小值或最大值
func (s *productInventoryDomainService) GetProductPackageInventoriesBatch(
	ctx context.Context,
	productPackageUuids []uint64,
	opts ...func(option *GetProductPackageInventoryOption),
) (map[uint64]float64, error) {
	// 解析选项
	option := &GetProductPackageInventoryOption{}
	for _, opt := range opts {
		opt(option)
	}

	// 1. 参数校验
	if len(productPackageUuids) == 0 {
		return make(map[uint64]float64), nil
	}

	// 2. 批量查询商品包信息，检查状态
	productPackages, err := s.productPackageRepo.FindByUuids(ctx, productPackageUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品包失败")
	}

	// 3. 过滤有效的商品包（未删除）
	validPackageUuids := make([]uint64, 0, len(productPackages))
	packageMap := make(map[uint64]*model.ProductPackage, len(productPackages))
	for _, productPackage := range productPackages {
		if productPackage != nil && !productPackage.IsDelete() {
			validPackageUuids = append(validPackageUuids, productPackage.Uuid)
			packageMap[productPackage.Uuid] = productPackage
		}
	}

	if len(validPackageUuids) == 0 {
		return make(map[uint64]float64), nil
	}

	// 4. 批量查询所有商品包下的BOM
	productBomInterfaces, err := s.productBomRepo.FindByProductPackageUuids(ctx, validPackageUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品包BOM列表失败")
	}

	// 5. 按商品包分组BOM，并收集所有BOM的UUID
	packageBomMap := make(map[uint64][]uint64) // map[商品包UUID][]BOM UUID
	allBomUuids := make([]uint64, 0)
	bomUuidSet := make(map[uint64]bool) // 用于去重

	for _, bomInterface := range productBomInterfaces {
		productBom, ok := bomInterface.(*model.ProductBom)
		if !ok {
			// 记录错误但继续处理其他BOM
			continue
		}

		packageUuid := productBom.ProductPackageUuid
		bomUuid := productBom.Uuid

		// 添加到对应商品包的BOM列表
		if _, exists := packageBomMap[packageUuid]; !exists {
			packageBomMap[packageUuid] = make([]uint64, 0)
		}
		packageBomMap[packageUuid] = append(packageBomMap[packageUuid], bomUuid)

		// 收集所有BOM UUID（去重）
		if !bomUuidSet[bomUuid] {
			allBomUuids = append(allBomUuids, bomUuid)
			bomUuidSet[bomUuid] = true
		}
	}

	// 6. 批量查询所有BOM的库存
	inventoryMap, err := s.GetProductInventoriesBatch(ctx, allBomUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品包BOM库存失败")
	}

	// 7. 使用策略计算每个商品包的库存
	// 如果未提供策略，使用默认策略
	strategy := option.Strategy
	if strategy == nil {
		strategy = s.defaultPackageInventoryStrategy
	}

	result := make(map[uint64]float64, len(validPackageUuids))
	for _, packageUuid := range validPackageUuids {
		bomUuids, exists := packageBomMap[packageUuid]
		if !exists || len(bomUuids) == 0 {
			// 商品包下没有BOM，库存为0
			result[packageUuid] = 0
			continue
		}

		// 收集该商品包下所有BOM的库存值
		inventories := make([]float64, 0, len(bomUuids))
		for _, bomUuid := range bomUuids {
			if inventory, exists := inventoryMap[bomUuid]; exists {
				inventories = append(inventories, inventory)
			}
		}

		// 如果没有有效的库存，设置为0
		if len(inventories) == 0 {
			result[packageUuid] = 0
			continue
		}

		// 使用策略计算商品包库存
		packageInventory, err := strategy.CalculatePackageInventory(ctx, inventories)
		if err != nil {
			// 记录错误但继续处理其他商品包
			logger.Logger.Warn("计算商品包库存失败", zap.Uint64("company_uuid", ctx.GetCompanyUuid()), zap.Uint64("product_package_uuid", packageUuid), zap.Error(err))
			result[packageUuid] = 0
			continue
		}

		result[packageUuid] = packageInventory
	}

	return result, nil
}
