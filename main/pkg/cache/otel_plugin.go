package cache

import (
	"log"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"ttpos-server-go/pkg/otlp"
)

// enableRedisTracing 为 Redis 客户端启用 OpenTelemetry 追踪
// 参数：
//   - client: Redis 单机客户端（可为 nil）
//   - clusterClient: Redis 集群客户端（可为 nil）
func enableRedisTracing(client *redis.Client, clusterClient *redis.ClusterClient) {
	cfg := otlp.Config()
	if cfg == nil || !cfg.Enabled {
		return
	}

	tracerProvider := otel.GetTracerProvider()

	// 为单机客户端启用追踪
	if client != nil {
		if err := redisotel.InstrumentTracing(client,
			redisotel.WithTracerProvider(tracerProvider),
			redisotel.WithDBSystem("redis"),
		); err != nil {
			log.Printf("[OTLP] register redis tracing failed: %v", err)
		}

		// TODO: 暂时禁用 Metrics，待验证 Tracing 稳定后再启用
		// if err := redisotel.InstrumentMetrics(client,
		// 	redisotel.WithMeterProvider(otel.GetMeterProvider()),
		// ); err != nil {
		// 	log.Printf("[OTLP] register redis metrics failed: %v", err)
		// }
	}

	// 为集群客户端启用追踪
	if clusterClient != nil {
		if err := redisotel.InstrumentTracing(clusterClient,
			redisotel.WithTracerProvider(tracerProvider),
			redisotel.WithDBSystem("redis"),
		); err != nil {
			log.Printf("[OTLP] register redis cluster tracing failed: %v", err)
		}

		// TODO: 暂时禁用 Metrics，待验证 Tracing 稳定后再启用
		// if err := redisotel.InstrumentMetrics(clusterClient,
		// 	redisotel.WithMeterProvider(otel.GetMeterProvider()),
		// ); err != nil {
		// 	log.Printf("[OTLP] register redis cluster metrics failed: %v", err)
		// }
	}
}
