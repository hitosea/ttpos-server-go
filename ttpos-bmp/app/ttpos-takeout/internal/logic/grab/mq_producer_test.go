package grab

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestNewRocketMQProducer 测试 RocketMQ 生产者创建
func TestNewRocketMQProducer(t *testing.T) {
	producer := NewRocketMQProducer()
	if producer == nil {
		t.Fatal("NewRocketMQProducer returned nil")
	}
}

// TestNewNoopMQProducer 测试空操作 MQ 生产者创建
func TestNewNoopMQProducer(t *testing.T) {
	producer := NewNoopMQProducer()
	if producer == nil {
		t.Fatal("NewNoopMQProducer returned nil")
	}
}

// TestNoopMQProducer_SendOrderEvent 测试空操作生产者发送消息
func TestNoopMQProducer_SendOrderEvent(t *testing.T) {
	producer := NewNoopMQProducer()
	ctx := context.Background()

	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    "test-uuid-123",
		OrderID:      "G-123456",
		MerchantID:   "M-001",
		Status:       "PENDING",
		Timestamp:    time.Now().Unix(),
	}

	err := producer.SendOrderEvent(ctx, event)
	if err != nil {
		t.Errorf("NoopMQProducer.SendOrderEvent() error = %v, want nil", err)
	}
}

// TestRocketMQProducer_SendOrderEvent 测试 RocketMQ 生产者发送消息
func TestRocketMQProducer_SendOrderEvent(t *testing.T) {
	producer := NewRocketMQProducer()
	ctx := context.Background()

	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    "test-uuid-456",
		OrderID:      "G-789012",
		MerchantID:   "M-002",
		Status:       "ACCEPTED",
		Timestamp:    time.Now().Unix(),
	}

	// 当前实现是 TODO 状态，不会返回错误
	err := producer.SendOrderEvent(ctx, event)
	if err != nil {
		t.Errorf("RocketMQProducer.SendOrderEvent() error = %v, want nil", err)
	}
}

// TestOrderEvent_JSON 测试订单事件 JSON 序列化
func TestOrderEvent_JSON(t *testing.T) {
	timestamp := time.Now().Unix()
	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    "uuid-123-456",
		OrderID:      "G-ORDER-001",
		MerchantID:   "M-MERCHANT-001",
		Status:       "PENDING",
		Timestamp:    timestamp,
	}

	// 序列化
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// 验证 JSON 字段
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	expectedFields := []string{"action", "providerName", "orderUuid", "orderId", "merchantId", "status", "timestamp"}
	for _, field := range expectedFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("JSON missing field: %s", field)
		}
	}

	// 验证字段值
	if decoded["action"] != "create" {
		t.Errorf("action = %v, want create", decoded["action"])
	}
	if decoded["providerName"] != "grab" {
		t.Errorf("providerName = %v, want grab", decoded["providerName"])
	}
}

// TestOrderEvent_Actions 测试不同订单事件类型
func TestOrderEvent_Actions(t *testing.T) {
	actions := []string{"create", "status_update", "cancel"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			event := &OrderEvent{
				Action:       action,
				ProviderName: "grab",
				OrderUUID:    "uuid-123",
				OrderID:      "G-123",
				MerchantID:   "M-001",
				Status:       "PENDING",
				Timestamp:    time.Now().Unix(),
			}

			if event.Action != action {
				t.Errorf("Action = %s, want %s", event.Action, action)
			}
		})
	}
}

// TestTopicGrabOrder_Constant 测试 Topic 常量
func TestTopicGrabOrder_Constant(t *testing.T) {
	if TopicGrabOrder == "" {
		t.Error("TopicGrabOrder should not be empty")
	}
	if TopicGrabOrder != "takeout_grab_order" {
		t.Errorf("TopicGrabOrder = %s, want takeout_grab_order", TopicGrabOrder)
	}
}

// TestMQProducer_Interface 测试 MQProducer 接口实现
func TestMQProducer_Interface(t *testing.T) {
	// 验证 RocketMQProducer 实现 MQProducer 接口
	var _ MQProducer = (*RocketMQProducer)(nil)

	// 验证 NoopMQProducer 实现 MQProducer 接口
	var _ MQProducer = (*NoopMQProducer)(nil)
}

// TestOrderEvent_StatusUpdate 测试状态更新事件
func TestOrderEvent_StatusUpdate(t *testing.T) {
	event := &OrderEvent{
		Action:       "status_update",
		ProviderName: "grab",
		OrderUUID:    "uuid-status-update",
		OrderID:      "G-STATUS-001",
		MerchantID:   "M-001",
		Status:       "ACCEPTED",
		Timestamp:    time.Now().Unix(),
	}

	if event.Action != "status_update" {
		t.Errorf("Action = %s, want status_update", event.Action)
	}
	if event.Status != "ACCEPTED" {
		t.Errorf("Status = %s, want ACCEPTED", event.Status)
	}
}

// BenchmarkOrderEvent_JSON 订单事件 JSON 序列化性能测试
func BenchmarkOrderEvent_JSON(b *testing.B) {
	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    "uuid-123-456-789",
		OrderID:      "G-ORDER-BENCHMARK",
		MerchantID:   "M-MERCHANT-BENCH",
		Status:       "PENDING",
		Timestamp:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(event)
	}
}

// BenchmarkNoopMQProducer_SendOrderEvent 空操作生产者性能测试
func BenchmarkNoopMQProducer_SendOrderEvent(b *testing.B) {
	producer := NewNoopMQProducer()
	ctx := context.Background()
	event := &OrderEvent{
		Action:       "create",
		ProviderName: "grab",
		OrderUUID:    "uuid-bench",
		OrderID:      "G-BENCH",
		MerchantID:   "M-BENCH",
		Status:       "PENDING",
		Timestamp:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = producer.SendOrderEvent(ctx, event)
	}
}
