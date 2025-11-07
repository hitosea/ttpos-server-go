package otlp

// 本文件仅用于演示如何使用 OpenTelemetry
// 使用官方 otelgin 包实现 Gin 框架的调用链跟踪
// 请勿在生产环境中直接使用此文件

/*
// ============================================================
// 使用示例 1: 在 main.go 中初始化 OpenTelemetry
// ============================================================

package main

import (
	"context"
	"ttpos-server-go/pkg/otlp"
	"github.com/jinzhu/copier"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	// 1. 加载配置
	config := otlp.LoadOtlpConfig(copier.Option{IgnoreEmpty: true})

	// 2. 初始化 OpenTelemetry
	if err := otlp.Init(ctx, config); err != nil {
		panic(err)
	}

	// 3. 确保在程序退出时关闭
	defer func() {
		if err := otlp.Shutdown(ctx); err != nil {
			logger.Logger.Error("Failed to shutdown OpenTelemetry", zap.Error(err))
		}
	}()

	// 4. 创建 Gin 引擎
	r := gin.Default()

	// 5. 应用 OpenTelemetry 中间件
	r.Use(otlp.OtlpMiddleware("ttpos-main"))

	// 6. 定义路由
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 7. 启动服务器
	r.Run(":8080")
}


// ============================================================
// 使用示例 2: 在 router.go 中应用中间件
// ============================================================

package router

import (
	"github.com/gin-gonic/gin"
	"ttpos-server-go/pkg/otlp"
)

func Setup(r *gin.Engine) {
	// 全局应用 OpenTelemetry 中间件
	r.Use(otlp.OtlpMiddleware("ttpos-main"))

	// 定义路由组
	apiV1 := r.Group("api/v1")
	{
		apiV1.GET("/users", GetUsers)
		apiV1.POST("/users", CreateUser)
	}
}


// ============================================================
// 使用示例 3: 在业务逻辑中创建子 Span
// ============================================================

package service

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type OrderService struct {}

func (s *OrderService) CreateOrder(ctx context.Context, orderID string, amount float64) error {
	// 创建新的 Span
	tracer := otel.Tracer("ttpos-main")
	ctx, span := tracer.Start(ctx, "OrderService.CreateOrder")
	defer span.End()

	// 添加业务属性
	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.Float64("order.amount", amount),
	)

	// 执行数据库操作
	if err := s.saveOrderToDB(ctx, orderID, amount); err != nil {
		// 记录错误
		span.RecordError(err)
		span.SetStatus(codes.Error, "保存订单失败")
		return err
	}

	// 发送通知
	if err := s.sendNotification(ctx, orderID); err != nil {
		// 记录错误但不中断流程
		span.RecordError(err)
		span.AddEvent("发送通知失败，但继续处理")
	}

	span.SetStatus(codes.Ok, "订单创建成功")
	return nil
}

func (s *OrderService) saveOrderToDB(ctx context.Context, orderID string, amount float64) error {
	// 创建子 Span
	tracer := otel.Tracer("ttpos-main")
	ctx, span := tracer.Start(ctx, "OrderService.saveOrderToDB")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "orders"),
	)

	// 数据库操作
	// ...

	return nil
}

func (s *OrderService) sendNotification(ctx context.Context, orderID string) error {
	// 创建子 Span
	tracer := otel.Tracer("ttpos-main")
	ctx, span := tracer.Start(ctx, "OrderService.sendNotification")
	defer span.End()

	span.SetAttributes(
		attribute.String("notification.type", "email"),
		attribute.String("order.id", orderID),
	)

	// 发送通知
	// ...

	return nil
}


// ============================================================
// 使用示例 4: 跨服务调用时传播 Trace Context
// ============================================================

package client

import (
	"context"
	"net/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/attribute"
)

type ERPClient struct {
	baseURL string
}

func (c *ERPClient) CreatePurchaseOrder(ctx context.Context, orderData interface{}) error {
	// 创建 Span
	tracer := otel.Tracer("ttpos-main")
	ctx, span := tracer.Start(ctx, "ERPClient.CreatePurchaseOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.url", c.baseURL+"/api/purchase-orders"),
	)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/purchase-orders", nil)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// 将 Trace Context 注入到请求头中
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	return nil
}


// ============================================================
// 使用示例 5: 在数据库查询中添加 Trace
// ============================================================

package repository

import (
	"context"
	"gorm.io/gorm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) FindByID(ctx context.Context, userID uint64) (*User, error) {
	// 创建 Span
	tracer := otel.Tracer("ttpos-main")
	ctx, span := tracer.Start(ctx, "UserRepository.FindByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "mysql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
		attribute.Int64("user.id", int64(userID)),
	)

	var user User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &user, nil
}


// ============================================================
// 使用示例 6: 在后台任务中传递 Context
// ============================================================

package worker

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func ProcessOrderAsync(ctx context.Context, orderID string) {
	// 在新的 goroutine 中创建 Span
	go func() {
		tracer := otel.Tracer("ttpos-main")
		ctx, span := tracer.Start(ctx, "ProcessOrderAsync")
		defer span.End()

		span.SetAttributes(
			attribute.String("order.id", orderID),
			attribute.String("worker.type", "async"),
		)

		// 执行异步任务
		// ...
	}()
}
*/
