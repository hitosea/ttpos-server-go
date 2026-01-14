package menu

import (
	"context"
	"fmt"

	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/app/ttpos-takeout/utility"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	api.UnimplementedMenuServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterMenuServiceServer(s.Server, &Controller{})
}

func (c *Controller) GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*takeout.ApiResponse, error) {
	res, err := service.ChannelMenu().GetMenuSnapshot(ctx, req)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 将 res 转换为 anypb.Any
	dataAny, err := anypb.New(res)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Debugf(ctx, "查询菜单快照成功:%+v", res)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

func (c *Controller) SaveMenuSnapshot(ctx context.Context, req *api.SaveMenuSnapshotReq) (*takeout.ApiResponse, error) {
	res, err := service.ChannelMenu().SaveMenuSnapshot(ctx, req)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// SaveMenuSnapshotResp 是空结构体，可以只返回 code 和 message
	// 如果需要，也可以将 res 转换为 anypb.Any
	var dataAny *anypb.Any
	if res != nil {
		dataAny, _ = anypb.New(res)
	}

	g.Log().Debugf(ctx, "保存菜单快照成功:%+v", res)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// UpdateMenuItem 更新菜单项
// 参数：shop_uuid、item_id 必填，其他字段可选
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) UpdateMenuItem(ctx context.Context, req *api.UpdateMenuItemReq) (*takeout.ApiResponse, error) {
	// 1. 基础参数校验
	if req.ShopUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "shop_uuid 不能为空",
		}, nil
	}
	if req.ItemId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgItemIDEmpty,
		}, nil
	}

	// 2. 获取 provider_name，默认为 grab
	providerName := "grab"
	if req.ProviderName != nil && *req.ProviderName != "" {
		providerName = *req.ProviderName
	}

	g.Log().Infof(ctx, "[Menu] UpdateMenuItem: provider=%s, shopUUID=%s, itemID=%s",
		providerName, req.ShopUuid, req.ItemId)

	// 3. 根据平台路由到对应处理逻辑
	switch providerName {
	case "grab":
		return c.handleGrabUpdate(ctx, req)
	case "lineman":
		return c.handleLinemanUpdate(ctx, req)
	default:
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "不支持的平台: " + providerName,
		}, nil
	}
}

// handleGrabUpdate Grab 平台菜单更新处理（保持现有逻辑）
func (c *Controller) handleGrabUpdate(ctx context.Context, req *api.UpdateMenuItemReq) (*takeout.ApiResponse, error) {
	// 1. 从 shop_provider_cfg 获取 Grab merchant_id
	merchantID, err := service.ShopProviderCfg().GetProviderMerchantID(ctx, req.ShopUuid, "grab")
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 2. Proto → DTO 转换（使用查询到的 merchantID）
	updateReq := &grabDto.UpdateMenuItemReq{
		MerchantID: merchantID,
		ItemID:     req.ItemId,
	}

	// 处理可选字段
	if req.Price != nil {
		price := *req.Price
		updateReq.Price = &price
	}
	if req.AvailableStatus != nil {
		updateReq.AvailableStatus = *req.AvailableStatus
	}
	if req.MaxStock != nil {
		stock := *req.MaxStock
		updateReq.MaxStock = &stock
	}

	// 转换高级定价配置
	for _, ap := range req.AdvancedPricings {
		updateReq.AdvancedPricings = append(updateReq.AdvancedPricings,
			grabDto.UpdateAdvancedPricingReq{
				Key:   ap.Key,
				Price: ap.Price,
			})
	}

	// 转换购买能力配置
	for _, p := range req.Purchasabilities {
		updateReq.Purchasabilities = append(updateReq.Purchasabilities,
			grabDto.UpdatePurchasabilityReq{
				Key:         p.Key,
				Purchasable: p.Purchasable,
			})
	}

	// 3. 调用 Service 层
	if err := service.Grab().UpdateMenuItem(ctx, updateReq); err != nil {
		g.Log().Errorf(ctx, "[Menu] UpdateMenuItem failed: shopUUID=%s, merchantID=%s, itemID=%s, error=%v",
			req.ShopUuid, merchantID, req.ItemId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "更新菜单项失败: " + err.Error(),
		}, nil
	}

	// 4. DTO → Proto 响应转换
	resp := &api.UpdateMenuItemResp{
		ShopUuid:   req.ShopUuid,
		RecordId:   req.ItemId,
		RecordType: string(grabDto.MenuItemUpdateFieldItem),
		// success, error_code 和 error_message 已移除，由 ApiResponse 统一处理
	}

	// 5. 包装为 ApiResponse
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] UpdateMenuItem success: shopUUID=%s, merchantID=%s, itemID=%s",
		req.ShopUuid, merchantID, req.ItemId)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// handleLinemanUpdate Lineman 平台菜单更新处理（新增）
func (c *Controller) handleLinemanUpdate(ctx context.Context, req *api.UpdateMenuItemReq) (*takeout.ApiResponse, error) {
	// 1. 从 shop_provider_cfg 获取 Lineman merchant_id (storeId)
	// merchantID, err := service.ShopProviderCfg().GetProviderMerchantID(ctx, req.ShopUuid, "lineman")
	// if err != nil {
	// 	return &takeout.ApiResponse{
	// 		Code:    string(consts.CodeServiceError),
	// 		Message: err.Error(),
	// 	}, nil
	// }

	// 2. Lineman 字段校验（仅允许 available_status）
	if err := c.validateLinemanRequest(req); err != nil {
		g.Log().Warningf(ctx, "[Menu] Lineman 字段校验失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: err.Error(),
		}, nil
	}

	// 3. 状态映射
	if req.AvailableStatus == nil || *req.AvailableStatus == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "available_status 不能为空",
		}, nil
	}

	linemanStatus, err := c.mapToLinemanStatus(*req.AvailableStatus)
	if err != nil {
		g.Log().Warningf(ctx, "[Menu] 状态映射失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: err.Error(),
		}, nil
	}

	// 4. 调用 Lineman Service 层
	if err := service.Lineman().UpdateMenuItemStatus(ctx, req.ShopUuid, req.ItemId, linemanStatus); err != nil {
		g.Log().Errorf(ctx, "[Menu] Lineman UpdateMenuItem failed: shopUUID=%s, itemID=%s, error=%v",
			req.ShopUuid, req.ItemId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "更新菜单项失败: " + err.Error(),
		}, nil
	}

	// 5. 构造响应
	resp := &api.UpdateMenuItemResp{
		ShopUuid:   req.ShopUuid,
		RecordId:   req.ItemId,
		RecordType: "ITEM",
	}

	// 6. 包装为 ApiResponse
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] Lineman UpdateMenuItem success: shopUUID=%s, itemID=%s, status=%s",
		req.ShopUuid, req.ItemId, linemanStatus)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// validateLinemanRequest 校验 Lineman 请求（仅允许 available_status）
func (c *Controller) validateLinemanRequest(req *api.UpdateMenuItemReq) error {
	if req.Price != nil {
		return gerror.New("Lineman 平台仅支持更新 available_status 字段，不支持 price 字段")
	}
	if req.MaxStock != nil {
		return gerror.New("Lineman 平台仅支持更新 available_status 字段，不支持 max_stock 字段")
	}
	if len(req.AdvancedPricings) > 0 {
		return gerror.New("Lineman 平台仅支持更新 available_status 字段，不支持 advanced_pricings 字段")
	}
	if len(req.Purchasabilities) > 0 {
		return gerror.New("Lineman 平台仅支持更新 available_status 字段，不支持 purchasabilities 字段")
	}
	if req.AvailableStatus == nil || *req.AvailableStatus == "" {
		return gerror.New("available_status 字段为必填")
	}
	return nil
}

// mapToLinemanStatus 映射状态到 Lineman
func (c *Controller) mapToLinemanStatus(status string) (string, error) {
	switch status {
	case "AVAILABLE":
		return "AVAILABLE", nil
	case "UNAVAILABLE":
		return "SUSPENDED", nil
	case "SOLD_OUT_TODAY":
		return "SOLD_OUT_TODAY", nil
	case "UNAVAILABLEHIDE":
		return "", gerror.New("Lineman 平台不支持 UNAVAILABLEHIDE 状态")
	default:
		return "", gerror.New("不支持的状态: " + status)
	}
}

// UpdateMenuModifier 更新菜单修饰符（支持多平台）
// 参数：shop_uuid、modifier_id、modifier_name 必填，其他字段可选
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) UpdateMenuModifier(ctx context.Context, req *api.UpdateMenuModifierReq) (*takeout.ApiResponse, error) {
	// 1. 参数校验
	if req.ShopUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "shop_uuid 不能为空",
		}, nil
	}
	if req.ModifierId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgModifierIDEmpty,
		}, nil
	}
	if req.ModifierName == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgModifierNameEmpty,
		}, nil
	}

	// 2. 获取 provider_name（默认为 "grab"）
	providerName := "grab"
	if req.ProviderName != nil && *req.ProviderName != "" {
		providerName = *req.ProviderName
	}

	// 3. 根据平台路由
	switch providerName {
	case "grab":
		return c.handleGrabModifierUpdate(ctx, req)
	case "lineman":
		return c.handleLinemanModifierUpdate(ctx, req)
	default:
		g.Log().Errorf(ctx, "[Menu] UpdateMenuModifier unsupported provider: %s", providerName)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "不支持的平台: " + providerName,
		}, nil
	}
}

// handleGrabModifierUpdate 处理 Grab 修饰符更新（现有逻辑）
func (c *Controller) handleGrabModifierUpdate(ctx context.Context, req *api.UpdateMenuModifierReq) (*takeout.ApiResponse, error) {
	// 1. 从 shop_provider_cfg 获取 Grab merchant_id
	merchantID, err := service.ShopProviderCfg().GetProviderMerchantID(ctx, req.ShopUuid, "grab")
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 2. Proto → DTO 转换（使用查询到的 merchantID）
	updateReq := &grabDto.UpdateMenuModifierReq{
		MerchantID:   merchantID,
		ModifierID:   req.ModifierId,
		ModifierName: req.ModifierName,
	}

	// 处理可选字段
	if req.Price != nil {
		price := *req.Price
		updateReq.Price = &price
	}
	if req.AvailableStatus != nil {
		updateReq.AvailableStatus = *req.AvailableStatus
	}
	if req.IsFree != nil {
		isFree := *req.IsFree
		updateReq.IsFree = &isFree
	}

	// 转换高级定价配置
	for _, ap := range req.AdvancedPricings {
		updateReq.AdvancedPricings = append(updateReq.AdvancedPricings,
			grabDto.UpdateAdvancedPricingReq{
				Key:   ap.Key,
				Price: ap.Price,
			})
	}

	// 调用 Service 层
	if err := service.Grab().UpdateMenuModifier(ctx, updateReq); err != nil {
		g.Log().Errorf(ctx, "[Menu] Grab UpdateMenuModifier failed: shopUUID=%s, merchantID=%s, modifierID=%s, error=%v",
			req.ShopUuid, merchantID, req.ModifierId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "更新菜单修饰符失败: " + err.Error(),
		}, nil
	}

	// DTO → Proto 响应转换
	resp := &api.UpdateMenuModifierResp{
		ShopUuid:   req.ShopUuid,
		RecordId:   req.ModifierId,
		RecordType: string(grabDto.MenuItemUpdateFieldModifier),
	}

	// 包装为 ApiResponse
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] Grab UpdateMenuModifier success: shopUUID=%s, merchantID=%s, modifierID=%s",
		req.ShopUuid, merchantID, req.ModifierId)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// handleLinemanModifierUpdate 处理 Lineman 修饰符更新
func (c *Controller) handleLinemanModifierUpdate(ctx context.Context, req *api.UpdateMenuModifierReq) (*takeout.ApiResponse, error) {
	// 1. 字段校验
	if err := c.validateLinemanModifierRequest(req); err != nil {
		g.Log().Errorf(ctx, "[Menu] Lineman UpdateMenuModifier validation failed: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: err.Error(),
		}, nil
	}

	// 2. 状态映射（string → int）
	availableStatus := ""
	if req.AvailableStatus != nil {
		availableStatus = *req.AvailableStatus
	}
	linemanStatus, err := utility.MapStatusToLinemanModifier(availableStatus)
	if err != nil {
		g.Log().Errorf(ctx, "[Menu] Lineman status mapping failed: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "状态映射失败: " + err.Error(),
		}, nil
	}

	// 3. 调用 Lineman Service
	err = service.Lineman().UpdateModifierStatus(
		ctx,
		req.ShopUuid,   // storeId (Lineman 场景下 shopUuid 就是 storeId)
		req.ModifierId, // modifierId
		linemanStatus,  // status (int)
	)
	if err != nil {
		g.Log().Errorf(ctx, "[Menu] Lineman UpdateMenuModifier failed: shopUUID=%s, modifierID=%s, error=%v",
			req.ShopUuid, req.ModifierId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "更新 Lineman 修饰符状态失败: " + err.Error(),
		}, nil
	}

	// 4. DTO → Proto 响应转换
	resp := &api.UpdateMenuModifierResp{
		ShopUuid:   req.ShopUuid,
		RecordId:   req.ModifierId,
		RecordType: string(grabDto.MenuItemUpdateFieldModifier),
	}

	// 5. 包装为 ApiResponse
	dataAny, err := anypb.New(resp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] Lineman UpdateMenuModifier success: shopUUID=%s, modifierID=%s",
		req.ShopUuid, req.ModifierId)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// validateLinemanModifierRequest 校验 Lineman 请求字段
func (c *Controller) validateLinemanModifierRequest(req *api.UpdateMenuModifierReq) error {
	// Lineman 仅支持更新 available_status
	// 禁止包含：price, is_free, advanced_pricings

	// 检查 available_status 是否为空
	if req.AvailableStatus == nil || *req.AvailableStatus == "" {
		return gerror.New("Lineman 平台的 available_status 字段为必填")
	}

	// 检查禁止字段
	if req.Price != nil {
		return gerror.New("Lineman 平台不支持更新 price 字段，仅支持 available_status")
	}
	if req.IsFree != nil {
		return gerror.New("Lineman 平台不支持更新 is_free 字段，仅支持 available_status")
	}
	if len(req.AdvancedPricings) > 0 {
		return gerror.New("Lineman 平台不支持更新 advanced_pricings 字段，仅支持 available_status")
	}

	return nil
}

// BatchUpdateMenu 批量更新菜单
// 参数：shop_uuid、field、menu_entities 必填
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) BatchUpdateMenu(ctx context.Context, req *api.BatchUpdateMenuReq) (*takeout.ApiResponse, error) {
	// 1. 获取 provider_name，默认为 grab
	providerName := "grab"
	if req.ProviderName != nil && *req.ProviderName != "" {
		providerName = *req.ProviderName
	}

	g.Log().Infof(ctx, "[Menu] BatchUpdateMenu: provider=%s, shopUUID=%s, field=%s, count=%d",
		providerName, req.ShopUuid, req.Field, len(req.MenuEntities))

	// 2. 根据平台路由到对应处理逻辑
	switch providerName {
	case "grab":
		return c.handleGrabBatchUpdate(ctx, req)
	case "lineman":
		return c.handleLinemanBatchUpdate(ctx, req)
	default:
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: fmt.Sprintf("不支持的平台: %s", providerName),
		}, nil
	}
}

// handleGrabBatchUpdate Grab 平台批量更新处理
func (c *Controller) handleGrabBatchUpdate(ctx context.Context, req *api.BatchUpdateMenuReq) (*takeout.ApiResponse, error) {
	// 调用 Grab Service 层
	protoResp, err := service.Grab().BatchUpdateMenuItems(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "[Menu] Grab BatchUpdateMenu failed: shopUUID=%s, field=%s, count=%d, error=%v",
			req.ShopUuid, req.Field, len(req.MenuEntities), err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "批量更新菜单失败: " + err.Error(),
		}, nil
	}

	// 包装为 ApiResponse
	dataAny, err := anypb.New(protoResp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] Grab BatchUpdateMenu success: shopUUID=%s, field=%s, status=%s, count=%d, errorCount=%d",
		req.ShopUuid, req.Field, protoResp.Status, len(req.MenuEntities), len(protoResp.Errors))
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// handleLinemanBatchUpdate Lineman 平台批量更新处理
func (c *Controller) handleLinemanBatchUpdate(ctx context.Context, req *api.BatchUpdateMenuReq) (*takeout.ApiResponse, error) {
	// 调用 Lineman Service 层
	protoResp, err := service.Lineman().BatchUpdateMenuItems(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "[Menu] Lineman BatchUpdateMenu failed: shopUUID=%s, field=%s, count=%d, error=%v",
			req.ShopUuid, req.Field, len(req.MenuEntities), err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "批量更新菜单失败: " + err.Error(),
		}, nil
	}

	// 包装为 ApiResponse
	dataAny, err := anypb.New(protoResp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] Lineman BatchUpdateMenu success: shopUUID=%s, field=%s, status=%s, count=%d, errorCount=%d",
		req.ShopUuid, req.Field, protoResp.Status, len(req.MenuEntities), len(protoResp.Errors))
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// NotifyMenuUpdate 通知菜单更新（统一入口）
// 根据 provider_name 路由到对应平台的菜单同步服务
func (c *Controller) NotifyMenuUpdate(ctx context.Context, req *api.NotifyMenuUpdateReq) (*takeout.ApiResponse, error) {
	return service.Menu().NotifyMenuUpdate(ctx, req)
}
