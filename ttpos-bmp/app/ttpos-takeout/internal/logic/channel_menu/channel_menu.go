package channel_menu

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

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
			dao.ChannelMenuSnapshot.Columns().MenuData:   menuData,
			dao.ChannelMenuSnapshot.Columns().UpdateTime: gtime.Now().Timestamp(),
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
			dao.ChannelMenuSnapshot.Columns().CreateTime:   gtime.Now().Timestamp(),
			dao.ChannelMenuSnapshot.Columns().UpdateTime:   gtime.Now().Timestamp(),
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
