# ERPNext 对接 - 物品管理增加默认销售单位 设计文档

> 本文档定义 ERPNext 对接 - 物品管理增加默认销售单位的技术设计和实现方案。

## 📋 概述

本功能在现有物品管理模块基础上，增加默认销售单位字段的同步、显示和编辑功能。主要涉及：
1. 数据库层面：在 `ttpos_material` 表中增加 `default_sales_unit` 字段
2. ERPNext 同步：同步 ERPNext 的 `Default Sales Unit of Measure` 字段
3. API 层面：物品创建/更新接口支持默认销售单位
4. 前端层面：物品详情页显示，创建/编辑表单支持设置

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口（如 MaterialService 依赖其他 Service）
- Repository 只持有 db 实例，不持有 DBManager
- URL 使用 snake_case（如 `/api/v1/shop/material_info`）
- data 字段必须是对象，不能是 null 或数组
- 不使用 panic，返回 error
- 接口以 `I` 开头，实现以 `Impl` 结尾

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式：`{code, message, data{}}`
- data 不能为 null 或数组
- 分页信息统一放在 meta 中

### 数据库规范 (database.mdc)

- 必需字段：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
- 时间字段使用 int 类型，`_time` 结尾，默认值 0
- UUID 字段使用 bigint unsigned
- 表名使用 `ttpos_` 前缀
- 字段名使用 snake_case

### Vue 前端规范 (vue.mdc)

- 使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 使用 Composition API
- 遵循命名规范

---

## 🔄 代码复用分析

### 可复用的现有组件

- **MaterialService**: `main/app/service/material.go` - 物品业务逻辑，已有同步、创建、更新方法
- **MaterialRepository**: `main/app/repository/material_repo.go` - 物品数据访问，已有 CRUD 方法
- **ERPNext Service**: `main/app/service/rpc/erp/material.go` - ERPNext 同步逻辑，已有 `GetMaterialList` 和 `AddMaterial` 方法
- **Material Model**: `main/app/model/material.go` - 物品数据模型，已有完整的字段定义
- **MaterialUnit Model**: `main/app/model/material.go` - 物品单位模型，已有单位相关字段

### 集成点

- **ERPNext 同步接口**: 在 `SyncMaterial` 方法中增加 `default_sales_unit` 字段的同步逻辑
- **物品创建/更新接口**: 在 `CreateMaterial` 和 `UpdateMaterial` 方法中支持 `default_sales_unit` 字段
- **物品详情接口**: 在响应中返回 `default_sales_unit` 字段
- **前端物品管理页面**: 在物品详情页和创建/编辑表单中增加默认销售单位字段

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (MaterialAPI)
  ↓ 依赖
业务层 (MaterialService)
  ↓ 依赖
数据层 (MaterialRepository)
```

**依赖规则**:
- ✅ API 层依赖 Service 层
- ✅ Service 层依赖 Repository 层
- ✅ Service 层可以依赖其他 Service 接口
- ❌ Service 不能直接依赖 Repository（通过 DBManager 获取）

### 模块划分

#### Go Main 模块

- **API 层**: `main/app/api/v1/shop/shop_material.go` - 物品 API 接口
- **Service 层**: `main/app/service/material.go` - 物品业务逻辑
- **Repository 层**: `main/app/repository/material_repo.go` - 物品数据访问
- **Model 层**: `main/app/model/material.go` - 物品数据模型
- **DTO 层**: `main/app/dto/req/material.go` - 请求参数
- **DTO 层**: `main/app/dto/resp/material_resp/material.go` - 响应数据

#### Vue 前端模块

- **Pages**: `admin/views/shop/pages/material/` - 物品管理页面
- **API**: `admin/views/shop/api/material.ts` - API 封装
- **Components**: `admin/views/shop/components/` - 可复用组件

---

## 🗄️ 数据库设计

### 数据表设计

#### 表: ttpos_material（物品表）

**新增字段**:

```sql
ALTER TABLE `ttpos_material` 
ADD COLUMN `default_sales_unit` bigint unsigned NOT NULL DEFAULT 0 COMMENT '默认销售单位ID（ERPNext），关联 ttpos_material_unit.uuid' 
AFTER `cost_unit_uuid`;
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| default_sales_unit | bigint unsigned | 默认销售单位ID（ERPNext） | DEFAULT 0，关联 MaterialUnit.uuid |

**索引设计**:
- 普通索引: `KEY idx_default_sales_unit (default_sales_unit)`（可选，用于查询优化）

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_default_sales_unit_to_material_table.php`

### 数据库迁移

**迁移脚本**:

```php
<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddDefaultSalesUnitToMaterialTable extends Migrator
{
    public function change()
    {
        $table = $this->table('ttpos_material');
        $table->addColumn('default_sales_unit', 'biginteger', [
            'limit' => 20,
            'unsigned' => true,
            'null' => false,
            'default' => 0,
            'comment' => '默认销售单位ID（ERPNext），关联 ttpos_material_unit.uuid',
            'after' => 'cost_unit_uuid'
        ])->addIndex(['default_sales_unit'], ['name' => 'idx_default_sales_unit'])
          ->update();
    }
}
```

**同步 Go Model**:

在 `main/app/model/material.go` 中增加字段定义。

---

## 📊 数据模型

### Go Model

```go
// main/app/model/material.go
type Material struct {
    BaseModel
    // ... 现有字段 ...
    CostUnitUuid          uint64   `gorm:"default:0;column:cost_unit_uuid;comment:'成本单位ID'"`
    DefaultSalesUnitUuid  uint64   `gorm:"default:0;column:default_sales_unit;comment:'默认销售单位ID（ERPNext）'"`
    // ... 其他字段 ...
    
    // 关联关系
    DefaultSalesUnit      *MaterialUnit `gorm:"foreignKey:default_sales_unit;references:uuid"` // 默认销售单位
}
```

### DTO 定义

#### Request DTO

```go
// main/app/dto/req/material.go
type MaterialCreateReq struct {
    // ... 现有字段 ...
    DefaultSalesUnitUuid *uint64 `json:"default_sales_unit_uuid"` // 默认销售单位UUID（可选）
}

type MaterialUpdateReq struct {
    // ... 现有字段 ...
    DefaultSalesUnitUuid *uint64 `json:"default_sales_unit_uuid"` // 默认销售单位UUID（可选）
}

type MaterialEditErpReq struct {
    // ... 现有字段 ...
    DefaultSalesUnit string `json:"default_sales_unit"` // 默认销售单位（ERPNext UOM）
}
```

#### Response DTO

```go
// main/app/dto/resp/material_resp/material.go
type MaterialResp struct {
    // ... 现有字段 ...
    DefaultSalesUnitUuid uint64 `json:"default_sales_unit_uuid"` // 默认销售单位UUID
    DefaultSalesUnit     *MaterialUnitResp `json:"default_sales_unit,omitempty"` // 默认销售单位信息
}
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 创建物品（支持默认销售单位）

**请求**:
- **URL**: `/api/v1/shop/material_info/create`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Body**:
  ```json
  {
    "name": "物品名称",
    "default_sales_unit_uuid": 123456
  }
  ```

**响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 123456,
    "default_sales_unit_uuid": 123456,
    "default_sales_unit": {
      "uuid": 123456,
      "name": "箱"
    }
  }
}
```

#### API 2: 更新物品（支持默认销售单位）

**请求**:
- **URL**: `/api/v1/shop/material_info/update`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "uuid": 123456,
    "default_sales_unit_uuid": 789012
  }
  ```

**响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

#### API 3: 获取物品详情（返回默认销售单位）

**请求**:
- **URL**: `/api/v1/shop/material_info/get`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "uuid": 123456
  }
  ```

**响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 123456,
    "name": "物品名称",
    "default_sales_unit_uuid": 123456,
    "default_sales_unit": {
      "uuid": 123456,
      "name": "箱",
      "unit_uuid": 789012,
      "erpnext_uom": "Box"
    },
    "is_headquarter": true
  }
}
```

---

## 🧩 组件和接口

### Service 层

#### Service 接口（无需修改）

```go
// main/app/service/i_material_service.go
// 接口定义保持不变，实现类增加字段处理逻辑
```

#### Service 实现修改点

```go
// main/app/service/material.go

// 1. SyncMaterial 方法 - 增加默认销售单位同步
func (s *materialSrv) SyncMaterial(ctx context.Context, syncHeadquarterData bool) error {
    // ... 现有代码 ...
    for _, itemInfo := range materialList.ItemList {
        // 获取默认销售单位UUID
        var defaultSalesUnitUuid uint64 = 0
        if itemInfo.DefaultSalesUnit != "" {
            // 通过 UOM 查找对应的 MaterialUnit
            unitUuid, err := s.getUnitUuidByUom(ctx, itemInfo.DefaultSalesUnit)
            if err == nil {
                defaultSalesUnitUuid = unitUuid
            }
        }
        
        // 在 UpdateMaterialByEprItem 中传入 defaultSalesUnitUuid
        // ...
    }
}

// 2. CreateMaterial 方法 - 支持默认销售单位
func (s *materialSrv) CreateMaterial(ctx *gin.Context, req *dto_req.MaterialCreateReq) (*dto_resp.MaterialResp, error) {
    // ... 现有代码 ...
    material := &model.Material{
        // ... 现有字段 ...
        DefaultSalesUnitUuid: 0,
    }
    
    // 如果传入了默认销售单位，验证并设置
    if req.DefaultSalesUnitUuid != nil && *req.DefaultSalesUnitUuid > 0 {
        // 验证单位是否属于该物品（在物品创建后设置单位时验证）
        material.DefaultSalesUnitUuid = *req.DefaultSalesUnitUuid
    }
    
    // ... 保存逻辑 ...
}

// 3. UpdateMaterial 方法 - 支持默认销售单位
func (s *materialSrv) UpdateMaterial(ctx *gin.Context, req *dto_req.MaterialUpdateReq) error {
    // ... 现有代码 ...
    
    // 如果传入了默认销售单位，验证并更新
    if req.DefaultSalesUnitUuid != nil {
        // 验证权限：总部来源的物品不允许子店修改
        material := materialRepo.GetMaterial(commonRepo.WhereByUuid(req.Uuid))
        if material.IsHeadquarter() {
            // 检查当前用户是否有权限修改总部物品
            // 如果没有权限，返回错误
        }
        
        // 验证单位是否属于该物品
        if *req.DefaultSalesUnitUuid > 0 {
            if !material.IsUnit(*req.DefaultSalesUnitUuid) {
                return errors.New("默认销售单位必须是该物品已配置的单位")
            }
        }
        
        updateData["default_sales_unit"] = *req.DefaultSalesUnitUuid
    }
    
    // ... 更新逻辑 ...
}
```

### Repository 层

#### Repository 接口（无需修改）

```go
// main/app/repository/i_material_repo.go
// 接口定义保持不变，现有方法已支持更新任意字段
```

#### Repository 实现（无需修改）

```go
// main/app/repository/material_repo.go
// 现有 UpdateMaterialData 方法已支持更新任意字段，无需修改
```

### API 层

```go
// main/app/api/v1/shop/shop_material.go
// API 层无需修改，DTO 已包含新字段，Service 层会自动处理
```

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:
- **Key 命名**: `ttpos:material:{uuid}`
- **过期时间**: 5 分钟
- **更新策略**: Cache-Aside Pattern

**缓存内容**:
- 物品基本信息（包含 default_sales_unit_uuid）
- 物品单位信息（包含默认销售单位）

**缓存更新时机**:
- 物品创建/更新时清除缓存
- ERPNext 同步时清除缓存

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 默认销售单位不属于该物品

- **处理方式**: 返回错误 "默认销售单位必须是该物品已配置的单位"
- **用户影响**: 提示用户选择正确的单位
- **代码示例**:
  ```go
  if !material.IsUnit(defaultSalesUnitUuid) {
      return errors.New("默认销售单位必须是该物品已配置的单位")
  }
  ```

#### 场景 2: 总部来源的物品不允许子店修改

- **处理方式**: 返回错误 "总部来源的物品不允许修改默认销售单位"
- **用户影响**: 前端显示字段为只读，后端拒绝修改请求
- **代码示例**:
  ```go
  if material.IsHeadquarter() && !hasPermission {
      return errors.New("总部来源的物品不允许修改默认销售单位")
  }
  ```

#### 场景 3: ERPNext 同步时单位不存在

- **处理方式**: 记录日志，设置 default_sales_unit 为 0
- **用户影响**: 物品详情页显示"无"
- **代码示例**:
  ```go
  unitUuid, err := s.getUnitUuidByUom(ctx, itemInfo.DefaultSalesUnit)
  if err != nil {
      logger.Logger.Warn("ERPNext同步-默认销售单位不存在", 
          zap.String("uom", itemInfo.DefaultSalesUnit),
          zap.Error(err))
      defaultSalesUnitUuid = 0
  }
  ```

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证（已有）

### 权限控制

- **总部物品只读**: 总部来源的物品默认销售单位为只读，子店无法修改
- **权限验证**: 后端 API 验证权限，拒绝未授权修改

### 数据安全

- **参数验证**: 验证默认销售单位必须是该物品已配置的单位
- **SQL 注入防护**: 使用 GORM 参数化查询（已有）

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:
- Service 层: 70%+
- Repository 层: 80%+

**测试内容**:
- 创建物品时设置默认销售单位
- 更新物品时修改默认销售单位
- 总部物品不允许子店修改
- 验证默认销售单位必须是该物品已配置的单位
- ERPNext 同步时处理默认销售单位

### API 测试

**测试内容**:
- 创建物品 API（包含默认销售单位）
- 更新物品 API（修改默认销售单位）
- 获取物品详情 API（返回默认销售单位）
- 权限控制测试（总部物品只读）

### 集成测试

**测试流程**:
- ERPNext 同步 → 查看详情 → 编辑 → 保存
- 创建物品 → 设置默认销售单位 → 查看详情

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 添加索引 `idx_default_sales_unit`（可选）
   - 查询时使用预加载（Preload）加载关联的单位信息

2. **缓存优化**:
   - 物品详情缓存包含默认销售单位信息
   - 缓存更新时清除相关缓存

3. **查询优化**:
   - 使用 GORM 的 Preload 预加载关联数据
   - 避免 N+1 查询问题

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms
- 缓存命中率: > 80%

---

## 🌐 浏览器兼容性

### 前端兼容性（Vue）

- Chrome 90+
- Safari 14+
- Firefox 88+
- Edge 90+

---

## 📚 实现清单

### Phase 1: 数据库和模型
- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 更新 Go Model
- [ ] 更新 DTO 定义

### Phase 2: 后端实现
- [ ] 修改 ERPNext 同步逻辑
- [ ] 修改物品创建逻辑
- [ ] 修改物品更新逻辑
- [ ] 修改物品详情响应

### Phase 3: 前端实现
- [ ] 物品详情页显示默认销售单位
- [ ] 创建物品表单添加默认销售单位字段
- [ ] 编辑物品表单添加默认销售单位字段
- [ ] 权限控制（总部物品只读）

### Phase 4: 测试
- [ ] 单元测试
- [ ] API 测试
- [ ] 集成测试
- [ ] 手动测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2026-01/2026-01-19.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-19  
**作者**: xiezhihuan  
**审核者**: {审核者}
