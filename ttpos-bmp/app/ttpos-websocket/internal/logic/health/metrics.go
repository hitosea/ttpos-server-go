// Package health 健康检查逻辑层
package health

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// RecordMetrics 记录健康检查指标
// 注意：当前为占位实现，实际应集成 Prometheus 或 OpenTelemetry
func RecordMetrics(ctx context.Context, response *HealthResponse) {
	// 记录总体状态（gauge）
	statusValue := 1
	if response.Status == StatusDown {
		statusValue = 0
	}
	g.Log().Debugf(ctx, "健康检查指标: ttpos_websocket_health_status{component=overall}=%d", statusValue)

	// 记录各组件状态和延迟
	for name, result := range response.Components {
		// 记录失败次数（counter）
		if result.Status == StatusDown {
			g.Log().Debugf(ctx, "健康检查指标: ttpos_websocket_health_fail_total{component=%s}++", name)
		}

		// 记录延迟（histogram）
		g.Log().Debugf(ctx, "健康检查指标: ttpos_websocket_health_latency_ms{component=%s}=%d", name, result.Latency)
	}

	// TODO: 实际集成 Prometheus 时，使用以下代码：
	// prometheus.MustRegister(...)
	// healthStatusGauge.WithLabelValues("overall").Set(float64(statusValue))
	// healthFailCounter.WithLabelValues(name).Inc()
	// healthLatencyHistogram.WithLabelValues(name).Observe(float64(result.Latency))
}
