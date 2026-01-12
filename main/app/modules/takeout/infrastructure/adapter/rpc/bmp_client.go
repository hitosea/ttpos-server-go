package rpc

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	grabApi "ttpos-bmp/app/ttpos-takeout/api/grab"
	menuApi "ttpos-bmp/app/ttpos-takeout/api/menu"
	orderApi "ttpos-bmp/app/ttpos-takeout/api/order"
	shopApi "ttpos-bmp/app/ttpos-takeout/api/shop"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
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

	// UpdateMenuItem 更新菜单项（商品）
	UpdateMenuItem(ctx context.Context, req *menuApi.UpdateMenuItemReq) error

	// UpdateMenuModifier 更新菜单修饰符
	UpdateMenuModifier(ctx context.Context, req *menuApi.UpdateMenuModifierReq) error

	// ActivateShop 激活门店外卖渠道
	ActivateShop(ctx context.Context, providerName string, shopUuid uint64) error

	//  ------------------------- Order Service -------------------------

	// GetOrderInfo 获取订单信息
	GetOrderInfo(ctx context.Context, shopUuid string, orderUuid string) (orderData map[string]interface{}, err error)

	// PrepareOrder 准备订单（接受/拒绝）
	PrepareOrder(ctx context.Context, takeoutOrderUuid string, toState string) error

	// MarkOrderReady 标记订单准备完成（呼叫骑手）
	MarkOrderReady(ctx context.Context, takeoutOrderUuid string) error

	// CheckOrderCancelable 检查订单是否可取消
	CheckOrderCancelable(ctx context.Context, takeoutOrderUuid string) (canCancel bool, reason string, rawData string, err error)

	// CancelOrder 取消订单
	CancelOrder(ctx context.Context, takeoutOrderUuid string, reasonCode string) error

	// Close 关闭连接
	Close() error
}

// BMPTakeoutClient BMP 外送 RPC 客户端实现
type BMPTakeoutClient struct {
	conn        *grpc.ClientConn
	client      api.TakeoutServiceClient
	grabClient  grabApi.GrabClient
	menuClient  menuApi.MenuServiceClient
	orderClient orderApi.OrderServiceClient
	shopClient  shopApi.ShopClient
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
	menuClient := menuApi.NewMenuServiceClient(conn)
	orderClient := orderApi.NewOrderServiceClient(conn)
	shopClient := shopApi.NewShopClient(conn)
	return &BMPTakeoutClient{
		conn:        conn,
		client:      client,
		grabClient:  grabClient,
		menuClient:  menuClient,
		orderClient: orderClient,
		shopClient:  shopClient,
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
		ProviderName: value_object.TakeoutPlatformGrab,
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
	req := &menuApi.GetMenuSnapshotReq{
		ProviderName: value_object.TakeoutPlatformGrab,
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

	return data.(*menuApi.GetMenuSnapshotResp).MenuData, nil
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

	req := &menuApi.SaveMenuSnapshotReq{
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

// UpdateMenuItem 更新菜单项（商品）
func (c *BMPTakeoutClient) UpdateMenuItem(ctx context.Context, req *menuApi.UpdateMenuItemReq) error {
	resp, err := c.menuClient.UpdateMenuItem(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 UpdateMenuItem 接口失败",
			zap.Error(err),
			zap.String("merchantId", req.MerchantId),
			zap.String("itemId", req.ItemId))
		return errors.WithMessage(err, "调用更新菜单项接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("UpdateMenuItem 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("merchantId", req.MerchantId),
			zap.String("itemId", req.ItemId))
		return errors.New(resp.Message)
	}

	return nil
}

// UpdateMenuModifier 更新菜单修饰符
func (c *BMPTakeoutClient) UpdateMenuModifier(ctx context.Context, req *menuApi.UpdateMenuModifierReq) error {
	resp, err := c.menuClient.UpdateMenuModifier(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 UpdateMenuModifier 接口失败",
			zap.Error(err),
			zap.String("merchantId", req.MerchantId),
			zap.String("modifierId", req.ModifierId))
		return errors.WithMessage(err, "调用更新菜单修饰符接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("UpdateMenuModifier 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("merchantId", req.MerchantId),
			zap.String("modifierId", req.ModifierId))
		return errors.New(resp.Message)
	}

	return nil
}

// ActivateShop 激活门店外卖渠道
func (c *BMPTakeoutClient) ActivateShop(ctx context.Context, providerName string, shopUuid uint64) error {
	req := &shopApi.ActivateShopReq{
		ProviderName: providerName,
		ShopUuid:     strconv.FormatUint(shopUuid, 10),
		RequestId:    uuid.New().String(),
	}

	resp, err := c.shopClient.ActivateShop(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 ActivateShop 接口失败",
			zap.Error(err),
			zap.String("providerName", providerName),
			zap.Uint64("shopUuid", shopUuid))
		return errors.WithMessage(err, "调用激活门店外卖渠道接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("ActivateShop 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("providerName", providerName),
			zap.Uint64("shopUuid", shopUuid))
		return errors.New(resp.Message)
	}

	return nil
}

// GetOrderInfo 获取订单信息
func (c *BMPTakeoutClient) GetOrderInfo(ctx context.Context, shopUuid string, orderUuid string) (orderData map[string]interface{}, err error) {
	req := &orderApi.GetOrderInfoReq{
		ShopUuid:  shopUuid,
		OrderUuid: orderUuid,
		RequestId: uuid.New().String(),
	}

	resp, err := c.orderClient.GetOrderInfo(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 GetOrderInfo 接口失败",
			zap.Error(err),
			zap.String("shopUuid", shopUuid),
			zap.String("orderUuid", orderUuid))
		return nil, errors.WithMessage(err, "调用获取订单信息接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("GetOrderInfo 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("shopUuid", shopUuid),
			zap.String("orderUuid", orderUuid))
		return nil, errors.New(resp.Message)
	}

	data, err := resp.Data.UnmarshalNew()
	if err != nil {
		logger.Logger.Error("GetOrderInfo 返回数据解析失败",
			zap.Error(err),
			zap.String("shopUuid", shopUuid),
			zap.String("orderUuid", orderUuid))
		return nil, errors.WithMessage(err, "获取订单信息数据解析失败")
	}

	orderResp := data.(*orderApi.GetOrderInfoResp)

	// 解析 raw_data JSON 字符串为 map
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(orderResp.RawData), &rawData); err != nil {
		logger.Logger.Error("解析订单原始数据失败",
			zap.Error(err),
			zap.String("shopUuid", shopUuid),
			zap.String("orderUuid", orderUuid))
		return nil, errors.WithMessage(err, "解析订单原始数据失败")
	}

	return rawData, nil
}

// PrepareOrder 准备订单（接受/拒绝）
// toState: "Accepted" 或 "Rejected"
func (c *BMPTakeoutClient) PrepareOrder(ctx context.Context, takeoutOrderUuid string, toState string) error {
	req := &orderApi.PrepareOrderReq{
		TakeoutOrderUuid: takeoutOrderUuid,
		ToState:          toState,
		RequestId:        uuid.New().String(),
	}

	resp, err := c.orderClient.PrepareOrder(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 PrepareOrder 接口失败",
			zap.Error(err),
			zap.String("takeoutOrderUuid", takeoutOrderUuid),
			zap.String("toState", toState))
		return errors.WithMessage(err, "调用订单接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("PrepareOrder 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("takeoutOrderUuid", takeoutOrderUuid),
			zap.String("toState", toState))
		return errors.New(resp.Message)
	}

	return nil
}

// MarkOrderReady 标记订单准备完成（呼叫骑手）
func (c *BMPTakeoutClient) MarkOrderReady(ctx context.Context, takeoutOrderUuid string) error {
	req := &orderApi.MarkOrderReadyReq{
		TakeoutOrderUuid: takeoutOrderUuid,
		RequestId:        uuid.New().String(),
	}

	resp, err := c.orderClient.MarkOrderReady(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 MarkOrderReady 接口失败", zap.Error(err), zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return errors.WithMessage(err, "调用标记订单准备完成接口失败")
	}

	if resp.Code != "0" {
		// 判断 resp.Message 是否包含 order already marked ready
		if strings.Contains(resp.Message, "order already marked ready") {
			return nil
		}
		if strings.Contains(resp.Message, "since status is in status: ORDER_DELIVERED") {
			return nil
		}
		if strings.Contains(resp.Message, "since status is in status: ORDER_COLLECTED") {
			return nil
		}
		logger.Logger.Warn("MarkOrderReady 返回错误", zap.String("code", resp.Code), zap.String("message", resp.Message), zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return errors.New(resp.Message)
	}

	return nil
}

// CheckOrderCancelable 检查订单是否可取消
func (c *BMPTakeoutClient) CheckOrderCancelable(ctx context.Context, takeoutOrderUuid string) (canCancel bool, reason string, rawData string, err error) {
	req := &orderApi.CheckOrderCancelableReq{
		TakeoutOrderUuid: takeoutOrderUuid,
		RequestId:        uuid.New().String(),
	}

	resp, err := c.orderClient.CheckOrderCancelable(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 CheckOrderCancelable 接口失败", zap.Error(err), zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return false, "", "", errors.WithMessage(err, "调用检查订单可取消接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("CheckOrderCancelable 返回错误",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
			zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return false, "", "", errors.New(resp.Message)
	}

	// 解析响应数据
	var checkResp orderApi.CheckOrderCancelableResp
	if err := resp.Data.UnmarshalTo(&checkResp); err != nil {
		logger.Logger.Error("解析 CheckOrderCancelable 响应失败", zap.Error(err), zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return false, "", "", errors.WithMessage(err, "解析响应数据失败")
	}

	// 如果可以取消，返回提示信息
	if checkResp.CanCancel {
		return true, "", checkResp.RawData, nil
	}

	// 如果不能取消，返回不可取消原因
	return false, checkResp.NonCancellationReason, "", nil
}

// CancelOrder 取消订单
func (c *BMPTakeoutClient) CancelOrder(ctx context.Context, takeoutOrderUuid string, reasonCode string) error {
	req := &orderApi.CancelOrderReq{
		TakeoutOrderUuid: takeoutOrderUuid,
		CancelCode:       reasonCode, // 将取消原因作为取消码传入
		RequestId:        uuid.New().String(),
	}

	resp, err := c.orderClient.CancelOrder(ctx, req)
	if err != nil {
		logger.Logger.Error("调用 CancelOrder 接口失败", zap.Error(err), zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return errors.WithMessage(err, "调用取消订单接口失败")
	}

	if resp.Code != "0" {
		logger.Logger.Warn("CancelOrder 返回错误", zap.String("code", resp.Code), zap.String("message", resp.Message), zap.String("takeoutOrderUuid", takeoutOrderUuid))
		return errors.New(resp.Message)
	}

	return nil
}

// Close 关闭连接

// GetMenuClient 获取菜单服务客户端
func (c *BMPTakeoutClient) GetMenuClient() menuApi.MenuServiceClient {
	return c.menuClient
}

func (c *BMPTakeoutClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
