package takeout

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

type ITakeout interface {
	// EstimatePrice 预估距离
	EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error)
	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error)
	// ConfirmOrder 商家确认订单
	ConfirmOrder(ctx context.Context, req *dto.ConfirmOrderInp) (res *api.ConfirmOrderResp, err error)
	// CancelOrder 取消订单
	CancelOrder(ctx context.Context, req *dto.CancelOrderInp) (res *api.CancelOrderResp, err error)
	// GetDriverInfo 获取司机信息
	GetDriverInfo(ctx context.Context, req *dto.GetDriverInfoInp) (res *api.GetDriverInfoResp, err error)
}

func GetService(name consts.ProviderName) ITakeout {
	switch name {
	case consts.ProviderSkootar:
		return service.Skootar()
	default:
		return service.Skootar()
	}
}

type sTakeout struct {
}

var Takeout = new(sTakeout)

// Get 根据 shop_order_uuid 查询订单（兼容旧接口，返回 SkootarJob DTO）
// 注意：此方法为兼容性方法，实际查询新表结构并转换为 DTO 格式
func (s *sTakeout) Get(ctx context.Context, shopOrderUuid string) (*dto.SkootarJob, error) {
	// 先尝试从旧表查询（兼容历史数据）
	var takeoutJob *entity.Job
	err := dao.Job.Ctx(ctx).Where(dao.Job.Columns().ShopRefNo, shopOrderUuid).Scan(&takeoutJob)
	if err == nil && takeoutJob != nil {
		// 转换 entity.Job 为 dto.SkootarJob
		return &dto.SkootarJob{
			Uuid:            takeoutJob.Uuid,
			ShopRefNo:       takeoutJob.ShopRefNo,
			TakeoutRefNo:    takeoutJob.TakeoutRefNo,
			ProviderName:    takeoutJob.ProviderName,
			JobStatus:       takeoutJob.JobStatus,
			SkootarId:       takeoutJob.SkootarId,
			SkootarName:     takeoutJob.SkootarName,
			SkootarPhone:    takeoutJob.SkootarPhone,
			SkootarImageUrl: takeoutJob.SkootarImageUrl,
			SkootarRating:   takeoutJob.SkootarRating,
		}, nil
	}

	// 从新表查询并转换为 DTO 格式
	orderWithDriver, err := s.GetWithDriver(ctx, shopOrderUuid)
	if err != nil {
		return nil, err
	}

	// 转换为 SkootarJob DTO（兼容性）
	return &dto.SkootarJob{
		Uuid:            orderWithDriver.OrderUuid,
		ShopRefNo:       shopOrderUuid,
		TakeoutRefNo:    orderWithDriver.PartnerOrderId,
		ProviderName:    orderWithDriver.ProviderName,
		JobStatus:       orderWithDriver.OrderStatus,
		SkootarId:       orderWithDriver.SkootarId,
		SkootarName:     orderWithDriver.SkootarName,
		SkootarPhone:    orderWithDriver.SkootarPhone,
		SkootarImageUrl: orderWithDriver.SkootarImageUrl,
		SkootarRating:   orderWithDriver.SkootarRating,
	}, nil
}

// GetWithDriver 根据 shop_order_uuid 查询订单及司机信息（新方法）
func (s *sTakeout) GetWithDriver(ctx context.Context, shopOrderUuid string) (*dto.OrderWithDriver, error) {
	// 从主表查询订单
	var order *entity.Order
	err := dao.Order.Ctx(ctx).Where(dao.Order.Columns().ShopRefNo, shopOrderUuid).Scan(&order)
	if err != nil {
		return nil, gerror.Wrap(err, "查询订单失败")
	}
	if order == nil {
		return nil, gerror.New("订单不存在")
	}

	result := &dto.OrderWithDriver{
		OrderUuid:      order.Uuid,
		ProviderName:   order.ProviderName,
		PartnerOrderId: order.PartnerOrderId,
		OrderStatus:    order.OrderStatus,
	}

	// 如果是 Skootar 订单，查询扩展表获取司机信息
	if order.ProviderName == string(consts.ProviderSkootar) {
		var orderSkootar *entity.OrderSkootar
		err = dao.OrderSkootar.Ctx(ctx).Where(dao.OrderSkootar.Columns().OrderUuid, order.Uuid).Scan(&orderSkootar)
		if err == nil && orderSkootar != nil {
			result.SkootarId = orderSkootar.SkootarId
			result.SkootarName = orderSkootar.SkootarName
			result.SkootarPhone = orderSkootar.SkootarPhone
			result.SkootarImageUrl = orderSkootar.SkootarImageUrl
			result.SkootarRating = orderSkootar.SkootarRating
		}
	}

	return result, nil
}

// GetMenuSnapshot 根据 request_id 查询菜单快照
func (s *sTakeout) GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*api.GetMenuSnapshotResp, error) {
	// 参数校验
	if req.ProviderName == "" {
		return nil, gerror.New("provider_name 不能为空")
	}
	if req.ShopUuid == "" {
		return nil, gerror.New("shop_uuid 不能为空")
	}
	if req.RequestId == "" {
		return nil, gerror.New("request_id 不能为空")
	}

	// 查询快照记录（使用字段名字符串，兼容迁移前后）
	record, err := dao.ChannelMenuSnapshot.Ctx(ctx).
		Where("request_id", req.RequestId).
		Where("provider_name", req.ProviderName).
		Where("shop_uuid", req.ShopUuid).
		Where("deleted_at IS NULL"). // 软删除过滤
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "查询菜单快照失败")
	}
	if record.IsEmpty() {
		return nil, gerror.New("菜单快照不存在")
	}

	// 从记录中提取字段值（使用字段名字符串，兼容迁移前后）
	content := record["menu_data"].String()
	// 尝试获取 updated_at，如果不存在则使用 update_time（兼容迁移前）
	updatedAt := record["updated_at"].Int64()
	if updatedAt == 0 {
		updatedAt = record["update_time"].Int64()
	}
	syncState := record["sync_state"].String()

	// 构建响应
	resp := &api.GetMenuSnapshotResp{
		ResponseInfo: &api.ResponseInfo{
			Code:    "1",
			Message: "success",
		},
		Content:   content,
		UpdatedAt: updatedAt,
		SyncState: syncState,
	}

	return resp, nil
}
