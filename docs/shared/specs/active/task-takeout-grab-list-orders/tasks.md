# task-takeout-grab-list-orders 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| **总 SP** | 2 |
| **总任务数** | 4 |
| **已完成** | 3 |
| **完成率** | 75% |

> **重要提示**: Grab 平台限制，ListOrders API **仅生产环境支持**，测试环境不可用。

> **设计变更 (v2.0)**: 移除 Proto/gRPC 方案，改为直接使用 SDK 类型。简化架构，提升类型安全。

---

## Phase 1: Logic 层 SDK 调用封装

### 1.1 在 grab_api.go 新增 ListOrders 方法

| 项目 | 内容 |
|------|------|
| **File** | `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_api.go` |
| **Purpose** | 封装 GrabFood SDK ListOrdersAPI 调用 |
| **Requirements** | Req-2, Req-3 |
| **Leverage** | grab.Default(), HandleSDKError(), 现有方法模式 |

**实现要点**：
1. 获取 Authorization Header
2. 构建 ApiListOrdersRequest
3. 设置可选参数（date, orderIDs, page）
4. 调用 Execute() 并处理错误
5. 直接返回 SDK `*grabfood.ListOrdersResponse` 类型
6. 记录日志（请求参数、响应订单数量）

- [x] 完成

---

### 1.2 更新 service 接口

| 项目 | 内容 |
|------|------|
| **File** | `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go` |
| **Purpose** | 更新 IGrab 接口，添加 ListOrders 方法签名 |
| **Requirements** | 接口一致性 |
| **Leverage** | 现有接口模式 |

**接口签名**：
```go
ListOrders(ctx context.Context, merchantID string, date string, orderIDs []string, page int32) (*grabfood.ListOrdersResponse, error)
```

> **注意**: 已手动更新接口定义，使用 SDK 类型作为返回值

- [x] 完成

---

## Phase 2: 测试

### 2.1 编写单元测试

| 项目 | 内容 |
|------|------|
| **File** | `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_api_test.go` |
| **Purpose** | ListOrders 方法单元测试 |
| **Requirements** | 非功能需求-测试要求 |
| **Leverage** | 现有测试模式 |

**测试用例**：
1. TestListOrders_Success - 正常查询
2. TestListOrders_WithDate - 日期过滤
3. TestListOrders_WithOrderIDs - 订单ID过滤
4. TestListOrders_Pagination - 分页验证
5. TestListOrders_EmptyMerchantID - 参数校验
6. TestListOrders_SDKError - SDK 错误处理

- [ ] 完成

---

## Phase 3: 代码质量检查

### 3.1 代码质量验证

| 项目 | 内容 |
|------|------|
| **Command** | `cd ttpos-bmp/app/ttpos-takeout && go mod tidy && go fmt ./... && go vet ./...` |
| **Purpose** | 确保代码质量符合规范 |

- [x] 完成

---

## 已移除任务

以下任务因设计变更（v2.0）而移除：

| 原任务 | 原因 |
|--------|------|
| Proto 定义 ListOrders RPC | 改为直接使用 SDK 类型，不需要 Proto 定义 |
| 执行 make pb 生成 Go 代码 | 无 Proto 变更 |
| 实现 gRPC Controller | 不暴露 gRPC 接口，仅供内部服务调用 |

---

## 提交清单

### 代码质量

- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [ ] 测试通过: `go test ./internal/logic/grab/...`

### 功能完整性

- [x] Logic 层实现完成
- [x] Service 接口更新
- [x] 文档更新

### BMP 规范

- [x] 遵循 go-rules.mdc
- [x] 日志包含关键参数
- [x] 使用 SDK 原生类型

---

## 执行命令汇总

```bash
# 代码质量检查
cd ttpos-bmp/app/ttpos-takeout && go mod tidy && go fmt ./... && go vet ./...

# 运行测试
cd ttpos-bmp/app/ttpos-takeout && go test -v ./internal/logic/grab/... -run TestListOrders
```

---

**版本**: v2.0.0
**创建日期**: 2026-01-23
**更新日期**: 2026-01-23
**更新说明**: 移除 Proto/gRPC 任务，更新完成状态
