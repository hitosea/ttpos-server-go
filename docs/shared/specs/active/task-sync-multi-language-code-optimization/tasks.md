# 任务分解 - 多语言同步功能代码优化

---

## 📋 任务分解原则

- **颗粒度**: 每个任务 10-30 分钟（SP ≤ 0.5）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **AI 友好**: 提供清晰的执行步骤

---

## 📊 进度总览

**总任务数**: 5（本次优化任务）  
**已完成**: 5  
**完成率**: 100% ✅

**关联任务 #36915 的重构工作**：
- ✅ 已完成10个同步方法的重构（移除多语言处理代码）
- ✅ 已创建 `SyncMultiLanguage` 统一同步方法
- ✅ 已添加 `SyncTaskTypeMultiLanguage` 常量

---

## Phase 1: 代码优化

### Task 1.1 添加配置包 Import

- [x] **添加配置包导入**
  - **File**: `main/app/service/sync.go`
  - **Purpose**: 引入 `ttpos-server-go/config` 包以使用表前缀配置
  - **Requirements**: R1
  - **执行步骤**:
    1. 在 import 块中添加 `"ttpos-server-go/config"`
    2. 按项目规范排列 import 顺序
  - **验收**: 编译通过，import 格式正确

---

### Task 1.2 替换硬编码表名

- [x] **使用配置化表前缀**
  - **File**: `main/app/service/sync.go`
  - **Method**: `SyncMultiLanguage`
  - **Purpose**: 将所有硬编码的表名改为使用 `config.Database.TablePrefix`
  - **Requirements**: R1
  - **执行步骤**:
    1. 找到 `tableConfigs` 定义（约第 502 行）
    2. 将每个 `"ttpos_xxx"` 改为 `config.Database.TablePrefix + "xxx"`
    3. 共需修改 12 个表名
  - **涉及的表**:
    ```
    - material
    - material_category
    - product_attribute
    - product_attribute_group
    - product_bom_card
    - product_category
    - product_flavor
    - product_package
    - product_package_group
    - product_sauce
    - product_unit
    - warehouse
    ```
  - **验收**: 所有表名使用配置化前缀，编译通过

---

### Task 1.3 更新类型定义

- [x] **使用 any 替代 interface{}**
  - **File**: `main/app/service/sync.go`
  - **Method**: `SyncMultiLanguage`
  - **Purpose**: 使用 Go 1.18+ 的 `any` 类型
  - **Requirements**: R2
  - **执行步骤**:
    1. 找到 `var records []map[string]interface{}` 定义（约第 522 行）
    2. 将 `interface{}` 改为 `any`
  - **代码变更**:
    ```go
    // 改进前
    var records []map[string]interface{}
    
    // 改进后
    var records []map[string]any
    ```
  - **验收**: 编译通过，类型定义使用 `any`

---

### Task 1.4 添加 Debug 支持

- [x] **添加 GORM Debug 调用**
  - **File**: `main/app/service/sync.go`
  - **Method**: `SyncMultiLanguage`
  - **Purpose**: 在查询语句中添加 `.Debug()` 以便调试
  - **Requirements**: R3
  - **执行步骤**:
    1. 找到查询多语言UUID的代码块（约第 523-528 行）
    2. 在 `.Find(&records).Error` 前添加 `.Debug()`
  - **代码变更**:
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
  - **验收**: 编译通过，运行时可以看到 SQL Debug 输出

---

## Phase 2: 验证测试

### Task 2.1 代码质量检查

- [x] **执行 linter 检查**
  - **Purpose**: 确保代码符合项目规范
  - **执行步骤**:
    1. 运行 golangci-lint 检查
    2. 修复任何 linter 错误
  - **命令**:
    ```bash
    golangci-lint run main/app/service/sync.go
    ```
  - **验收**: 无 linter 错误或警告

---

## 📝 完整代码变更示例

### 变更前

```go
func (s *SyncSrv) SyncMultiLanguage(ctx context.Context) error {
    // ... 前面的代码
    
    // 所有需要同步多语言的表配置（按表名字母顺序排列）
    tableConfigs := []tableConfig{
        {tableName: "ttpos_material", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_material_category", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_attribute", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_attribute_group", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_bom_card", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_category", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_flavor", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_package", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_package_group", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_sauce", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_product_unit", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: "ttpos_warehouse", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
    }
    
    // 从总部表中查询所有总部数据的多语言UUID
    for _, config := range tableConfigs {
        var records []map[string]interface{}
        err := headquarterDB.Table(config.tableName).
            Select(config.multiLanguageUuidColumn).
            Where("delete_time = 0").
            Where("headquarter_uuid = 0").
            Where(config.multiLanguageUuidColumn + " > 0").
            Find(&records).Error
        // ... 处理代码
    }
}
```

### 变更后

```go
import (
    // ... 其他 import
    "ttpos-server-go/config"  // 新增
)

func (s *SyncSrv) SyncMultiLanguage(ctx context.Context) error {
    // ... 前面的代码
    
    // 所有需要同步多语言的表配置（按表名字母顺序排列）
    tableConfigs := []tableConfig{
        {tableName: config.Database.TablePrefix + "material", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "material_category", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_attribute", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_attribute_group", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_bom_card", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_category", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_flavor", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_package", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_package_group", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_sauce", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "product_unit", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
        {tableName: config.Database.TablePrefix + "warehouse", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
    }
    
    // 从总部表中查询所有总部数据的多语言UUID
    for _, config := range tableConfigs {
        var records []map[string]any
        err := headquarterDB.Table(config.tableName).
            Select(config.multiLanguageUuidColumn).
            Where("delete_time = 0").
            Where("headquarter_uuid = 0").
            Where(config.multiLanguageUuidColumn + " > 0").Debug().
            Find(&records).Error
        // ... 处理代码
    }
}
```

---

## ✅ 最终验收

### 代码质量
- [x] 所有表名使用配置化表前缀
- [x] 类型定义使用 `any`
- [x] 添加 `.Debug()` 支持
- [x] 通过 golangci-lint 检查
- [x] Import 格式正确

### 功能验证
- [x] 编译通过
- [x] 多语言同步功能正常
- [x] Debug 输出正常

---

## 📊 变更统计

### 本次优化（task-sync-multi-language-code-optimization）

| 项目 | 数量 |
|------|------|
| 文件变更 | 1 个 |
| 代码行变更 | ~15 行 |
| 新增 Import | 1 个 |
| 表名修改 | 12 处 |
| 类型修改 | 1 处 |
| Debug 添加 | 1 处 |

### 关联重构（任务 #36915）

| 项目 | 数量 |
|------|------|
| 文件变更 | 4 个 |
| 同步方法重构 | 10 个 |
| 新增同步任务类型 | 1 个 |
| 新增统一同步方法 | 1 个 |
| 代码行变更 | ~500+ 行 |

---

## 📋 完整文件变更清单

### 任务 #36915 主要重构（已完成）

#### 1. `main/app/constant/sync_task.go`
```go
+ const SyncTaskTypeMultiLanguage = "multi_language"
+ SyncTaskTypeMultiLanguage: "多语言",
```

#### 2. `main/app/service/sync.go`
```go
+ import "gorm.io/gorm"
+ 
+ // SyncMultiLanguage 同步多语言数据
+ func (s *SyncSrv) SyncMultiLanguage(ctx context.Context) error {
+     // 统一处理12个表的多语言同步
+ }
+ 
+ allTasks: []string{
+     SyncTaskTypeMultiLanguage, // 最优先执行
+     // ... 其他任务
+ }
```

#### 3. `main/app/service/material.go`（3个方法）
```go
// SyncMaterialCategory
- // 创建/更新多语言代码（~30行）
+ // 同步总部物品分类到子店（多语言由 SyncMultiLanguage 任务处理）

// SyncMaterial  
- // 多语言创建/删除代码（~60行）
+ // 同步总部物品到子店（多语言由 SyncMultiLanguage 任务处理）

// SyncProductBomCard
- // 多语言创建代码（~15行）
+ // 同步总部成本卡到子店（多语言由 SyncMultiLanguage 任务处理）
```

#### 4. `main/app/service/product.go`（6个方法）
```go
// SyncProductShopCategory
- // 多语言创建/更新代码（~25行）
+ // 只引用 MultiLanguageNameUuid

// SyncUnit
- // 总部单位的多语言创建（~20行）
+ // 保留ERP同步的多语言，移除总部同步的多语言

// SyncSauce
- // 多语言创建/删除代码（~40行）
+ // 只引用 MultiLanguageNameUuid

// SyncAttributeGroup
- // 属性组和属性的多语言处理（~50行）
+ // 只引用 MultiLanguageNameUuid

// SyncProductFlavor
- // 多语言创建/删除代码（~35行）
+ // 只引用 MultiLanguageNameUuid

// SyncProduct
- // 总部商品的多语言创建（~25行）
+ // 保留ERP同步的多语言，移除总部同步的多语言
```

#### 5. `main/app/service/warehouse.go`（1个方法）
```go
// SyncWarehouse
- // 总部仓库的多语言创建（~15行）
+ // 同步ttpos总店数据（多语言由 SyncMultiLanguage 任务处理）
+ // 保留ERP同步的多语言，移除总部同步的多语言
```

### 本次优化（已完成）

#### `main/app/service/sync.go` - SyncMultiLanguage 方法
```diff
+ import "ttpos-server-go/config"

  tableConfigs := []tableConfig{
-     {tableName: "ttpos_material", ...},
+     {tableName: config.Database.TablePrefix + "material", ...},
      // ... 其他11个表
  }
  
- var records []map[string]interface{}
+ var records []map[string]any

  err := headquarterDB.Table(config.tableName).
      // ... 查询条件
-     Find(&records).Error
+     .Debug().Find(&records).Error
```

---

## 🎯 设计原则总结

1. **多语言数据统一管理**
   - 所有从总部同步的多语言数据由 `SyncMultiLanguage` 统一处理
   - 在任务列表中最优先执行，确保其他任务执行时多语言已存在

2. **区分数据来源**
   - **总部数据**：只同步 `MultiLanguageNameUuid` 引用，不创建多语言
   - **ERP数据**（子店自己的）：仍需创建和管理多语言

3. **代码质量**
   - 使用配置化表前缀，避免硬编码
   - 使用现代 Go 特性（`any` 类型）
   - 添加 Debug 支持，方便问题排查

4. **注释规范**
   - 在同步方法中明确注释"多语言由 SyncMultiLanguage 任务处理"
   - 保留必要的业务逻辑注释

---

**创建时间**: 2025-11-25  
**完成时间**: 2025-11-25  
**维护者**: 曾振华

