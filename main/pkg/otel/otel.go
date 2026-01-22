package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// DefaultTracerName 默认 tracer 名称
	DefaultTracerName = "xie/ttpos"
)

// SpanEndFunc Span 结束函数类型
type SpanEndFunc func()

// AddSpanEvent 向当前 Span 添加事件
// ctx: 上下文
// name: 事件名称
// attrs: 事件属性
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// RecordSpanError 记录错误到当前 Span
// ctx: 上下文
// err: 错误对象
// message: 可选的错误消息
func RecordSpanError(ctx context.Context, err error, message ...string) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.RecordError(err)

	msg := "执行失败"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	span.SetStatus(codes.Error, msg)
	span.SetAttributes(attribute.String("error.message", err.Error()))
}
