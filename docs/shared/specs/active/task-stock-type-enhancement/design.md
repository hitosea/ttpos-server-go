# 盘点单类型字段扩展 设计文档

> 本文档定义盘点单类型字段扩展的技术设计和实施方案。

## 📋 概述

扩展盘点单表（`ttpos_stock_reconciliation`）的 `type` 字段类型定义，新增日盘（3）、周盘（4）、月盘（5）三种周期性盘点类型。本次变更仅涉及字段注释更新，不涉及数据迁移和业务逻辑变更。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- 仅修改 Model 文件注释，保持代码结构不变
- 注释使用中文，清晰说明每个类型的含义

### 数据库规范 (database.mdc)

- 必须同步更新 `admin/database/seeds/shop_01.sql`
- 迁移文件必须幂等（检查是否已修改）
- 字段类型保持 `int` 不变
- 必须包含清晰的注释说明

---

## 🔄 代码复用分析

### 可复用的现有组件

- **迁移文件模板**: `admin/database/migrations/*` - 参考现有迁移文件的结构和写法
- **Model 文件**: `main/app/model/stock_reconciliation.go` - 直接修改注释

### 集成点

- **数据库表**: `ttpos_stock_reconciliation` - 修改 `type` 字段注释
- **种子文件**: `admin/database/seeds/shop_01.sql` - 同步更新表结构定义

---

## 🏗️ 架构设计

### 变更范围

本次变更仅涉及数据定义层，不影响其他层次：

```
❌ API 层 (Controller/API)   - 不涉及
❌ 业务层 (Service)          - 不涉及
❌ 数据层 (Repository)       - 不涉及
✅ 模型层 (Model)            - 更新注释
✅ 数据库层 (Schema)         - 更新字段注释
```

---

## 🗄️ 数据库设计

### 数据表变更

#### 表: ttpos_stock_reconciliation

**变更内容**: 仅修改 `type` 字段的注释

**变更前**:

```sql
`type` int NOT NULL DEFAULT 1 COMMENT '盘点类型 1-指定物品盘点 2-全部物品盘点'
```

**变更后**:

```sql
`type` int NOT NULL DEFAULT 1 COMMENT '盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘'
```

**说明**:

- 字段类型: `int` (保持不变)
- 默认值: `1` (保持不变)
- 允许值: `1`, `2`, `3`, `4`, `5`
- 现有数据: 不受影响（类型 1、2 保持原有含义）

### 数据库迁移

**迁移文件命名**:

```bash
{YYYYMMDDHHMMSS}_add_type_options_to_stock_reconciliation_table.php
```

**迁移逻辑**:

```php
<?php
declare(strict_types=1);

use think\migration\Migrator;

final class AddTypeOptionsToStockReconciliationTable extends Migrator
{
    public function change(): void
    {
        $table = $this->table('stock_reconciliation');
        
        // 仅修改字段注释
        if ($table->hasColumn('type')) {
            $table->changeColumn('type', 'integer', [
                'null' => false,
                'default' => 1,
                'comment' => '盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘',
                'signed' => true,
            ])->update();
        }
    }
}
```

**同步 Seeds 文件**: 必须同步更新 `admin/database/seeds/shop_01.sql` 中 `ttpos_stock_reconciliation` 表的 `type` 字段定义。

---

## 📊 数据模型

### Go Model 变更

**文件**: `main/app/model/stock_reconciliation.go`

**变更前**:

```go
Type int `gorm:"column:type;not null;default:1;comment:盘点类型 1-指定物品盘点 2-全部物品盘点" json:"type"`
```

**变更后**:

```go
Type int `gorm:"column:type;not null;default:1;comment:盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘" json:"type"`
```

### DTO 参数验证变更

**文件**: `main/app/dto/req/stock_reconciliation.go`

**变更前**:

```go
Type int `json:"type"` // 盘点类型 1-指定物品盘点 2-全部物品盘点
```

**变更后**:

```go
Type int `json:"type" binding:"oneof=1 2 3 4 5"` // 盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘
```

**说明**:

- 使用 `binding:"oneof=1 2 3 4 5"` 验证标签，确保只能传入合法的类型值
- Gin 框架会自动验证参数，传入非法值时返回 400 错误
- 与数据库字段注释保持一致

---

## 🔍 类型枚举定义

虽然本次仅修改注释，但为便于后续开发，建议在代码中定义常量：

```go
// 盘点类型常量（建议在后续开发中使用）
const (
    StockReconciliationTypeSpecific = 1 // 指定物品盘点
    StockReconciliationTypeAll      = 2 // 全部物品盘点
    StockReconciliationTypeDaily    = 3 // 日盘
    StockReconciliationTypeWeekly   = 4 // 周盘
    StockReconciliationTypeMonthly  = 5 // 月盘
)
```

---

## 🚨 错误处理

### 场景 1: 迁移文件已执行

- **处理方式**: 使用 `hasColumn()` 检查字段存在性，确保幂等
- **用户影响**: 无影响，迁移可重复执行
- **代码示例**:
  ```php
  if ($table->hasColumn('type')) {
      // 仅在字段存在时修改
  }
  ```

### 场景 2: 种子文件未同步

- **处理方式**: 代码审查时重点检查
- **用户影响**: 新商户初始化时表结构不一致
- **缓解措施**: 严格遵循数据库规范，强制同步更新

### 场景 3: 传入非法类型值

- **处理方式**: Gin 框架自动参数验证，返回 400 错误
- **用户影响**: 用户收到明确的参数错误提示
- **代码示例**:
  ```go
  // DTO 定义
  Type int `json:"type" binding:"oneof=1 2 3 4 5"`
  
  // 错误响应示例
  {
    "code": 0,
    "message": "Key: 'StockReconciliationSaveReq.Type' Error:Field validation for 'Type' failed on the 'oneof' tag",
    "data": {}
  }
  ```

---

## 🔒 安全设计

### 数据安全

- 本次变更仅修改注释，不涉及数据变更，无数据安全风险
- 迁移前建议备份数据库

---

## 🧪 测试策略

### 迁移测试

**测试内容**:

- [x] 迁移文件执行成功
- [x] 字段注释已更新
- [x] 现有数据不受影响
- [x] 可重复执行（幂等性）

**测试步骤**:

```bash
# 1. 执行迁移
cd admin
php think migrate:run

# 2. 验证字段注释
mysql -e "SHOW FULL COLUMNS FROM ttpos_stock_reconciliation WHERE Field='type';"

# 3. 验证现有数据
mysql -e "SELECT uuid, type FROM ttpos_stock_reconciliation LIMIT 10;"

# 4. 重复执行迁移（验证幂等性）
php think migrate:run
```

### 功能测试

**测试内容**:

- [x] 查询现有盘点单，类型显示正确
- [x] 创建新类型盘点单（类型 3、4、5），保存成功
- [x] 查询新类型盘点单，类型显示正确
- [x] 传入非法类型值（如 0、6、-1），返回参数验证错误

**测试步骤**:

```bash
# 1. 创建日盘盘点单（type=3）
curl -X POST http://localhost:8080/api/v1/stock_reconciliation/save \
  -H "Content-Type: application/json" \
  -d '{"warehouse_uuid": 123, "type": 3, "purpose": 1, "items": []}'

# 2. 创建周盘盘点单（type=4）
curl -X POST http://localhost:8080/api/v1/stock_reconciliation/save \
  -H "Content-Type: application/json" \
  -d '{"warehouse_uuid": 123, "type": 4, "purpose": 1, "items": []}'

# 3. 创建月盘盘点单（type=5）
curl -X POST http://localhost:8080/api/v1/stock_reconciliation/save \
  -H "Content-Type: application/json" \
  -d '{"warehouse_uuid": 123, "type": 5, "purpose": 1, "items": []}'

# 4. 测试非法类型值（应返回 400 错误）
curl -X POST http://localhost:8080/api/v1/stock_reconciliation/save \
  -H "Content-Type: application/json" \
  -d '{"warehouse_uuid": 123, "type": 6, "purpose": 1, "items": []}'
```

---

## 📈 性能优化

本次变更仅修改注释，不涉及性能优化。

---

## 📚 实施清单

### Phase 1: 数据库迁移

- [x] 创建迁移文件
- [ ] 执行迁移（修改字段注释）
- [x] 同步更新 `admin/database/seeds/shop_01.sql`
- [ ] 验证迁移成功

### Phase 2: 模型和参数同步

- [x] 更新 Go Model 注释
- [x] 更新 DTO 参数验证
- [x] 验证代码格式

### Phase 3: 测试验证

- [ ] 迁移测试
- [ ] 兼容性测试
- [ ] 新类型功能测试
- [ ] 参数验证测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-29  
**作者**: 曾振华  
**审核者**: 曾振华

