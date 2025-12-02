# BMP WebSocket 健康检查接口 任务分解

> 覆盖 requirements.md 中的功能需求（Req1~Req3）以及非功能项，所有任务完成后方可归档。

## 📊 进度总览

- **总任务数**: 12  
- **已完成**: 12  
- **完成率**: 100%

---

## Phase 1 - 框架与配置 (Req1.x)

- [x] **1.1 健康检查配置模板与 README 示意**  
  - File: `ttpos-bmp/app/ttpos-websocket/manifest/config/config.tpl.yaml`, `README.MD`  
  - Purpose: 新增 `health.*` 配置块（enable/path/token/cache_ttl/dependencies），并在 README 中描述默认值与启用方式。  
  - Requirements: 1.1 ~ 1.4, 非功能-安全。

- [x] **1.2 Boot 注册健康端点**  
  - File: `internal/boot/boot.go`  
  - Purpose: 读取配置并在 `InitHttp` 阶段绑定 `/internal/health`（可配置）；若关闭则不注册。  
  - Requirements: 1.1。

- [x] **1.3 HTTP 控制器实现**  
  - File: `internal/controller/http/health.go`  
  - Purpose: 解析 query/header、调用逻辑层、返回 200/503、写入 `Retry-After`。  
  - Requirements: 1.1~1.4, 3.4。

- [x] **1.4 安全与访问控制**  
  - File: 同 1.3 及可能的 `internal/logic/health/config.go`  
  - Purpose: 实现 Token/IP 白名单校验、detail=false 逻辑。  
  - Requirements: 1.3, 非功能-安全。

---

## Phase 2 - 依赖检测核心 (Req2.x)

- [x] **2.1 设计 Checker 接口与缓存组件**  
  - File: `internal/logic/health/service.go`、`cache.go`、`types.go`  
  - Purpose: 定义 `Checker`、`CheckResult`、结果缓存、并发执行/超时控制。  
  - Requirements: 1.2, 2.1~2.4。

- [x] **2.2 MySQL Checker**  
  - File: `internal/logic/health/checker_mysql.go`  
  - Purpose: 复用 g.DB 执行 `SELECT 1`，计算耗时，返回状态。  
  - Requirements: 2.2。

- [x] **2.3 Redis Checker**  
  - File: `internal/logic/health/checker_redis.go`  
  - Purpose: cluster/单机兼容，执行 `PING` + 可选 `CLUSTER INFO`。  
  - Requirements: 2.1。

- [x] **2.4 可插拔依赖配置**  
  - File: `internal/logic/health/service.go` (registerCheckers)  
  - Purpose: 根据 `health.dependencies` 动态注册 RocketMQ/Nacos 占位 Checker（默认 disabled，返回 SKIP）。  
  - Requirements: 2.3。

---

## Phase 3 - 观测与降级 (Req3.x)

- [x] **3.1 指标与日志**  
  - File: `internal/logic/health/metrics.go`  
  - Purpose: 记录 gauge/counter/histogram，日志包含 trace_id/component。  
  - Requirements: 3.1, 非功能-可观测。

- [x] **3.2 降级钩子与 webhook 预留**  
  - File: `internal/logic/health/notifier.go`  
  - Purpose: 当 overall=DOWN 时触发日志、可选 webhook，预留暂停推送接口。  
  - Requirements: 3.2, 3.3。

---

## Phase 4 - 文档 & 测试

- [x] **4.1 单元测试**  
  - File: `internal/logic/health/service_test.go`  
  - Purpose: 覆盖并发聚合、缓存命中、各 Checker 成功/失败路径。  
  - Requirements: 验收-测试。

- [x] **4.2 集成测试/自测脚本**  
  - File: `tests/health_integration_test.sh`  
  - Purpose: 使用 curl 验证健康检查端点功能与响应格式。  
  - Requirements: 验收-探针/观测。

- [x] **4.3 README 与运维文档**  
  - File: `ttpos-bmp/app/ttpos-websocket/README.MD`  
  - Purpose: 更新使用指南、探针示例、Grafana 指标说明。  
  - Requirements: 3.2, 非功能-文档。

---

> 所有任务完成后需将 `[ ]` 改为 `[x]`，并在提交说明中引用具体任务编号。
