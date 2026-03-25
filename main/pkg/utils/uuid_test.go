package utils

import (
	"sync"
	"testing"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestInitIdGenerator 测试ID生成器初始化
func TestInitIdGenerator(t *testing.T) {
	// 设置有效的SERVER_ID
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	// 验证可以生成ID
	id1, err := GetID()
	assert.NoError(t, err)
	assert.Greater(t, id1, uint64(0))

	id2 := MustGetID()
	assert.Greater(t, id2, uint64(0))
	assert.NotEqual(t, id1, id2)
}

// TestInitIdGeneratorWithInvalidServerId 测试无效SERVER_ID的处理
func TestInitIdGeneratorWithInvalidServerId(t *testing.T) {
	// 测试超出范围的SERVER_ID
	testCases := []int{0, -1, 1024, 9999}

	for _, serverId := range testCases {
		viper.Set("SERVER_ID", serverId)
		InitIdGenerator()

		// 验证仍然可以生成唯一ID
		id1, err := GetID()
		assert.NoError(t, err)
		assert.Greater(t, id1, uint64(0))
	}
}

// TestIdUniqueness 测试ID唯一性
func TestIdUniqueness(t *testing.T) {
	count := 100000
	if testing.Short() {
		t.Log("唯一性测试（使用 -short 标志）")
		count = 10000
	}
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	ids := make(map[uint64]bool, count)

	// 生成大量ID并检查唯一性
	for i := 0; i < count; i++ {
		id := MustGetID()
		assert.False(t, ids[id], "发现重复ID: %d", id)
		ids[id] = true
	}

	assert.Equal(t, count, len(ids), "生成的唯一ID数量应该等于请求数量")
}

// TestConcurrentIdGeneration 测试并发生成ID的唯一性
func TestConcurrentIdGeneration(t *testing.T) {
	goroutines := 100
	idsPerGoroutine := 1000
	if testing.Short() {
		t.Log("并发生成测试（使用 -short 标志）")
		goroutines = 20
		idsPerGoroutine = 500 // 20 * 500 = 10,000 IDs in short mode
	}
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	var wg sync.WaitGroup
	idsChan := make(chan uint64, goroutines*idsPerGoroutine)

	// 启动多个goroutine并发生成ID
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := MustGetID()
				idsChan <- id
			}
		}()
	}

	wg.Wait()
	close(idsChan)

	// 检查所有ID的唯一性
	ids := make(map[uint64]bool)
	for id := range idsChan {
		assert.False(t, ids[id], "并发场景下发现重复ID: %d", id)
		ids[id] = true
	}

	expectedCount := goroutines * idsPerGoroutine
	assert.Equal(t, expectedCount, len(ids), "并发生成的唯一ID数量应该等于请求总数")
}

// TestMultipleInstanceSimulation 模拟多实例场景
func TestMultipleInstanceSimulation(t *testing.T) {
	// 模拟3个不同的实例
	serverIds := []int{1, 2, 3}
	allIds := make(map[uint64]bool)
	var mu sync.Mutex

	for _, serverId := range serverIds {
		viper.Set("SERVER_ID", serverId)
		InitIdGenerator()

		// 每个实例生成1000个ID
		for i := 0; i < 1000; i++ {
			id := MustGetID()
			mu.Lock()
			assert.False(t, allIds[id], "多实例场景下发现重复ID: %d (SERVER_ID=%d)", id, serverId)
			allIds[id] = true
			mu.Unlock()
		}
	}

	assert.Equal(t, 3000, len(allIds), "多实例生成的ID应该全部唯一")
}

// BenchmarkMustGetID 性能基准测试
func BenchmarkMustGetID(b *testing.B) {
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MustGetID()
	}
}

// BenchmarkGetID 性能基准测试（带错误检查）
func BenchmarkGetID(b *testing.B) {
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetID()
	}
}

// BenchmarkConcurrentMustGetID 并发性能基准测试
func BenchmarkConcurrentMustGetID(b *testing.B) {
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			MustGetID()
		}
	})
}

// TestMultipleNodesConcurrentGeneration 测试多个节点并发生成ID的唯一性
// 模拟真实场景：多个服务器节点同时生成大量ID，验证是否有重复
func TestMultipleNodesConcurrentGeneration(t *testing.T) {
	// 配置测试参数
	const (
		nodeCount        = 10   // 模拟10个节点
		idsPerNode       = 5000 // 每个节点生成5000个ID
		totalExpectedIds = nodeCount * idsPerNode
	)

	// 使用 map 收集所有生成的ID，key为ID值，value为生成该ID的节点ID
	type idInfo struct {
		nodeID int
		count  int
	}
	allIds := make(map[uint64]*idInfo)
	var mu sync.Mutex

	var wg sync.WaitGroup

	// 为每个节点启动一个goroutine，模拟多个节点同时生成ID
	for nodeID := 1; nodeID <= nodeCount; nodeID++ {
		wg.Add(1)
		go func(serverID int) {
			defer wg.Done()

			// 为当前节点创建独立的ID生成器实例
			// 注意：这里创建独立的生成器实例来模拟不同节点的独立进程
			nodeGenerator := goid.NewID()
			nodeGenerator.SetNode(uint32(serverID), totalNodeBits)

			// 当前节点生成指定数量的ID
			nodeIds := make([]uint64, 0, idsPerNode)
			for i := 0; i < idsPerNode; i++ {
				id := uint64(nodeGenerator.Generate())
				nodeIds = append(nodeIds, id)
			}

			// 将当前节点生成的ID添加到全局集合中
			mu.Lock()
			for _, id := range nodeIds {
				if info, exists := allIds[id]; exists {
					// 发现重复ID
					info.count++
					t.Errorf("发现重复ID: %d (节点%d和节点%d都生成了此ID)", id, info.nodeID, serverID)
				} else {
					allIds[id] = &idInfo{
						nodeID: serverID,
						count:  1,
					}
				}
			}
			mu.Unlock()
		}(nodeID)
	}

	// 等待所有节点完成ID生成
	wg.Wait()

	// 验证结果
	mu.Lock()
	defer mu.Unlock()

	// 检查是否有重复ID
	duplicateCount := 0
	for id, info := range allIds {
		if info.count > 1 {
			duplicateCount++
			t.Errorf("ID %d 被重复生成了 %d 次", id, info.count)
		}
	}

	// 验证生成的ID总数
	actualCount := len(allIds)
	if actualCount != totalExpectedIds {
		t.Errorf("生成的ID总数不匹配: 期望 %d, 实际 %d (发现 %d 个重复ID)",
			totalExpectedIds, actualCount, duplicateCount)
	}

	// 如果没有重复，输出成功信息
	if duplicateCount == 0 {
		t.Logf("✅ 测试通过: %d 个节点并发生成了 %d 个唯一ID，无重复", nodeCount, actualCount)
	} else {
		t.Errorf("❌ 测试失败: 发现 %d 个重复ID", duplicateCount)
	}
}

// TestMultipleNodesConcurrentGenerationStress 压力测试：更多节点和更多ID
func TestMultipleNodesConcurrentGenerationStress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试（使用 -short 标志时）")
	}

	// 配置更严格的测试参数
	const (
		nodeCount        = 50   // 模拟50个节点
		idsPerNode       = 2000 // 每个节点生成2000个ID
		totalExpectedIds = nodeCount * idsPerNode
	)

	allIds := make(map[uint64]int) // key: ID, value: 生成次数
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 启动多个节点并发生成ID
	for nodeID := 1; nodeID <= nodeCount; nodeID++ {
		wg.Add(1)
		go func(serverID int) {
			defer wg.Done()

			// 创建独立的ID生成器实例
			nodeGenerator := goid.NewID()
			nodeGenerator.SetNode(uint32(serverID), totalNodeBits)

			// 生成ID并收集
			localIds := make([]uint64, idsPerNode)
			for i := 0; i < idsPerNode; i++ {
				localIds[i] = uint64(nodeGenerator.Generate())
			}

			// 添加到全局集合
			mu.Lock()
			for _, id := range localIds {
				allIds[id]++
			}
			mu.Unlock()
		}(nodeID)
	}

	wg.Wait()

	// 检查重复
	mu.Lock()
	duplicateCount := 0
	for id, count := range allIds {
		if count > 1 {
			duplicateCount++
			t.Errorf("ID %d 被重复生成了 %d 次", id, count)
		}
	}
	actualCount := len(allIds)
	mu.Unlock()

	// 验证结果
	if duplicateCount > 0 {
		t.Errorf("压力测试失败: 发现 %d 个重复ID", duplicateCount)
	}
	if actualCount != totalExpectedIds {
		t.Errorf("生成的ID总数不匹配: 期望 %d, 实际 %d", totalExpectedIds, actualCount)
	}

	if duplicateCount == 0 {
		t.Logf("✅ 压力测试通过: %d 个节点并发生成了 %d 个唯一ID", nodeCount, actualCount)
	}
}
