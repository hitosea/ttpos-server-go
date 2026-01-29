// Package lineman Lineman API 客户端
package lineman

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gogf/gf/v2/test/gtest"
)

// TestWithRetry_Success 测试首次成功
func TestWithRetry_Success(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		callCount := 0

		err := WithRetry(context.Background(), func() error {
			callCount++
			return nil // 首次成功
		})

		t.AssertNil(err)
		t.Assert(callCount, 1) // 只调用一次
	})
}

// TestWithRetry_SuccessAfterRetry 测试第二次成功
func TestWithRetry_SuccessAfterRetry(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		callCount := 0

		err := WithRetry(context.Background(), func() error {
			callCount++
			if callCount == 1 {
				return errors.New("第一次失败")
			}
			return nil // 第二次成功
		})

		t.AssertNil(err)
		t.Assert(callCount, 2) // 调用两次
	})
}

// TestWithRetry_AllFailures 测试所有重试都失败
func TestWithRetry_AllFailures(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		callCount := 0

		err := WithRetry(context.Background(), func() error {
			callCount++
			return errors.New("失败")
		})

		t.AssertNE(err, nil)
		t.Assert(callCount, 3) // 最多调用 3 次
		t.AssertIN("失败", err.Error())
	})
}

// TestWithRetry_ExponentialBackoff 测试指数退避
func TestWithRetry_ExponentialBackoff(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		callCount := 0
		var callTimes []time.Time

		startTime := time.Now()

		err := WithRetry(context.Background(), func() error {
			callCount++
			callTimes = append(callTimes, time.Now())
			return errors.New("失败")
		})

		t.AssertNE(err, nil)
		t.Assert(callCount, 3)

		// 验证总耗时（第1次 + 2s + 第2次 + 4s + 第3次）
		// 总耗时应该 >= 6s（2s + 4s）
		elapsed := time.Since(startTime)
		t.AssertGE(elapsed.Seconds(), 6.0)

		// 验证间隔时间
		if len(callTimes) == 3 {
			// 第1次到第2次：约 2s
			interval1 := callTimes[1].Sub(callTimes[0])
			t.AssertGE(interval1.Seconds(), 1.9)
			t.AssertLE(interval1.Seconds(), 2.5)

			// 第2次到第3次：约 4s
			interval2 := callTimes[2].Sub(callTimes[1])
			t.AssertGE(interval2.Seconds(), 3.9)
			t.AssertLE(interval2.Seconds(), 4.5)
		}
	})
}

// TestWithRetry_MaxRetries 测试最大重试次数
func TestWithRetry_MaxRetries(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		callCount := 0

		// 模拟第 3 次成功
		err := WithRetry(context.Background(), func() error {
			callCount++
			if callCount < 3 {
				return errors.New("失败")
			}
			return nil // 第 3 次成功
		})

		t.AssertNil(err)
		t.Assert(callCount, 3)
	})
}
