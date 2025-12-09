package grab

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// StoreService 门店服务
// 内部使用，通过 sGrab 统一管理
type StoreService struct {
	verifier *SignatureVerifier
}

// HandleIntegrationStatus 处理门店集成状态回调
// 使用 SDK grabfood.PushIntegrationStatusWebhookRequest
func (s *StoreService) HandleIntegrationStatus(ctx context.Context, signature, timestamp string, body []byte) error {
	// 1. 验证签名
	if err := s.verifier.VerifySignature(signature, timestamp, body); err != nil {
		g.Log().Errorf(ctx, "Grab signature verification failed: %v", err)
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 2. 解析请求 - 使用 SDK Model
	var req grabfood.PushIntegrationStatusWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		g.Log().Errorf(ctx, "Failed to parse integration status request: %v", err)
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// 3. 记录状态变更
	g.Log().Infof(ctx, "Store integration status changed: merchant=%s, status=%s",
		req.GetPartnerMerchantID(), req.GetIntegrationStatus())

	// TODO: 可以保存到数据库或触发业务流程
	// 例如：当状态变为 INACTIVE 时，通知商户

	return nil
}

// 门店暂停/恢复功能已迁移到 sGrab，直接使用 SDKWrapper
