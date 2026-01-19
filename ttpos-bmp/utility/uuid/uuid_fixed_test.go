package uuid

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/gogf/gf/v2/frame/g"
)

// TestFixedMultiInstance 测试修复后的多实例 ID 生成
func TestFixedMultiInstance(t *testing.T) {
	ctx := context.Background()

	// 模拟多个实例，使用不同的 appType 和 instanceID
	instances := []struct {
		name       string
		appType    uint32
		instanceID uint32
	}{
		{"erp-1", 1, 1},
		{"message-1", 2, 1},
		{"manager-1", 3, 1},
		{"shop-1", 4, 1},
	}

	var generators []*goid.ID
	var nodeIDs []uint32

	// 初始化所有实例
	for _, inst := range instances {
		gen := goid.NewID()
		nodeID := (inst.appType << 6) | inst.instanceID
		gen.SetNode(nodeID, 10)

		counterBits := 21 - 10
		maxRandomDelta := uint32((1 << counterBits) - 1)
		randomDelta := uint32((time.Now().UnixNano() % int64(maxRandomDelta-1)) + 2)
		gen.SetRandomDelta(randomDelta)

		generators = append(generators, gen)
		nodeIDs = append(nodeIDs, nodeID)

		g.Log().Infof(ctx, "[Fixed] %s: nodeID=%d, randomDelta=%d", inst.name, nodeID, randomDelta)
	}

	// 同步启动，在同一秒内生成多个 ID
	var wg sync.WaitGroup
	var mu sync.Mutex
	idMap := make(map[int64]string) // ID -> 实例名称

	for i, gen := range generators {
		wg.Add(1)
		go func(idx int, generator *goid.ID, name string) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				id := generator.Generate()
				mu.Lock()
				if existing, exists := idMap[id]; exists {
					t.Errorf("[Fixed] ID 重复！%s 和 %s 都生成了 ID: %d", existing, name, id)
				}
				idMap[id] = name
				mu.Unlock()
			}
		}(i, gen, instances[i].name)
	}

	wg.Wait()

	g.Log().Infof(ctx, "[Fixed] 成功生成 %d 个唯一 ID", len(idMap))
}

// TestRandomDeltaEffect 测试随机增量的效果
func TestRandomDeltaEffect(t *testing.T) {
	ctx := context.Background()

	// 创建 10 个实例，检查 randomDelta 的分布
	var deltas []uint32

	for i := 0; i < 10; i++ {
		counterBits := 21 - 10
		maxRandomDelta := uint32((1 << counterBits) - 1)
		randomDelta := uint32((time.Now().UnixNano() % int64(maxRandomDelta-1)) + 2)

		deltas = append(deltas, randomDelta)
	}

	g.Log().Infof(ctx, "[RandomDelta] 生成的增量值: %v", deltas)

	uniqueDeltas := make(map[uint32]bool)
	for _, d := range deltas {
		uniqueDeltas[d] = true
	}

	g.Log().Infof(ctx, "[RandomDelta] 唯一增量数量: %d/%d", len(uniqueDeltas), len(deltas))

	if len(uniqueDeltas) < len(deltas)/2 {
		t.Errorf("[RandomDelta] 增量值过于集中，可能影响去重效果")
	}
}

// TestConcurrentSameInstance 测试同一实例并发生成 ID
func TestConcurrentSameInstance(t *testing.T) {
	ctx := context.Background()

	gen := goid.NewID()
	gen.SetNode(65, 10)
	gen.SetRandomDelta(1)

	var wg sync.WaitGroup
	var mu sync.Mutex
	ids := make(map[int64]bool)

	// 10 个 goroutine，每个生成 100 个 ID
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				id := gen.Generate()
				mu.Lock()
				if ids[id] {
					t.Errorf("[Concurrent] 发现重复 ID: %d", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	g.Log().Infof(ctx, "[Concurrent] 成功生成 %d 个唯一 ID", len(ids))
	if len(ids) != 1000 {
		t.Errorf("[Concurrent] 预期 1000 个 ID，实际 %d", len(ids))
	}
}

// TestNodeIDDistribution 测试 nodeID 分布
func TestNodeIDDistribution(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		appType    uint32
		instanceID uint32
		expected   uint32
	}{
		{1, 1, 65},
		{2, 1, 129},
		{2, 2, 130},
		{3, 1, 193},
		{4, 1, 257},
		{15, 63, 1023},
	}

	for _, tc := range testCases {
		nodeID := (tc.appType << 6) | tc.instanceID
		if nodeID != tc.expected {
			t.Errorf("[NodeID] 计算错误: appType=%d, instanceID=%d, 期望=%d, 实际=%d",
				tc.appType, tc.instanceID, tc.expected, nodeID)
		}

		g.Log().Infof(ctx, "[NodeID] appType=%d, instanceID=%d -> nodeID=%d (0b%s)",
			tc.appType, tc.instanceID, nodeID, fmt.Sprintf("%010b", nodeID))
	}
}
