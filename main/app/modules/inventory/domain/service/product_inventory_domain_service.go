package service

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/pkg/context"
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

	// GetProductPackageInventory 获取商品包库存
	// productPackageUuid: 商品包的UUID
	// opts: 可选参数，使用 WithStrategy 设置策略
	// 返回: 库存数量（float64），根据策略计算该商品包下所有BOM库存的最小值或最大值
	GetProductPackageInventory(ctx context.Context, productPackageUuid uint64, opts ...func(option *GetProductPackageInventoryOption)) (float64, error)
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
		NewNonBomCardProductInventoryStrategy(),
		NewMaxProductPackageInventoryStrategy(), // 默认使用最大值策略
	)
}

// NewProductInventoryDomainServiceWithStrategies 创建商品库存领域服务（支持策略注入，用于测试）
func NewProductInventoryDomainServiceWithStrategies(
	productBomRepo repository.IProductBomRepository,
	productPackageRepo repository.IProductPackageRepository,
	bomCardStrategy IInventoryStrategy,
	nonBomCardStrategy IInventoryStrategy,
	packageInventoryStrategy IProductPackageInventoryStrategy,
) IProductInventoryDomainService {
	strategies := make(map[string]IInventoryStrategy)
	strategies["bom_card"] = bomCardStrategy
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

	// 4. 遍历每个BOM，收集库存
	inventories := make([]float64, 0)
	for _, bomInterface := range productBomInterfaces {
		productBom, ok := bomInterface.(*model.ProductBom)
		if !ok {
			// 记录错误但继续处理其他BOM
			continue
		}

		inventory, err := s.GetProductInventory(ctx, productBom.Uuid)
		if err != nil {
			// 记录错误但继续处理其他BOM
			continue
		}

		inventories = append(inventories, inventory)
	}

	// 5. 如果没有有效的库存，返回错误
	if len(inventories) == 0 {
		return 0, errors.New("无法计算商品包库存：所有BOM库存查询失败")
	}

	// 6. 使用策略计算商品包库存
	// 如果未提供策略，使用默认策略
	strategy := option.Strategy
	if strategy == nil {
		strategy = s.defaultPackageInventoryStrategy
	}

	return strategy.CalculatePackageInventory(ctx, inventories)
}
