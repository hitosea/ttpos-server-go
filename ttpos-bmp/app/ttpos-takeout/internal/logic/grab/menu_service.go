package grab

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
)

// MenuService 菜单服务
// 内部使用，通过 sGrab 统一管理
type MenuService struct {
	verifier *SignatureVerifier
}

// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
func (s *MenuService) HandleGetMenu(ctx context.Context, signature, timestamp string, merchantID string) (*grabDto.GetMenuResponse, error) {
	// 注意: GET 请求通常没有 body，签名验证可能仅基于 timestamp
	// 这里假设 Grab 对 GET 请求也要求签名验证

	g.Log().Infof(ctx, "Received GetMenu request for merchant: %s", merchantID)

	// TODO: 从 POS 系统获取菜单数据并转换为 Grab 格式
	// 这里返回示例结构，实际需要对接 POS 菜单服务

	menu := &grabDto.GetMenuResponse{
		MerchantID:        merchantID,
		PartnerMerchantID: merchantID,
		Currency: grabDto.Currency{
			Code:     "THB",
			Symbol:   "฿",
			Exponent: 2,
		},
		SellingTimes: []grabDto.SellingTime{
			{
				ID:   "default",
				Name: "All Day",
				ServiceHours: grabDto.ServiceHours{
					Mon: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
					Tue: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
					Wed: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
					Thu: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
					Fri: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
					Sat: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
					Sun: &grabDto.DayHours{Periods: []grabDto.Period{{StartTime: "00:00", EndTime: "23:59"}}},
				},
			},
		},
		Categories: []grabDto.Category{
			// 实际菜单数据从 POS 获取
		},
	}

	return menu, nil
}

// HandleMenuSyncState 处理菜单同步状态回调
// 使用 SDK grabfood.MenuSyncWebhookRequest
func (s *MenuService) HandleMenuSyncState(ctx context.Context, signature, timestamp string, body []byte) error {
	// 1. 验证签名
	if err := s.verifier.VerifySignature(signature, timestamp, body); err != nil {
		g.Log().Errorf(ctx, "Grab signature verification failed: %v", err)
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 2. 解析请求 - 使用 SDK Model
	var req grabfood.MenuSyncWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		g.Log().Errorf(ctx, "Failed to parse menu sync state request: %v", err)
		return fmt.Errorf("failed to parse request: %w", err)
	}

	requestID := req.GetRequestID()
	status := req.GetStatus()

	// 3. 更新菜单日志状态
	_, err := dao.MenuLog.Ctx(ctx).
		Where(dao.MenuLog.Columns().RequestId, requestID).
		Data(g.Map{
			dao.MenuLog.Columns().Status:    status,
			dao.MenuLog.Columns().UpdatedAt: gtime.Now(),
		}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to update menu log status: %v", err)
		return fmt.Errorf("failed to update menu log: %w", err)
	}

	// 4. 如果失败，记录错误信息
	// SDK 的 Errors 是 []string，直接合并
	if status == grabDto.MenuSyncStatusFail && len(req.GetErrors()) > 0 {
		errMsg := strings.Join(req.GetErrors(), "; ")
		_, _ = dao.MenuLog.Ctx(ctx).
			Where(dao.MenuLog.Columns().RequestId, requestID).
			Data(g.Map{
				dao.MenuLog.Columns().ErrorMsg: errMsg,
			}).Update()
	}

	g.Log().Infof(ctx, "Menu sync state updated: requestId=%s, status=%s", requestID, status)
	return nil
}

// SyncMenu 主动同步菜单到 Grab
func (s *MenuService) SyncMenu(ctx context.Context, merchantID string, menu *grabDto.GetMenuResponse, notifier MenuNotifier) error {
	// 1. 保存菜单快照
	menuSnapshot, _ := json.Marshal(menu)
	logUUID := uuid.New().String()

	logDo := &do.MenuLog{
		Uuid:         logUUID,
		MerchantId:   merchantID,
		ProviderName: "grab",
		SyncType:     "FULL",
		Status:       grabDto.MenuSyncStatusQueued,
		MenuSnapshot: string(menuSnapshot),
		CreatedAt:    gtime.Now(),
		UpdatedAt:    gtime.Now(),
	}

	_, err := dao.MenuLog.Ctx(ctx).Data(logDo).Insert()
	if err != nil {
		return fmt.Errorf("failed to save menu log: %w", err)
	}

	// 2. 调用 Grab API 通知菜单更新
	requestID, err := notifier.NotifyMenuUpdate(ctx, merchantID)
	if err != nil {
		// 更新日志状态为失败
		_, _ = dao.MenuLog.Ctx(ctx).
			Where(dao.MenuLog.Columns().Uuid, logUUID).
			Data(g.Map{
				dao.MenuLog.Columns().Status:    grabDto.MenuSyncStatusFail,
				dao.MenuLog.Columns().ErrorMsg:  err.Error(),
				dao.MenuLog.Columns().UpdatedAt: gtime.Now(),
			}).Update()
		return fmt.Errorf("failed to notify Grab: %w", err)
	}

	// 3. 更新请求 ID
	_, err = dao.MenuLog.Ctx(ctx).
		Where(dao.MenuLog.Columns().Uuid, logUUID).
		Data(g.Map{
			dao.MenuLog.Columns().RequestId: requestID,
			dao.MenuLog.Columns().Status:    grabDto.MenuSyncStatusProcessing,
			dao.MenuLog.Columns().UpdatedAt: gtime.Now(),
		}).Update()
	if err != nil {
		g.Log().Warningf(ctx, "Failed to update menu log with requestID: %v", err)
	}

	g.Log().Infof(ctx, "Menu sync initiated: merchant=%s, requestId=%s", merchantID, requestID)
	return nil
}

// MenuNotifier 菜单通知接口 (仅用于 SyncMenu)
type MenuNotifier interface {
	// NotifyMenuUpdate 通知 Grab 菜单已更新
	NotifyMenuUpdate(ctx context.Context, merchantID string) (requestID string, err error)
}
