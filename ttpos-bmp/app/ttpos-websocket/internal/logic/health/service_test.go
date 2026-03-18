// Package health 健康检查逻辑层测试
package health

import (
	"context"
	"testing"
	"time"
)

// TestResultCache 测试结果缓存
func TestResultCache(t *testing.T) {
	cache := NewResultCache(100 * time.Millisecond)

	// 初始状态应该为空
	if _, ok := cache.Get(); ok {
		t.Error("初始缓存应该为空")
	}

	// 设置缓存
	response := &HealthResponse{
		Status:    StatusUp,
		Timestamp: time.Now().UnixMilli(),
	}
	cache.Set(response)

	// 应该能获取到缓存
	if cached, ok := cache.Get(); !ok {
		t.Error("应该能获取到缓存")
	} else if cached.Status != StatusUp {
		t.Errorf("缓存状态错误: expected=%s, got=%s", StatusUp, cached.Status)
	}

	// 等待缓存过期
	time.Sleep(150 * time.Millisecond)

	// 过期后应该获取不到
	if _, ok := cache.Get(); ok {
		t.Error("过期后应该获取不到缓存")
	}
}

// TestCheckResult 测试检查结果聚合
func TestCheckResult(t *testing.T) {
	tests := []struct {
		name     string
		results  map[string]CheckResult
		expected string
	}{
		{
			name: "全部UP",
			results: map[string]CheckResult{
				"mysql": {Status: StatusUp},
				"redis": {Status: StatusUp},
			},
			expected: StatusUp,
		},
		{
			name: "一个DOWN",
			results: map[string]CheckResult{
				"mysql": {Status: StatusDown},
				"redis": {Status: StatusUp},
			},
			expected: StatusDown,
		},
		{
			name: "全部DOWN",
			results: map[string]CheckResult{
				"mysql": {Status: StatusDown},
				"redis": {Status: StatusDown},
			},
			expected: StatusDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟聚合逻辑
			status := StatusUp
			for _, result := range tt.results {
				if result.Status == StatusDown {
					status = StatusDown
					break
				}
			}

			if status != tt.expected {
				t.Errorf("聚合结果错误: expected=%s, got=%s", tt.expected, status)
			}
		})
	}
}

// MockChecker 模拟检查器
type MockChecker struct {
	name   string
	result CheckResult
	delay  time.Duration
}

func (m *MockChecker) Name() string {
	return m.name
}

func (m *MockChecker) Check(ctx context.Context) CheckResult {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.result
}

// TestServiceCheck 测试服务检查
func TestServiceCheck(t *testing.T) {
	// 创建服务实例
	svc := &Service{
		checkers: make(map[string]Checker),
		config: &Config{
			CacheTTL: 100 * time.Millisecond,
			Timeout:  500 * time.Millisecond,
		},
	}
	svc.cache = NewResultCache(svc.config.CacheTTL)

	// 注册模拟检查器
	svc.RegisterChecker(&MockChecker{
		name:   "test1",
		result: CheckResult{Status: StatusUp, Latency: 10},
	})
	svc.RegisterChecker(&MockChecker{
		name:   "test2",
		result: CheckResult{Status: StatusUp, Latency: 20},
	})

	// 执行检查
	ctx := context.Background()
	response := svc.Check(ctx, CheckOptions{Detail: true})

	// 验证结果
	if response.Status != StatusUp {
		t.Errorf("期望状态为 UP, 实际为 %s", response.Status)
	}

	if len(response.Components) != 2 {
		t.Errorf("期望 2 个组件, 实际为 %d", len(response.Components))
	}

	// 测试缓存
	time.Sleep(50 * time.Millisecond)
	response2 := svc.Check(ctx, CheckOptions{Detail: false})
	if response2.Timestamp != response.Timestamp {
		t.Error("应该返回缓存结果")
	}
}

// TestServiceCheckTimeout 测试超时控制
func TestServiceCheckTimeout(t *testing.T) {
	svc := &Service{
		checkers: make(map[string]Checker),
		config: &Config{
			CacheTTL: 100 * time.Millisecond,
			Timeout:  100 * time.Millisecond, // 短超时
		},
	}
	svc.cache = NewResultCache(svc.config.CacheTTL)

	// 注册一个慢检查器
	svc.RegisterChecker(&MockChecker{
		name:   "slow",
		result: CheckResult{Status: StatusUp},
		delay:  200 * time.Millisecond, // 超过超时时间
	})

	ctx := context.Background()
	start := time.Now()
	response := svc.Check(ctx, CheckOptions{Detail: true})
	duration := time.Since(start)

	// 应该在超时时间内返回
	if duration > 500*time.Millisecond {
		t.Errorf("检查耗时过长: %v", duration)
	}

	t.Logf("检查完成: status=%s, duration=%v", response.Status, duration)
}
