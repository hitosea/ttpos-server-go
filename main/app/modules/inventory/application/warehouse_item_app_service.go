package inventory

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	domainService "ttpos-server-go/app/modules/inventory/domain/service"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// WarehouseItemAppService 库存物品应用服务
type WarehouseItemAppService struct {
	warehouseItemRepo repository.IWarehouseItemRepository
	domainService     domainService.IWarehouseItemDomainService
	dbm               *database.DBManager
}

// NewWarehouseItemAppService 创建库存物品应用服务
func NewWarehouseItemAppService(
	warehouseItemRepo repository.IWarehouseItemRepository,
	domainService domainService.IWarehouseItemDomainService,
	dbm *database.DBManager,
) *WarehouseItemAppService {
	return &WarehouseItemAppService{
		warehouseItemRepo: warehouseItemRepo,
		domainService:     domainService,
		dbm:               dbm,
	}
}

// GetMaterialStockRequest 查询物品库存请求
type GetMaterialStockRequest struct {
	MaterialUuid uint64 // 物品UUID
}

// GetMaterialStockResponse 查询物品库存响应
type GetMaterialStockResponse struct {
	MaterialUuid   uint64  // 物品UUID
	TotalStock     float64 // 总库存（普通仓库）
	AvailableStock float64 // 可用库存
}

// GetMaterialStock 查询物品总库存
func (s *WarehouseItemAppService) GetMaterialStock(ctx context.Context, materialUuid uint64) (*GetMaterialStockResponse, error) {
	totalStock, err := s.domainService.GetMaterialStock(ctx, materialUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询物品库存失败")
	}

	return &GetMaterialStockResponse{
		MaterialUuid:   materialUuid,
		TotalStock:     totalStock,
		AvailableStock: totalStock, // 简化处理，可扩展为计算预留库存后的可用量
	}, nil
}

// GetMaterialStockBatchRequest 批量查询物品库存请求
type GetMaterialStockBatchRequest struct {
	MaterialUuids []uint64 // 物品UUID列表
}

// GetMaterialStockBatchResponse 批量查询物品库存响应
type GetMaterialStockBatchResponse struct {
	StockMap map[uint64]float64 // map[物品UUID]总库存
}

// GetMaterialStockBatch 批量查询物品总库存
func (s *WarehouseItemAppService) GetMaterialStockBatch(ctx context.Context, materialUuids []uint64) (*GetMaterialStockBatchResponse, error) {
	stockMap, err := s.domainService.GetMaterialStockBatch(ctx, materialUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询物品库存失败")
	}

	return &GetMaterialStockBatchResponse{
		StockMap: stockMap,
	}, nil
}

// MaterialStockDetailItem 物品库存详情项
type MaterialStockDetailItem struct {
	WarehouseUuid  uint64  // 仓库UUID
	WarehouseName  string  // 仓库名称
	WarehouseCode  string  // 仓库编码
	WarehouseType  string  // 仓库类型
	Stock          float64 // 库存数量
	ReservedStock  float64 // 预留库存
	AvailableStock float64 // 可用库存
}

// GetMaterialStockDetailResponse 物品库存详情响应
type GetMaterialStockDetailResponse struct {
	MaterialUuid uint64                    // 物品UUID
	TotalStock   float64                   // 总库存
	Details      []MaterialStockDetailItem // 各仓库库存详情
}

// GetMaterialStockDetail 查询物品在各仓库的库存详情
func (s *WarehouseItemAppService) GetMaterialStockDetail(ctx context.Context, materialUuid uint64) (*GetMaterialStockDetailResponse, error) {
	stockList, err := s.domainService.GetMaterialStockByWarehouse(ctx, materialUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询物品库存详情失败")
	}

	var totalStock float64
	details := make([]MaterialStockDetailItem, 0, len(stockList))
	for _, stock := range stockList {
		totalStock += stock.Stock
		details = append(details, MaterialStockDetailItem{
			WarehouseUuid:  stock.WarehouseUuid,
			WarehouseName:  stock.WarehouseName,
			WarehouseCode:  stock.WarehouseCode,
			WarehouseType:  stock.WarehouseType,
			Stock:          stock.Stock,
			ReservedStock:  stock.ReservedStock,
			AvailableStock: stock.Stock - stock.ReservedStock,
		})
	}

	return &GetMaterialStockDetailResponse{
		MaterialUuid: materialUuid,
		TotalStock:   totalStock,
		Details:      details,
	}, nil
}

// WarehouseItemInfo 仓库物品信息
type WarehouseItemInfo struct {
	Uuid           uint64  // 库存记录UUID
	WarehouseUuid  uint64  // 仓库UUID
	MaterialUuid   uint64  // 物品UUID
	MaterialCode   string  // 物品编码
	Stock          float64 // 库存数量
	ReservedStock  float64 // 预留库存
	AvailableStock float64 // 可用库存
	Valuation      float64 // 估值单价
	TotalValue     float64 // 库存总价值
}

// GetWarehouseItemListRequest 获取仓库物品列表请求
type GetWarehouseItemListRequest struct {
	WarehouseUuid uint64 // 仓库UUID
	Keyword       string // 关键字搜索
	HasStock      bool   // 是否只显示有库存的
	PageNo        int    // 页码
	PageSize      int    // 每页数量
}

// GetWarehouseItemListResponse 获取仓库物品列表响应
type GetWarehouseItemListResponse struct {
	List []WarehouseItemInfo // 物品列表
	Meta dto.PageResponse    // 分页信息
}

// GetWarehouseItemList 获取仓库物品列表
func (s *WarehouseItemAppService) GetWarehouseItemList(ctx context.Context, listReq GetWarehouseItemListRequest) (*GetWarehouseItemListResponse, error) {
	// 构建查询规格
	spec := repository.NewWarehouseItemQuerySpec()
	if listReq.WarehouseUuid > 0 {
		spec = spec.WithWarehouseUuid(listReq.WarehouseUuid)
	}
	if listReq.Keyword != "" {
		spec = spec.WithKeyword(listReq.Keyword)
	}
	if listReq.HasStock {
		spec = spec.WithHasStock(true)
	}

	// 查询库存列表
	items, total, err := s.warehouseItemRepo.FindWithPagination(ctx, spec, listReq.PageNo, listReq.PageSize)
	if err != nil {
		return nil, errors.WithMessage(err, "查询仓库物品列表失败")
	}

	// 转换为响应
	list := make([]WarehouseItemInfo, 0, len(items))
	for _, item := range items {
		list = append(list, WarehouseItemInfo{
			Uuid:           item.Uuid(),
			WarehouseUuid:  item.WarehouseUuid(),
			MaterialUuid:   item.MaterialUuid(),
			MaterialCode:   item.MaterialCode(),
			Stock:          item.Stock().Value(),
			ReservedStock:  item.ReservedStock().Value(),
			AvailableStock: item.AvailableStock().Value(),
			Valuation:      item.Valuation(),
			TotalValue:     item.TotalValue(),
		})
	}

	return &GetWarehouseItemListResponse{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
	}, nil
}

// AddStockRequest 入库请求
type AddStockRequest struct {
	WarehouseUuid uint64  // 仓库UUID
	MaterialUuid  uint64  // 物品UUID
	MaterialCode  string  // 物品编码
	Quantity      float64 // 入库数量
	Valuation     float64 // 估值单价
}

// AddStock 入库
func (s *WarehouseItemAppService) AddStock(ctx context.Context, addReq AddStockRequest) error {
	return s.domainService.AddStock(ctx, addReq.WarehouseUuid, addReq.MaterialUuid, addReq.MaterialCode, addReq.Quantity, addReq.Valuation)
}

// ReduceStockRequest 出库请求
type ReduceStockRequest struct {
	WarehouseUuid uint64  // 仓库UUID
	MaterialUuid  uint64  // 物品UUID
	Quantity      float64 // 出库数量
}

// ReduceStock 出库
func (s *WarehouseItemAppService) ReduceStock(ctx context.Context, reduceReq ReduceStockRequest) error {
	return s.domainService.ConsumeStock(ctx, reduceReq.WarehouseUuid, reduceReq.MaterialUuid, reduceReq.Quantity)
}

// TransferStockRequest 调拨请求
type TransferStockRequest struct {
	FromWarehouseUuid uint64  // 源仓库UUID
	ToWarehouseUuid   uint64  // 目标仓库UUID
	MaterialUuid      uint64  // 物品UUID
	Quantity          float64 // 调拨数量
}

// TransferStock 仓库间调拨
func (s *WarehouseItemAppService) TransferStock(ctx context.Context, transferReq TransferStockRequest) error {
	return s.domainService.TransferStock(ctx, transferReq.FromWarehouseUuid, transferReq.ToWarehouseUuid, transferReq.MaterialUuid, transferReq.Quantity)
}
