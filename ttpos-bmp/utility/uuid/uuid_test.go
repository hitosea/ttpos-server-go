package uuid

import (
	"testing"

	"github.com/gogf/gf/v2/os/gctx"
)

// TestNodeIdCalculation 测试节点 ID 计算逻辑
func TestNodeIdCalculation(t *testing.T) {
	tests := []struct {
		name       string
		appType    uint32
		instanceID uint32
		expected   uint32
	}{
		{"ERP实例1", AppTypeERP, 1, 65},          // (1 << 6) | 1 = 65
		{"ERP实例2", AppTypeERP, 2, 66},          // (1 << 6) | 2 = 66
		{"Message实例1", AppTypeMessage, 1, 129}, // (2 << 6) | 1 = 129
		{"Message实例2", AppTypeMessage, 2, 130}, // (2 << 6) | 2 = 130
		{"Manager实例1", AppTypeManager, 1, 193}, // (3 << 6) | 1 = 193
		{"最大值", 15, 63, 1023},                  // (15 << 6) | 63 = 1023
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeID := (tt.appType << instanceIDBits) | tt.instanceID
			if nodeID != tt.expected {
				t.Errorf("节点ID计算错误: appType=%d, instanceID=%d, got=%d, want=%d",
					tt.appType, tt.instanceID, nodeID, tt.expected)
			}
		})
	}
}

// TestIdUniquenessWithinApp 测试同一应用内 ID 唯一性
func TestIdUniquenessWithinApp(t *testing.T) {
	ctx := gctx.New()
	InitIdGenerator(ctx, AppTypeERP)

	ids := make(map[uint64]bool)
	count := 10000

	for i := 0; i < count; i++ {
		id := MustGetID()
		if ids[id] {
			t.Errorf("发现重复ID: %d (第 %d 次生成)", id, i)
		}
		ids[id] = true
	}

	t.Logf("成功生成 %d 个唯一ID", count)
}

// TestIdUniquenessAcrossApps 测试跨应用 ID 唯一性
func TestIdUniquenessAcrossApps(t *testing.T) {
	ctx := gctx.New()

	// 收集 ERP 应用的 ID
	InitIdGenerator(ctx, AppTypeERP)
	erpIDs := make(map[uint64]bool)
	for i := 0; i < 1000; i++ {
		id := MustGetID()
		if erpIDs[id] {
			t.Errorf("ERP内部发现重复ID: %d", id)
		}
		erpIDs[id] = true
	}

	// 切换到 Message 应用，收集 ID
	InitIdGenerator(ctx, AppTypeMessage)
	msgIDs := make(map[uint64]bool)
	for i := 0; i < 1000; i++ {
		id := MustGetID()
		if msgIDs[id] {
			t.Errorf("Message内部发现重复ID: %d", id)
		}
		if erpIDs[id] {
			t.Errorf("跨应用发现重复ID: %d (ERP和Message冲突)", id)
		}
		msgIDs[id] = true
	}

	t.Logf("ERP生成 %d 个ID，Message生成 %d 个ID，无冲突", len(erpIDs), len(msgIDs))
}

// TestAppTypeConstants 测试应用类型常量定义正确
func TestAppTypeConstants(t *testing.T) {
	// 验证常量值在有效范围内
	appTypes := []struct {
		name    string
		value   uint32
		isValid bool
	}{
		{"Unknown", AppTypeUnknown, false}, // 0 不应该使用
		{"ERP", AppTypeERP, true},
		{"Message", AppTypeMessage, true},
		{"Manager", AppTypeManager, true},
		{"Shop", AppTypeShop, true},
		{"Takeout", AppTypeTakeout, true},
		{"Websocket", AppTypeWebsocket, true},
	}

	for _, tt := range appTypes {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid && (tt.value < 1 || tt.value > 15) {
				t.Errorf("%s 的值 %d 超出有效范围 (1-15)", tt.name, tt.value)
			}
		})
	}

	// 验证常量值互不相同
	values := make(map[uint32]string)
	for _, tt := range appTypes {
		if tt.value == 0 {
			continue // 跳过 Unknown
		}
		if existing, ok := values[tt.value]; ok {
			t.Errorf("常量值冲突: %s 和 %s 都是 %d", tt.name, existing, tt.value)
		}
		values[tt.value] = tt.name
	}
}

// TestBitsConfiguration 测试位数配置正确
func TestBitsConfiguration(t *testing.T) {
	// 验证位数配置
	if appTypeBits+instanceIDBits != totalNodeBits {
		t.Errorf("位数配置错误: appTypeBits(%d) + instanceIDBits(%d) != totalNodeBits(%d)",
			appTypeBits, instanceIDBits, totalNodeBits)
	}

	// 验证应用类型最大值
	maxAppType := uint32((1 << appTypeBits) - 1) // 15
	if maxAppType != 15 {
		t.Errorf("应用类型最大值错误: got %d, want 15", maxAppType)
	}

	// 验证实例 ID 最大值
	maxInstanceID := uint32((1 << instanceIDBits) - 1) // 63
	if maxInstanceID != 63 {
		t.Errorf("实例ID最大值错误: got %d, want 63", maxInstanceID)
	}

	// 验证节点 ID 最大值
	maxNodeID := uint32((1 << totalNodeBits) - 1) // 1023
	if maxNodeID != 1023 {
		t.Errorf("节点ID最大值错误: got %d, want 1023", maxNodeID)
	}
}

// TestInitIdGeneratorWithInvalidAppType 测试无效应用类型处理
func TestInitIdGeneratorWithInvalidAppType(t *testing.T) {
	ctx := gctx.New()

	// 测试超出范围的应用类型
	InitIdGenerator(ctx, 100) // 应该降级为 Unknown

	// 应该仍然能正常生成 ID
	id := MustGetID()
	if id == 0 {
		t.Error("使用无效应用类型后无法生成ID")
	}

	t.Logf("无效应用类型降级后生成的ID: %d", id)
}

// BenchmarkMustGetID 性能基准测试
func BenchmarkMustGetID(b *testing.B) {
	ctx := gctx.New()
	InitIdGenerator(ctx, AppTypeERP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MustGetID()
	}
}
