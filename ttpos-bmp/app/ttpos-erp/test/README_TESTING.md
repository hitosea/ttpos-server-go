# ERP 模块测试指南

## 📋 测试概述

本文档说明如何运行 ttpos-erp 模块的各类测试。

## 🧪 测试分类

### 1. 单元测试 (Unit Tests)

**位置**: `internal/logic/selling/`

**已实现的单元测试**:
- ✅ `paymentid_test.go` - PaymentID 生成逻辑测试
- ✅ `permission_test.go` - 权限校验测试

**运行方式**:
```bash
# 运行所有单元测试
cd ttpos-bmp/app/ttpos-erp
go test -v ./internal/logic/selling/paymentid_test.go
go test -v ./internal/logic/selling/permission_test.go

# 运行基准测试
go test -bench=. ./internal/logic/selling/paymentid_test.go
```

**特点**:
- ✅ 不依赖外部服务
- ✅ 不需要配置文件
- ✅ 快速执行（< 1秒）
- ✅ 适合 CI/CD 流水线

### 2. gRPC 接口测试 (Integration Tests)

**位置**: `test/grpc/`

**待实现的 gRPC 测试**:
- ⏳ 创建支付方式测试
- ⏳ 查询支付方式测试  
- ⏳ 更新支付方式测试

**前置条件**:
- 需要运行 ttpos-erp gRPC 服务器
- 需要配置 ERPNext 连接
- 需要测试数据库

**运行方式**:
```bash
# 1. 启动 ttpos-erp 服务
cd ttpos-bmp/app/ttpos-erp
go run main.go

# 2. 在另一个终端运行测试
cd ttpos-bmp/app/ttpos-erp
go test -v -tags=integration ./test/grpc/...
```

### 3. 端到端测试 (E2E Tests)

**位置**: `test/e2e/`

**测试场景**:
- 完整的创建 → 查询 → 更新流程
- PaymentID 自动生成验证
- 权限校验验证

**前置条件**:
- 需要完整的运行环境
- 需要 ERPNext 测试环境
- 需要网络连接

## 🚀 快速开始

### 运行所有单元测试

```bash
cd /home/coder/workspaces/ttpos-server-go/ttpos-bmp/app/ttpos-erp

# 运行 PaymentID 生成测试
go test -v ./internal/logic/selling/paymentid_test.go

# 运行权限校验测试
go test -v ./internal/logic/selling/permission_test.go

# 查看测试覆盖率
go test -cover ./internal/logic/selling/paymentid_test.go
```

### 性能基准测试

```bash
cd /home/coder/workspaces/ttpos-server-go/ttpos-bmp/app/ttpos-erp

# 运行基准测试
go test -bench=BenchmarkPaymentIDGeneration ./internal/logic/selling/paymentid_test.go

# 带内存分配统计
go test -bench=. -benchmem ./internal/logic/selling/paymentid_test.go
```

## 📊 测试覆盖率

### 当前覆盖率

- **PaymentID 生成逻辑**: ✅ 100% (核心逻辑)
- **权限校验逻辑**: ✅ 100% (核心逻辑)
- **gRPC 接口**: ⏳ 待实现
- **端到端流程**: ⏳ 待实现

### 目标覆盖率

- 核心业务逻辑: ≥ 80%
- 数据处理逻辑: ≥ 70%
- 总体覆盖率: ≥ 60%

## 🔍 测试结果示例

### PaymentID 生成测试

```
=== RUN   TestPaymentIDGeneration
    paymentid_test.go:48: ✓ 生成的 PaymentID: PID3704524594350081 (长度: 19)
--- PASS: TestPaymentIDGeneration (0.00s)

=== RUN   TestPaymentIDUniqueness
    paymentid_test.go:72: ✓ 成功生成 1000 个唯一的 PaymentID
--- PASS: TestPaymentIDUniqueness (0.00s)

=== RUN   TestPaymentIDFormat
--- PASS: TestPaymentIDFormat (0.00s)
```

### 权限校验测试

```
=== RUN   TestPermissionValidation
=== RUN   TestPermissionValidation/SameCompany_ShouldAllow
=== RUN   TestPermissionValidation/DifferentCompany_ShouldDeny
--- PASS: TestPermissionValidation (0.00s)
```

## 🛠️ 故障排查

### 问题 1: 配置文件未找到

**错误信息**:
```
panic: possible config files "config" not found
```

**解决方案**:
- 单元测试不应该依赖配置文件
- 使用 `*_test.go` 中的 `init()` 函数初始化必要的依赖
- 参考 `paymentid_test.go` 的初始化方式

### 问题 2: UUID 生成器未初始化

**错误信息**:
```
panic: invalid memory address or nil pointer dereference
```

**解决方案**:
```go
func init() {
    ctx := context.Background()
    uuid.InitIdGenerator(ctx, uuid.AppTypeERP)
}
```

## 📚 相关文档

- [单元测试最佳实践](https://go.dev/doc/tutorial/add-a-test)
- [GoFrame 测试文档](https://goframe.org/pages/viewpage.action?pageId=1114367)
- [gRPC 测试指南](https://grpc.io/docs/languages/go/testing/)

## 🔄 持续集成

### CI/CD 配置示例

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.23'
      - name: Run Unit Tests
        run: |
          cd ttpos-bmp/app/ttpos-erp
          go test -v ./internal/logic/selling/paymentid_test.go
          go test -v ./internal/logic/selling/permission_test.go
```

---

**最后更新**: 2025-12-23
**维护者**: ttpos-erp 开发组

