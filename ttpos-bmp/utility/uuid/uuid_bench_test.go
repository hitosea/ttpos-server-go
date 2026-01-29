package uuid

import (
	"sync"
	"testing"

	goid "github.com/ace-zhaoy/go-id"
)

// BenchmarkGenerate_WithRandomDelta 测试使用 randomDelta 的性能（问题配置）
func BenchmarkGenerate_WithRandomDelta(b *testing.B) {
	id := goid.NewID()
	id.SetNode(65, 10)
	id.SetRandomDelta(2000) // 这会导致每次调用 crypto/rand

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id.Generate()
	}
}

// BenchmarkGenerate_WithoutRandomDelta 测试不使用 randomDelta 的性能（修复后配置）
func BenchmarkGenerate_WithoutRandomDelta(b *testing.B) {
	id := goid.NewID()
	id.SetNode(65, 10)
	// 不设置 randomDelta，使用默认 delta=1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id.Generate()
	}
}

// BenchmarkGenerate_Parallel_WithRandomDelta 并发测试（问题配置）
func BenchmarkGenerate_Parallel_WithRandomDelta(b *testing.B) {
	id := goid.NewID()
	id.SetNode(65, 10)
	id.SetRandomDelta(2000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id.Generate()
		}
	})
}

// BenchmarkGenerate_Parallel_WithoutRandomDelta 并发测试（修复后配置）
func BenchmarkGenerate_Parallel_WithoutRandomDelta(b *testing.B) {
	id := goid.NewID()
	id.SetNode(65, 10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id.Generate()
		}
	})
}

// TestIDUniqueness 验证修复后 ID 仍然唯一
func TestIDUniqueness(t *testing.T) {
	id := goid.NewID()
	id.SetNode(65, 10)

	const count = 100000
	ids := make(map[int64]struct{}, count)

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < count/10; j++ {
				newID := id.Generate()
				mu.Lock()
				if _, exists := ids[newID]; exists {
					t.Errorf("Duplicate ID: %d", newID)
				}
				ids[newID] = struct{}{}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}
