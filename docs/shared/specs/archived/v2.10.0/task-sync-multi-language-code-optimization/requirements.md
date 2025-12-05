> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 需求文档 - 多语言同步功能代码优化

---

## 📋 需求背景

在实现任务 #36915（总店基础数据同步到子店）时，我们进行了大规模的代码重构：

1. **重构多语言同步功能**：创建了统一的 `SyncMultiLanguage` 方法
2. **重构8个同步方法**：从 `SyncMaterialCategory`、`SyncMaterial`、`SyncProductBomCard`、`SyncProductShopCategory`、`SyncUnit`、`SyncSauce`、`SyncAttributeGroup`、`SyncProductFlavor`、`SyncProduct`、`SyncWarehouse` 等方法中移除多语言处理代码
3. **区分数据来源**：明确区分"从总部同步的数据"和"从ERP同步的数据"的多语言处理策略

在手动完成这些重构后，代码审查中发现以下可以进一步优化的技术问题：

### 当前问题

1. **硬编码表名**
   - 表名使用硬编码字符串 `"ttpos_material"`、`"ttpos_warehouse"` 等
   - 不符合项目配置化原则
   - 与其他代码风格不一致（如 `printer_tasks/tasks.go` 中使用 `config.Database.TablePrefix`）

2. **类型定义过时**
   - 使用 `map[string]interface{}` 定义记录类型
   - Go 1.18+ 引入了 `any` 作为 `interface{}` 的别名，更简洁
   - 项目已使用 Go 1.23+，应采用现代特性

3. **特殊表处理缺失**
   - `product_package_group` 表没有 `headquarter_uuid` 字段
   - 使用统一的 `headquarter_uuid = 0` 条件会导致查询失败
   - 需要通过关联 `product_package` 表来筛选总部数据

---

## 🎯 需求目标

### 主要目标

优化 `SyncMultiLanguage` 方法的代码质量，提升可维护性。

### 具体需求

#### R1. 配置化表前缀

**需求描述**：  
将所有硬编码的表名改为使用配置文件中的表前缀。

**实现要求**：
- 引入 `ttpos-server-go/config` 包
- 使用 `config.Database.TablePrefix` 替代硬编码的 `"ttpos_"`
- 所有表配置都使用配置化的表前缀

**示例**：
```go
// 改进前
{tableName: "ttpos_material", ...}

// 改进后
{tableName: config.Database.TablePrefix + "material", ...}
```

**验收标准**：
- ✅ 所有表名使用配置化表前缀
- ✅ 与项目其他代码风格保持一致
- ✅ 通过 golangci-lint 检查

---

#### R2. 使用现代类型定义

**需求描述**：  
使用 Go 1.18+ 的 `any` 类型替代 `interface{}`。

**实现要求**：
- 将 `[]map[string]interface{}` 改为 `[]map[string]any`
- 保持功能不变，仅优化类型定义

**示例**：
```go
// 改进前
var records []map[string]interface{}

// 改进后
var records []map[string]any
```

**验收标准**：
- ✅ 类型定义使用 `any`
- ✅ 编译通过，功能正常
- ✅ 符合 Go 1.18+ 最佳实践

---

#### R3. 支持特殊表的自定义筛选条件

**需求描述**：  
`product_package_group` 表没有 `headquarter_uuid` 字段，需要通过关联表筛选。

**实现要求**：
- 在 `tableConfig` 结构体中添加 `filterCondition` 字段
- 对于没有 `headquarter_uuid` 的表，使用子查询筛选
- 保持代码的统一性和可扩展性

**示例**：
```go
// 结构体定义
type tableConfig struct {
    tableName               string
    multiLanguageUuidColumn string
    entityUuidColumn        string
    preloadRelations        []string
    filterCondition         string   // 自定义筛选条件（可选）
}

// product_package_group 使用子查询
{tableName: config.Database.TablePrefix + "product_package_group", ..., 
 filterCondition: "product_package_uuid IN (SELECT uuid FROM " + 
     config.Database.TablePrefix + "product_package WHERE headquarter_uuid = 0)"}

// 查询逻辑
if cfg.filterCondition != "" {
    query = query.Where(cfg.filterCondition)
} else {
    query = query.Where("headquarter_uuid = 0")
}
```

**验收标准**：
- ✅ `tableConfig` 结构体支持 `filterCondition` 字段
- ✅ `product_package_group` 表使用子查询筛选
- ✅ 其他表仍使用默认的 `headquarter_uuid = 0` 条件

---

## 📦 完整重构范围（任务 #36915）

### 涉及的文件和方法

#### 1. 常量定义（`main/app/constant/sync_task.go`）
- ✅ 新增 `SyncTaskTypeMultiLanguage` 常量
- ✅ 添加 "多语言" 任务类型名称

#### 2. 同步服务（`main/app/service/sync.go`）
- ✅ 新增 `SyncMultiLanguage` 方法
- ✅ 在任务列表中添加多语言任务（最优先执行）
- 🔧 本次优化：配置化表前缀、`any` 类型、自定义筛选条件

#### 3. 物品同步（`main/app/service/material.go`）

| 方法 | 重构内容 | 原则 |
|------|---------|------|
| `SyncMaterialCategory` | 移除多语言创建/更新，只同步 `MultiLanguageNameUuid` | 总部数据：只引用 |
| `SyncMaterial` | 移除从总部同步时的多语言处理 | 总部数据：只引用<br>ERP数据：仍创建 |
| `SyncProductBomCard` | 移除多语言创建，保留存在性检查 | 总部数据：只引用 |

#### 4. 商品同步（`main/app/service/product.go`）

| 方法 | 重构内容 | 原则 |
|------|---------|------|
| `SyncProductShopCategory` | 移除多语言创建/更新 | 总部数据：只引用 |
| `SyncUnit` | 移除从总部同步时的多语言创建 | 总部数据：只引用<br>ERP数据：仍创建 |
| `SyncSauce` | 移除多语言创建/删除 | 总部数据：只引用 |
| `SyncAttributeGroup` | 移除属性组和属性的多语言处理 | 总部数据：只引用 |
| `SyncProductFlavor` | 移除多语言创建/删除 | 总部数据：只引用 |
| `SyncProduct` | 移除从总部同步时的多语言创建 | 总部数据：只引用<br>ERP数据：仍创建 |

#### 5. 仓库同步（`main/app/service/warehouse.go`）

| 方法 | 重构内容 | 原则 |
|------|---------|------|
| `SyncWarehouse` | 移除从总部同步时的多语言创建 | 总部数据：只引用<br>ERP数据：仍创建 |

### 重构设计原则

```yaml
数据来源: 总部
  多语言处理: ❌ 不创建/不更新
  引用方式: ✅ 只引用 multi_language_name_uuid
  由谁同步: ✅ SyncMultiLanguage 统一处理

数据来源: ERP（子店自己的数据）
  多语言处理: ✅ 仍需创建
  引用方式: ✅ 创建并引用 uuid
  由谁同步: ✅ 各自同步方法处理
```

---

## 📊 影响范围

### 代码变更
- **文件**: `main/app/service/sync.go`
- **方法**: `SyncMultiLanguage`
- **变更类型**: 代码优化，无功能变更
- **关联重构**: 10个同步方法已在 #36915 中重构

### 功能影响
- ✅ 无功能变更
- ✅ 向后兼容
- ✅ 不影响现有数据

### 依赖变更
- 新增 import: `ttpos-server-go/config`

---

## ✅ 验收标准

### 代码质量
1. 所有表名使用 `config.Database.TablePrefix`
2. 类型定义使用 `any` 替代 `interface{}`
3. 特殊表（`product_package_group`）使用自定义筛选条件
4. 通过 golangci-lint 检查
5. 符合项目代码规范

### 功能验证
1. 多语言同步功能正常运行
2. `product_package_group` 表查询正常
3. 与总部数据保持一致
4. 无性能回退

---

## 📝 非功能需求

1. **性能要求**
   - 不影响同步性能

2. **可维护性要求**
   - 代码符合项目规范
   - 注释清晰准确
   - 与其他代码风格一致

3. **兼容性要求**
   - 向后兼容
   - 不影响现有功能
   - Go 1.23+ 运行环境

---

## 🔗 相关文档

- [提案文档](../../../../team/proposals/2025-11/sync-multi-language-code-optimization.md)
- [设计文档](./design.md)
- [任务分解](./tasks.md)

---

**创建时间**: 2025-11-25  
**维护者**: 曾振华

