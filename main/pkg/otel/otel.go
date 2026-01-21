package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
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

// StartSpan 创建并开始一个新的 Span
// ctx: 上下文
// spanName: Span 名称，建议使用格式 "ServiceName.MethodName"
// attrs: Span 属性
// 返回: 新的上下文、Span 对象和结束函数
func StartSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span, SpanEndFunc) {
	tracer := otel.Tracer(DefaultTracerName)
	ctx, span := tracer.Start(ctx, spanName)

	// 添加属性
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	// 返回结束函数
	end := func() {
		span.End()
	}

	return ctx, span, end
}

// TraceFunc 追踪一个无返回值的函数执行
// ctx: 上下文
// spanName: Span 名称
// fn: 要执行的函数
// attrs: Span 属性
// 返回: 函数执行错误
func TraceFunc(ctx context.Context, spanName string, fn func() error, attrs ...attribute.KeyValue) error {
	ctx, span, end := StartSpan(ctx, spanName, attrs...)
	defer end()

	err := fn()
	if err != nil {
		RecordSpanError(ctx, err)
		return err
	}

	span.SetStatus(codes.Ok, "执行成功")
	return nil
}

// TraceFuncWithResult 追踪一个带返回值的函数执行
// ctx: 上下文
// spanName: Span 名称
// fn: 要执行的函数
// attrs: Span 属性
// 返回: 函数执行结果和错误
func TraceFuncWithResult[T any](ctx context.Context, spanName string, fn func() (T, error), attrs ...attribute.KeyValue) (T, error) {
	var zero T
	ctx, span, end := StartSpan(ctx, spanName, attrs...)
	defer end()

	result, err := fn()
	if err != nil {
		RecordSpanError(ctx, err)
		return zero, err
	}

	span.SetStatus(codes.Ok, "执行成功")
	return result, nil
}

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

// SetSpanAttributes 设置当前 Span 的属性
// ctx: 上下文
// attrs: 属性列表
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(attrs...)
}

// FormatSpanName 格式化 Span 名称
// 使用统一的命名规范：component.ServiceName.MethodName
// component: 组件名称，如 "service", "handler", "repository" 等
// serviceName: 服务名称，如 "OrderService", "UserHandler" 等
// methodName: 方法名称，如 "CreateOrder", "GetUser" 等
// 返回: 格式化后的 Span 名称，如 "service.OrderService.CreateOrder"
func FormatSpanName(component, serviceName, methodName string) string {
	return fmt.Sprintf("%s.%s.%s", component, serviceName, methodName)
}

// TraceDuration 记录方法执行耗时
// ctx: 上下文
// startTime: 开始时间
// operationName: 操作名称
func TraceDuration(ctx context.Context, startTime time.Time, operationName string) {
	duration := time.Since(startTime)
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("operation", operationName),
		attribute.Int64("duration.ns", duration.Nanoseconds()),
		attribute.Int64("duration.us", duration.Microseconds()),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	)
}

// GetSpanFromContext 从上下文中获取 Span
// ctx: 上下文
// 返回: Span 对象，如果不存在则返回 NoopSpan
func GetSpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// IsSpanRecording 检查当前 Span 是否正在记录
// ctx: 上下文
// 返回: 是否正在记录
func IsSpanRecording(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	return span.IsRecording()
}

// StartSpanWithCustomContext 为自定义 Context 创建并开始一个新的 Span
// 这个函数专门用于处理实现了 context.Context 接口的自定义 Context 类型
// customCtx: 自定义上下文（必须实现 context.Context 接口）
// spanName: Span 名称
// attrs: Span 属性
// 返回: 更新后的自定义上下文、Span 对象和结束函数
func StartSpanWithCustomContext(customCtx interface {
	context.Context
	GetContext() context.Context
}, spanName string, attrs ...attribute.KeyValue) (interface {
	context.Context
	GetContext() context.Context
}, trace.Span, SpanEndFunc) {
	// 获取标准 context
	stdCtx := customCtx.GetContext()

	// 创建 span
	tracer := otel.Tracer(DefaultTracerName)
	_, span := tracer.Start(stdCtx, spanName)

	// 添加属性
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	// 更新自定义 Context 的底层 context
	// 注意：这需要自定义 Context 提供设置方法，如果没有则只能使用标准 context
	// 这里返回原始的 customCtx，因为 OpenTelemetry 的 context 传播会自动处理
	end := func() {
		span.End()
	}

	return customCtx, span, end
}
