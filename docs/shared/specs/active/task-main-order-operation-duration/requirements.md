# 订单操作耗时记录 需求文档

## 📋 基本信息

| 项目              | 内容                                                                              |
| ----------------- | --------------------------------------------------------------------------------- |
| **Spec ID**       | task-main-order-operation-duration                                                |
| **来源 Proposal** | [all-order-operation-duration](../../../team/proposals/2026-02/all-order-operation-duration.md) |
| **创建日期**      | 2026-02-05                                                                        |
| **负责人**        | xiezhihuan                                                                        |
| **目标 Sprint**   | 待定                                                                              |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | xiezhihuan |
| **审核日期** | 2026-02-05 |

---

## 📝 用户故事

**作为** 后端开发人员/运维人员/产品经理
**我想** 记录每个订单操作的耗时数据（包括 company、账单、操作类型、服务实例等信息）
**以便于** 快速定位性能瓶颈、支撑容量规划、排查分布式环境下的故障、进行操作审计

---

## 功能需求

### Requirement 1: 耗时记录器

**用户故事**: 作为后端开发人员，我想在 Handler 层记录操作开始/结束时间，以便于统计每个操作的耗时

#### 验收标准

1. **WHEN** 订单操作开始 **THEN** 系统 **SHALL** 记录开始时间戳（毫秒级）
2. **WHEN** 订单操作结束 **THEN** 系统 **SHALL** 计算耗时并推送到内存队列
3. **WHEN** 操作失败 **THEN** 系统 **SHALL** 记录错误信息和失败状态

#### 记录字段

| 字段            | 类型   | 说明                              |
| --------------- | ------ | --------------------------------- |
| company_uuid    | uint64 | 商户标识                          |
| sale_bill_uuid  | uint64 | 账单 UUID                         |
| sale_order_uuid | uint64 | 订单 UUID                         |
| action          | string | 操作类型 (cancel_order/checkout…) |
| source          | string | 来源终端 (cashier/assistant/h5…)  |
| staff_uuid      | uint64 | 操作员工                          |
| device_sn       | string | 设备序列号                        |
| instance_id     | string | 服务实例标识                      |
| start_time      | int64  | 开始时间戳（毫秒）                |
| end_time        | int64  | 结束时间戳（毫秒）                |
| duration_ms     | int    | 耗时（毫秒）                      |
| request_path    | string | 请求路径                          |
| status          | int    | 1 成功 0 失败                     |
| error_msg       | string | 错误信息                          |

---

### Requirement 2: 内存队列缓冲

**用户故事**: 作为后端开发人员，我想使用内存队列缓冲记录数据，以便于不阻塞业务请求

#### 验收标准

1. **WHEN** 推送记录到队列 **THEN** 系统 **SHALL** 使用非阻塞方式（select + default）
2. **WHEN** 队列未满 **THEN** 系统 **SHALL** 成功推送记录
3. **WHEN** 队列已满 **THEN** 系统 **SHALL** 丢弃记录并输出 WARN 日志，不阻塞业务
4. **IF** 服务重启 **THEN** 系统 **SHALL** 接受队列中未消费记录丢失（可接受的数据损失）

#### 队列配置

| 参数       | 默认值 | 说明               |
| ---------- | ------ | ------------------ |
| 队列容量   | 10000  | 带缓冲 channel     |
| 批量阈值   | 100    | 累积多少条后写入   |
| 刷新间隔   | 5 秒   | 超时强制刷新       |

---

### Requirement 3: 异步批量写入

**用户故事**: 作为后端开发人员，我想使用单消费者协程批量写入数据库，以便于减少数据库连接开销

#### 验收标准

1. **WHEN** 队列中记录数达到批量阈值（100 条） **THEN** 系统 **SHALL** 批量写入数据库
2. **WHEN** 队列中有记录且超过刷新间隔（5 秒） **THEN** 系统 **SHALL** 批量写入数据库
3. **WHEN** 批量写入 **THEN** 系统 **SHALL** 使用同一个 SaaS 库连接
4. **WHEN** 写入失败 **THEN** 系统 **SHALL** 输出 ERROR 日志（不重试，避免积压）

---

### Requirement 4: 分布式追踪

**用户故事**: 作为运维人员，我想查看记录来自哪个服务实例，以便于定位分布式环境下的问题

#### 验收标准

1. **WHEN** 推送记录 **THEN** 系统 **SHALL** 自动填充当前服务实例 ID
2. **WHEN** 查询记录 **THEN** 系统 **SHALL** 支持按 instance_id 筛选
3. **IF** 环境变量 INSTANCE_ID 存在 **THEN** 系统 **SHALL** 使用该值
4. **IF** 环境变量不存在 **THEN** 系统 **SHALL** 使用 hostname 作为实例标识

---

## 非功能需求

### 性能要求

- [ ] Handler 层记录耗时 < 1ms（仅 channel 推送）
- [ ] 不影响业务请求响应时间
- [ ] 批量写入效率 > 1000 条/秒

### 可靠性要求

- [ ] 队列满时不阻塞业务
- [ ] 消费者异常时自动恢复（使用 utils.Go 包裹）
- [ ] 服务重启时接受未消费数据丢失

### 平台兼容性

- [x] Go Main 模块（main/app/）
- [ ] Go BMP 模块（后续可扩展）

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- 协程必须使用 `utils.Go` 方法（内置 recover）
- 数据库写入到 SaaS 主库（constant.DefaultDB）
- 日志必须包含 `company_uuid` 字段

### 数据库约束

- 表名: `ttpos_order_operation_duration`
- 存储位置: SaaS 主库（`const TARGET = 'main';`）
- 必须包含 `delete_time` 字段支持软删除

### 资源约束

- Story Point: 3

---

## 数据表设计

```sql
CREATE TABLE `ttpos_order_operation_duration` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `action` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作类型',
    `source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源终端',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `device_sn` VARCHAR(255) NOT NULL DEFAULT '',
    `instance_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '服务实例标识',
    `start_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '开始时间戳(毫秒)',
    `end_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '结束时间戳(毫秒)',
    `duration_ms` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)',
    `request_path` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1成功 0失败',
    `error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '错误信息',
    `create_time` INT UNSIGNED NOT NULL DEFAULT 0,
    `update_time` INT UNSIGNED NOT NULL DEFAULT 0,
    `delete_time` INT UNSIGNED NOT NULL DEFAULT 0,
    INDEX `idx_company_uuid` (`company_uuid`),
    INDEX `idx_action` (`action`),
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`),
    INDEX `idx_instance_id` (`instance_id`),
    INDEX `idx_create_time` (`create_time`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '订单操作耗时记录';
```

---

## 风险和缓解

### 风险 1: 内存队列满时数据丢失

**影响**: 低
**缓解措施**:
- 队列容量设置为 10000，足够缓冲突发流量
- 丢弃时输出 WARN 日志，可监控告警
- 后续可扩展降级写入 Redis

### 风险 2: 大量写入影响 SaaS 库性能

**影响**: 中
**缓解措施**:
- 批量写入减少连接开销
- 单消费者控制写入速率
- 可考虑独立数据库或按月分表

### 风险 3: 服务重启数据丢失

**影响**: 低
**缓解措施**:
- 耗时记录为辅助数据，可接受少量丢失
- 正常关闭时可等待队列消费完成

---

**版本**: v1.0.0
**创建日期**: 2026-02-05
