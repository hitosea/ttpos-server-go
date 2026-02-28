package gormcache

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ttpos-server-go/pkg/otlp"
)

const (
	tracerName = "gorm-cache"
)

// startCacheSpan 创建带有 gorm-cache 标识的 span
// 用于在追踪中区分 gorm-cache 的 Redis 操作和普通业务 Redis 操作
func startCacheSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	// 检查 OTLP 是否启用
	cfg := otlp.Config()
	if cfg == nil || !cfg.Enabled {
		return ctx, nil
	}

	tracer := otel.Tracer(tracerName)

	// 添加默认属性
	defaultAttrs := []attribute.KeyValue{
		attribute.String("cache.system", "gorm-cache"),
		attribute.String("cache.operation", operation),
	}

	// 合并自定义属性
	allAttrs := append(defaultAttrs, attrs...)

	ctx, span := tracer.Start(ctx, "gorm-cache:"+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(allAttrs...),
	)

	return ctx, span
}

// endSpan 安全地结束 span
func endSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}

// addSpanAttributes 为当前 span 添加属性
func addSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span != nil {
		span.SetAttributes(attrs...)
	}
}
