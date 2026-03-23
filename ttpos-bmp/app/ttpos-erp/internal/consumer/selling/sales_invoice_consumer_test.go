//go:build unit

package selling

import (
	"testing"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	erp "ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/model/mq"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// ========== maxRetryCount 常量测试 ==========

func TestMaxRetryCount_值为5(t *testing.T) {
	assert.Equal(t, 5, maxRetryCount)
}

// ========== Consumer Topic/Concurrency 测试 ==========

func TestSaveSalesInvoiceConsumer_GetTopic(t *testing.T) {
	c := &SaveSalesInvoiceConsumer{}
	assert.Equal(t, "save-sales-invoice", c.GetTopic())
}

func TestSaveSalesInvoiceConsumer_GetConcurrency(t *testing.T) {
	c := &SaveSalesInvoiceConsumer{}
	assert.Equal(t, 10, c.GetConcurrency())
}

func TestCancelSalesInvoiceConsumer_GetTopic(t *testing.T) {
	c := &CancelSalesInvoiceConsumer{}
	assert.Equal(t, "cancel-sales-invoice", c.GetTopic())
}

func TestReturnSalesInvoiceConsumer_GetTopic(t *testing.T) {
	c := &ReturnSalesInvoiceConsumer{}
	assert.Equal(t, "return-sales-invoice", c.GetTopic())
}

// ========== AsyncSalesInvoiceMsg 消息解析测试 ==========

func TestAsyncSalesInvoiceMsg_JSON解析(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *mq.AsyncSalesInvoiceMsg
		wantErr bool
	}{
		{
			name: "完整消息",
			json: `{"record_id":130,"msg_type":"save-sales-invoice","site_code":"2"}`,
			want: &mq.AsyncSalesInvoiceMsg{
				RecordId: 130,
				MsgType:  mq.MsgTypeSaveSalesInvoice,
				SiteCode: "2",
			},
		},
		{
			name: "仅必填字段",
			json: `{"msg_type":"cancel-sales-invoice"}`,
			want: &mq.AsyncSalesInvoiceMsg{
				MsgType: mq.MsgTypeCancelSalesInvoice,
			},
		},
		{
			name: "return类型",
			json: `{"record_id":99,"msg_type":"return-sales-invoice","site_code":"5"}`,
			want: &mq.AsyncSalesInvoiceMsg{
				RecordId: 99,
				MsgType:  mq.MsgTypeReturnSalesInvoice,
				SiteCode: "5",
			},
		},
		{
			name:    "无效JSON",
			json:    `{invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j, err := gjson.DecodeToJson([]byte(tt.json))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			msg := &mq.AsyncSalesInvoiceMsg{}
			err = j.Scan(msg)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.RecordId, msg.RecordId)
			assert.Equal(t, tt.want.MsgType, msg.MsgType)
			assert.Equal(t, tt.want.SiteCode, msg.SiteCode)
		})
	}
}

// ========== ExtractReqMetadata 测试 ==========

func TestExtractReqMetadata_正常提取(t *testing.T) {
	// 构建一个 protobuf 序列化的请求
	req := &selling.SaveSalesInvoiceReq{
		CompanyUuid: "8267304538112000",
		OrderType:   "sale_order",
		OrderNo:     "202603191177393082",
	}
	buf, err := proto.Marshal(req)
	assert.NoError(t, err)

	record := &entity.ReceiveSalesInvoice{
		ReqMessage: gbase64.EncodeToString(buf),
	}

	companyUuid, orderType := ExtractReqMetadata(record)
	assert.Equal(t, "8267304538112000", companyUuid)
	assert.Equal(t, "sale_order", orderType)
}

func TestExtractReqMetadata_充值订单(t *testing.T) {
	req := &selling.SaveSalesInvoiceReq{
		CompanyUuid: "1234567890",
		OrderType:   "recharge",
	}
	buf, _ := proto.Marshal(req)

	record := &entity.ReceiveSalesInvoice{
		ReqMessage: gbase64.EncodeToString(buf),
	}

	companyUuid, orderType := ExtractReqMetadata(record)
	assert.Equal(t, "1234567890", companyUuid)
	assert.Equal(t, "recharge", orderType)
}

func TestExtractReqMetadata_外卖订单(t *testing.T) {
	req := &selling.SaveSalesInvoiceReq{
		CompanyUuid: "9999",
		OrderType:   "takeout",
	}
	buf, _ := proto.Marshal(req)

	record := &entity.ReceiveSalesInvoice{
		ReqMessage: gbase64.EncodeToString(buf),
	}

	companyUuid, orderType := ExtractReqMetadata(record)
	assert.Equal(t, "9999", companyUuid)
	assert.Equal(t, "takeout", orderType)
}

func TestExtractReqMetadata_nil记录(t *testing.T) {
	companyUuid, orderType := ExtractReqMetadata(nil)
	assert.Empty(t, companyUuid)
	assert.Empty(t, orderType)
}

func TestExtractReqMetadata_空ReqMessage(t *testing.T) {
	record := &entity.ReceiveSalesInvoice{ReqMessage: ""}
	companyUuid, orderType := ExtractReqMetadata(record)
	assert.Empty(t, companyUuid)
	assert.Empty(t, orderType)
}

func TestExtractReqMetadata_无效base64(t *testing.T) {
	record := &entity.ReceiveSalesInvoice{ReqMessage: "!!!invalid-base64!!!"}
	companyUuid, orderType := ExtractReqMetadata(record)
	assert.Empty(t, companyUuid)
	assert.Empty(t, orderType)
}

func TestExtractReqMetadata_无效protobuf(t *testing.T) {
	record := &entity.ReceiveSalesInvoice{
		ReqMessage: gbase64.EncodeToString([]byte("not-protobuf-data")),
	}
	// protobuf 对无效数据容错性很强，不一定报错，但不应 panic
	companyUuid, orderType := ExtractReqMetadata(record)
	_ = companyUuid
	_ = orderType
}

// ========== MsgTyp 常量测试 ==========

func TestMsgTypes_值正确(t *testing.T) {
	assert.Equal(t, mq.MsgTyp("save-sales-invoice"), mq.MsgTypeSaveSalesInvoice)
	assert.Equal(t, mq.MsgTyp("cancel-sales-invoice"), mq.MsgTypeCancelSalesInvoice)
	assert.Equal(t, mq.MsgTyp("return-sales-invoice"), mq.MsgTypeReturnSalesInvoice)
}

// ========== MsgTyp 与 Topic 名称一致性测试 ==========

func TestMsgType_与Topic字符串匹配(t *testing.T) {
	// 确保 MsgType 和 Topic 名称一致，防止 retry 时推送到错误 topic
	assert.Equal(t, string(mq.MsgTypeSaveSalesInvoice), "save-sales-invoice")
	assert.Equal(t, string(mq.MsgTypeCancelSalesInvoice), "cancel-sales-invoice")
	assert.Equal(t, string(mq.MsgTypeReturnSalesInvoice), "return-sales-invoice")
}

// ========== Docstatus 常量测试 ==========

func TestDocstatus_常量值(t *testing.T) {
	assert.Equal(t, "0", erp.DocstatusDraft)
	assert.Equal(t, "1", erp.DocstatusSubmitted)
	assert.Equal(t, "2", erp.DocstatusCancelled)
}

// ========== MaxRetryCount 测试 ==========

func TestMaxRetryCount_与BMP常量一致(t *testing.T) {
	assert.Equal(t, 5, erp.MaxRetryCount)
	assert.Equal(t, erp.MaxRetryCount, maxRetryCount)
}

// ========== SyncStatus 常量测试 ==========

func TestSyncStatus_常量值(t *testing.T) {
	assert.Equal(t, 3, erp.SyncStatusSuccess)
	assert.Equal(t, 4, erp.SyncStatusFailed)
	assert.Equal(t, 5, erp.SyncStatusExternalCancel)
}

// ========== Consumer 配置一致性测试 ==========

func TestAllConsumers_Concurrency一致(t *testing.T) {
	// 三个 SI consumer 应有相同并发度
	save := &SaveSalesInvoiceConsumer{}
	cancel := &CancelSalesInvoiceConsumer{}
	ret := &ReturnSalesInvoiceConsumer{}

	assert.Equal(t, save.GetConcurrency(), cancel.GetConcurrency())
	assert.Equal(t, save.GetConcurrency(), ret.GetConcurrency())
}

// ========== ExtractReqMetadata 更多边界测试 ==========

func TestExtractReqMetadata_默认OrderType(t *testing.T) {
	// OrderType 为空时仍应正常提取 CompanyUuid
	req := &selling.SaveSalesInvoiceReq{
		CompanyUuid: "8267304538112000",
		// OrderType 不设置，默认 ""
	}
	buf, _ := proto.Marshal(req)

	record := &entity.ReceiveSalesInvoice{
		ReqMessage: gbase64.EncodeToString(buf),
	}

	companyUuid, orderType := ExtractReqMetadata(record)
	assert.Equal(t, "8267304538112000", companyUuid)
	assert.Empty(t, orderType) // 默认值为空字符串
}

// ========== AsyncSalesInvoiceMsg 序列化往返测试 ==========

func TestAsyncSalesInvoiceMsg_序列化往返(t *testing.T) {
	original := &mq.AsyncSalesInvoiceMsg{
		RecordId: 42,
		MsgType:  mq.MsgTypeSaveSalesInvoice,
		SiteCode: "site-001",
	}

	data, err := gjson.Encode(original)
	assert.NoError(t, err)

	j, err := gjson.DecodeToJson(data)
	assert.NoError(t, err)

	decoded := &mq.AsyncSalesInvoiceMsg{}
	err = j.Scan(decoded)
	assert.NoError(t, err)
	assert.Equal(t, original.RecordId, decoded.RecordId)
	assert.Equal(t, original.MsgType, decoded.MsgType)
	assert.Equal(t, original.SiteCode, decoded.SiteCode)
}

// ========== Topic 不重复测试 ==========

func TestTopics_不重复(t *testing.T) {
	topics := []string{
		(&SaveSalesInvoiceConsumer{}).GetTopic(),
		(&CancelSalesInvoiceConsumer{}).GetTopic(),
		(&ReturnSalesInvoiceConsumer{}).GetTopic(),
	}

	seen := make(map[string]bool)
	for _, topic := range topics {
		assert.False(t, seen[topic], "Topic 重复: %s", topic)
		seen[topic] = true
	}
}
