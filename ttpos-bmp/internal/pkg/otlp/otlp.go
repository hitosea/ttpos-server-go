package otlp

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gipv4"
	"github.com/gogf/gf/v2/os/gctx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	globalCtx = gctx.GetInitCtx()
	config    = &OtlpConfig{}
)

type OtlpConfig struct {
	ServiceName string `json:"serviceName"`
	Endpoint    string `json:"endpoint"`
	Path        string `json:"path"`
	Enabled     bool   `json:"enabled"`
	Token       string `json:"token"`
}

func init() {
	// 初始化调用链监控
	// 从配置文件中读取调用链监控配置
	err := g.Cfg().MustGet(globalCtx, "otlp").Scan(&config)
	if err != nil {
		g.Log().Warningf(globalCtx, "otlp config error: %v", err)
		return
	}
	g.Log().Infof(globalCtx, "otlp config: enabled=%v service=%s endpoint=%s path=%s has_token=%v",
		config.Enabled,
		config.ServiceName,
		config.Endpoint,
		config.Path,
		config.Token != "",
	)
}

func InitOtlp() func(context.Context) {
	if config.Enabled {
		shutdown, err := initOtlpHTTP(config.ServiceName, config.Endpoint, config.Path, config.Token)
		if err != nil {
			g.Log().Warningf(globalCtx, "otlp init error: %v", err)
			return nil
		}
		g.Log().Infof(globalCtx, "otlp init success")
		return shutdown
	}
	return func(ctx context.Context) {
		g.Log().Info(ctx, "otlp not enabled, skip shutdown")
	}
}

func initOtlpHTTP(serviceName, endpoint, path, token string) (func(ctx context.Context), error) {
	var (
		intranetIPArray, err = gipv4.GetIntranetIpArray()
		hostIP               = "NoHostIpFound"
	)

	if err != nil {
		return nil, err
	}

	if len(intranetIPArray) == 0 {
		if intranetIPArray, err = gipv4.GetIpArray(); err != nil {
			return nil, err
		}
	}
	if len(intranetIPArray) > 0 {
		hostIP = intranetIPArray[0]
	}

	endpoint, insecure := normalizeEndpoint(endpoint)
	ctx := context.Background()
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithCompression(1),
	}
	if insecure {
		options = append(options, otlptracehttp.WithInsecure())
	}
	if token != "" {
		options = append(options, otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + token,
		}))
	}
	traceExp, err := otlptrace.New(ctx, otlptracehttp.NewClient(options...))
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.HostNameKey.String(hostIP),
			attribute.String("hostname", hostIP),
		),
	)

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExp)),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)

	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err = tracerProvider.Shutdown(ctx); err != nil {
			g.Log().Errorf(ctx, "Shutdown tracerProvider failed err:%+v", err)
		} else {
			g.Log().Debug(ctx, "Shutdown tracerProvider success")
		}
	}, nil
}

func normalizeEndpoint(endpoint string) (string, bool) {
	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://"), false
	}
	if strings.HasPrefix(endpoint, "http://") {
		return strings.TrimPrefix(endpoint, "http://"), true
	}
	return endpoint, true
}
