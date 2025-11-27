# Bug-251127-003 修复方案

> 外卖/国籍管理-排序和删除功能缺陷

---

## 问题概述

新管理端的外卖和国籍管理功能存在两个核心问题：

1. **列表排序问题**：新增配置项按创建时间正序显示，最新记录在底部不便查看
2. **删除限制问题**：已使用的配置项无法删除，影响数据维护

---

## 根本原因

### 问题 1：排序问题

**根因**：
- 后端 API 查询配置列表时，未指定排序规则
- 数据库查询默认使用主键 `id` 排序（等同于创建时间正序）
- 前端直接渲染后端返回的数据，未做额外排序处理

**涉及代码**（推测）：
```go
// main/app/service/order_source.go
func (s *OrderSourceService) GetList(shopId int) ([]*model.OrderSource, error) {
    var list []*model.OrderSource
    // 当前查询：未指定排序，默认按 id 排序
    err := s.db.Where("shop_id = ?", shopId).Find(&list).Error
    // 应改为：ORDER BY create_time DESC
    return list, err
}
```

---

### 问题 2：删除限制

**根因**：
- 数据表 `ttpos_order_source` 和 `ttpos_nationality` **虽然已有 `delete_time` 字段**，但**代码层面未使用软删除机制**
- 删除接口使用硬删除（`DELETE FROM`），直接删除数据库记录，而不是更新 `delete_time`
- 订单表 `ttpos_order` 中存储了外键引用（`order_source_id`, `nationality_id`）
- 硬删除会导致历史订单的外键关联失效，引发数据完整性问题
- 后端在删除前检查是否被订单引用，如果被引用则禁止删除

**数据库现状**（已正确设计）：

根据 `admin/database/seeds/shop_01.sql`，表结构已包含软删除字段：

```sql
-- ttpos_order_source 表（外卖来源配置）
CREATE TABLE IF NOT EXISTS `ttpos_order_source` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `sort` INT(11) NOT NULL DEFAULT 0,
  `status` INT(3) NOT NULL DEFAULT 1,
  `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0,
  `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0,
  `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0,  -- ✅ 已有软删除字段
  UNIQUE KEY `uk_uuid` (`uuid`),
  INDEX `idx_delete_time` (`delete_time`)  -- ✅ 已有索引
) COMMENT = '外卖来源配置表';

-- ttpos_nationality 表（国籍配置）同样已有 delete_time 字段和索引
```

**涉及代码**（推测）：
```
main/app/api/v1/shop/order_source.go        # 删除接口 API
main/app/service/order_source.go            # 删除服务层逻辑（当前硬删除）
main/app/repository/order_source.go         # 数据访问层（如果使用）
main/app/service/order_manage.go            # 订单查询相关
```

**问题链**：
1. 管理员删除配置 → **代码执行硬删除**（`DELETE FROM`）而非更新 `delete_time`
2. 历史订单的 `order_source_id` / `nationality_id` 变为无效值
3. 订单详情查询时关联查询失败，显示空白或报错
4. 系统为保护数据完整性，禁止删除已使用的配置

**核心问题**：数据库设计正确，但**代码层面未启用软删除**

**修复方向**：
1. Service 层删除方法改为软删除（更新 `delete_time` 字段）
2. Repository 层实现软删除逻辑（如果使用 Repository 模式）
3. 查询时过滤已删除的配置项（`WHERE delete_time = 0`）
4. 历史订单展示时不过滤，保留完整信息

**数据库现状**：
- ✅ 表已有 `delete_time` 字段（无需迁移）

---

## 修复方案

### 方案对比

#### 方案 1：启用软删除机制（推荐）

**实施内容**：
1. ✅ 数据表已有 `delete_time` 字段（无需迁移）
2. Go Service 层实现软删除逻辑
3. 删除接口改为软删除（更新 `delete_time = 当前时间戳`）
4. 查询接口过滤已删除记录（`WHERE delete_time = 0`）
5. 历史订单详情仍可查询到已删除的配置项（不过滤 `delete_time`）

**优点**：
- ✅ 数据库已正确设计，**无需迁移，只需修改代码**
- ✅ 符合 TTPOS 数据库规范（表已包含 `delete_time`）
- ✅ 保护历史数据完整性，订单详情仍可显示原配置名称
- ✅ 支持未来恢复误删除的配置
- ✅ 符合 Go Main 规范（Service 层实现业务逻辑）
- ✅ 实施简单快速，风险极低

**缺点**：
- ⚠️ 需要更新 Go 代码逻辑（Service + Repository 层）

**风险**：
- 🟢 **极低**：无数据库变更，只修改代码逻辑
- 🟢 **极低**：历史订单不受影响（订单表不变）
- 🟢 **极低**：表已有字段和索引，性能无影响

---

#### 方案 2：保持硬删除，放开删除限制

**实施内容**：
1. 删除接口移除"已使用检查"
2. 允许删除已被订单引用的配置
3. 订单详情查询时处理关联失败情况（显示"已删除"或 ID）

**优点**：
- ✅ 无需数据库迁移
- ✅ 实现简单

**缺点**：
- ❌ 历史订单丢失配置名称，用户体验差
- ❌ 违反数据库规范（未使用软删除）
- ❌ 无法恢复误删除的配置
- ❌ 破坏数据完整性

**风险**：
- 🔴 **高**：历史订单数据不完整
- 🟡 **中**：可能引发其他关联查询问题

---

### ✅ 最终选择：方案 1（软删除机制）

**理由**：
1. **符合规范**：表已包含 `delete_time` 字段，符合 `database.mdc` 规范
2. **保护数据**：历史订单仍可查询到原配置名称
3. **最佳实践**：与 `story-order-source-nationality` 设计文档保持一致
4. **Graphiti 经验**：知识库明确指出"设计数据模型时应考虑软删除需求"
5. **Go Main 规范**：Service 层负责业务逻辑，Repository 只持有 db 实例

---

## 实施步骤

### 阶段 1：Go Main 后端修改

> **注意**：数据表已有 `delete_time` 字段，无需数据库迁移，只需修改 Go 代码

**步骤 1.1**: 修改外卖来源 Service
- 文件: `main/app/service/order_source.go`
- 修改 `GetList()` 方法：添加 `WHERE delete_time = 0` + `ORDER BY create_time DESC`
- 修改 `Delete()` 方法：改为更新 `delete_time = time.Now().Unix()`
- 新增 `GetByIdWithDeleted()` 方法：用于订单详情查询（不过滤 `delete_time`）

**步骤 1.2**: 修改国籍 Service
- 文件: `main/app/service/nationality.go`
- 同外卖来源的修改

**步骤 1.3**: 修改 Repository 层（如果存在）
- 文件: `main/app/repository/order_source.go`, `main/app/repository/nationality.go`
- 更新查询和删除方法，支持软删除

**步骤 1.4**: 修改订单详情查询
- 文件: `main/app/service/order_manage.go`
- 订单详情关联查询时调用 `GetByIdWithDeleted()` 方法
- 保证历史订单仍可显示已删除配置

**步骤 1.5**: 修改 API 层（如果需要）
- 文件: `main/app/api/v1/shop/order_source.go`, `main/app/api/v1/shop/nationality.go`
- 确保 API 调用正确的 Service 方法

---

### 阶段 2：前端适配（可选优化）

**步骤 2.1**: 前端列表排序备份
- 如果后端排序失败，前端可添加本地排序逻辑
- 文件: Flutter 新管理端配置页面（推测）

---

## 技术方案

> **重要说明**：数据表已有 `delete_time` 字段和索引，无需数据库迁移，直接修改 Go 代码即可。

---

### Go Service 层修改

```go
// main/app/service/order_source.go

// GetOrderSourceList 获取外卖来源列表（过滤已删除 + 倒序排序）
func (s *OrderSourceService) GetList(shopId int) ([]*model.OrderSource, error) {
    var list []*model.OrderSource
    
    err := s.db.Where("shop_id = ? AND delete_time = 0", shopId).
        Order("create_time DESC").  // ✅ 倒序排序
        Find(&list).Error
    
    return list, err
}

// Delete 软删除外卖来源（更新 delete_time）
func (s *OrderSourceService) Delete(id int) error {
    now := time.Now().Unix()
    
    // ✅ 软删除：更新 delete_time 字段
    err := s.db.Model(&model.OrderSource{}).
        Where("id = ?", id).
        Update("delete_time", now).Error
    
    return err
}

// GetByIdWithDeleted 获取外卖来源（包含已删除）- 用于订单详情
func (s *OrderSourceService) GetByIdWithDeleted(id int) (*model.OrderSource, error) {
    var item model.OrderSource
    
    // ⚠️ 不过滤 delete_time，保留历史记录
    err := s.db.Where("id = ?", id).First(&item).Error
    
    return &item, err
}
```

---

### Go Repository 层修改（如果使用 Repository 模式）

```go
// main/app/repository/order_source.go

type OrderSourceRepository struct {
    db *gorm.DB
}

// FindList 查询列表（过滤已删除）
func (r *OrderSourceRepository) FindList(shopId int) ([]*model.OrderSource, error) {
    var list []*model.OrderSource
    
    err := r.db.Where("shop_id = ? AND delete_time = 0", shopId).
        Order("create_time DESC").
        Find(&list).Error
    
    return list, err
}

// SoftDelete 软删除
func (r *OrderSourceRepository) SoftDelete(id int) error {
    return r.db.Model(&model.OrderSource{}).
        Where("id = ?", id).
        Update("delete_time", time.Now().Unix()).Error
}

// FindByIdWithDeleted 查询单条（不过滤已删除）
func (r *OrderSourceRepository) FindByIdWithDeleted(id int) (*model.OrderSource, error) {
    var item model.OrderSource
    err := r.db.Where("id = ?", id).First(&item).Error
    return &item, err
}
```

---

### 订单详情查询修改

```go
// main/app/service/order_manage.go

// GetOrderDetail 获取订单详情
func (s *OrderManageService) GetOrderDetail(orderId int) (*OrderDetail, error) {
    // ... 查询订单基本信息
    
    // 关联查询外卖来源（使用 GetByIdWithDeleted，不过滤已删除）
    if order.OrderSourceId > 0 {
        orderSource, err := s.orderSourceService.GetByIdWithDeleted(order.OrderSourceId)
        if err == nil {
            detail.OrderSourceName = orderSource.Name
        }
    }
    
    // 关联查询国籍（同理）
    if order.NationalityId > 0 {
        nationality, err := s.nationalityService.GetByIdWithDeleted(order.NationalityId)
        if err == nil {
            detail.NationalityName = nationality.Name
        }
    }
    
    return detail, nil
}
```

---

## 影响分析

### 兼容性

| 项目 | 影响 | 说明 |
| ---- | ---- | ---- |
| **历史订单** | ✅ 无影响 | 订单详情仍可查询到原配置名称 |
| **API 接口** | ✅ 无影响 | 响应格式不变，只是数据顺序调整 |
| **前端页面** | ✅ 无影响 | 前端无需改动 |
| **数据库** | ✅ 无变更 | 表已有字段，无需修改 |

---

### 性能影响

| 项目 | 影响 | 说明 |
| ---- | ---- | ---- |
| **查询性能** | 🟢 无影响 | 添加索引 `idx_delete_time`，查询性能不降低 |
| **写入性能** | 🟢 无影响 | 软删除只更新一个字段 |
| **存储空间** | 🟢 极小 | 每条记录增加 4 字节（int 类型） |

---

### 安全风险

| 风险 | 等级 | 缓解措施 |
| ---- | ---- | -------- |
| **代码逻辑错误** | 🟡 中 | 充分的单元测试和集成测试 |
| **数据丢失** | 🟢 极低 | 软删除不删除数据，支持恢复 |
| **权限问题** | 🟢 极低 | 删除接口已有权限校验 |

---

## 测试计划

### 单元测试

**测试 1**: 配置列表排序
- **文件**: `main/app/service/order_source_test.go`
- **用例**: 新增配置后，查询列表，验证最新记录在第一位

**测试 2**: 软删除功能
- **文件**: `main/app/service/order_source_test.go`
- **用例**: 删除配置后，查询列表，验证该配置不在列表中

**测试 3**: 历史订单查询
- **文件**: `main/app/service/order_manage_test.go`
- **用例**: 查询包含已删除配置的订单，验证仍可显示配置名称

---

### 集成测试

**测试场景 1**: 完整流程
1. 新增外卖来源配置"美团外卖"
2. 创建订单，选择"美团外卖"
3. 删除"美团外卖"配置
4. 查询配置列表，验证"美团外卖"不在列表
5. 查询订单详情，验证仍显示"美团外卖"

**测试场景 2**: 国籍配置同样流程

---

### 手动测试清单

**功能测试**：
- [ ] 新增配置后，列表最新记录在顶部
- [ ] 删除未使用的配置，成功删除
- [ ] 删除已使用的配置，成功删除
- [ ] 删除后，配置列表不显示该配置
- [ ] 历史订单详情仍显示已删除的配置名称

**边界测试**：
- [ ] 删除同一配置两次，验证幂等性
- [ ] 大量配置时的排序性能
- [ ] 并发删除配置的情况

**兼容性测试**：
- [ ] 历史订单（迁移前创建）的详情展示
- [ ] 未开启功能的门店不受影响

---

## 上线计划

### 发布时间

**预计版本**: v2.10.10  
**预计发布**: 2025-11-28（本周四）

**发布步骤**：
1. **11-27（周三）**: 完成 Go 代码开发和单元测试
2. **11-28（周四上午）**: 测试环境验证和集成测试
3. **11-28（周四下午）**: 生产环境发布

---

### 部署流程

**Step 1**: 停服维护（可选，低峰期发布可不停服）

**Step 2**: 部署 Go Main 模块
```bash
cd main
git pull
go build -o ttpos-main
systemctl restart ttpos-main
```

**Step 3**: 验证
- 检查配置列表排序
- 测试删除功能
- 验证历史订单详情

**Step 4**: 恢复服务（如有停服）

---

### 回滚方案

**场景 1**: 功能异常
- 回滚 Go Main 代码到上一版本
- 重新编译并重启服务
```bash
cd main
git checkout <上一个版本tag>
go build -o ttpos-main
systemctl restart ttpos-main
```

**数据保护**：
- 无数据库变更，回滚风险极低
- 软删除不丢失数据，误删除可恢复

---

### 监控指标

**关键指标**：
1. **配置查询响应时间**: < 100ms
2. **删除操作成功率**: > 99%
3. **订单详情查询成功率**: > 99.9%
4. **数据库慢查询**: 监控是否有新增慢查询

**监控工具**：
- SkyWalking：监控 API 响应时间
- Prometheus：监控数据库查询性能
- 日志系统：监控错误日志

---

## 预防措施

### 如何避免类似问题再次发生

1. **数据库设计阶段**
   - ✅ 所有配置表必须包含 `delete_time` 字段
   - ✅ 遵循 `database.mdc` 规范检查清单
   - ✅ Code Review 时检查表结构

2. **API 设计阶段**
   - ✅ 列表查询接口必须明确指定排序规则
   - ✅ 默认使用倒序排序（最新记录在前）
   - ✅ 文档中明确说明排序逻辑

3. **需求评审阶段**
   - ✅ 涉及配置管理的需求，必须讨论删除策略
   - ✅ 明确软删除 vs 硬删除的场景
   - ✅ 评估历史数据关联影响

4. **测试阶段**
   - ✅ 配置管理功能必须测试删除场景
   - ✅ 必须测试历史数据关联展示
   - ✅ 包含边界测试（大量数据、并发操作）

---

## 经验总结

### 关键教训

1. **设计先行**：数据表设计时就应考虑软删除需求
2. **规范为王**：严格遵循 `database.mdc` 规范可避免此类问题
3. **历史保护**：涉及外键关联的配置表，必须使用软删除
4. **用户体验**：列表排序看似小事，实则直接影响用户体验

### 最佳实践

1. **配置表标配**：
   - `delete_time` 软删除字段
   - `sort` 排序字段
   - `status` 启用/禁用字段

2. **删除接口设计**：
   - 配置表统一使用软删除
   - 关联查询时根据场景决定是否过滤 `delete_time`

3. **排序规范**：
   - 列表查询默认倒序（最新在前）
   - 明确指定 `ORDER BY` 子句，不依赖数据库默认行为

---

## 相关链接

### 关联文档
- **Bug 报告**: `bug.md`
- **任务清单**: `tasks.md`
- **关联 Spec**: `docs/shared/specs/active/story-order-source-nationality/`
- **数据库规范**: `.cursor/rules/database.mdc`
- **Go Main 规范**: `.cursor/rules/go-main.mdc`

### Graphiti 知识库
- 无法删除已使用的配置项 - 根因是删除功能未实现软删除且外键约束阻止删除
- 修改删除接口实现软删除并在查询时过滤已删除记录
- 设计数据模型时应考虑软删除需求以避免历史数据引用问题
- Go GORM 实现软删除：使用 `.Update("delete_time", time.Now().Unix())`

### DooTask
- [任务 #37152 - 新管理端业务设置-外卖+国籍控制相关](http://t.hitosea.com/project/task/37152)

---

**创建时间**: 2025-11-27  
**创建者**: weifashi  
**审核者**: 待指定  
**状态**: ✅ 方案完成，待审核

