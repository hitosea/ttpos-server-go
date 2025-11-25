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

---

## 🎯 需求目标

### 主要目标

优化 `SyncMultiLanguage` 方法的代码质量，提升可维护性和调试友好性。

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

#### R3. 添加调试支持

**需求描述**：  
在查询语句中添加 GORM 的 `.Debug()` 调用，方便开发调试。

**实现要求**：
- 在查询多语言UUID的 SQL 语句中添加 `.Debug()`
- 开发环境可以看到实际执行的 SQL
- 不影响生产环境性能

**示例**：
```go
// 改进前
err := headquarterDB.Table(config.tableName).
    Select(config.multiLanguageUuidColumn).
    Where("delete_time = 0").
    Where("headquarter_uuid = 0").
    Where(config.multiLanguageUuidColumn + " > 0").
    Find(&records).Error

// 改进后
err := headquarterDB.Table(config.tableName).
    Select(config.multiLanguageUuidColumn).
    Where("delete_time = 0").
    Where("headquarter_uuid = 0").
    Where(config.multiLanguageUuidColumn + " > 0").Debug().
    Find(&records).Error
```

**验收标准**：
- ✅ 添加 `.Debug()` 调用
- ✅ 开发环境可以查看 SQL 输出
- ✅ 方便问题排查和调试

---

## 📦 完整重构范围（任务 #36915）

### 涉及的文件和方法

#### 1. 常量定义（`main/app/constant/sync_task.go`）
- ✅ 新增 `SyncTaskTypeMultiLanguage` 常量
- ✅ 添加 "多语言" 任务类型名称

#### 2. 同步服务（`main/app/service/sync.go`）
- ✅ 新增 `SyncMultiLanguage` 方法
- ✅ 在任务列表中添加多语言任务（最优先执行）
- 🔧 本次优化：配置化表前缀、`any` 类型、`.Debug()` 调试

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
3. 查询语句添加 `.Debug()` 支持
4. 通过 golangci-lint 检查
5. 符合项目代码规范

### 功能验证
1. 多语言同步功能正常运行
2. 开发环境可以看到 SQL Debug 输出
3. 与总部数据保持一致
4. 无性能回退

---

## 📝 非功能需求

1. **性能要求**
   - 不影响同步性能
   - Debug 输出仅在开发环境启用

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

