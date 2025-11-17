---
Status: Deprecated
---

# 数据库迁移冲突（Phinx + Go Model）

> 适用范围：`admin/database/migrations/`, `main/app/model/`, `admin/database/seeds/`

## 问题现象

- 执行 `php think migrate:run` 提示表已存在/字段重复。
- CI 中 `go test ./...` 报 `unknown column`，说明 Go Model 与数据库结构不一致。
- 多名开发者在同一表新增字段，生成的迁移顺序不同导致冲突。

## 问题原因

1. 迁移文件缺少存在性判断，例如未使用 `$this->hasTable()` / `$table->hasColumn()`。
2. 新增字段后未同步更新 `main/app/model/*.go`，导致 struct 与 DB 不一致。
3. 手工修改数据库而未写迁移，下一次运行导致重复建表。

## 解决方案

1. **回滚并定位冲突**
   ```bash
   cd /Users/benbige/Projects/ttpos-server-go/admin
   php think migrate:status          # 查看执行顺序
   php think migrate:rollback --step=1
   ```
2. **为迁移添加存在性判断**

   ```php
   public function change()
   {
       if ($this->hasTable('ttpos_order')) {
           return;
       }
       $table = $this->table('ttpos_order', [...]);
       ...
   }
   ```

   或在 `up/down` 方法中使用 `$table->hasColumn('is_quick_payment')`。

3. **同步 Go Model**

   ```go
   // main/app/model/order.go
   type Order struct {
       IsQuickPayment uint8 `gorm:"column:is_quick_payment" json:"is_quick_payment"`
   }
   func (Order) TableName() string { return "ttpos_order" }
   ```

   - 确保新增字段/索引与数据库一致，运行 `go test ./main/app/model/...`.

4. **合并迁移顺序**
   - 若多人改同表，合并到同一个迁移文件或按日期重新排序（`YYYYMMDDHHMMSS_description.php`）。
   - `git rebase` 后重新执行 `php think migrate:run` 以验证顺序。

## 预防措施

- 在 PR 检查清单中新增：**迁移文件存在性判断 + Go Model 同步 + Seeds 更新**。
- 约定先运行 `php think migrate:status` 并与团队确认是否已有相同字段。
- 对共享表的改动在 Spec/Design 中写清楚字段名、默认值、约束。

## 相关资源

- `.cursor/rules/database.mdc`
- `.cursor/rules/php.mdc`
- `docs/agent/workflows/database-migration.md`
- `admin/database/migrations/`
- `main/app/model/`

Related Episode: `[待补充 by @责任人]`
