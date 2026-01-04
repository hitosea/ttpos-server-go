# 高并发 POS 系统缓存框架 任务分解

> 本文档定义 高并发缓存框架 的详细执行任务清单。

## 📊 进度总览

**总任务数**: 8  
**已完成**: 8  
**进行中**: 已全部完成 ✅
**完成率**: 100%

---

## Phase 1: 核心定义与类型

- [x] 1.1 定义核心接口与任务模型

  - File: `main/pkg/cache/types.go`
  - Purpose: 定义泛型 Task 接口和 ICacheGroup 接口
  - Requirements: 1.1, 1.2
  - Success: 接口定义符合设计文档，泛型约束正确

- [x] 1.2 实现 Singleflight 包装引擎

  - File: `main/pkg/cache/engine.go`
  - Purpose: 封装 golang.org/x/sync/singleflight 的合并逻辑
  - Requirements: 1.1, 1.3
  - Leverage: `golang.org/x/sync/singleflight`
  - Success: 并发请求下 Exec 仅执行一次

---

## Phase 2: 多级缓存实现

- [x] 2.1 实现 L1 本地缓存（内存）

  - File: `main/pkg/cache/local.go`
  - Purpose: 提供基于内存的极速缓存层
  - Requirements: 2.1
  - Leverage: `github.com/patrickmn/go-cache`
  - Success: 单元测试验证 L1 读写及过期逻辑

- [x] 2.2 实现 L2 分布式缓存（Redis）

  - File: `main/pkg/cache/redis.go`
  - Purpose: 提供分布式一致的缓存层
  - Requirements: 2.1
  - Leverage: `main/pkg/cache/cache.go` (Global Cache)
  - Success: 验证 Redis Set/Get 序列化正确

---

## Phase 3: 框架集成与 Read-Through

- [x] 3.1 实现 CacheGroup.Do 核心闭环

  - File: `main/pkg/cache/group.go`
  - Purpose: 串联 L1 -> L2 -> Singleflight -> DB 的完整逻辑
  - Requirements: 2.1, 2.3
  - Success: 实现 Read-Through 模式，自动加载并填充多级缓存

- [x] 3.2 实现负缓存（防穿透）逻辑

  - File: `main/pkg/cache/group.go`
  - Purpose: 缓存空结果，防止缓存穿透
  - Requirements: 2.3
  - Success: 当 DB 返回空时，后续请求在短时间内直接返回空，不查库

---

## Phase 4: 测试与验证

- [x] 4.1 编写高并发压力测试

  - File: `main/pkg/cache/group_test.go`
  - Purpose: 验证框架在极端压力下的表现
  - Requirements: 性能要求, 1.3
  - Success: 1000 并发下无 Data Race，所有请求最终结果一致

- [x] 4.2 接入业务场景试点（ERP 同步）

  - File: `main/app/service/sync_product_to_erp.go`
  - Purpose: 验证框架在真实业务中的效果
  - Requirements: 业务对齐
  - Success: 并发同步请求被正确合并，减少对 ERP API 的调用

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-20  
**维护者**: xiezhihuan
