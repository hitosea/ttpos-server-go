// Package lineman Lineman API 客户端
package lineman

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	maxRetries   = 3
	retryDelay   = 2 * time.Second
	retryBackoff = 2.0
)

// WithRetry 带重试的请求执行
func WithRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := retryDelay

	for i := 0; i < maxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if i < maxRetries-1 {
			g.Log().Warningf(ctx, "[LinemanClient] 请求失败（第 %d/%d 次），%s 后重试: %v", i+1, maxRetries, delay, err)
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * retryBackoff)
		}
	}

	g.Log().Errorf(ctx, "[LinemanClient] 重试 %d 次后仍然失败: %v", maxRetries, lastErr)
	return lastErr
}
