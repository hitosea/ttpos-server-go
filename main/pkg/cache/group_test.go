package cache

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"ttpos-server-go/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Logger = zap.NewNop()
	os.Exit(m.Run())
}

// mockTask 模拟一个耗时任务
type mockTask struct {
	key     string
	val     string
	delay   time.Duration
	execCnt *int32
	err     error
}

func (m *mockTask) Hash() string {
	return m.key
}

func (m *mockTask) Exec(ctx context.Context) (string, error) {
	atomic.AddInt32(m.execCnt, 1)
	time.Sleep(m.delay)
	return m.val, m.err
}

func (m *mockTask) TTL() time.Duration {
	return 1 * time.Minute
}

func TestCacheGroup_Do_Singleflight(t *testing.T) {
	config := GroupConfig{
		Name:             "test-group",
		EnableLocalCache: true,
		EnableRedisCache: false, // 测试中暂不开启 Redis
	}
	group := NewCacheGroup[string](config)

	var execCnt int32
	concurrent := 100
	wg := sync.WaitGroup{}
	wg.Add(concurrent)

	// 100 个并发请求同一个 Key
	task := &mockTask{
		key:     "hot-key",
		val:     "some-data",
		delay:   50 * time.Millisecond,
		execCnt: &execCnt,
	}

	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			val, err := group.Do(context.Background(), task)
			assert.NoError(t, err)
			assert.Equal(t, "some-data", val)
		}()
	}

	wg.Wait()

	// 验证 Exec 仅被调用了 1 次
	assert.Equal(t, int32(1), atomic.LoadInt32(&execCnt))
}

func TestCacheGroup_L1Cache(t *testing.T) {
	config := GroupConfig{
		Name:             "test-l1",
		EnableLocalCache: true,
	}
	group := NewCacheGroup[string](config)

	var execCnt int32
	task := &mockTask{
		key:     "l1-key",
		val:     "l1-data",
		execCnt: &execCnt,
	}

	// 第一次执行
	val, err := group.Do(context.Background(), task)
	assert.NoError(t, err)
	assert.Equal(t, "l1-data", val)
	assert.Equal(t, int32(1), atomic.LoadInt32(&execCnt))

	// 第二次执行，应该命中 L1
	val, err = group.Do(context.Background(), task)
	assert.NoError(t, err)
	assert.Equal(t, "l1-data", val)
	assert.Equal(t, int32(1), atomic.LoadInt32(&execCnt), "Should hit L1 cache")
}

func TestCacheGroup_NegativeCaching(t *testing.T) {
	config := GroupConfig{
		Name:             "test-neg",
		EnableLocalCache: true,
		NegativeTTL:      100 * time.Millisecond,
	}
	group := NewCacheGroup[string](config)

	var execCnt int32
	// 模拟一个返回空数据（zero value）且有错误的场景（如 Record Not Found）
	task := &mockTask{
		key:     "empty-key",
		val:     "",
		err:     fmt.Errorf("record not found"),
		execCnt: &execCnt,
	}

	// 第一次执行，触发错误并缓存零值
	val1, err1 := group.Do(context.Background(), task)
	assert.Error(t, err1)
	assert.Equal(t, "", val1)
	assert.Equal(t, int32(1), atomic.LoadInt32(&execCnt))

	// 第二次执行，在 NegativeTTL 内，由于缓存了零值，Do 会直接从缓存返回 (zero, nil)
	// 这符合“空结果缓存”的预期：不再查库，直接返回空
	val2, err2 := group.Do(context.Background(), task)
	assert.NoError(t, err2)
	assert.Equal(t, "", val2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&execCnt), "Should hit negative cache (zero value)")

	// 等待负缓存过期
	time.Sleep(150 * time.Millisecond)

	// 第三次执行，负缓存过期，应再次触发 Exec
	_, err3 := group.Do(context.Background(), task)
	assert.Error(t, err3)
	assert.Equal(t, int32(2), atomic.LoadInt32(&execCnt))
}
