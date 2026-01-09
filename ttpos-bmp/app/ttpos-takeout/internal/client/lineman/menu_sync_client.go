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
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// MenuSyncClient Lineman 菜单同步 API 客户端
type MenuSyncClient struct {
	endpoint string        // API Endpoint
	timeout  time.Duration // 请求超时时间
}

// NewMenuSyncClient 创建菜单同步客户端
func NewMenuSyncClient() *MenuSyncClient {
	cfg := MustConfig(context.Background())
	return &MenuSyncClient{
		endpoint: cfg.Endpoint,
		timeout:  30 * time.Second,
	}
}

// SyncMenu 同步菜单到 Lineman
func (c *MenuSyncClient) SyncMenu(
	ctx context.Context,
	storeId string,
	menuData *dto.MenuSyncRequest,
) (*dto.MenuSyncResponse, error) {
	// 1. 从配置获取 partnerId
	cfg := MustConfig(ctx)
	partnerId := cfg.PartnerId
	if partnerId == "" {
		return nil, gerror.New("[LinemanClient] 配置中未设置 partnerId")
	}

	// 2. 获取 Authorization Header
	authHeader, err := c.getAuthorizationHeader(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "[LinemanClient] 获取 Authorization Header 失败")
	}

	url := fmt.Sprintf("%s/v2/partners/%s/stores/%s/menus", c.endpoint, partnerId, storeId)
	g.Log().Infof(ctx, "[LinemanClient] 开始同步菜单: partnerId=%s, storeId=%s", partnerId, storeId)

	client := g.Client().SetTimeout(c.timeout)
	resp, err := client.
		SetHeader("Authorization", authHeader).
		SetHeader("Content-Type", "application/json").
		ContentJson().
		Put(ctx, url, menuData)

	if err != nil {
		return nil, gerror.Wrapf(err, "[LinemanClient] 请求失败")
	}
	defer resp.Close()

	respBytes := resp.ReadAll()
	g.Log().Debugf(ctx, "[LinemanClient] 响应: status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		return nil, gerror.Newf("[LinemanClient] API 返回错误: status=%d", resp.StatusCode)
	}

	var syncResp dto.MenuSyncResponse
	if err := gjson.Unmarshal(respBytes, &syncResp); err != nil {
		return nil, gerror.Wrapf(err, "[LinemanClient] 响应解析失败")
	}

	if syncResp.Status != "ok" {
		return nil, gerror.Newf("[LinemanClient] 同步失败: code=%s, msg=%s", syncResp.Code, syncResp.Message)
	}

	g.Log().Infof(ctx, "[LinemanClient] 同步成功: requestId=%s", syncResp.MenuSyncRequestId)

	return &syncResp, nil
}

// SyncMenuWithRetry 带重试机制的菜单同步
func (c *MenuSyncClient) SyncMenuWithRetry(
	ctx context.Context,
	storeId string,
	menuData *dto.MenuSyncRequest,
) (*dto.MenuSyncResponse, error) {
	var lastResp *dto.MenuSyncResponse
	var lastErr error

	err := WithRetry(ctx, func() error {
		resp, err := c.SyncMenu(ctx, storeId, menuData)
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
func (c *MenuSyncClient) getAuthorizationHeader(ctx context.Context) (string, error) {
	// 调用 Lineman Service 获取 Authorization Header
	return service.Lineman().GetAuthorizationHeader(ctx)
}
