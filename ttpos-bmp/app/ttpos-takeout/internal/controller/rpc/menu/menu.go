package menu

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
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
// 参数：merchant_id、item_id 必填，其他字段可选
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) UpdateMenuItem(ctx context.Context, req *api.UpdateMenuItemReq) (*takeout.ApiResponse, error) {
	// 1. 参数校验
	if req.MerchantId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgMerchantIDEmpty,
		}, nil
	}
	if req.ItemId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgItemIDEmpty,
		}, nil
	}

	// 2. Proto → DTO 转换
	updateReq := &grabDto.UpdateMenuItemReq{
		MerchantID: req.MerchantId,
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
		g.Log().Errorf(ctx, "[Menu] UpdateMenuItem failed: merchantID=%s, itemID=%s, error=%v",
			req.MerchantId, req.ItemId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "更新菜单项失败: " + err.Error(),
		}, nil
	}

	// 4. DTO → Proto 响应转换
	resp := &api.UpdateMenuItemResp{
		MerchantId: req.MerchantId,
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

	g.Log().Infof(ctx, "[Menu] UpdateMenuItem success: merchantID=%s, itemID=%s",
		req.MerchantId, req.ItemId)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// UpdateMenuModifier 更新菜单修饰符
// 参数：merchant_id、modifier_id、modifier_name 必填，其他字段可选
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) UpdateMenuModifier(ctx context.Context, req *api.UpdateMenuModifierReq) (*takeout.ApiResponse, error) {
	// 1. 参数校验
	if req.MerchantId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgMerchantIDEmpty,
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

	// 2. Proto → DTO 转换
	updateReq := &grabDto.UpdateMenuModifierReq{
		MerchantID:   req.MerchantId,
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

	// 3. 调用 Service 层
	if err := service.Grab().UpdateMenuModifier(ctx, updateReq); err != nil {
		g.Log().Errorf(ctx, "[Menu] UpdateMenuModifier failed: merchantID=%s, modifierID=%s, error=%v",
			req.MerchantId, req.ModifierId, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "更新菜单修饰符失败: " + err.Error(),
		}, nil
	}

	// 4. DTO → Proto 响应转换
	resp := &api.UpdateMenuModifierResp{
		MerchantId: req.MerchantId,
		RecordId:   req.ModifierId,
		RecordType: string(grabDto.MenuItemUpdateFieldModifier),
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

	g.Log().Infof(ctx, "[Menu] UpdateMenuModifier success: merchantID=%s, modifierID=%s",
		req.MerchantId, req.ModifierId)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// BatchUpdateMenu 批量更新菜单
// 参数：merchant_id、field、menu_entities 必填
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) BatchUpdateMenu(ctx context.Context, req *api.BatchUpdateMenuReq) (*takeout.ApiResponse, error) {
	// 1. 参数校验
	if req.MerchantId == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: consts.MsgMerchantIDEmpty,
		}, nil
	}
	if req.Field == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "field 不能为空",
		}, nil
	}
	if req.Field != grabDto.MenuItemUpdateFieldItem && req.Field != grabDto.MenuItemUpdateFieldModifier {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "field 必须是 ITEM 或 MODIFIER",
		}, nil
	}
	if len(req.MenuEntities) == 0 {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "menu_entities 不能为空",
		}, nil
	}
	if len(req.MenuEntities) > 100 {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "menu_entities 数量不能超过 100 个",
		}, nil
	}

	// 2. Proto → DTO 转换
	dtoReq := &grabDto.BatchUpdateMenuReq{
		MerchantID:   req.MerchantId,
		Field:        req.Field,
		MenuEntities: make([]grabDto.MenuEntity, 0, len(req.MenuEntities)),
	}

	// 转换 MenuEntities
	for _, protoEntity := range req.MenuEntities {
		dtoEntity := grabDto.MenuEntity{
			ID: protoEntity.Id,
		}

		// 处理可选字段 - Price
		if protoEntity.Price != nil {
			price := *protoEntity.Price
			dtoEntity.Price = &price
		}

		// 处理可选字段 - AvailableStatus
		if protoEntity.AvailableStatus != nil {
			dtoEntity.AvailableStatus = *protoEntity.AvailableStatus
		}

		// 处理可选字段 - MaxStock（仅商品支持）
		if protoEntity.MaxStock != nil {
			stock := *protoEntity.MaxStock
			dtoEntity.MaxStock = &stock
		}

		// 转换高级定价配置
		for _, ap := range protoEntity.AdvancedPricings {
			dtoEntity.AdvancedPricings = append(dtoEntity.AdvancedPricings,
				grabDto.UpdateAdvancedPricingReq{
					Key:   ap.Key,
					Price: ap.Price,
				})
		}

		// 转换购买能力配置（仅商品支持）
		for _, p := range protoEntity.Purchasabilities {
			dtoEntity.Purchasabilities = append(dtoEntity.Purchasabilities,
				grabDto.UpdatePurchasabilityReq{
					Key:         p.Key,
					Purchasable: p.Purchasable,
				})
		}

		dtoReq.MenuEntities = append(dtoReq.MenuEntities, dtoEntity)
	}

	// 3. 调用 Service 层
	dtoResp, err := service.Grab().BatchUpdateMenuItems(ctx, dtoReq)
	if err != nil {
		g.Log().Errorf(ctx, "[Menu] BatchUpdateMenu failed: merchantID=%s, field=%s, count=%d, error=%v",
			req.MerchantId, req.Field, len(req.MenuEntities), err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "批量更新菜单失败: " + err.Error(),
		}, nil
	}

	// 4. DTO → Proto 响应转换
	protoResp := &api.BatchUpdateMenuResp{
		MerchantId: dtoResp.MerchantID,
		Status:     dtoResp.Status,
	}

	// 转换错误列表
	if len(dtoResp.Errors) > 0 {
		protoResp.Errors = make([]*api.MenuEntityError, 0, len(dtoResp.Errors))
		for _, dtoErr := range dtoResp.Errors {
			protoResp.Errors = append(protoResp.Errors, &api.MenuEntityError{
				Id:           dtoErr.ID,
				ErrorCode:    dtoErr.ErrorCode,
				ErrorMessage: dtoErr.ErrorMessage,
			})
		}
	}

	// 5. 包装为 ApiResponse
	dataAny, err := anypb.New(protoResp)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "[Menu] BatchUpdateMenu success: merchantID=%s, field=%s, status=%s, count=%d, errorCount=%d",
		req.MerchantId, req.Field, dtoResp.Status, len(req.MenuEntities), len(dtoResp.Errors))
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
