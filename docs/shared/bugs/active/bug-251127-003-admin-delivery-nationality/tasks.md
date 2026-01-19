# Bug-251127-003 修复任务清单

> **Bug**: 外卖/国籍管理-排序和删除功能缺陷  
> **当前状态**: 🟢 开发中  
> **开始时间**: 2025-11-27  
> **预计完成**: 2025-11-29

---

## 📋 任务列表

> **重要说明**：数据表已有 `delete_time` 字段，无需数据库迁移，直接修改代码即可。

---

### 阶段 1: Go Main 后端修改（1.5 天）

#### 1.1 修改 Service 层 - 外卖来源

- [x] **修改列表查询方法** `main/app/repository/order_source_repository.go::FindList()`
  - **需求**: 过滤已删除 + 倒序排序
  - **修改点**:
    - 已有 `WHERE delete_time = 0`
    - ✅ 修改 `ORDER BY sort ASC` → `ORDER BY create_time DESC`
  - **实际时间**: 0.1 小时
  - **负责人**: AI Agent
  - **说明**: 在 Repository 层完成

- [x] **修改删除方法** `main/app/service/order_source_service.go::Delete()`
  - **需求**: 改为软删除
  - **修改点**:
    - ✅ 软删除已在 `SoftDelete()` 实现
    - ✅ 移除 `CheckCanDelete()` 调用
  - **实际时间**: 0.1 小时
  - **负责人**: AI Agent

- [x] **新增查询方法（包含已删除）** `main/app/repository/order_source_repository.go::FindByUuidWithDeleted()`
  - **需求**: 用于订单详情查询，不过滤 `delete_time`
  - **内容**: ✅ 新增方法 `FindByUuidWithDeleted(uuid uint64) (*model.OrderSource, error)`
  - **实际时间**: 0.2 小时
  - **负责人**: AI Agent

#### 1.2 修改 Service 层 - 国籍

- [x] **修改国籍查询服务** `main/app/repository/nationality_repository.go` & `main/app/service/nationality_service.go`
  - **需求**: 同外卖来源的 3 个修改
    1. ✅ `FindList()` - 修改排序为倒序
    2. ✅ `Delete()` - 移除 CheckCanDelete 调用
    3. ✅ `FindByUuidWithDeleted()` - 新增方法
  - **实际时间**: 0.3 小时
  - **负责人**: AI Agent

#### 1.3 修改 Repository 层（如果使用 Repository 模式）

- [x] **修改 Repository 层** `main/app/repository/order_source_repository.go`, `main/app/repository/nationality_repository.go`
  - **需求**: 更新查询和删除方法
  - **修改点**:
    - ✅ `FindList()` - 修改 `ORDER BY sort ASC` → `ORDER BY create_time DESC`
    - ✅ `SoftDelete()` - 已存在且正确实现
    - ✅ `FindByUuidWithDeleted()` - 新增方法，不过滤 `delete_time`
  - **实际时间**: 0.5 小时
  - **负责人**: AI Agent
  - **说明**: 项目使用 Repository 模式

#### 1.4 修改订单详情查询

- [x] **修改订单详情 Repository** `main/app/repository/order.go`
  - **需求**: 订单详情关联查询时不过滤已删除配置
  - **修改点**:
    - ✅ `GetSaleBillDetails()` - 移除 OrderSource/Nationality Preload 的 `delete_time` 过滤
    - ✅ `GetSaleBillAllInfo()` - 移除 OrderSource/Nationality Preload 的 `delete_time` 过滤
    - ✅ `GetSaleBillWithProducts()` - 移除 OrderSource/Nationality Preload 的 `delete_time` 过滤
  - **实际时间**: 0.3 小时
  - **负责人**: AI Agent
  - **说明**: 在 Repository 层的 Preload 修改，而非 Service 层

#### 1.5 单元测试（Go）

- [ ] **编写列表排序测试** `main/app/service/order_source_test.go::TestGetListSorting()`
  - **需求**: 验证列表倒序排序
  - **测试用例**:
    1. 创建配置 A（时间 T1）
    2. 创建配置 B（时间 T2）
    3. 调用 GetList()
    4. 验证 B 在 A 前面
  - **预计时间**: 1 小时
  - **负责人**: 

- [ ] **编写软删除测试** `main/app/service/order_source_test.go::TestSoftDelete()`
  - **需求**: 验证软删除和过滤
  - **测试用例**:
    1. 创建配置
    2. 调用 Delete()
    3. 调用 GetList()，验证列表不包含该配置
    4. 直接查询数据库，验证 `delete_time` 不为 0
  - **预计时间**: 1 小时
  - **负责人**: 

- [ ] **编写历史订单查询测试** `main/app/service/order_manage_test.go::TestGetOrderDetailWithDeletedConfig()`
  - **需求**: 验证历史订单仍可显示已删除的配置
  - **测试用例**:
    1. 创建配置
    2. 创建订单，关联该配置
    3. 软删除配置
    4. 调用 GetOrderDetail()
    5. 验证订单详情仍显示配置名称
  - **预计时间**: 1 小时
  - **负责人**: 

---

### 阶段 2: 集成测试（0.5 天）

#### 2.1 完整流程测试

- [ ] **外卖来源完整流程测试**
  - **测试步骤**:
    1. 新管理端新增外卖来源"美团外卖"
    2. 验证列表最新记录在顶部
    3. 收银端创建订单，选择"美团外卖"
    4. 新管理端删除"美团外卖"配置
    5. 验证配置列表不显示"美团外卖"
    6. 收银端查询订单详情，验证仍显示"美团外卖"
  - **预计时间**: 1 小时
  - **负责人**: 

- [ ] **国籍配置完整流程测试**
  - **测试步骤**: （同外卖来源）
  - **预计时间**: 1 小时
  - **负责人**: 

#### 2.2 边界测试

- [ ] **删除幂等性测试**
  - **测试场景**: 删除同一配置两次，验证不报错
  - **预计时间**: 0.5 小时
  - **负责人**: 

- [ ] **大量数据排序性能测试**
  - **测试场景**: 创建 100+ 配置，验证列表查询性能 < 100ms
  - **预计时间**: 0.5 小时
  - **负责人**: 

- [ ] **并发删除测试**
  - **测试场景**: 多个用户同时删除配置，验证数据一致性
  - **预计时间**: 0.5 小时
  - **负责人**: 

---

### 阶段 3: 文档更新（0.5 天）

#### 3.1 更新技术文档

- [ ] **更新 Bug 修复方案** `solution.md`
  - **需求**: 确认根因分析正确（表已有字段，只需修改代码）
  - **预计时间**: 0.5 小时
  - **负责人**: 

- [ ] **创建故障排查指南** `docs/shared/troubleshooting/admin-config-soft-delete.md`
  - **需求**: 记录软删除机制和常见问题
  - **内容**:
    - 软删除原理
    - 如何恢复误删除的配置
    - 常见问题 FAQ
  - **预计时间**: 1 小时
  - **负责人**: 

#### 3.2 记录经验到 Graphiti

- [ ] **创建 Graphiti Episode**
  - **需求**: 记录本次 Bug 修复经验
  - **内容**:
    - 问题根因分析
    - 软删除实施方案
    - 历史数据保护策略
    - 预防措施
  - **预计时间**: 1 小时
  - **负责人**: 

---

### 阶段 4: 部署上线（0.5 天）

#### 4.1 预发布准备

- [ ] **数据库备份（可选）**
  - **需求**: 备份 `ttpos_order_source` 和 `ttpos_nationality` 表（预防性措施）
  - **说明**: 本次修复无数据库变更，备份可选
  - **预计时间**: 0.5 小时
  - **负责人**: 

- [ ] **Code Review**
  - **需求**: 通过团队 Code Review
  - **检查点**:
    - 迁移脚本正确性
    - 软删除实现规范性
    - 测试覆盖率
  - **预计时间**: 1 小时
  - **负责人**: 

#### 4.2 生产发布

- [ ] **部署 Main 模块**
  - **需求**: 部署 Go 后端代码
  - **步骤**:
    1. `git pull`
    2. `go build`
    3. `systemctl restart ttpos-main`
  - **预计时间**: 0.5 小时
  - **负责人**: 

#### 4.3 发布验证

- [ ] **冒烟测试**
  - **测试用例**:
    - 查询配置列表，验证排序正确
    - 删除配置，验证成功
    - 查询历史订单，验证配置显示正常
  - **预计时间**: 0.5 小时
  - **负责人**: 

- [ ] **监控观察**
  - **需求**: 发布后 2 小时内持续监控
  - **监控指标**:
    - API 响应时间
    - 错误率
    - 数据库慢查询
  - **预计时间**: 2 小时
  - **负责人**: 

---

### 阶段 5: Bug 归档（0.5 天）

#### 5.1 最终验证

- [ ] **生产环境验证**
  - **需求**: 在生产环境复现原问题，验证已修复
  - **验证点**:
    - 新增配置后，列表最新记录在顶部 ✅
    - 删除已使用的配置，成功删除 ✅
    - 历史订单详情仍显示配置名称 ✅
  - **预计时间**: 1 小时
  - **负责人**: 

#### 5.2 Bug 归档

- [ ] **更新 Bug 状态为「已修复」**
  - **需求**: 更新 `bug.md` 状态
  - **预计时间**: 0.5 小时
  - **负责人**: 

- [ ] **执行 `/bug-archive`**
  - **需求**: 将 Bug 归档到 `docs/shared/bugs/resolved/v2.10.10/`
  - **预计时间**: 0.5 小时
  - **负责人**: 

---

## 📊 任务统计

| 阶段 | 总任务数 | 预计时间 | 负责人 | 状态 |
| ---- | -------- | -------- | ------ | ---- |
| **阶段 1: Go Main 后端** | 10 | 1.5 天 | Go 工程师 | ⏸️ 待开始 |
| **阶段 2: 集成测试** | 4 | 0.5 天 | 测试工程师 | ⏸️ 待开始 |
| **阶段 3: 文档更新** | 2 | 0.5 天 | 开发工程师 | ⏸️ 待开始 |
| **阶段 4: 部署上线** | 4 | 0.5 天 | 运维+后端 | ⏸️ 待开始 |
| **阶段 5: Bug 归档** | 2 | 0.5 天 | 开发工程师 | ⏸️ 待开始 |
| **总计** | **22** | **3 天** | | ⏸️ 待开始 |

---

## 📈 进度跟踪

### 总体进度

- **总任务数**: 22
- **已完成**: 0
- **进行中**: 0
- **未开始**: 22
- **完成率**: 0%

### 里程碑

| 里程碑 | 预计完成日期 | 状态 |
| ------ | ------------ | ---- |
| 后端开发完成 | 2025-11-27 下午 | ⏸️ 待开始 |
| 测试完成 | 2025-11-28 上午 | ⏸️ 待开始 |
| 生产发布 | 2025-11-28 下午 | ⏸️ 待开始 |
| Bug 归档 | 2025-11-28 下午 | ⏸️ 待开始 |

---

## 🔗 相关链接

- **Bug 报告**: `bug.md`
- **修复方案**: `solution.md`
- **关联 Spec**: `docs/shared/specs/active/story-order-source-nationality/`
- **DooTask**: [任务 #37152](http://t.hitosea.com/project/task/37152)
- **数据库规范**: `.cursor/rules/database.mdc`
- **PHP 规范**: `.cursor/rules/php.mdc`
- **Go Main 规范**: `.cursor/rules/go-main.mdc`

---

## 📝 备注

### 任务分配建议

- **Go Main 后端修改**：Go 工程师（熟悉 GORM）
- **单元测试**：Go 工程师
- **集成测试**：测试工程师 + Go 工程师
- **部署上线**：运维工程师 + Go 工程师

### 风险提示

1. **代码逻辑修改**：需要充分测试软删除逻辑
2. **生产发布时间**：建议选择低峰期（如周四晚上或周五上午）
3. **回滚准备**：只需回滚代码，无数据库变更
4. **监控预警**：发布后持续监控 2 小时，发现异常立即回滚

### 优势说明

✅ **无数据库迁移风险**：表已有 `delete_time` 字段，只需修改代码  
✅ **实施快速**：预计 2-3 天完成（比原计划少 0.5 天）  
✅ **回滚简单**：无数据库变更，回滚代码即可

---

**创建时间**: 2025-11-27  
**创建者**: weifashi  
**当前阶段**: 规划中 → 待分配任务和开始实施

