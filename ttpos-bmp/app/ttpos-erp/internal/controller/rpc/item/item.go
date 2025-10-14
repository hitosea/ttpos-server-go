package item

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"

	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 物品服务控制器
type Controller struct {
	item.UnimplementedItemServiceServer
}

// Register 注册物品相关服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	item.RegisterItemServiceServer(s.Server, &Controller{})
	item.RegisterItemGroupServiceServer(s.Server, &GroupController{})
}

// GetItemList 获取物品列表
// 参数：ctx 上下文，req 获取物品列表请求
// 返回：响应信息和错误
func (c *Controller) GetItemList(ctx context.Context, req *item.GetItemListReq) (*api.ResponseInfo, error) {

	// 调用服务层获取数据
	dataList, err := service.Item().GetItemList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品列表成功", dataList), nil
}

// SaveItem 保存物品信息
// 参数：ctx 上下文，req 物品信息
// 返回：响应信息和错误
func (c *Controller) SaveItem(ctx context.Context, req *item.ItemInfo) (*api.ResponseInfo, error) {
	// 参数验证
	if req.ItemName == "" {
		return rpc.ApiError("物品名称不能为空"), nil
	}
	if req.StockUom == "" {
		return rpc.ApiError("库存单位不能为空"), nil
	}
	if len(req.TemplateItemCode) > 0 && req.ItemSpecification == "" {
		return rpc.ApiError("多规格商品时，物品规格不能为空"), nil
	}

	// 调用服务层保存数据
	savedItem, err := service.Item().SaveItem(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("保存物品成功", savedItem), nil
}

// GetUomList 获取单位列表
// 参数：ctx 上下文，req 获取单位列表请求
// 返回：响应信息和错误
func (c *Controller) GetUomList(ctx context.Context, req *item.GetUomListReq) (*api.ResponseInfo, error) {
	// 调用服务层获取数据
	dataList, err := service.Uom().GetUomList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取单位列表成功", dataList), nil
}

// GetUom 根据单位名称获取单个单位信息
// 参数：ctx 上下文，req 包含单位名称
// 返回：响应信息和错误
func (c *Controller) GetUom(ctx context.Context, req *item.GetUomReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req.UomName == "" {
		return rpc.ApiError("单位名称不能为空"), nil
	}

	// 调用服务层获取数据
	uomInfo, err := service.Uom().GetUom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换为 protobuf 响应格式
	resp := &item.GetUomResp{
		UomInfo: &item.UomInfo{
			UomName:           uomInfo.UomName,
			AliasName:         uomInfo.CustomAlias,
			Company:           uomInfo.CustomCompany,
			Branch:            uomInfo.CustomBranch,
			MustBeWholeNumber: uomInfo.MustBeWholeNumber == 1,
		},
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取单位信息成功", resp), nil
}

// SaveUom 保存单位信息
func (*Controller) SaveUom(ctx context.Context, req *item.UomInfo) (*api.ResponseInfo, error) {
	if req.UomName == "" {
		return rpc.ApiError("单位名称不能为空"), nil
	}

	// 调用服务层保存数据
	err := service.Uom().SaveUom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("保存单位成功"), nil
}

// GetAttributeList 获取属性列表
// 参数：ctx 上下文，req 获取属性列表请求
// 返回：响应信息和错误
func (c *Controller) GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (*api.ResponseInfo, error) {
	// 调用服务层获取数据
	dataList, err := service.Stock().GetAttributeList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取属性列表成功", dataList), nil
}

// SaveAttribute 保存属性信息
func (*Controller) SaveAttribute(ctx context.Context, req *item.AttributeInfo) (*api.ResponseInfo, error) {
	if req.AttributeName == "" {
		return rpc.ApiError("属性名称不能为空"), nil
	}

	// 调用服务层保存数据
	err := service.Stock().SaveAttribute(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccess("保存属性成功"), nil
}

// GetItemStock 获取物品库存信息
func (*Controller) GetItemStock(ctx context.Context, req *item.GetItemStockReq) (res *api.ResponseInfo, err error) {
	if req.CompanyAbbr == "" {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	//if req.Branch == "" {
	//	return rpc.ApiError("分支机构不能为空"), nil
	//}

	// 调用服务层获取数据
	dataList, err := service.Item().GetItemStock(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品库存成功", dataList), nil
}

// GetItem 根据物品编码获取单个物品信息
// 参数：
//
//	ctx - 请求上下文，包含认证信息和调用链追踪
//	req - 获取物品请求，包含物品编码和可选的公司、分支信息
//
// 返回：
//
//	封装了物品详细信息的响应对象和可能的错误
func (c *Controller) GetItem(ctx context.Context, req *item.GetItemReq) (*api.ResponseInfo, error) {
	// 参数验证：确保物品编码不为空
	if req.ItemCode == "" {
		return rpc.ApiError("物品编码不能为空"), nil
	}

	// 调用服务层获取物品的完整信息
	itemInfo, err := service.Item().GetItem(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换单位列表数据结构：从内部模型转换为protobuf响应模型
	uomDetails := make([]*item.UomDetail, 0, len(itemInfo.Uoms))
	for _, uom := range itemInfo.Uoms {
		uomDetails = append(uomDetails, &item.UomDetail{
			Uom:              uom.Uom,              // 单位名称
			ConversionFactor: uom.ConversionFactor, // 转换因子
		})
	}

	// 转换属性列表数据结构：从内部模型转换为protobuf响应模型
	attrList := make([]*item.ItemAttribute, 0, len(itemInfo.Attributes))
	for _, attr := range itemInfo.Attributes {
		attrList = append(attrList, &item.ItemAttribute{
			AttributeName:  attr.CustomAttributeName,  // 属性名称
			AttributeValue: attr.CustomAttributeValue, // 属性值
		})
	}

	// 构建最终的物品信息响应对象
	// 将内部itemInfo模型的各个字段映射到protobuf响应模型中
	respItem := &item.ItemInfo{
		ItemName:           itemInfo.ItemName,                 // 物品名称
		StockUom:           itemInfo.StockUom,                 // 库存单位
		ItemCode:           itemInfo.ItemCode,                 // 物品编码
		ValuationRate:      itemInfo.ValuationRate,            // 估价率
		IsStockItem:        itemInfo.IsStockItem == 1,         // 是否为库存物品（数值型转布尔型）
		Branch:             itemInfo.CustomBranch,             // 分支机构
		Company:            itemInfo.CustomCompany,            // 公司
		ItemSpecification:  itemInfo.CustomSpecification,      // 物品规格
		Disabled:           itemInfo.Disabled == 1,            // 是否禁用（数值型转布尔型）
		Classification:     itemInfo.CustomClassification,     // 分类
		ClassificationCode: itemInfo.CustomClassificationCode, // 分类编码
		InternalCode:       itemInfo.CustomInternalCode,       // 内部编码
		PurchaseUom:        itemInfo.PurchaseUom,              // 采购单位
		Uoms:               uomDetails,                        // 单位列表（已转换）
		OpeningStock:       itemInfo.OpeningStock,             // 期初库存
		Attributes:         attrList,                          // 属性列表（已转换）
	}

	// 返回成功响应，包含转换后的物品详细信息
	return rpc.ApiSuccessWithData("获取物品信息成功", respItem), nil
}

// SavePosAttribute 保存 POS 系统中的属性物品
// 参数：ctx 上下文，req 保存属性物品请求
// 返回：响应信息和错误
func (c *Controller) SavePosAttribute(ctx context.Context, req *item.SavePosAttributeReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req.Item == nil {
		return rpc.ApiError("属性物品信息不能为空"), nil
	}
	if req.Item.ItemName == "" {
		return rpc.ApiError("属性物品名称不能为空"), nil
	}

	// 调用服务层保存数据
	savedItem, err := service.Item().SavePosAttribute(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("保存属性物品成功", savedItem), nil
}

// SavePosAddon 保存 POS 系统中的加料物品
// 参数：ctx 上下文，req 保存加料物品请求
// 返回：响应信息和错误
func (c *Controller) SavePosAddon(ctx context.Context, req *item.SavePosAddonReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req.Item == nil {
		return rpc.ApiError("加料物品信息不能为空"), nil
	}
	if req.Item.ItemName == "" {
		return rpc.ApiError("加料物品名称不能为空"), nil
	}

	// 调用服务层保存数据
	savedItem, err := service.Item().SavePosAddon(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("保存加料物品成功", savedItem), nil
}

func (*Controller) DeleteItem(ctx context.Context, req *item.DeleteItemReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req.ItemCode == "" {
		return rpc.ApiError("物品编码不能为空"), nil
	}
	// 调用服务层删除数据
	resp, err := service.Item().DeleteItem(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	// 返回成功响应
	return rpc.ApiSuccessWithData("删除物品成功", resp), nil
}

func (*Controller) DeleteUom(ctx context.Context, req *item.DeleteUomReq) (*api.ResponseInfo, error) {
	// 参数验证
	if req.Uom == "" {
		return rpc.ApiError("单位编码不能为空"), nil
	}
	// 调用服务层删除数据
	resp, err := service.Uom().DeleteUom(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	// 返回成功响应
	return rpc.ApiSuccessWithData("删除单位成功", resp), nil
}

func (*Controller) CreateSingleVariantItem(ctx context.Context, req *item.CreateSingleVariantItemReq) (*api.ResponseInfo, error) {
	// 获取基准商品
	itemInfo, err := service.Item().GetItem(ctx, &item.GetItemReq{
		ItemCode: req.VariantsOf,
	})
	if err != nil {
		return rpc.ApiError("创建单规格商品失败，获取基准商品失败"), nil
	}
	variant := make(map[string]string, len(req.Attributes))
	for _, attr := range req.Attributes {
		// 将每种属性名称和属性值的组合添加到itemVarList中
		variant[attr.AttributeName] = attr.AttributeValue
	}
	multiSpecItemCode, err := service.Item().CreateSingleVariantItem(ctx, &erp.CreateSingleVariantItemReq{
		TemplateItem: req.VariantsOf,
		Args:         variant,
		InternalCode: req.InternalCode,
	}, &item.ItemInfo{
		ClassificationCode: itemInfo.CustomClassificationCode,
		Classification:     itemInfo.CustomClassification,
		Company:            itemInfo.CustomCompany,
		Branch:             itemInfo.CustomBranch,
	})
	if err != nil {
		return rpc.ApiError("创建单规格商品失败"), nil
	}
	return rpc.ApiSuccessWithData("创建成功", &item.CreateSingleVariantItemResp{
		ItemCode: multiSpecItemCode,
	}), nil
}
