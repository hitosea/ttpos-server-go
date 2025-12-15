package rpc

import (
	"context"
	"encoding/json"
	"strconv"
	grabApi "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/api/menu"
	menuApi "ttpos-bmp/app/ttpos-takeout/api/menu"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TakeoutRPCClient 外送 RPC 客户端接口
type TakeoutRPCClient interface {
	// GetShopProviderCfg 检查绑定状态
	GetShopProviderCfg(ctx context.Context, platform string, companyUuid uint64) (status string, updatedAt int64, merchantId string, merchantName string, err error)

	// GetGrabBindingLink 获取 Grab 绑定链接
	GetGrabBindingLink(ctx context.Context, companyUuid uint64) (bindingLink string, err error)

	// GetGrabMenu 获取 Grab 商品菜单
	GetMenuSnapshot(ctx context.Context, companyUuid uint64) (menuData interface{}, err error)

	// SaveGrabMenu 保存 Grab 菜单
	SaveMenuSnapshot(ctx context.Context, providerName string, shopUuid string, requestId string, menuData interface{}) (err error)

	// Close 关闭连接
	Close() error
}

// BMPTakeoutClient BMP 外送 RPC 客户端实现
type BMPTakeoutClient struct {
	conn       *grpc.ClientConn
	client     api.TakeoutServiceClient
	grabClient grabApi.GrabClient
	menuClient menuApi.MenuServiceClient
}

// NewBMPTakeoutClient 创建 BMP 外送 RPC 客户端
func NewBMPTakeoutClient() (TakeoutRPCClient, error) {
	// 建立gRPC连接
	conn, err := cloud.GetRpcConnWithName(cloud.TakeOutServiceName)
	if err != nil {
		logger.Logger.Error("创建 BMP 外送服务 gRPC 连接失败", zap.Error(err))
		return nil, errors.WithMessage(err, "创建 BMP 外送服务 gRPC 连接失败")
	}

	client := api.NewTakeoutServiceClient(conn)
	grabClient := grabApi.NewGrabClient(conn)
	menuClient := menu.NewMenuServiceClient(conn)
	return &BMPTakeoutClient{
		conn:       conn,
		client:     client,
		grabClient: grabClient,
		menuClient: menuClient,
	}, nil
}

// GetShopProviderCfg 检查绑定状态
func (c *BMPTakeoutClient) GetShopProviderCfg(ctx context.Context, platform string, companyUuid uint64) (status string, updatedAt int64, merchantId string, merchantName string, err error) {
	resp, err := c.grabClient.GetShopProviderCfg(ctx, &grabApi.GetShopProviderCfgReq{
		ShopUuid:     companyUuid,
		ProviderName: platform,
	})
	if err != nil {
		logger.Logger.Error("调用 CheckBindingStatus 接口失败",
			zap.Error(err),
			zap.String("platform", platform),
			zap.Uint64("companyUuid", companyUuid))
		return "", 0, "", "", errors.WithMessage(err, "调用获取门店第三方配置接口失败")
	}
	if resp.Code != "0" {
		logger.Logger.Warn("GetShopProviderCfg 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.Uint64("companyUuid", companyUuid))
		return "", 0, "", "", errors.New(resp.Message)
	}
	data, err := resp.Data.UnmarshalNew()
	if err != nil {
		logger.Logger.Warn("GetShopProviderCfg 返回数据解析失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return "", 0, "", "", errors.WithMessage(err, "获取门店第三方配置数据解析失败")
	}

	return data.(*grabApi.GetShopProviderCfgResp).ProviderShopStatus,
		data.(*grabApi.GetShopProviderCfgResp).UpdatedAt,
		data.(*grabApi.GetShopProviderCfgResp).ProviderMerchantId,
		data.(*grabApi.GetShopProviderCfgResp).ProviderName,
		nil
}

// GetGrabBindingLink 获取 Grab 绑定链接
func (c *BMPTakeoutClient) GetGrabBindingLink(ctx context.Context, companyUuid uint64) (bindingLink string, err error) {
	resp, err := c.grabClient.CreateSelfServeJourney(ctx, &grabApi.CreateSelfServeJourneyReq{
		ProviderName: "grab",
		ShopUuid:     strconv.FormatUint(companyUuid, 10),
		RequestId:    uuid.New().String(),
	})
	if err != nil {
		logger.Logger.Error("调用 GetGrabBindingLink 接口失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return "", errors.WithMessage(err, "调用获取绑定链接接口失败")
	}
	if resp.Code != "0" {
		logger.Logger.Warn("GetGrabBindingLink 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.Uint64("companyUuid", companyUuid))
		return "", errors.New(resp.Message)
	}

	data, err := resp.Data.UnmarshalNew()
	if err != nil {
		logger.Logger.Warn("GetGrabBindingLink 返回数据解析失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return "", errors.WithMessage(err, "获取绑定链接数据解析失败")
	}

	return data.(*grabApi.CreateSelfServeJourneyResp).SelfServeUrl, nil
}

// GetMenuSnapshot 获取 Grab 商品菜单
func (c *BMPTakeoutClient) GetMenuSnapshot(ctx context.Context, companyUuid uint64) (menuData interface{}, err error) {
	req := &menu.GetMenuSnapshotReq{
		ProviderName: "grab",
		ShopUuid:     strconv.FormatUint(companyUuid, 10),
		RequestId:    uuid.New().String(),
	}

	resp, err := c.menuClient.GetMenuSnapshot(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 GetGrabMenu 接口失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return nil, errors.WithMessage(err, "调用获取 Grab 菜单接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("GetGrabMenu 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.Uint64("companyUuid", companyUuid))
		return nil, errors.New(resp.Message)
	}
	data, err := resp.Data.UnmarshalNew()
	if err != nil {
		logger.Logger.Warn("GetGrabMenu 返回数据解析失败",
			zap.Error(err),
			zap.Uint64("companyUuid", companyUuid))
		return nil, errors.WithMessage(err, "获取 Grab 菜单数据解析失败")
	}

	return data.(*menu.GetMenuSnapshotResp).MenuData, nil
}

// SaveGrabMenu 保存 Grab 菜单
func (c *BMPTakeoutClient) SaveMenuSnapshot(ctx context.Context, providerName string, shopUuid string, requestId string, menuData interface{}) (err error) {
	// 将 menuData 序列化为 JSON 字符串
	var menuDataStr string
	switch v := menuData.(type) {
	case string:
		// 如果已经是字符串，直接使用
		menuDataStr = v
	default:
		// 否则序列化为 JSON
		jsonBytes, err := json.Marshal(menuData)
		if err != nil {
			logger.Logger.Error("序列化菜单数据失败",
				zap.Error(err),
				zap.String("providerName", providerName),
				zap.String("shopUuid", shopUuid))
			return errors.WithMessage(err, "序列化菜单数据失败")
		}
		menuDataStr = string(jsonBytes)
	}

	req := &menu.SaveMenuSnapshotReq{
		ProviderName: providerName,
		ShopUuid:     shopUuid,
		RequestId:    requestId,
		MenuData:     menuDataStr,
	}

	resp, err := c.menuClient.SaveMenuSnapshot(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 SaveGrabMenu 接口失败",
			zap.Error(err),
			zap.String("providerName", providerName),
			zap.String("shopUuid", shopUuid),
			zap.String("requestId", requestId))
		return errors.WithMessage(err, "调用保存 Grab 菜单接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("SaveGrabMenu 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("providerName", providerName),
			zap.String("shopUuid", shopUuid),
			zap.String("requestId", requestId))
		return errors.New(resp.Message)
	}

	return nil
}

// Close 关闭连接
func (c *BMPTakeoutClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
