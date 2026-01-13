package channel_menu

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"

	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/utility/uuid"
)

type sChannelMenu struct{}

func init() {
	service.RegisterChannelMenu(New())
	// 同时注册为 Menu Service（统一菜单路由入口）
	service.RegisterMenu(New())
}

func New() *sChannelMenu {
	return &sChannelMenu{}
}

// GetChannelMenu 读取外卖渠道菜单快照
func (s *sChannelMenu) GetChannelMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error) {
	record, err := dao.ChannelMenuSnapshot.Ctx(ctx).
		Fields(dao.ChannelMenuSnapshot.Columns().MenuData).
		Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUUID).
		Where(dao.ChannelMenuSnapshot.Columns().ProviderName, providerName).
		One()
	if err != nil {
		return "", err
	}
	if record.IsEmpty() {
		return "", nil // Not found
	}
	return record[dao.ChannelMenuSnapshot.Columns().MenuData].String(), nil
}

// GetTtposMenu 读取TTPOS菜单快照
func (s *sChannelMenu) GetTtposMenu(ctx context.Context, shopUUID uint64, providerName string) (string, error) {
	record, err := dao.ChannelMenuSnapshot.Ctx(ctx).
		Fields(dao.ChannelMenuSnapshot.Columns().TtposMenuData).
		Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUUID).
		Where(dao.ChannelMenuSnapshot.Columns().ProviderName, providerName).
		One()
	if err != nil {
		return "", err
	}
	if record.IsEmpty() {
		return "", nil // Not found
	}
	return record[dao.ChannelMenuSnapshot.Columns().TtposMenuData].String(), nil
}

// GetMenuSnapshot
func (s *sChannelMenu) GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*api.GetMenuSnapshotResp, error) {
	// 参数校验
	if req.ProviderName == "" {
		return nil, gerror.New("provider_name 不能为空")
	}
	if req.ShopUuid == "" {
		return nil, gerror.New("shop_uuid 不能为空")
	}

	// 查询快照记录（使用字段名字符串，兼容迁移前后）
	record, err := dao.ChannelMenuSnapshot.Ctx(ctx).
		Where("provider_name", req.ProviderName).
		Where("shop_uuid", req.ShopUuid).
		Where("deleted_at = 0 "). // 软删除过滤
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "查询菜单快照失败")
	}
	if record.IsEmpty() {
		return nil, gerror.New("菜单快照不存在")
	}

	// 从记录中提取字段值（使用字段名字符串，兼容迁移前后）
	menuData := record[dao.ChannelMenuSnapshot.Columns().MenuData].String()
	updatedAt := record[dao.ChannelMenuSnapshot.Columns().UpdatedAt].Int64()
	syncState := record[dao.ChannelMenuSnapshot.Columns().SyncState].String()

	// 构建响应
	resp := &api.GetMenuSnapshotResp{
		MenuData:  menuData,
		UpdatedAt: updatedAt,
		SyncState: syncState,
	}

	return resp, nil
}

// SaveMenuSnapshot 保存菜单快照
func (s *sChannelMenu) SaveMenuSnapshot(ctx context.Context, req *api.SaveMenuSnapshotReq) (*api.SaveMenuSnapshotResp, error) {
	// 1. 参数校验
	if req.ProviderName == "" {
		return nil, gerror.New("provider_name 不能为空")
	}
	if req.ShopUuid == "" {
		return nil, gerror.New("shop_uuid 不能为空")
	}
	if req.MenuData == "" {
		return nil, gerror.New("menu_data 不能为空")
	}

	// 2. 解析 shop_uuid
	shopUuidInt, err := strconv.ParseUint(req.ShopUuid, 10, 64)
	if err != nil {
		return nil, gerror.Wrap(err, "shop_uuid 格式错误")
	}

	// 3. 查找是否已存在记录（根据 provider_name + shop_uuid）
	// nowTs := int(time.Now().Unix())
	record, err := dao.ChannelMenuSnapshot.Ctx(ctx).
		Where(dao.ChannelMenuSnapshot.Columns().ProviderName, req.ProviderName).
		Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUuidInt).
		Where("deleted_at = 0 ").
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "查询菜单快照失败")
	}

	if record.IsEmpty() {
		// 4a. 不存在，创建新记录
		_, err = dao.ChannelMenuSnapshot.Ctx(ctx).Data(g.Map{
			dao.ChannelMenuSnapshot.Columns().Uuid:          uuid.MustGetID(),
			dao.ChannelMenuSnapshot.Columns().ShopUuid:      shopUuidInt,
			dao.ChannelMenuSnapshot.Columns().ProviderName:  req.ProviderName,
			dao.ChannelMenuSnapshot.Columns().TtposMenuData: req.MenuData,
			// dao.ChannelMenuSnapshot.Columns().TtposUpdatedAt: nowTs,
		}).Insert()
	} else {
		// 4b. 存在，更新记录
		_, err = dao.ChannelMenuSnapshot.Ctx(ctx).
			Where(dao.ChannelMenuSnapshot.Columns().Id, record["id"].Uint64()).
			Data(g.Map{
				dao.ChannelMenuSnapshot.Columns().TtposMenuData: req.MenuData,
				// dao.ChannelMenuSnapshot.Columns().TtposUpdatedAt: nowTs,
			}).Update()
	}

	if err != nil {
		return nil, gerror.Wrap(err, "保存菜单快照失败")
	}

	g.Log().Infof(ctx, "SaveMenuSnapshot: 保存成功, provider=%s, shop_uuid=%s", req.ProviderName, req.ShopUuid)

	// 5. 如果是 Grab 渠道，异步通知菜单更新
	if req.ProviderName == string(consts.ProviderGrab) {
		go s.notifyGrabMenuUpdate(context.Background(), shopUuidInt)
	}
	// 6. 如果是 Lineman 渠道， 实时同步调用lineman SyncMenu
	if req.ProviderName == string(consts.ProviderLineman) {
		err = service.Lineman().SyncMenu(ctx, shopUuidInt)
		if err != nil {
			g.Log().Errorf(ctx, "SyncMenu failed: shop_uuid=%d, err=%v", shopUuidInt, err)
			return nil, gerror.Wrap(err, "同步菜单失败")
		}
	}

	return &api.SaveMenuSnapshotResp{}, nil
}

// LogMenuSync 记录菜单同步日志
// 通用方法，供各个渠道（grab、lineman 等）调用
// 参数：
//   - ctx: 上下文
//   - merchantID: 商户ID
//   - providerName: 渠道名称（grab/lineman）
//   - syncType: 同步类型（FULL/PARTIAL/BATCH_UPDATE_ITEM等）
//   - requestID: 请求ID（来自第三方API响应）
//   - success: 是否成功
//   - menuSnapshot: 菜单快照（JSON 字符串，可选）
//   - errMsg: 错误信息（失败时）
func (s *sChannelMenu) LogMenuSync(ctx context.Context, merchantID, providerName, syncType, requestID string, success bool, menuSnapshot, errMsg string) error {
	logUUID := uuid.MustGetID()
	status := "SUCCESS"
	if !success {
		status = "FAIL"
	}

	logDo := g.Map{
		dao.MenuLog.Columns().Uuid:         logUUID,
		dao.MenuLog.Columns().MerchantId:   merchantID,
		dao.MenuLog.Columns().ProviderName: providerName,
		dao.MenuLog.Columns().SyncType:     syncType,
		dao.MenuLog.Columns().RequestId:    requestID,
		dao.MenuLog.Columns().Status:       status,
		dao.MenuLog.Columns().MenuSnapshot: menuSnapshot,
		dao.MenuLog.Columns().ErrorMsg:     errMsg,
	}

	_, err := dao.MenuLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "[ChannelMenu] 插入菜单同步日志失败: merchantID=%s, provider=%s, syncType=%s, error=%v",
			merchantID, providerName, syncType, err)
		return gerror.Wrap(err, "插入菜单同步日志失败")
	}

	g.Log().Debugf(ctx, "[ChannelMenu] 菜单同步日志已插入: logUUID=%d, merchantID=%s, provider=%s, syncType=%s, success=%v",
		logUUID, merchantID, providerName, syncType, success)

	return nil
}

// notifyGrabMenuUpdate 异步通知 Grab 菜单更新
func (s *sChannelMenu) notifyGrabMenuUpdate(ctx context.Context, shopUuid uint64) {
	// 1. 获取门店的 Grab 配置
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUuid, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "notifyGrabMenuUpdate: 获取门店第三方配置失败: shop_uuid=%d, err=%v", shopUuid, err)
		return
	}
	if cfg == nil || cfg.ProviderMerchantId == "" {
		g.Log().Warningf(ctx, "notifyGrabMenuUpdate: 未找到 merchant_id: shop_uuid=%d", shopUuid)
		return
	}

	// 2. 调用 Grab NotifyMenuUpdate
	requestId, err := service.Grab().NotifyMenuUpdate(ctx, cfg.ProviderMerchantId)
	if err != nil {
		g.Log().Errorf(ctx, "notifyGrabMenuUpdate: 通知 Grab 失败: shop_uuid=%d, merchant_id=%s, err=%v", shopUuid, cfg.ProviderMerchantId, err)
		return
	}

	g.Log().Infof(ctx, "notifyGrabMenuUpdate: 成功, shop_uuid=%d, merchant_id=%s, request_id=%s", shopUuid, cfg.ProviderMerchantId, requestId)
}

// NotifyMenuUpdate 通知菜单更新（统一路由入口）
// 实现 IMenu 接口，根据 provider_name 路由到对应平台的菜单同步服务
func (s *sChannelMenu) NotifyMenuUpdate(ctx context.Context, req *api.NotifyMenuUpdateReq) (*takeout.ApiResponse, error) {
	// 1. 参数校验
	if req.ShopUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "shop_uuid 不能为空",
		}, nil
	}
	if req.ProviderName == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "provider_name 不能为空",
		}, nil
	}

	// 2. 查询店铺平台配置
	shopUUID, err := strconv.ParseUint(req.ShopUuid, 10, 64)
	if err != nil {
		g.Log().Errorf(ctx, "[菜单服务] shop_uuid 格式无效: %s, 错误: %v", req.ShopUuid, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "shop_uuid 格式无效: " + err.Error(),
		}, nil
	}

	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, req.ProviderName)
	if err != nil {
		g.Log().Errorf(ctx, "[菜单服务] 获取店铺平台配置失败: shop=%s, provider=%s, 错误: %v", req.ShopUuid, req.ProviderName, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: fmt.Sprintf("获取店铺平台配置失败: %v", err),
		}, nil
	}

	// 检查配置是否存在
	if cfg == nil {
		g.Log().Warningf(ctx, "[菜单服务] 未找到店铺平台配置: shop=%s, provider=%s", req.ShopUuid, req.ProviderName)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: fmt.Sprintf("店铺 %s 未配置平台 %s", req.ShopUuid, req.ProviderName),
		}, nil
	}

	// 检查平台状态（ProviderShopStatus 应该是 ACTIVE）
	if cfg.ProviderShopStatus != string(consts.ProviderShopStatusActive) {
		g.Log().Warningf(ctx, "[菜单服务] 平台未激活: shop=%s, provider=%s, status=%s", req.ShopUuid, req.ProviderName, cfg.ProviderShopStatus)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: fmt.Sprintf("平台 %s 在店铺 %s 中未激活", req.ProviderName, req.ShopUuid),
		}, nil
	}

	// 3. 记录日志
	g.Log().Infof(ctx, "[菜单服务] 通知菜单更新: shop_uuid=%s, provider=%s, request_id=%s, provider_merchant_id=%s",
		req.ShopUuid, req.ProviderName, req.RequestId, cfg.ProviderMerchantId)

	// 4. 根据 provider_name 路由到对应服务
	switch req.ProviderName {
	case "grab":
		return s.notifyGrabMenuUpdateWithResponse(ctx, cfg.ProviderMerchantId, req.RequestId)

	case "lineman":
		return s.notifyLinemanMenuUpdateWithResponse(ctx, shopUUID, req.RequestId)

	default:
		errMsg := fmt.Sprintf("不支持的平台: %s，支持的平台: grab, lineman", req.ProviderName)
		g.Log().Warningf(ctx, "[菜单服务] %s", errMsg)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: errMsg,
		}, nil
	}
}

// notifyGrabMenuUpdateWithResponse 通知 Grab 菜单更新（带响应）
func (s *sChannelMenu) notifyGrabMenuUpdateWithResponse(ctx context.Context, merchantID string, requestID string) (*takeout.ApiResponse, error) {
	// 调用 Grab Service
	grabRequestID, err := service.Grab().NotifyMenuUpdate(ctx, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "[菜单服务] 通知 Grab 失败: merchant_id=%s, 错误: %v", merchantID, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "通知 Grab 菜单更新失败: " + err.Error(),
		}, nil
	}

	// 构建响应数据
	respData := g.Map{
		"sync_status": "QUEUED",
		"request_id":  grabRequestID,
		"provider":    "grab",
	}

	// 转换为 anypb.Any
	dataAny, err := anypb.New(&takeout.ApiResponse{})
	if err == nil {
		// 如果需要，可以使用更合适的消息类型
		g.Log().Debugf(ctx, "[菜单服务] Grab 菜单更新响应: %+v", respData)
	}

	g.Log().Infof(ctx, "[菜单服务] Grab 菜单更新通知成功: merchant_id=%s, request_id=%s", merchantID, grabRequestID)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// notifyLinemanMenuUpdateWithResponse 通知 Lineman 菜单更新（带响应）
func (s *sChannelMenu) notifyLinemanMenuUpdateWithResponse(ctx context.Context, shopUUID uint64, requestID string) (*takeout.ApiResponse, error) {
	// 调用 Lineman Service
	err := service.Lineman().SyncMenu(ctx, shopUUID)
	if err != nil {
		g.Log().Errorf(ctx, "[菜单服务] 通知 Lineman 失败: shop_uuid=%d, 错误: %v", shopUUID, err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: "通知 Lineman 菜单更新失败: " + err.Error(),
		}, nil
	}

	// 构建响应数据
	respData := g.Map{
		"sync_status": "SUCCESS",
		"request_id":  requestID,
		"provider":    "lineman",
	}

	// 转换为 anypb.Any
	dataAny, err := anypb.New(&takeout.ApiResponse{})
	if err == nil {
		g.Log().Debugf(ctx, "[菜单服务] Lineman 菜单更新响应: %+v", respData)
	}

	g.Log().Infof(ctx, "[菜单服务] Lineman 菜单更新通知成功: shop_uuid=%d", shopUUID)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}
