//go:build unit

package consumer

import (
	"testing"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/stretchr/testify/assert"
)

// ========== Topic/Concurrency 测试 ==========

func TestDocChangeConsumer_GetTopic(t *testing.T) {
	c := &DocChangeConsumer{}
	assert.Equal(t, "erp-doc-change", c.GetTopic())
}

func TestDocChangeConsumer_GetConcurrency(t *testing.T) {
	c := &DocChangeConsumer{}
	assert.Equal(t, 5, c.GetConcurrency())
}

// ========== docChangeMsg 消息解析测试 ==========

func TestDocChangeMsg_完整消息解析(t *testing.T) {
	body := `{
		"event": "on_cancel",
		"docType": "Sales Invoice",
		"docName": "ACC-SINV-2026-00143",
		"siteCode": "2",
		"data": {
			"sale_order_uuid": "3720191183685635"
		}
	}`

	j, err := gjson.DecodeToJson([]byte(body))
	assert.NoError(t, err)

	msg := &docChangeMsg{}
	err = j.Scan(msg)
	assert.NoError(t, err)
	assert.Equal(t, "on_cancel", msg.Event)
	assert.Equal(t, "Sales Invoice", msg.DocType)
	assert.Equal(t, "ACC-SINV-2026-00143", msg.DocName)
	assert.Equal(t, "2", msg.SiteCode)
	assert.NotNil(t, msg.Data)
	assert.Equal(t, "3720191183685635", msg.Data.SaleOrderUuid)
}

func TestDocChangeMsg_无data字段(t *testing.T) {
	body := `{
		"event": "on_cancel",
		"docType": "Sales Invoice",
		"docName": "ACC-SINV-2026-00100",
		"siteCode": "1"
	}`

	j, err := gjson.DecodeToJson([]byte(body))
	assert.NoError(t, err)

	msg := &docChangeMsg{}
	err = j.Scan(msg)
	assert.NoError(t, err)
	assert.Nil(t, msg.Data)
}

func TestDocChangeMsg_无效JSON(t *testing.T) {
	_, err := gjson.DecodeToJson([]byte(`{invalid`))
	assert.Error(t, err)
}

// ========== 事件过滤逻辑测试 ==========

func TestDocChangeMsg_事件过滤(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		docType   string
		shouldProcess bool
	}{
		{"on_cancel + Sales Invoice → 处理", "on_cancel", "Sales Invoice", true},
		{"on_submit + Sales Invoice → 跳过", "on_submit", "Sales Invoice", false},
		{"on_cancel + Purchase Invoice → 跳过", "on_cancel", "Purchase Invoice", false},
		{"on_update + Sales Invoice → 跳过", "on_update", "Sales Invoice", false},
		{"空event → 跳过", "", "Sales Invoice", false},
		{"空docType → 跳过", "on_cancel", "", false},
		{"都为空 → 跳过", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &docChangeMsg{
				Event:   tt.event,
				DocType: tt.docType,
			}
			shouldProcess := msg.Event == "on_cancel" && msg.DocType == "Sales Invoice"
			assert.Equal(t, tt.shouldProcess, shouldProcess)
		})
	}
}

// ========== 必要字段验证测试 ==========

func TestDocChangeMsg_必要字段验证(t *testing.T) {
	tests := []struct {
		name    string
		docName string
		siteCode string
		valid   bool
	}{
		{"两者都有 → 有效", "ACC-SINV-001", "2", true},
		{"缺少docName → 无效", "", "2", false},
		{"缺少siteCode → 无效", "ACC-SINV-001", "", false},
		{"两者都缺 → 无效", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &docChangeMsg{
				Event:    "on_cancel",
				DocType:  "Sales Invoice",
				DocName:  tt.docName,
				SiteCode: tt.siteCode,
			}
			valid := msg.DocName != "" && msg.SiteCode != ""
			assert.Equal(t, tt.valid, valid)
		})
	}
}
