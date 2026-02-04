# task-main-multi-shop-sync 任务清单

## 📊 进度总览

| 项目     | 数值 |
| -------- | ---- |
| 总 SP    | 3    |
| 总任务数 | 7    |
| 已完成   | 7    |
| 完成率   | 100% |

---

## Phase 1: CLI 命令框架

### 1.1 创建批量同步命令文件

| 项目         | 内容                                      |
| ------------ | ----------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go`     |
| Purpose      | 定义 sync-erp-data-batch 命令结构         |
| Requirements | Req1: CLI 命令行接口                      |
| Leverage     | 参考 `main/command/sync_erp_data.go` 结构 |

**实现要点**:
- 使用 Cobra 命令框架
- 定义 `--companies` 参数（[]uint64，逗号分隔）
- 定义 `--workers` 参数（int，默认 5）
- PreRun 中完成配置、缓存、日志、ID 生成器初始化

- [x] 完成

### 1.2 实现门店验证逻辑

| 项目         | 内容                                          |
| ------------ | --------------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go`         |
| Purpose      | 验证所有 company_uuid 有效性                  |
| Requirements | Req1.3: 无效 UUID 终止执行                    |
| Leverage     | 复用 `repository.NewCompanyRepo` 查询公司信息 |

**实现要点**:
- 遍历验证每个 company_uuid
- 检查数据库连接是否成功
- 检查公司是否存在且开启 ERP
- 任一无效则终止，输出无效 UUID 列表

- [x] 完成

---

## Phase 2: 并发执行引擎

### 2.1 实现 Worker Pool

| 项目         | 内容                                  |
| ------------ | ------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go` |
| Purpose      | 并发控制和任务分发                    |
| Requirements | Req2: 并行同步执行                    |
| Leverage     | 使用 `utils.Go` 启动协程              |

**实现要点**:
- 创建 jobs channel 分发任务
- 创建 results channel 收集结果
- 使用 sync.WaitGroup 等待完成
- workers 数量由 `--workers` 参数控制

- [x] 完成

### 2.2 实现单门店同步逻辑

| 项目         | 内容                                      |
| ------------ | ----------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go`     |
| Purpose      | 封装单门店同步流程                        |
| Requirements | Req4: 失败处理与容错                      |
| Leverage     | 复用 `SyncSrv.Sync()` 和上下文设置逻辑    |

**实现要点**:
- 连接门店数据库
- 获取公司信息并设置上下文
- 调用 `syncSrv.Sync(ctx, req.SyncReq{IsSyncExecute: true})`
- 捕获错误，返回 SyncResult 结构

- [x] 完成

---

## Phase 3: 结果汇总与输出

### 3.1 实现进度显示

| 项目         | 内容                                  |
| ------------ | ------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go` |
| Purpose      | 实时显示同步进度                      |
| Requirements | Req3.1, Req3.2: 进度和单门店结果显示  |

**实现要点**:
- 每完成一个门店输出进度 `[N/Total]`
- 显示成功/失败状态图标
- 显示门店 UUID、名称、耗时

- [x] 完成

### 3.2 实现结果汇总

| 项目         | 内容                                  |
| ------------ | ------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go` |
| Purpose      | 汇总显示同步结果                      |
| Requirements | Req3.3, Req4.2: 汇总统计和失败列表    |

**实现要点**:
- 统计成功/失败数量
- 输出失败门店列表（UUID、名称、错误原因）
- 格式化输出便于查看

- [x] 完成

---

## Phase 4: 同步锁与幂等性保护

### 4.1 实现同步锁检测与跳过逻辑

| 项目         | 内容                                  |
| ------------ | ------------------------------------- |
| File         | `main/command/sync_erp_data_batch.go` |
| Purpose      | 防止重复同步，跳过正在同步中的门店    |
| Requirements | Req5: 同步锁与幂等性保护              |
| Leverage     | 复用 `pkg/lock` 分布式锁              |

**实现要点**:
- 定义同步锁 key 前缀：`sync_erp_lock:`
- 同步前使用 `TryLockUuidString` 非阻塞获取锁
- 获取失败则标记为 `Skipped=true`，不执行同步
- 同步完成（成功或失败）后使用 `defer` 释放锁
- 汇总中单独显示跳过列表（区别于失败列表）

- [x] 完成

---

## 提交清单

### 代码质量

- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [ ] 测试通过: `go test ./command/...`（待补充单元测试）

### 功能完整性

- [x] Req1: CLI 命令行接口 - 参数解析、无效 UUID 终止
- [x] Req2: 并行同步执行 - workers 控制、默认并发数 5
- [x] Req3: 进度显示与结果汇总 - 实时进度、最终统计
- [x] Req4: 失败处理与容错 - 失败隔离、失败列表输出
- [x] Req5: 同步锁与幂等性保护 - 跳过正在同步的门店

### 日志要求

- [x] 同步开始/结束记录日志
- [x] 日志包含 company_uuid 字段
- [x] 失败原因记录详细错误信息

---

## 验收测试

### 手动测试用例

```bash
# 1. 基本功能测试 - 同步 2 个有效门店
./main sync-erp-data-batch --companies 1234567890,2345678901

# 2. 并发控制测试 - 指定 workers=2
./main sync-erp-data-batch -c 1234567890,2345678901,3456789012 -w 2

# 3. 无效 UUID 测试 - 应终止并提示
./main sync-erp-data-batch --companies 9999999999

# 4. 空参数测试 - 应显示帮助信息
./main sync-erp-data-batch

# 5. 部分失败测试 - 验证失败隔离
./main sync-erp-data-batch --companies 1234567890,9999999999

# 6. 同步锁测试 - 同时启动两个同步任务，验证跳过逻辑
# 终端1:
./main sync-erp-data-batch -c 1234567890 -w 1
# 终端2（在终端1执行期间）:
./main sync-erp-data-batch -c 1234567890 -w 1
# 预期：终端2 应显示门店被跳过（正在同步中）
```

---

**版本**: v1.0.0
**创建日期**: 2026-02-03
