// Package lineman Lineman API 客户端
package lineman

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// ModifierStatusClient Lineman 修饰符状态更新 API 客户端
type ModifierStatusClient struct {
	endpoint string        // API Endpoint
	timeout  time.Duration // 请求超时时间
}

// NewModifierStatusClient 创建修饰符状态客户端
func NewModifierStatusClient() *ModifierStatusClient {
	cfg := MustConfig(context.Background())
	return &ModifierStatusClient{
		endpoint: cfg.Endpoint,
		timeout:  30 * time.Second,
	}
}

// UpdateModifierStatus 更新修饰符状态
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID（对应 Lineman storeId）
//   - req: 修饰符状态更新请求
//
// 返回:
//   - *dto.ModifierStatusUpdateResp: 响应
//   - error: 错误信息
func (c *ModifierStatusClient) UpdateModifierStatus(
	ctx context.Context,
	storeId string,
	req *dto.ModifierStatusUpdateReq,
) (*dto.ModifierStatusUpdateResp, error) {
	// 1. 从配置获取 partnerId
	cfg := MustConfig(ctx)
	partnerId := cfg.PartnerId
	if partnerId == "" {
		return nil, gerror.New("[ModifierStatusClient] 配置中未设置 partnerId")
	}

	// 2. 获取 Authorization Header
	authHeader, err := c.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "[ModifierStatusClient] 获取 Authorization Header 失败")
	}

	// 3. 构造 URL
	url := fmt.Sprintf("%s/v1/partners/%s/stores/%s/menu/property/values/status",
		c.endpoint, partnerId, storeId)
	g.Log().Infof(ctx, "[ModifierStatusClient] 更新修饰符状态: partnerId=%s, storeId=%s, count=%d",
		partnerId, storeId, len(req.PropertyValues))

	// 4. 发送请求
	client := g.Client().SetTimeout(c.timeout)
	resp, err := client.
		SetHeader("Authorization", authHeader).
		SetHeader("Content-Type", "application/json").
		ContentJson().
		Put(ctx, url, req)

	if err != nil {
		return nil, gerror.Wrapf(err, "[ModifierStatusClient] 请求失败")
	}
	defer resp.Close()

	respBytes := resp.ReadAll()
	g.Log().Debugf(ctx, "[ModifierStatusClient] 响应: status=%d, body=%s",
		resp.StatusCode, string(respBytes))

	// 5. 检查 HTTP 状态码
	if resp.StatusCode != 200 {
		return nil, gerror.Newf("[ModifierStatusClient] API 返回错误: status=%d, body=%s",
			resp.StatusCode, string(respBytes))
	}

	// 6. 解析响应
	var statusResp dto.ModifierStatusUpdateResp
	if err := gjson.Unmarshal(respBytes, &statusResp); err != nil {
		return nil, gerror.Wrapf(err, "[ModifierStatusClient] 响应解析失败")
	}

	// 7. 检查业务状态
	if statusResp.Status != "ok" {
		return nil, gerror.Newf("[ModifierStatusClient] 更新失败: code=%s, msg=%s",
			statusResp.Code, statusResp.Message)
	}

	g.Log().Infof(ctx, "[ModifierStatusClient] 更新成功")
	return &statusResp, nil
}

// UpdateModifierStatusWithRetry 带重试机制的修饰符状态更新
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID
//   - req: 修饰符状态更新请求
//
// 返回:
//   - *dto.ModifierStatusUpdateResp: 响应
//   - error: 错误信息
func (c *ModifierStatusClient) UpdateModifierStatusWithRetry(
	ctx context.Context,
	storeId string,
	req *dto.ModifierStatusUpdateReq,
) (*dto.ModifierStatusUpdateResp, error) {
	var lastResp *dto.ModifierStatusUpdateResp
	var lastErr error

	err := WithRetry(ctx, func() error {
		resp, err := c.UpdateModifierStatus(ctx, storeId, req)
		if err != nil {
			lastErr = err
			return err
		}
		lastResp = resp
		return nil
	})

	if err != nil {
		return nil, lastErr
	}

	return lastResp, nil
}

// getAuthorizationHeader 获取 Authorization Header
func (c *ModifierStatusClient) getAuthorizationHeader(ctx context.Context) (string, error) {
	// 直接调用 OAuthTokenClient
	client := NewOAuthTokenClient()
	return client.GetAuthorizationHeader(ctx)
}
