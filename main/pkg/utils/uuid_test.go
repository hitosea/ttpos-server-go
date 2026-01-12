package utils

import (
	"sync"
	"testing"

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
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	const count = 100000
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
	viper.Set("SERVER_ID", 1)
	InitIdGenerator()

	const goroutines = 100
	const idsPerGoroutine = 1000

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
