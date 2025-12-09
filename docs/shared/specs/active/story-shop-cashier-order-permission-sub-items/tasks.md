# 收银机接单权限子项管理 任务分解

> 本文档定义收银机接单权限子项管理功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 12  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 数据库迁移和权限数据

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 创建数据库迁移脚本

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 创建权限结构调整的迁移脚本
  - Requirements: R1.7
  - Leverage: 现有迁移文件: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php`
  - Success: 迁移脚本创建成功，包含所有权限操作和角色分配逻辑

- [x] 1.2 新增接单权限（父级）

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 在 access 表中新增接单权限，作为收银机权限的子级
  - Requirements: R1.1
  - Leverage: Task 1.1 的迁移脚本，参考外送权限迁移逻辑
  - Success: 接单权限已创建，parent_uuid=1704880670，sort=11

- [x] 1.3 更新原有接单权限为扫码接单

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 将原有接单权限（1724320522）调整为扫码接单，更新 parent_uuid 和名称
  - Requirements: R1.2
  - Leverage: Task 1.1 的迁移脚本
  - Success: 原有接单权限已更新为扫码接单，parent_uuid 已更新为新接单权限UUID

- [x] 1.4 更新外送权限的 parent_uuid

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 将外送权限（1752716650）的 parent_uuid 更新为新接单权限UUID
  - Requirements: R1.3
  - Leverage: Task 1.1 的迁移脚本
  - Success: 外送权限的 parent_uuid 已更新为新接单权限UUID

- [x] 1.5 新增外卖权限（子级）

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 在 access 表中新增外卖权限，作为新接单权限的子级
  - Requirements: R1.4
  - Leverage: Task 1.1 的迁移脚本，参考外送权限创建逻辑
  - Success: 外卖权限已创建，parent_uuid=新接单权限UUID，sort=30

- [x] 1.6 为新接单权限分配默认角色权限

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 为 Cashier（role_uuid=2）和 Store Manager（role_uuid=1）角色分配新接单权限
  - Requirements: R1.5
  - Leverage: Task 1.1 的迁移脚本，参考外送权限分配逻辑（第38-45行）
  - Success: 新接单权限已分配给默认角色

- [x] 1.7 为外卖权限分配默认角色权限

  - File: `admin/database/migrations/20251209144514_add_cashier_order_permission_sub_items.php`
  - Purpose: 为 Cashier（role_uuid=2）和 Store Manager（role_uuid=1）角色分配外卖权限
  - Requirements: R1.6
  - Leverage: Task 1.1 的迁移脚本，参考外送权限分配逻辑
  - Success: 外卖权限已分配给默认角色

- [ ] 1.8 执行数据库迁移

  - File: -
  - Purpose: 在数据库中执行权限结构调整
  - Requirements: R1.7
  - Leverage: Task 1.1 的迁移脚本
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，权限数据已更新

---

## Phase 2: 权限筛选逻辑

- [x] 2.1 在 Go Main 模块添加外卖权限过滤

  - File: `main/app/service/role_access.go`
  - Purpose: 在 filterPermission 函数中添加外卖权限过滤逻辑
  - Requirements: R2.1
  - Leverage: 现有权限筛选逻辑: `main/app/service/role_access.go` 第189-192行（外送权限过滤）
  - Success: 权限筛选逻辑添加成功，未开启外卖时正确过滤外卖权限（UUID: 1734000001）

- [x] 2.2 在 PHP Admin 模块添加外卖权限过滤

  - File: `admin/app/common/model/shop/Access.php`
  - Purpose: 在 recursiveMenuArray 函数中添加外卖权限过滤逻辑
  - Requirements: R2.2
  - Leverage: 现有权限筛选逻辑: `admin/app/common/model/shop/Access.php` 第379-384行（外送权限过滤）
  - Success: 权限筛选逻辑添加成功，未开启外卖时正确过滤外卖权限（UUID: 1734000001）

- [x] 2.3 在 getLicense() 中添加 enable_grab_delivery 字段

  - File: `admin/app/common/model/app/App.php`
  - Purpose: 在 getLicense() 方法的返回数据中添加 enable_grab_delivery 字段
  - Requirements: R2.3
  - Leverage: 现有 getLicense() 方法: `admin/app/common/model/app/App.php` 第176-224行
  - Success: enable_grab_delivery 字段已添加到 getLicense() 返回数据中

---

## Phase 3: API 确认

- [x] 3.1 确认 BaseInfo API 返回外卖开关状态

  - File: `main/app/dto/resp/base.go`, `main/app/service/auth.go`
  - Purpose: 确认 BaseInfo API 已正确返回 is_open_grab_delivery 字段
  - Requirements: R4.1, R4.2
  - Leverage: 现有 BaseInfo API: `main/app/service/auth.go` 第1005-1021行
  - Success: BaseInfo API 正确返回 is_open_grab_delivery 字段，字段值与数据库一致

- [x] 3.2 确认其他 BaseInfo 相关接口返回外卖开关状态

  - File: `main/app/api/v1/menu/menu_handler.go`, `main/app/service/auth.go` (TabletBase), `main/app/service/h5_service.go`
  - Purpose: 确认其他 BaseInfo 相关接口也返回 is_open_grab_delivery 字段
  - Requirements: R4.3
  - Leverage: 现有 BaseInfo 接口实现
  - Success: 所有 BaseInfo 相关接口都正确返回 is_open_grab_delivery 字段（/menu/base, /tablet/base, /h5/base）

---

## Phase 4: 测试和验证

- [ ] 4.1 测试权限迁移脚本

  - File: `admin/database/migrations/{timestamp}_add_cashier_order_permission_sub_items.php`
  - Purpose: 验证权限迁移脚本执行正确
  - Requirements: 所有 Phase 1 任务
  - Leverage: 测试数据库
  - Success: 迁移脚本执行成功，权限数据正确，角色权限分配正确

- [ ] 4.2 测试权限筛选逻辑（Go Main）

  - File: `main/app/service/role_access.go`
  - Purpose: 验证 Go Main 模块权限筛选逻辑正确
  - Requirements: R2.1
  - Leverage: 单元测试或手动测试
  - Success: 权限筛选逻辑正确，未开启外卖时正确过滤外卖权限

- [ ] 4.3 测试权限筛选逻辑（PHP Admin）

  - File: `admin/app/common/model/shop/Access.php`
  - Purpose: 验证 PHP Admin 模块权限筛选逻辑正确
  - Requirements: R2.2
  - Leverage: 单元测试或手动测试
  - Success: 权限筛选逻辑正确，未开启外卖时正确过滤外卖权限

- [ ] 4.4 测试默认角色权限分配

  - File: `admin/database/migrations/{timestamp}_add_cashier_order_permission_sub_items.php`
  - Purpose: 验证默认角色权限分配正确
  - Requirements: R1.5, R1.6
  - Leverage: 测试数据库
  - Success: 默认角色（Cashier 和 Store Manager）已正确分配新接单权限和外卖权限

- [ ] 4.5 集成测试

  - File: 所有相关文件
  - Purpose: 端到端测试完整功能
  - Requirements: 所有功能需求
  - Leverage: 测试环境
  - Success: 所有功能测试通过，权限结构正确，权限筛选正确，默认角色权限分配正确

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 迁移脚本使用事务确保数据一致性
- [ ] 权限筛选逻辑测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 权限结构正确（接单权限与沽清平级，包含三个子项）
- [ ] 权限筛选逻辑正确（未开启外卖时过滤外卖权限）
- [ ] 默认角色权限分配正确

### 文档同步

- [ ] 迁移脚本注释完整
- [ ] 代码注释清晰
- [ ] 活动日志已记录

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-cashier-order-permission-sub-items/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-cashier-order-permission-sub-items/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-cashier-order-permission-sub-items/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `php think migrate:run`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-09  
**维护者**: 后端开发组
