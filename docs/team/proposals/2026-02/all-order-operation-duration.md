# 订单操作耗时记录 需求提案

## 📋 提案信息

| 项目          | 内容                       |
| ------------- | -------------------------- |
| **提案人**    | xiezhihuan                 |
| **日期**      | 2026-02-05                 |
| **目标版本**  | 待定                       |
| **状态**      | 待评审                     |
| **关联 Spec** | -                          |

---

## 🎯 背景和动机

### 问题描述

当前系统存在以下问题：

1. **无法追踪性能瓶颈**：不知道哪个订单操作慢、慢在哪里，难以定位和优化性能问题
2. **缺乏操作审计数据**：没有完整记录谁、什么时间、做了什么操作，以及操作耗时多久
3. **分布式服务难排查**：多实例部署时，不知道问题发生在哪个服务实例，排查困难

### 业务价值

1. **快速定位性能问题**：通过耗时数据快速发现并优化慢操作
2. **提升系统稳定性**：及时发现异常，防止大规模故障
3. **支撑容量规划**：基于性能数据制定扩容策略
4. **审计合规**：满足操作可追溯的合规要求

### 目标用户

- [x] 后端开发人员（性能调优和问题排查）
- [x] 运维人员（监控和告警）
- [x] 产品经理（操作效率和用户行为分析）

---

## 💡 解决方案概述

### 方案描述

在 Handler 层记录每个订单操作的开始时间和结束时间，计算耗时后推送到内存队列（带缓冲的 channel）。消费者协程从队列中批量读取记录，使用同一个数据库连接写入 SaaS 主库。

该方案的核心特点：
- **零阻塞**：Handler 只做 `chan <- record`，不等待数据库写入
- **批量优化**：累积 100 条或超时 5 秒后批量 INSERT，减少数据库连接开销
- **分布式支持**：记录服务实例标识，支持跨实例问题追踪

### 核心功能点

1. **耗时记录**：记录 company_uuid、sale_bill_uuid、sale_order_uuid、操作类型、开始/结束时间、耗时
2. **服务实例标识**：记录 instance_id，支持分布式环境下的问题定位
3. **内存队列缓冲**：使用 channel 作为本地缓存，队列满时丢弃并告警（保证业务不受影响）
4. **异步批量写入**：单消费者协程持有 SaaS 库连接，批量写入数据库

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [x] Shop 商家管理端
- [x] KDS 厨显端
- [x] QDS 排号端
- [x] Assistant 助手端
- [x] Tablet 平板端
- [x] Mobile 扫码端
- [x] Menu 电子菜单端
- [x] Member 会员端
- [x] Kiosk 自助点餐机

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（Handler 层）
- [x] 数据模型（新增表）
- [x] 业务逻辑（队列和消费者）
- [x] 其他: Main 模块后端服务

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整,无业务逻辑变更
- [x] **中**：需要前后端联调,基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预估 SP**: 5（待技术评审确认）

### 拆分预估

**是否需要拆分**：
- [x] **否**：单终端，SP ≤ 5，可直接创建 1 个 Spec
- [ ] **是**：需要拆分为多个 Spec

**预估 Spec 数量**：1 个

**预估 Spec 列表**：
1. `story-main-order-operation-duration` - 订单操作耗时记录功能实现

### 风险识别

**潜在风险**：
1. 内存队列满时数据丢失
2. 大量写入可能影响 SaaS 库性能

**缓解措施**：
1. 队列满时记录告警日志，可后续扩展降级写入 Redis
2. 批量写入 + 异步处理，与业务隔离；可考虑独立数据库或分表

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名       | 签名/日期 |
| ---------- | ---------- | --------- |
| 产品经理   | -          |           |
| 技术负责人 | xiezhihuan |           |
| 开发代表   | -          |           |
| 测试代表   | -          |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
待评审
```

**下一步行动**：

- [ ] 创建 Spec：`story-main-order-operation-duration`
- [ ] 分配负责人：待定
- [ ] 目标 Sprint：待定

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发人员/运维人员
**我想** 查看每个订单操作的耗时记录（包括 company、账单、操作类型、服务实例等信息）
**以便于** 快速定位性能瓶颈、排查分布式环境下的问题、进行容量规划

### AC 验收标准（初稿）

1. **WHEN** 订单操作完成 **THEN** 系统 **SHALL** 异步记录操作耗时到数据库
2. **WHEN** 队列满 **THEN** 系统 **SHALL** 丢弃记录并输出告警日志，不阻塞业务
3. **WHEN** 查询耗时记录 **THEN** 系统 **SHALL** 支持按 company_uuid、action、时间范围筛选

### 技术方案摘要

**数据表（SaaS 主库）**：
```sql
CREATE TABLE `ttpos_order_operation_duration` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `action` VARCHAR(100) NOT NULL DEFAULT '',
    `source` VARCHAR(50) NOT NULL DEFAULT '',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `device_sn` VARCHAR(255) NOT NULL DEFAULT '',
    `instance_id` VARCHAR(128) NOT NULL DEFAULT '',
    `start_time` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `end_time` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `duration_ms` INT UNSIGNED NOT NULL DEFAULT 0,
    `request_path` VARCHAR(255) NOT NULL DEFAULT '',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `error_msg` VARCHAR(500) NOT NULL DEFAULT '',
    `create_time` INT UNSIGNED NOT NULL DEFAULT 0,
    `update_time` INT UNSIGNED NOT NULL DEFAULT 0,
    `delete_time` INT UNSIGNED NOT NULL DEFAULT 0,
    INDEX `idx_company_uuid` (`company_uuid`),
    INDEX `idx_action` (`action`),
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`),
    INDEX `idx_instance_id` (`instance_id`),
    INDEX `idx_create_time` (`create_time`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
```

**架构流程**：
```
Handler → 内存 Channel (10000) → 消费者协程 → 批量写入 SaaS 库
```

---

**版本**: v1.0.0
