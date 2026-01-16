package lineman

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	lineman_dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// IMenuStatusClient 菜单状态客户端接口（用于测试 Mock）
type IMenuStatusClient interface {
	UpdateMenuStatusWithRetry(ctx context.Context, storeId string, req *lineman_dto.MenuStatusUpdateReq) (*lineman_dto.MenuStatusUpdateResp, error)
}

// MapStatusToLineman 将 TTPOS 状态映射为 Lineman 状态
// TTPOS 状态 -> Lineman 状态:
// AVAILABLE -> AVAILABLE
// UNAVAILABLE -> SUSPENDED
// SOLD_OUT_TODAY -> SOLD_OUT_TODAY
// UNAVAILABLEHIDE -> 不支持（返回错误）
//
// 参考: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633
func MapStatusToLineman(ttposStatus string) (string, error) {
	switch ttposStatus {
	case "AVAILABLE":
		return "AVAILABLE", nil
	case "UNAVAILABLE":
		return "SUSPENDED", nil
	case "SOLD_OUT_TODAY":
		return "SOLD_OUT_TODAY", nil
	case "UNAVAILABLEHIDE":
		return "", gerror.New("Lineman 平台不支持 UNAVAILABLEHIDE 状态")
	case "":
		return "", gerror.New("available_status 不能为空")
	default:
		return "", gerror.Newf("不支持的状态: %s", ttposStatus)
	}
}

// UpdateMenuItemStatus 更新菜单商品状态（单个商品）
//
// 参数:
//   - ctx: 上下文
//   - shopUuid: 店铺 UUID（内部会查询对应的 Lineman storeId）
//   - itemId: 商品 ID (partner item id)
//   - menuStatus: Lineman 状态 (AVAILABLE, SUSPENDED, SOLD_OUT_TODAY)
//
// 返回:
//   - error: 错误信息
func (s *sLineman) UpdateMenuItemStatus(ctx context.Context, shopUuid string, itemId string, menuStatus string) error {
	// 1. 参数校验
	if shopUuid == "" {
		return gerror.New("shopUuid 不能为空")
	}
	if itemId == "" {
		return gerror.New("itemId 不能为空")
	}
	if menuStatus == "" {
		return gerror.New("menuStatus 不能为空")
	}

	// 2. 构造请求
	req := &lineman_dto.MenuStatusUpdateReq{
		MenuItems: []lineman_dto.MenuItemStatus{
			{
				ID:         itemId,
				MenuStatus: menuStatus,
			},
		},
	}

	// 3. 调用 Lineman Client（shopUuid 对应 Lineman 的 storeId）
	resp, err := s.menuStatusClient.UpdateMenuStatusWithRetry(ctx, shopUuid, req)
	if err != nil {
		return gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 检查响应
	if resp.Status != "ok" {
		return gerror.Newf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
	}

	return nil
}

// BatchUpdateMenuItems 批量更新 Lineman 菜单项
//
// 参数:
//   - ctx: 上下文
//   - req: 批量更新请求（Proto 类型）
//
// 返回:
//   - *api.BatchUpdateMenuResp: 更新结果
//   - error: 错误信息
func (s *sLineman) BatchUpdateMenuItems(ctx context.Context, req *api.BatchUpdateMenuReq) (*api.BatchUpdateMenuResp, error) {
	g.Log().Infof(ctx, "[Lineman] 批量更新菜单: shopUUID=%s, field=%s, count=%d",
		req.ShopUuid, req.Field, len(req.MenuEntities))

	// 1. 从 shop_provider_cfg 获取 Lineman storeId
	// 注意：在 Lineman 场景下，shopUuid 本身就是 storeId
	storeId := req.ShopUuid

	// 2. 参数校验
	if req.Field == "" {
		return nil, gerror.New("field 不能为空")
	}
	if req.Field != "ITEM" {
		return nil, gerror.Newf("Lineman 平台仅支持 field=ITEM，不支持: %s", req.Field)
	}
	if len(req.MenuEntities) == 0 {
		return nil, gerror.New("menu_entities 不能为空")
	}
	if len(req.MenuEntities) > 100 {
		return nil, gerror.Newf("menu_entities 数量不能超过 100 个，当前: %d", len(req.MenuEntities))
	}

	// 3. 校验字段支持（Lineman 仅支持 available_status）
	if err := s.validateLinemanBatchRequest(req); err != nil {
		return nil, err
	}

	// 4. Proto → Lineman DTO 转换
	linemanReq, err := convertProtoToLinemanDTO(ctx, req)
	if err != nil {
		return nil, gerror.Wrap(err, "转换请求参数失败")
	}

	// 5. 调用 Lineman API
	resp, err := s.menuStatusClient.UpdateMenuStatusWithRetry(ctx, storeId, linemanReq)
	if err != nil {
		g.Log().Errorf(ctx, "[Lineman] 批量更新菜单 API 调用失败: storeId=%s, error=%v",
			storeId, err)
		return nil, gerror.Wrap(err, "调用 Lineman API 失败")
	}

	// 6. 检查响应
	if resp.Status != "ok" {
		g.Log().Errorf(ctx, "[Lineman] 批量更新菜单失败: storeId=%s, code=%s, message=%s",
			storeId, resp.Code, resp.Message)
		return nil, gerror.Newf("Lineman API 返回错误: %s - %s", resp.Code, resp.Message)
	}

	// 7. 构造成功响应
	protoResp := &api.BatchUpdateMenuResp{
		ShopUuid: req.ShopUuid,
		Status:   "success",
		Errors:   make([]*api.MenuEntityError, 0), // Lineman API 不返回部分失败，成功即全部成功
	}

	g.Log().Infof(ctx, "[Lineman] 批量更新菜单完成: storeId=%s, count=%d",
		storeId, len(req.MenuEntities))

	return protoResp, nil
}

// validateLinemanBatchRequest 校验 Lineman 批量请求（仅允许 available_status）
func (s *sLineman) validateLinemanBatchRequest(req *api.BatchUpdateMenuReq) error {
	for i, entity := range req.MenuEntities {
		// 检查是否有不支持的字段
		if entity.Price != nil {
			return gerror.Newf("Lineman 平台仅支持更新 available_status，不支持 price 字段 (entity[%d])", i)
		}
		if entity.MaxStock != nil {
			return gerror.Newf("Lineman 平台仅支持更新 available_status，不支持 max_stock 字段 (entity[%d])", i)
		}
		if len(entity.AdvancedPricings) > 0 {
			return gerror.Newf("Lineman 平台仅支持更新 available_status，不支持 advanced_pricings 字段 (entity[%d])", i)
		}
		if len(entity.Purchasabilities) > 0 {
			return gerror.Newf("Lineman 平台仅支持更新 available_status，不支持 purchasabilities 字段 (entity[%d])", i)
		}

		// available_status 必须提供
		if entity.AvailableStatus == nil {
			return gerror.Newf("Lineman 平台 available_status 字段为必填 (entity[%d])", i)
		}
	}

	return nil
}

// convertProtoToLinemanDTO 转换 Proto BatchUpdateMenuReq 到 Lineman DTO
func convertProtoToLinemanDTO(ctx context.Context, protoReq *api.BatchUpdateMenuReq) (*lineman_dto.MenuStatusUpdateReq, error) {
	linemanReq := &lineman_dto.MenuStatusUpdateReq{
		MenuItems: make([]lineman_dto.MenuItemStatus, 0, len(protoReq.MenuEntities)),
	}

	// 转换 MenuEntities
	for _, protoEntity := range protoReq.MenuEntities {
		// 映射状态到 Lineman
		linemanStatus, err := MapStatusToLineman(*protoEntity.AvailableStatus)
		if err != nil {
			return nil, gerror.Wrapf(err, "映射状态失败 (entity id=%s)", protoEntity.Id)
		}

		linemanReq.MenuItems = append(linemanReq.MenuItems, lineman_dto.MenuItemStatus{
			ID:         protoEntity.Id,
			MenuStatus: linemanStatus,
		})
	}

	return linemanReq, nil
}
