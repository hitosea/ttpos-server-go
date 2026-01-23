package otlp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type OtlpConfig struct {
	ServiceName   string  `json:"serviceName"`
	Endpoint      string  `json:"endpoint"`
	Path          string  `json:"path"`
	Token         string  `json:"token"`
	LogHeaders    string  `json:"logHeaders"`
	Enabled       bool    `json:"enabled"`
	SamplingRatio float64 `json:"samplingRatio"`
	SlowQueryMs   int     `json:"slowQueryMs"`
}

var (
	tracerProvider *sdktrace.TracerProvider
	currentConfig  *OtlpConfig
	Inited         = false // 是否已初始化
	logHeaders     []string
)

// parseEndpoint 解析 endpoint，支持 http:// 或 https:// 格式
// 参数：
//   - rawEndpoint: 原始 endpoint 字符串，可以是 "host:port" 或 "http://host:port" 或 "https://host:port"
//
// 返回：
//   - endpoint: 解析后的 host:port 格式
//   - useInsecure: 是否使用不安全连接（HTTP）
func parseEndpoint(rawEndpoint string) (endpoint string, useInsecure bool) {
	// 默认使用不安全连接（向后兼容）
	useInsecure = true

	// 如果不包含协议前缀，直接返回原始值
	if !strings.Contains(rawEndpoint, "://") {
		return rawEndpoint, useInsecure
	}

	// 解析 URL
	u, err := url.Parse(rawEndpoint)
	if err != nil {
		// 解析失败，返回原始值
		return rawEndpoint, useInsecure
	}

	// 根据协议决定是否使用 TLS
	if strings.EqualFold(u.Scheme, "https") {
		useInsecure = false
	}

	// 返回 host:port 格式（WithEndpoint 不需要协议前缀）
	return u.Host, useInsecure
}

// LoadOtlpConfig 加载 OpenTelemetry 配置
// 参数：
//   - opt: copier 配置选项
//
// 返回：
//   - *OtlpConfig: OTLP 配置对象
func LoadOtlpConfig(opt copier.Option) *OtlpConfig {
	otlpConfig := &OtlpConfig{
		ServiceName:   "ttpos-main",
		Endpoint:      "localhost:4318",
		Path:          "/v1/traces",
		Enabled:       false,
		SamplingRatio: 1,
		SlowQueryMs:   200,
	}

	err := copier.CopyWithOption(otlpConfig, OtlpConfig{
		ServiceName:   viper.GetString("OTLP_SERVICE_NAME"),
		Endpoint:      viper.GetString("OTLP_ENDPOINT"),
		Path:          viper.GetString("OTLP_PATH"),
		Token:         viper.GetString("OTLP_TOKEN"),
		LogHeaders:    viper.GetString("OTLP_LOG_HEADERS"),
		Enabled:       viper.GetBool("OTLP_ENABLED"),
		SamplingRatio: viper.GetFloat64("OTLP_SAMPLING_RATIO"),
		SlowQueryMs:   viper.GetInt("OTLP_SLOW_QUERY_MS"),
	}, opt)
	if err != nil {
		log.Printf("[OTLP] Config error: %v", err)
		return nil
	}

	return otlpConfig
}

// Init 初始化 OpenTelemetry
// 参数：
//   - ctx: 上下文对象
//   - config: OTLP 配置
//
// 返回：
//   - error: 错误信息
func Init(ctx context.Context, config *OtlpConfig) error {
	if config == nil {
		log.Println("[OTLP] Config is nil, skipping initialization")
		return nil
	}
	if !config.Enabled {
		log.Println("[OTLP] Disabled, skipping initialization")
		return nil
	}
	// 配置特殊 headers 记录
	if config.LogHeaders != "" {
		logHeaders = strings.Split(config.LogHeaders, ",")
	}

	currentConfig = config

	// 解析 endpoint，支持 http:// 或 https:// 格式
	endpoint, useInsecure := parseEndpoint(config.Endpoint)

	// 创建 OTLP HTTP Exporter 选项
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath(config.Path),
	}
	if useInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if config.Token != "" {
		opts = append(opts, otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + config.Token,
		}))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("创建 OTLP exporter 失败: %w", err)
	}

	// 创建 Resource，标识服务
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			attribute.String("environment", viper.GetString("SERVER_MODE")),
			attribute.String("appid", viper.GetString("APP_ID")),
		),
	)
	if err != nil {
		return fmt.Errorf("创建 resource 失败: %w", err)
	}

	// 创建 TracerProvider
	samplingRatio := config.SamplingRatio
	if samplingRatio < 0 {
		samplingRatio = 0
	}
	if samplingRatio > 1 {
		samplingRatio = 1
	}
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingRatio))

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithMaxExportBatchSize(512), sdktrace.WithBatchTimeout(3*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tracerProvider)

	// 设置全局 Propagator，用于跨服务传播 trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("[OTLP] Initialized successfully (service=%s, endpoint=%s, tls=%v, has_token=%v)",
		config.ServiceName, config.Endpoint, !useInsecure, config.Token != "")
	Inited = true
	return nil
}

// Shutdown 关闭 OpenTelemetry TracerProvider
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - error: 错误信息
func Shutdown(ctx context.Context) error {
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}

// Config 返回当前 OTLP 配置
func Config() *OtlpConfig {
	return currentConfig
}

// OtlpMiddleware 返回一个 Gin 的 OpenTelemetry 中间件，用于自动生成 trace
// 使用官方的 otelgin 包实现，自动处理：
//   - Trace Context 传播
//   - Span 创建和管理
//   - HTTP 请求/响应属性记录
//   - 所有 HTTP Headers 记录
//   - 错误记录
//
// 参数：
//   - serviceName: 服务名称，用于区分不同服务的 trace
//
// 返回：
//   - gin.HandlerFunc: Gin 中间件函数
func OtlpMiddleware(serviceName string) gin.HandlerFunc {
	// 先使用 otelgin 中间件创建基础的 trace
	return otelgin.Middleware(serviceName)
}

// RecordSpecialHeaders 记录特殊 HTTP headers 到 OpenTelemetry span
// 特殊 headers 包括：
//   - device-id: 设备 ID，记录为字符串属性
//
// 返回：
//   - gin.HandlerFunc: Gin 中间件函数
func RecordSpecialHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前的 span，记录所有 HTTP headers
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			// 遍历所有 HTTP headers 并添加到 span 中
			for key, values := range c.Request.Header {
				// 检查是否需要记录该 header
				if len(logHeaders) > 0 {
					found := false
					for _, h := range logHeaders {
						if strings.ToLower(key) == strings.ToLower(h) {
							span.SetAttributes(attribute.String(strings.ToLower(h), strings.Join(values, ",")))
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
			}
		}

		// 继续处理请求
		c.Next()
	}
}
