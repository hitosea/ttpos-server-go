package rpc

import (
	"context"
	"fmt"
	"strconv"
	menuApi "ttpos-bmp/app/ttpos-takeout/api/menu"
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

// CheckBindingStatus 检查绑定状态（保持向后兼容）
// GetBMPClient 获取 BMP 客户端（用于直接调用客户端方法）
func (s *TakeoutRPCService) GetBMPClient() (*BMPTakeoutClient, error) {
	client, err := NewBMPTakeoutClient()
	if err != nil {
		return nil, err
	}

	// 类型断言
	bmpClient, ok := client.(*BMPTakeoutClient)
	if !ok {
		return nil, fmt.Errorf("无法将客户端转换为 BMPTakeoutClient 类型")
	}

	return bmpClient, nil
}

func (s *TakeoutRPCService) CheckBindingStatus(ctx context.Context, platform string, companyUuid uint64) (bool, error) {
	status, _, _, _, err := s.CheckBindingStatusWithMerchantId(ctx, platform, companyUuid)
	return status, err
}

// CheckBindingStatusWithMerchantId 检查绑定状态并返回 MerchantID
func (s *TakeoutRPCService) CheckBindingStatusWithMerchantId(ctx context.Context, platform string, companyUuid uint64) (status bool, updatedAt int64, merchantId string, merchantName string, err error) {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败",
			zap.Error(err),
			zap.String("platform", platform),
			zap.Uint64("companyUuid", companyUuid))
		return false, 0, "", "", errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	providerStatus, updatedAt, merchantId, merchantName, err := client.GetShopProviderCfg(ctx, platform, companyUuid)
	if err != nil {
		if errors.Is(err, errors.New("no rows in result set")) {
			return false, 0, "", "", nil
		}
		logger.Logger.Error("检查绑定状态失败",
			zap.Error(err),
			zap.String("platform", platform),
			zap.Uint64("companyUuid", companyUuid))
		return false, 0, "", "", errors.WithMessage(err, "检查绑定状态失败")
	}

	// 门店集成状态 (INACTIVE/ACTIVE/SYNCING/FAILED)
	if providerStatus != "ACTIVE" {
		return false, updatedAt, merchantId, merchantName, nil // 未绑定
	}

	return true, updatedAt, merchantId, merchantName, nil
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

// UpdateMenuItem 更新菜单项（商品）
func (s *TakeoutRPCService) UpdateMenuItem(ctx context.Context, merchantId string, itemId string, price *int64, availableStatus string, maxStock *int64) error {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败", zap.Error(err))
		return errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 构造请求
	req := &menuApi.UpdateMenuItemReq{
		MerchantId:      merchantId,
		ItemId:          itemId,
		Price:           price,
		AvailableStatus: &availableStatus,
		MaxStock:        maxStock,
		RequestId:       uuid.New().String(),
	}

	// 调用 RPC 接口
	err = client.UpdateMenuItem(ctx, req)
	if err != nil {
		logger.Logger.Error("更新菜单项失败",
			zap.Error(err),
			zap.String("merchantId", merchantId),
			zap.String("itemId", itemId))
		return errors.WithMessage(err, "更新菜单项失败")
	}

	return nil
}

// UpdateMenuModifier 更新菜单修饰符
func (s *TakeoutRPCService) UpdateMenuModifier(ctx context.Context, merchantId string, modifierId string, modifierName string, price *int64, availableStatus string) error {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败", zap.Error(err))
		return errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 构造请求
	req := &menuApi.UpdateMenuModifierReq{
		MerchantId:      merchantId,
		ModifierId:      modifierId,
		ModifierName:    modifierName,
		Price:           price,
		AvailableStatus: &availableStatus,
		RequestId:       uuid.New().String(),
	}

	// 调用 RPC 接口
	err = client.UpdateMenuModifier(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// ActivateShop 激活门店外卖渠道
func (s *TakeoutRPCService) ActivateShop(ctx context.Context, providerName string, shopUuid uint64) error {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败",
			zap.Error(err),
			zap.String("providerName", providerName),
			zap.Uint64("shopUuid", shopUuid))
		return errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	err = client.ActivateShop(ctx, providerName, shopUuid)
	if err != nil {
		logger.Logger.Error("激活门店外卖渠道失败",
			zap.Error(err),
			zap.String("providerName", providerName),
			zap.Uint64("shopUuid", shopUuid))
		return errors.WithMessage(err, "激活门店外卖渠道失败")
	}

	return nil
}

// GetOrderInfo 获取订单信息
func (s *TakeoutRPCService) GetOrderInfo(ctx context.Context, shopUuid string, orderUuid string) (map[string]interface{}, error) {
	// 创建客户端
	client, err := NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败", zap.Error(err), zap.String("shopUuid", shopUuid), zap.String("orderUuid", orderUuid))
		return nil, errors.WithMessage(err, "创建 RPC 客户端失败")
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	// 调用 RPC 接口
	orderData, err := client.GetOrderInfo(ctx, shopUuid, orderUuid)
	if err != nil {
		logger.Logger.Error("获取订单信息失败", zap.Error(err), zap.String("shopUuid", shopUuid), zap.String("orderUuid", orderUuid))
		return nil, errors.WithMessage(err, "获取订单信息失败")
	}

	return orderData, nil
}
