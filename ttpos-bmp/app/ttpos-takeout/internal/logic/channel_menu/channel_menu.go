package channel_menu

import (
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"ttpos-bmp/utility/uuid"
)

type sChannelMenu struct{}

func init() {
	service.RegisterChannelMenu(New())
}

func New() *sChannelMenu {
	return &sChannelMenu{}
}

// SaveChannelMenu 保存外卖渠道菜单快照
func (s *sChannelMenu) SaveChannelMenu(ctx context.Context, shopUUID uint64, providerName string, menuData string) error {
	// 1. 检查是否存在
	count, err := dao.ChannelMenuSnapshot.Ctx(ctx).
		Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUUID).
		Where(dao.ChannelMenuSnapshot.Columns().ProviderName, providerName).
		Count()
	if err != nil {
		return err
	}

	if count > 0 {
		// Update
		_, err := dao.ChannelMenuSnapshot.Ctx(ctx).Data(g.Map{
			dao.ChannelMenuSnapshot.Columns().MenuData: menuData,
		}).Where(dao.ChannelMenuSnapshot.Columns().ShopUuid, shopUUID).
			Where(dao.ChannelMenuSnapshot.Columns().ProviderName, providerName).Update()
		return err
	} else {
		// Insert
		_, err := dao.ChannelMenuSnapshot.Ctx(ctx).Data(g.Map{
			dao.ChannelMenuSnapshot.Columns().Uuid:         uuid.MustGetID(),
			dao.ChannelMenuSnapshot.Columns().ShopUuid:     shopUUID,
			dao.ChannelMenuSnapshot.Columns().ProviderName: providerName,
			dao.ChannelMenuSnapshot.Columns().MenuData:     menuData,
		}).Insert()
		return err
	}
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

// GetMenuSnapshot 根据 request_id 查询菜单快照
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
	content := record[dao.ChannelMenuSnapshot.Columns().TtposMenuData].String()
	// 尝试获取 ttpos_updated_at，如果不存在则使用 updated_at
	updatedAt := record[dao.ChannelMenuSnapshot.Columns().TtposUpdatedAt].Int64()
	if updatedAt == 0 {
		updatedAt = record[dao.ChannelMenuSnapshot.Columns().UpdatedAt].Int64()
	}
	syncState := record[dao.ChannelMenuSnapshot.Columns().SyncState].String()

	// 构建响应
	resp := &api.GetMenuSnapshotResp{

		Content:   content,
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
	nowTs := int(time.Now().Unix())
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
			dao.ChannelMenuSnapshot.Columns().ShopUuid:       shopUuidInt,
			dao.ChannelMenuSnapshot.Columns().ProviderName:   req.ProviderName,
			dao.ChannelMenuSnapshot.Columns().TtposMenuData:  req.MenuData,
			dao.ChannelMenuSnapshot.Columns().TtposUpdatedAt: nowTs,
		}).Insert()
	} else {
		// 4b. 存在，更新记录
		_, err = dao.ChannelMenuSnapshot.Ctx(ctx).
			Where(dao.ChannelMenuSnapshot.Columns().Id, record["id"].Uint64()).
			Data(g.Map{
				dao.ChannelMenuSnapshot.Columns().TtposMenuData:  req.MenuData,
				dao.ChannelMenuSnapshot.Columns().TtposUpdatedAt: nowTs,
			}).Update()
	}

	if err != nil {
		return nil, gerror.Wrap(err, "保存菜单快照失败")
	}

	g.Log().Infof(ctx, "SaveMenuSnapshot: saved successfully, provider=%s, shop_uuid=%s", req.ProviderName, req.ShopUuid)

	// 5. 如果是 Grab 渠道，异步通知菜单更新
	if req.ProviderName == string(consts.ProviderGrab) {
		go s.notifyGrabMenuUpdate(context.Background(), shopUuidInt)
	}

	return &api.SaveMenuSnapshotResp{}, nil
}

// notifyGrabMenuUpdate 异步通知 Grab 菜单更新
func (s *sChannelMenu) notifyGrabMenuUpdate(ctx context.Context, shopUuid uint64) {
	// 1. 获取门店的 Grab 配置
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUuid, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "notifyGrabMenuUpdate: get shop provider cfg failed: shop_uuid=%d, err=%v", shopUuid, err)
		return
	}
	if cfg == nil || cfg.ProviderMerchantId == "" {
		g.Log().Warningf(ctx, "notifyGrabMenuUpdate: merchant_id not found: shop_uuid=%d", shopUuid)
		return
	}

	// 2. 调用 Grab NotifyMenuUpdate
	requestId, err := service.Grab().NotifyMenuUpdate(ctx, cfg.ProviderMerchantId)
	if err != nil {
		g.Log().Errorf(ctx, "notifyGrabMenuUpdate: notify grab failed: shop_uuid=%d, merchant_id=%s, err=%v", shopUuid, cfg.ProviderMerchantId, err)
		return
	}

	g.Log().Infof(ctx, "notifyGrabMenuUpdate: success, shop_uuid=%d, merchant_id=%s, request_id=%s", shopUuid, cfg.ProviderMerchantId, requestId)
}
