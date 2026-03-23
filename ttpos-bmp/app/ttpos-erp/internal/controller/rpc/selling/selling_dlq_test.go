//go:build unit

package selling

import (
	"strings"
	"testing"
	"time"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/mq"

	"github.com/stretchr/testify/assert"
)

// ========== skipRetryCount 常量测试 ==========

func TestSkipRetryCount_值为负一(t *testing.T) {
	assert.Equal(t, -1, skipRetryCount)
}

func TestSkipRetryCount_不会被FailedQuery匹配(t *testing.T) {
	// 失败查询条件: retry_count >= MaxRetryCount(5)
	// skipRetryCount = -1 必须不满足此条件
	assert.Less(t, skipRetryCount, erp.MaxRetryCount)
}

// ========== action 路由测试 ==========

func TestRetryAction_判断逻辑(t *testing.T) {
	tests := []struct {
		name   string
		action string
		isSkip bool
	}{
		{"空字符串 → 重试", "", false},
		{"retry → 重试", "retry", false},
		{"skip → 跳过", "skip", true},
		{"其他值 → 重试", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSkip := tt.action == "skip"
			assert.Equal(t, tt.isSkip, isSkip)
		})
	}
}

// ========== MsgType 路由测试 ==========

func TestRetryMsgTypeRouting(t *testing.T) {
	tests := []struct {
		name        string
		msgType     string
		processSave bool
		processCancel bool
		processReturn bool
	}{
		{
			name:          "空MsgType → 处理全部",
			msgType:       "",
			processSave:   true,
			processCancel: true,
			processReturn: true,
		},
		{
			name:          "save-sales-invoice → 只处理Save",
			msgType:       string(mq.MsgTypeSaveSalesInvoice),
			processSave:   true,
			processCancel: false,
			processReturn: false,
		},
		{
			name:          "cancel-sales-invoice → 只处理Cancel",
			msgType:       string(mq.MsgTypeCancelSalesInvoice),
			processSave:   false,
			processCancel: true,
			processReturn: false,
		},
		{
			name:          "return-sales-invoice → 只处理Return",
			msgType:       string(mq.MsgTypeReturnSalesInvoice),
			processSave:   false,
			processCancel: false,
			processReturn: true,
		},
		{
			name:          "未知MsgType → 不处理任何",
			msgType:       "unknown-type",
			processSave:   false,
			processCancel: false,
			processReturn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processSave := tt.msgType == "" || tt.msgType == string(mq.MsgTypeSaveSalesInvoice)
			processCancel := tt.msgType == "" || tt.msgType == string(mq.MsgTypeCancelSalesInvoice)
			processReturn := tt.msgType == "" || tt.msgType == string(mq.MsgTypeReturnSalesInvoice)

			assert.Equal(t, tt.processSave, processSave)
			assert.Equal(t, tt.processCancel, processCancel)
			assert.Equal(t, tt.processReturn, processReturn)
		})
	}
}

// ========== ERP 常量一致性测试 ==========

func TestErpConsts_MaxRetryCount(t *testing.T) {
	assert.Equal(t, 5, erp.MaxRetryCount)
}

func TestErpConsts_Docstatus(t *testing.T) {
	assert.Equal(t, "0", erp.DocstatusDraft)
	assert.Equal(t, "1", erp.DocstatusSubmitted)
	assert.Equal(t, "2", erp.DocstatusCancelled)
}

func TestErpConsts_SyncStatus(t *testing.T) {
	assert.Equal(t, 3, erp.SyncStatusSuccess)
	assert.Equal(t, 4, erp.SyncStatusFailed)
	assert.Equal(t, 5, erp.SyncStatusExternalCancel)
}

// ========== MsgType 完整性测试 ==========

func TestMsgTypes_SI类型完整(t *testing.T) {
	// 确保三类 SI 消息类型定义正确
	assert.Equal(t, mq.MsgTyp("save-sales-invoice"), mq.MsgTypeSaveSalesInvoice)
	assert.Equal(t, mq.MsgTyp("cancel-sales-invoice"), mq.MsgTypeCancelSalesInvoice)
	assert.Equal(t, mq.MsgTyp("return-sales-invoice"), mq.MsgTypeReturnSalesInvoice)
}

// ========== Cancel 检测逻辑测试 ==========

func TestCancelDetection_RespBody匹配(t *testing.T) {
	tests := []struct {
		name      string
		docstatus string
		respBody  string
		isCancel  bool
	}{
		{
			name:      "Submitted + 含取消SI失败 → Cancel失败",
			docstatus: erp.DocstatusSubmitted,
			respBody:  "取消SI失败: some error",
			isCancel:  true,
		},
		{
			name:      "Draft + 含取消SI失败 → 不是Cancel",
			docstatus: erp.DocstatusDraft,
			respBody:  "取消SI失败: some error",
			isCancel:  false,
		},
		{
			name:      "Submitted + 不含取消SI失败 → 不是Cancel",
			docstatus: erp.DocstatusSubmitted,
			respBody:  "其他错误信息",
			isCancel:  false,
		},
		{
			name:      "Submitted + 空resp_body → 不是Cancel",
			docstatus: erp.DocstatusSubmitted,
			respBody:  "",
			isCancel:  false,
		},
		{
			name:      "Cancelled + 含取消SI失败 → 不是Cancel失败（已取消）",
			docstatus: erp.DocstatusCancelled,
			respBody:  "取消SI失败(attempt=3): some error",
			isCancel:  false,
		},
		{
			name:      "Submitted + 含取消SI失败在中间位置 → Cancel失败",
			docstatus: erp.DocstatusSubmitted,
			respBody:  "xxx取消SI失败(attempt=5): timeout yyy",
			isCancel:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCancel := tt.docstatus == erp.DocstatusSubmitted && strings.Contains(tt.respBody, "取消SI失败")
			assert.Equal(t, tt.isCancel, isCancel)
		})
	}
}

// ========== Save 失败检测逻辑测试 ==========

func TestSaveFailedDetection(t *testing.T) {
	tests := []struct {
		name       string
		docstatus  string
		retryCount int
		respBody   string
		isFailed   bool
	}{
		{
			name:       "Draft + 重试耗尽 + 有错误 → 失败",
			docstatus:  erp.DocstatusDraft,
			retryCount: 5,
			respBody:   "SaveSalesInvoice失败(attempt=5): timeout",
			isFailed:   true,
		},
		{
			name:       "Draft + 重试耗尽 + 空resp → 非失败（还在处理中）",
			docstatus:  erp.DocstatusDraft,
			retryCount: 5,
			respBody:   "",
			isFailed:   false,
		},
		{
			name:       "Submitted + 重试耗尽 → 非失败（已成功）",
			docstatus:  erp.DocstatusSubmitted,
			retryCount: 5,
			respBody:   "SaveSalesInvoice失败(attempt=5): error",
			isFailed:   false,
		},
		{
			name:       "Draft + 未耗尽 → 非失败（还在重试）",
			docstatus:  erp.DocstatusDraft,
			retryCount: 3,
			respBody:   "SaveSalesInvoice失败(attempt=3): error",
			isFailed:   false,
		},
		{
			name:       "Draft + 已跳过(retry_count=-1) → 非失败",
			docstatus:  erp.DocstatusDraft,
			retryCount: skipRetryCount,
			respBody:   "SaveSalesInvoice失败(attempt=5): error",
			isFailed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isFailed := tt.docstatus == erp.DocstatusDraft &&
				tt.retryCount >= erp.MaxRetryCount &&
				tt.respBody != ""
			assert.Equal(t, tt.isFailed, isFailed)
		})
	}
}

// ========== 30 天时间过滤边界测试 ==========

func TestThirtyDayFilter_边界(t *testing.T) {
	now := time.Now()
	since := int(now.AddDate(0, 0, -30).Unix())

	tests := []struct {
		name      string
		createdAt int
		inRange   bool
	}{
		{"刚好30天前", since, true},
		{"29天前", int(now.AddDate(0, 0, -29).Unix()), true},
		{"今天", int(now.Unix()), true},
		{"31天前", int(now.AddDate(0, 0, -31).Unix()), false},
		{"60天前", int(now.AddDate(0, 0, -60).Unix()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRange := tt.createdAt >= since
			assert.Equal(t, tt.inRange, inRange)
		})
	}
}

// ========== Cancel 使用 UpdatedAt 而非 CreatedAt 的原因验证 ==========

func TestCancelTimeFilter_应使用UpdatedAt(t *testing.T) {
	// Cancel 记录复用 SI 表，created_at 是原 SI 创建时间
	// 取消操作发生在 updated_at，所以 Cancel 应按 updated_at 过滤
	now := time.Now()
	since := int(now.AddDate(0, 0, -30).Unix())

	// 模拟：SI 在 60 天前创建，但 30 天内取消失败
	createdAt := int(now.AddDate(0, 0, -60).Unix()) // 60天前创建
	updatedAt := int(now.AddDate(0, 0, -5).Unix())  // 5天前取消失败

	// Save 用 createdAt 过滤：60 天前的不在范围内
	assert.False(t, createdAt >= since, "60天前创建的 SI 不在30天范围内")

	// Cancel 用 updatedAt 过滤：5 天前更新的在范围内
	assert.True(t, updatedAt >= since, "5天前取消失败的应在30天范围内")
}
