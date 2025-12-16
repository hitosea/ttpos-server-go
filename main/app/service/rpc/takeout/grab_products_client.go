package takeout

import (
	"context"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// TODO: 待对齐 bmp 的 Grab 商品拉取 RPC 接口

// GrabProductRPCItem 假定的 Grab 商品结构（字段需与 bmp 对齐）
type GrabProductRPCItem struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Category   string            `json:"category"`
	Price      int64             `json:"price"`
	Attributes map[string]string `json:"attributes"`
	Unit       string            `json:"unit"`
	RawPayload interface{}       `json:"rawPayload"`
}

// GrabProductListResponse 假定的分页响应
type GrabProductListResponse struct {
	Items    []GrabProductRPCItem `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
	HasMore  bool                 `json:"hasMore"`
}

// GrabProductClient 占位客户端
type GrabProductClient struct {
	// 待补充真实 rpc client
}

// NewGrabProductClient 创建占位客户端
func NewGrabProductClient() *GrabProductClient {
	return &GrabProductClient{}
}

// GetGrabProducts 拉取 Grab 商品列表（占位实现）
func (c *GrabProductClient) GetGrabProducts(ctx context.Context, page int, pageSize int) (*GrabProductListResponse, error) {
	logger.Logger.Warn("GetGrabProducts 调用占位实现，需对齐 bmp RPC 接口", zap.Int("page", page), zap.Int("pageSize", pageSize))
	return nil, errors.New("GetGrabProducts RPC 尚未实现，对齐后补充")
}
