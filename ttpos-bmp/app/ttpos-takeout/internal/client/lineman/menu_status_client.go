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

// MenuStatusClient Lineman 菜单状态更新 API 客户端
type MenuStatusClient struct {
	endpoint string        // API Endpoint
	timeout  time.Duration // 请求超时时间
}

// NewMenuStatusClient 创建菜单状态更新客户端
func NewMenuStatusClient() *MenuStatusClient {
	cfg := MustConfig(context.Background())
	return &MenuStatusClient{
		endpoint: cfg.Endpoint,
		timeout:  30 * time.Second,
	}
}

// UpdateMenuStatus 更新菜单商品状态
// 参考: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633
//
// API: PUT /v1/partners/{partnerId}/stores/{storeId}/menu/items/status
//
// 参数:
//   - ctx: 上下文
//   - storeId: 店铺 ID
//   - req: 菜单状态更新请求
//
// 返回:
//   - resp: 菜单状态更新响应
//   - err: 错误信息
func (c *MenuStatusClient) UpdateMenuStatus(
	ctx context.Context,
	storeId string,
	req *dto.MenuStatusUpdateReq,
) (*dto.MenuStatusUpdateResp, error) {
	// 1. 参数校验
	if storeId == "" {
		return nil, gerror.New("[LinemanMenuStatusClient] storeId 不能为空")
	}
	if req == nil || len(req.MenuItems) == 0 {
		return nil, gerror.New("[LinemanMenuStatusClient] menuItems 不能为空")
	}
	if len(req.MenuItems) > 100 {
		return nil, gerror.New("[LinemanMenuStatusClient] menuItems 最多支持 100 个商品")
	}

	// 2. 从配置获取 partnerId
	cfg := MustConfig(ctx)
	partnerId := cfg.PartnerId
	if partnerId == "" {
		return nil, gerror.New("[LinemanMenuStatusClient] 配置中未设置 partnerId")
	}

	// 3. 获取 Authorization Header
	authHeader, err := c.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "[LinemanMenuStatusClient] 获取 Authorization Header 失败")
	}

	// 4. 构造请求 URL
	url := fmt.Sprintf("%s/v1/partners/%s/stores/%s/menu/items/status", c.endpoint, partnerId, storeId)
	g.Log().Infof(ctx, "[LinemanMenuStatusClient] 开始更新菜单状态: partnerId=%s, storeId=%s, itemCount=%d",
		partnerId, storeId, len(req.MenuItems))

	// 5. 发送 HTTP 请求
	client := g.Client().SetTimeout(c.timeout)
	resp, err := client.
		SetHeader("Authorization", authHeader).
		SetHeader("Content-Type", "application/json").
		ContentJson().
		Put(ctx, url, req)

	if err != nil {
		return nil, gerror.Wrapf(err, "[LinemanMenuStatusClient] 请求失败")
	}
	defer resp.Close()

	respBytes := resp.ReadAll()
	g.Log().Debugf(ctx, "[LinemanMenuStatusClient] 响应: status=%d, body=%s", resp.StatusCode, string(respBytes))

	// 6. 检查 HTTP 状态码（2xx 都视为成功，如 200 OK、201 Created）
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, gerror.Newf("[LinemanMenuStatusClient] API 返回错误: status=%d, body=%s",
			resp.StatusCode, string(respBytes))
	}

	// 7. 解析响应
	var statusResp dto.MenuStatusUpdateResp
	if err := gjson.Unmarshal(respBytes, &statusResp); err != nil {
		return nil, gerror.Wrapf(err, "[LinemanMenuStatusClient] 响应解析失败")
	}

	// 8. 检查业务响应
	if statusResp.Status != "ok" {
		return nil, gerror.Newf("[LinemanMenuStatusClient] 更新失败: code=%s, msg=%s",
			statusResp.Code, statusResp.Message)
	}

	g.Log().Infof(ctx, "[LinemanMenuStatusClient] 更新成功: itemCount=%d", len(req.MenuItems))

	return &statusResp, nil
}

// UpdateMenuStatusWithRetry 带重试机制的菜单状态更新
func (c *MenuStatusClient) UpdateMenuStatusWithRetry(
	ctx context.Context,
	storeId string,
	req *dto.MenuStatusUpdateReq,
) (*dto.MenuStatusUpdateResp, error) {
	var lastResp *dto.MenuStatusUpdateResp
	var lastErr error

	err := WithRetry(ctx, func() error {
		resp, err := c.UpdateMenuStatus(ctx, storeId, req)
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
func (c *MenuStatusClient) getAuthorizationHeader(ctx context.Context) (string, error) {
	// 直接调用 OAuthTokenClient
	client := NewOAuthTokenClient()
	return client.GetAuthorizationHeader(ctx)
}
