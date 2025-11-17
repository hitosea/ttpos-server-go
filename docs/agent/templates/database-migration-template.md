# 数据库迁移模板

> 🤖 Agent 视角模板：用于在 `admin/database/migrations/` 中创建 PHP Phinx 迁移，并同步 Go Model、Seeds、文档。

---

## 元信息

| 字段      | 内容                                                |
| --------- | --------------------------------------------------- | -------------------------- |
| 迁移名称  | `{YYYYMMDDHHMMSS}_add_{table_or_column}_fields.php` |
| 触发 Spec | `docs/shared/specs/{story                           | task}-{module}-{feature}/` |
| 负责人    | `@`                                                 |
| 关联任务  | `tasks.md` 中的编号                                 |

---

## 1. 变更摘要

- **表/字段**：`ttpos_*`
- **类型**：`新增表 / 新增字段 / 修改字段 / 删除字段`
- **目的**：`[待补充]`
- **影响范围**：Go Model / PHP Admin / Seeds / 其他服务

---

## 2. 规范对齐

- `.cursor/rules/database.mdc`
  - 时间字段：`int`，`_time` 结尾，默认 0
  - 金额字段：`decimal(20,8)`
  - 必备字段：`uuid`, `create_time`, `update_time`, `delete_time`
- `.cursor/rules/go-main.mdc` / `.cursor/rules/php.mdc`（模型同步、命名规则）
- `.cursor/rules/security.mdc`（敏感数据处理）

---

## 3. 迁移脚本

```php
<?php
use think\migration\Migrator;
use think\migration\db\Column;

class {ClassName} extends Migrator
{
    public function change()
    {
        // 示例：为 ttpos_company 增加 default_payment_method
        $table = $this->table('ttpos_company');
        if (!$table->hasColumn('default_payment_method')) {
            $table->addColumn('default_payment_method', 'integer', [
                'limit' => 1,
                'default' => 1,
                'comment' => '默认支付方式：1-现金，2-微信...',
                'after' => 'status',
            ])->update();
        }
    }
}
```

> **必填**：字段存在性检查、注释、默认值。若涉及索引/外键，同步 addIndex / addForeignKey。

---

## 4. Go Model & Seeds

- **Go Model** (`main/app/model/*.go`)
  ```go
  type Company struct {
      DefaultPaymentMethod uint8 `gorm:"column:default_payment_method" json:"default_payment_method"`
  }
  ```
- **Seeds** (`admin/database/seeds/*.sql`)
  ```sql
  UPDATE ttpos_company SET default_payment_method = 1 WHERE default_payment_method = 0;
  ```
- **同步检查**
  - [ ] Model 字段 + `TableName()` 方法
  - [ ] JSON / gorm 标签
  - [ ] 相关 DTO/Service/Repository 更新

---

## 5. 执行步骤

```bash
cd admin
php think migrate:status
php think migrate:run

# 回滚（如需）
php think migrate:rollback --step=1
```

**验证**：

- [ ] 迁移成功，无错误。
- [ ] 本地/测试环境数据库字段已生效。
- [ ] Seeds/默认值符合预期。

---

## 6. 测试清单

- [ ] 受影响 API / Service 单元测试通过。
- [ ] 数据迁移前后数据一致性验证。
- [ ] 回滚脚本验证（如必要）。
- [ ] 性能评估（大表添加字段需关注锁表时间）。

---

## 7. 文档 & 链接

- 更新 `docs/shared/specs/{spec}/design.md` 的数据库章节。
- 如涉及接口，更新 `docs/shared/api/{module}_api.md`。
- 在 `requirements.md` / `design.md` / `tasks.md` 的 `Graphiti & 活动日志` 区域记录：
  - Related Episode: `[待补充]`
  - 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

## 8. 回顾

- 风险与缓解措施
- 需要补充的预防性任务
- 是否需要新增 Graphiti Episode（是/否）

---

**最后更新**：2025-11-17
