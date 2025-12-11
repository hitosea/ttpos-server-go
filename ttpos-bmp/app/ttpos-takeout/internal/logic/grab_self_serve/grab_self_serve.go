// Package grab_self_serve 提供 GrabFood 自助激活链接服务的业务逻辑
package grab_self_serve

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	grabLogic "ttpos-bmp/app/ttpos-takeout/internal/logic/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// sGrabSelfServe 自助激活链接服务
type sGrabSelfServe struct {
	sdkWrapper *grabLogic.SDKWrapper
}

func init() {
	service.RegisterGrabSelfServe(New())
}

// New 创建自助激活链接服务实例
func New() *sGrabSelfServe {
	return &sGrabSelfServe{}
}

// getSdkWrapper 获取 SDK Wrapper (懒加载)
func (s *sGrabSelfServe) getSdkWrapper() *grabLogic.SDKWrapper {
	if s.sdkWrapper == nil {
		conf := service.Grab().MustConf()
		s.sdkWrapper = grabLogic.NewSDKWrapper(&grabLogic.SDKConfig{
			ClientID:     conf.ClientID,
			ClientSecret: conf.ClientSecret,
			Environment:  conf.Environment,
		})
	}
	return s.sdkWrapper
}

// CreateSelfServeJourney 创建自助激活链接
// 根据 shop_uuid 获取 Grab 配置，调用 SDK 生成激活链接
func (s *sGrabSelfServe) CreateSelfServeJourney(ctx context.Context, req *grab.CreateSelfServeJourneyReq) (*grab.CreateSelfServeJourneyResp, error) {
	// 1. 参数校验
	if req.ProviderName == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "provider_name 不能为空")
	}
	if req.ShopUuid == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "shop_uuid 不能为空")
	}

	// 2. 验证 provider_name 是否为 grab
	if req.ProviderName != string(consts.ProviderGrab) {
		return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "不支持的 provider_name: %s，当前仅支持 grab", req.ProviderName)
	}

	// 3. 获取 Grab 配置（包含 Environment, Credentials）
	conf := service.Grab().MustConf()
	if conf == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "Grab 平台配置未加载，请检查配置文件 app.provider.grab.platform")
	}
	if conf.ClientID == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "Grab 配置不完整：缺少 ClientID，请检查配置文件 app.provider.grab.platform.clientId")
	}
	if conf.ClientSecret == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "Grab 配置不完整：缺少 ClientSecret，请检查配置文件 app.provider.grab.platform.clientSecret")
	}
	if conf.Environment == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "Grab 配置不完整：缺少 Environment，请检查配置文件 app.provider.grab.platform.environment")
	}

	// 4. 获取或创建 SDK Wrapper
	sdkWrapper := s.getSdkWrapper()

	// 5. 调用 SDK 创建自助激活链接
	// 注意：shop_uuid 作为 merchantID 传入，实际应该根据业务需求从数据库查询 Grab Merchant ID
	// 当前实现假设 shop_uuid 就是 Grab Merchant ID
	activationURL, requestID, err := sdkWrapper.CreateSelfServeJourney(ctx, req.ShopUuid)
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] CreateSelfServeJourney failed: shop_uuid=%s, error=%v", req.ShopUuid, err)
		// 错误映射：根据错误类型返回不同的错误码
		return nil, s.mapGrabError(err)
	}

	// 5.1 旅程创建成功，落库 shop_provider_cfg（状态 SYNCING）
	shopUUID := g.NewVar(req.ShopUuid).Uint64()
	if shopUUID > 0 {
		if upsertErr := service.ShopProviderCfg().UpsertShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab), "", consts.ProviderShopStatusSyncing); upsertErr != nil {
			// 记录错误但不中断流程，旅程创建已成功
			g.Log().Warningf(ctx, "[Grab] CreateSelfServeJourney upsert shop_provider_cfg failed: shop_uuid=%d, error=%v", shopUUID, upsertErr)
		}
	}

	// 6. 构建响应
	resp := &grab.CreateSelfServeJourneyResp{
		ProviderName: req.ProviderName,
		SelfServeUrl: activationURL,
		RequestId:    requestID,
	}

	// 如果请求中没有 request_id，使用 SDK 返回的 request_id
	if req.RequestId == "" && requestID != "" {
		resp.RequestId = requestID
	} else if req.RequestId != "" {
		resp.RequestId = req.RequestId
	}

	g.Log().Infof(ctx, "[Grab] CreateSelfServeJourney success: shop_uuid=%s, request_id=%s", req.ShopUuid, resp.RequestId)
	return resp, nil
}

// mapGrabError 映射 Grab SDK 错误到业务错误
func (s *sGrabSelfServe) mapGrabError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	// 授权错误
	if contains(errStr, "401") || contains(errStr, "unauthorized") || contains(errStr, "authentication") {
		return gerror.NewCode(gcode.CodeNotAuthorized, fmt.Sprintf("Grab 授权失败: %v", err))
	}

	// 参数错误
	if contains(errStr, "400") || contains(errStr, "bad request") || contains(errStr, "invalid") {
		return gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("Grab 参数错误: %v", err))
	}

	// 网络超时
	if contains(errStr, "timeout") || contains(errStr, "deadline exceeded") {
		return gerror.NewCode(gcode.CodeInternalError, fmt.Sprintf("Grab API 调用超时: %v", err))
	}

	// 其他错误
	return gerror.Wrap(err, "Grab API 调用失败")
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
