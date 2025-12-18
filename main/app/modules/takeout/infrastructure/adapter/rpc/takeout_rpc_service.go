package rpc

import (
	"context"
	"strconv"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TakeoutRPCService 外送 RPC 服务
type TakeoutRPCService struct {
}

// NewTakeoutRPCService 创建外送 RPC 服务
func NewTakeoutRPCService() *TakeoutRPCService {
	return &TakeoutRPCService{}
}

// CheckBindingStatus 检查绑定状态
func (s *TakeoutRPCService) CheckBindingStatus(ctx context.Context, platform string, companyUuid uint64) (status bool, err error) {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败",
			zap.Error(err),
			zap.String("platform", platform),
			zap.Uint64("companyUuid", companyUuid))
		return false, errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	providerStatus, _, _, _, err := client.GetShopProviderCfg(ctx, platform, companyUuid)
	if err != nil {
		if errors.Is(err, errors.New("no rows in result set")) {
			return false, nil
		}
		logger.Logger.Error("检查绑定状态失败",
			zap.Error(err),
			zap.String("platform", platform),
			zap.Uint64("companyUuid", companyUuid))
		return false, errors.WithMessage(err, "检查绑定状态失败")
	}

	// 门店集成状态 (INACTIVE/ACTIVE/SYNCING/FAILED)
	if providerStatus != "ACTIVE" {
		return false, nil // 未绑定
	}

	return true, nil
}

// GetGrabBindingLink 获取 Grab 绑定链接
func (s *TakeoutRPCService) GetGrabBindingLink(ctx context.Context, companyUuid uint64) (bindingLink string, err error) {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return "", errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	bindingLink, err = client.GetGrabBindingLink(ctx, companyUuid)
	if err != nil {
		logger.Logger.Error("获取 Grab 绑定链接失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return "", errors.WithMessage(err, "获取 Grab 绑定链接失败")
	}

	return bindingLink, nil
}

// GetGrabMenu 获取 Grab 商品菜单
func (s *TakeoutRPCService) GetGrabMenu(ctx context.Context, companyUuid uint64) (menuData interface{}, err error) {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return nil, errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	menuData, err = client.GetMenuSnapshot(ctx, companyUuid)
	if err != nil {
		logger.Logger.Error("获取 Grab 商品菜单失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return nil, errors.WithMessage(err, "获取 Grab 商品菜单失败")
	}

	return menuData, nil
}

// SaveMenuSnapshot 推送菜单到Grab
func (s *TakeoutRPCService) SaveMenuSnapshot(ctx context.Context, providerName string, companyUuid uint64, menu interface{}) (err error) {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	err = client.SaveMenuSnapshot(ctx, providerName, strconv.FormatUint(companyUuid, 10), uuid.New().String(), menu)
	if err != nil {
		logger.Logger.Error("保存 Grab 菜单失败", zap.Error(err), zap.Uint64("companyUuid", companyUuid))
		return errors.WithMessage(err, "保存 Grab 菜单失败")
	}

	return nil
}
