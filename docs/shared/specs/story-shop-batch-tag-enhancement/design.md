# Shop 端分批类型管理功能增强 设计文档

> 本文档定义 Shop 端分批类型管理功能增强的技术设计和实现方案。

## 📋 概述

在 Shop 端的分批类型管理功能中，增加名称缩写字段。主要涉及数据库表结构调整、Model 更新、API 接口调整和业务逻辑增强。

**注意**：多语言名称支持已在 v2.9.0 版本中实现，本次无需修改。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error

### API 设计规范 (api.mdc)

- URL 使用 snake_case（如：`/api/v1/shop/batch/tag/add`）
- 响应格式统一：`{code, message, data{}}`
- data 不能为 null 或数组

### 数据库规范 (database.mdc)

- 必需字段完整（id, uuid, create_time, update_time, delete_time）
- 时间字段使用 int
- UUID 字段使用 bigint unsigned
- 字段名使用 snake_case
- 多语言名称通过 `multi_language_name_uuid` 关联 `ttpos_multi_language_name` 表

---

## 🔄 代码复用分析

### 可复用的现有组件

- **BatchTagRepo**: `main/app/repository/batch_tag.go` - 分批类型的 CRUD 操作（已支持多语言）
- **ProductService**: `main/app/service/product.go` - 分批类型相关的业务逻辑（AddBatchTag, EditBatchTag, GetBatchTag，已支持多语言）
- **BatchProductHandler**: `main/app/api/v1/shop/shop_batch_product.go` - 分批类型 API 接口（已支持多语言）

### 集成点

- **分批类型表**: `ttpos_batch_tag` - 需要增加 `abbreviation` 字段
- **多语言名称表**: `ttpos_multi_language_name` - 已在 v2.9.0 实现，本次无需修改

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (Controller/API)
  ↓ 依赖
业务层 (Service)
  ↓ 依赖
数据层 (Repository)
```

**依赖规则**:
- ✅ 上层可依赖下层
- ❌ 禁止下层依赖上层
- ❌ 禁止跨层调用
- ❌ Service 不能依赖 Repository
- ✅ Service 可以依赖其他 Service 接口

### 模块划分

#### Go Main 模块

- **API 层**: `main/app/api/v1/shop/shop_batch_product.go` - 路由处理、参数校验
- **Service 层**: `main/app/service/product.go` - 业务逻辑、事务管理
- **Repository 层**: `main/app/repository/batch_tag.go` - 数据访问、数据库操作
- **Model 层**: `main/app/model/product.go` - 数据模型
- **DTO 层**: `main/app/dto/` - 数据传输对象
  - `req/product.go` - 请求参数
  - `resp/product_resp/product.go` - 响应数据

---

## 🗄️ 数据库设计

### 数据表设计

#### 表: ttpos_batch_tag（修改）

**表结构调整**：

```sql
ALTER TABLE `ttpos_batch_tag` 
ADD COLUMN `abbreviation` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '名称缩写' AFTER `multi_language_name_uuid`;
```

**完整表结构**：

```sql
CREATE TABLE IF NOT EXISTS `ttpos_batch_tag` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '分批类型名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称UUID',
    `abbreviation` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '名称缩写',
    `color` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '颜色,如#FF0000',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '分批类型表';
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | INT(11) UNSIGNED | 主键 ID | AUTO_INCREMENT |
| uuid | BIGINT UNSIGNED | 唯一标识 | DEFAULT 0, UNIQUE |
| name | VARCHAR(255) | 分批类型名称（JSON格式） | DEFAULT '' |
| multi_language_name_uuid | BIGINT UNSIGNED | 多语言名称UUID | DEFAULT 0 |
| **abbreviation** | **VARCHAR(10)** | **名称缩写（新增）** | **NOT NULL, DEFAULT ''** |
| color | VARCHAR(255) | 颜色值 | DEFAULT '' |
| sort | INT(11) | 排序 | DEFAULT 0 |
| create_time | INT(10) UNSIGNED | 创建时间 | DEFAULT 0 |
| update_time | INT(10) UNSIGNED | 更新时间 | DEFAULT 0 |
| delete_time | INT(10) UNSIGNED | 删除时间 | DEFAULT 0 |

**索引设计**:
- 主键索引: `PRIMARY KEY (id)`
- 唯一索引: `UNIQUE KEY unique_uuid (uuid)`

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_abbreviation_to_batch_tag_table.php`

### 数据库迁移

**迁移脚本**:

```php
<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddAbbreviationToBatchTagTable extends Migrator
{
    public function change()
    {
        $table = $this->table('batch_tag');
        
        // 添加 abbreviation 字段
        if (!$table->hasColumn('abbreviation')) {
            $table->addColumn('abbreviation', 'string', [
                'limit' => 10,
                'null' => false,
                'default' => '',
                'comment' => '名称缩写',
                'after' => 'multi_language_name_uuid'
            ])->update();
        }
    }
}
```

**数据迁移脚本**:

```php
<?php
// 为现有分批类型创建多语言记录（如果还没有）
// 为现有分批类型设置默认缩写
```

**同步 Go Model**:

在 `main/app/model/product.go` 中更新 `BatchTag` 结构体

**参考**: `docs/agent/workflows/database-migration.md`

---

## 📊 数据模型

### Go Model

```go
// main/app/model/product.go
// BatchTag 分批类型表,定义分批类型的相关信息 ttpos_batch_tag
type BatchTag struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`
	Abbreviation          string `gorm:"default:'';column:abbreviation;comment:'名称缩写'"` // 新增字段
	Color                 string `gorm:"default:'';column:color;comment:'颜色值，如#FF0000'"`
	Sort                  int    `gorm:"default:0;column:sort;comment:'排序(数字越小越靠前)';NOT NULL" json:"sort"`

	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}
```

### DTO 定义

#### Request DTO

```go
// main/app/dto/req/product.go

// ProductBatchTypeAddReq 分批类型添加请求
type BatchTagAddReq struct {
	LocaleName   dto.LocaleResponse `json:"locale_name" binding:"required"`   // 分批类型名称，多语言
	Abbreviation string             `json:"abbreviation" binding:"required"` // 名称缩写（新增）
	Color        string             `json:"color" binding:"required"`         // 颜色值，如#FF0000
}

// ProductBatchTypeEditReq 分批类型编辑请求
type BatchTagEditReq struct {
	Uuid         uint64             `json:"uuid" binding:"required"`           // 分批类型UUID
	LocaleName   dto.LocaleResponse `json:"locale_name" binding:"required"`   // 分批类型名称，多语言
	Abbreviation string             `json:"abbreviation" binding:"required"` // 名称缩写（新增）
	Color        string             `json:"color" binding:"required"`         // 颜色值，如#FF0000
}
```

#### Response DTO

```go
// main/app/dto/resp/product_resp/product.go

// ProductBatchType 分批类型
type BatchTag struct {
	Uuid         uint64             `json:"uuid"`         // 分批类型UUID
	LocaleName   dto.LocaleResponse `json:"locale_name"` // 分批类型名称，多语言
	Abbreviation string             `json:"abbreviation"` // 名称缩写（新增）
	Color        string             `json:"color"`        // 颜色值，如#FF0000
	Sort         int                `json:"sort"`         // 排序，数字越小越靠前
}

// ProductBatchTypeDetail 分批类型详情
type BatchTagDetail struct {
	Uuid         uint64             `json:"uuid"`         // 分批类型UUID
	LocaleName   dto.LocaleResponse `json:"locale_name"` // 分批类型名称，多语言
	Abbreviation string             `json:"abbreviation"` // 名称缩写（新增）
	Color        string             `json:"color"`        // 颜色值，如#FF0000
	Sort         int                `json:"sort"`         // 排序，数字越小越靠前
}
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 添加分批类型

**请求**:

- **URL**: `/api/v1/shop/batch/tag/add`
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
    "locale_name": {
      "zh": "主食",
      "en": "Main Course",
      "th": "อาหารหลัก"
    },
    "abbreviation": "主食",
    "color": "#FF0000"
  }
  ```
  **注意**：`locale_name` 字段已在 v2.9.0 实现，本次新增 `abbreviation` 字段。

**响应**:

```json
{
  "code": 1,
  "message": "添加成功",
  "data": {}
}
```

**错误响应**:

```json
{
  "code": 0,
  "message": "名称缩写不能为空",
  "data": {}
}
```

#### API 2: 编辑分批类型

**请求**:

- **URL**: `/api/v1/shop/batch/tag/edit`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "uuid": 123456,
    "locale_name": {
      "zh": "主食",
      "en": "Main Course",
      "th": "อาหารหลัก"
    },
    "abbreviation": "主食",
    "color": "#FF0000"
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "保存成功",
  "data": {}
}
```

#### API 3: 获取分批类型详情

**请求**:

- **URL**: `/api/v1/shop/batch/tag?uuid=123456`
- **Method**: `GET`

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 123456,
    "locale_name": {
      "zh": "主食",
      "en": "Main Course",
      "th": "อาหารหลัก"
    },
    "abbreviation": "主食",
    "color": "#FF0000",
    "sort": 1
  }
}
```

---

## 🧩 组件和接口

### Service 层

#### Service 接口（已存在）

```go
// main/app/service/i_product_service.go
type IProductSrv interface {
    AddBatchTag(ctx context.Context, req req.BatchTagAddReq) error
    EditBatchTag(ctx context.Context, req req.BatchTagEditReq) error
    GetBatchTag(ctx context.Context, req req.BatchTagReq) (*product_resp.BatchTagDetail, error)
    // ... 其他方法
}
```

#### Service 实现（需要修改）

```go
// main/app/service/product.go

// AddBatchTag 添加分批类型
func (s *productSrv) AddBatchTag(ctx context.Context, req req.BatchTagAddReq) error {
    db := s.dbm.GetDB(ctx.GetDbId())
    if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        batchTagRepo := repository.NewBatchTagRepo(tx)
        
        // 验证缩写字段（必填，长度 1-10 个字符）
        if req.Abbreviation == "" {
            return errors.New("名称缩写不能为空")
        }
        if len(req.Abbreviation) > 10 {
            return errors.New("名称缩写不能超过10个字符")
        }
        
        // 检查颜色是否已被使用
        if batchTagRepo.CheckColorExists(req.Color, 0) {
            return errors.New("该颜色已被其他分批类型使用")
        }

        // 获取下一个排序值
        maxSort, err := batchTagRepo.GetMaxSort()
        if err != nil {
            return errors.WithMessage(err, "获取当前最大的排序值失败")
        }
        nextSort := maxSort + 1

        // 创建多语言名称（已在 v2.9.0 实现，保持不变）
        multiLanguageName := model.MultiLanguageName{}
        multiLanguageName.InitByLocaleResponse(req.LocaleName)
        multiLanguageNameUuid, err := repository.NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(multiLanguageName)
        if err != nil {
            return errors.WithMessage(err, "创建多语言名称失败")
        }

        // 创建分批类型
        batchTag := model.BatchTag{
            Name:                  req.LocaleName.ToJson(),
            MultiLanguageNameUuid: multiLanguageNameUuid,
            Abbreviation:          req.Abbreviation, // 新增字段
            Color:                 req.Color,
            Sort:                  nextSort,
        }

        err = batchTagRepo.CreateBatchTag(batchTag)
        if err != nil {
            return err
        }

        return nil
    }); err != nil {
        return errors.WithMessage(err, "添加分批类型失败")
    }

    return nil
}

// EditBatchTag 编辑分批类型
func (s *productSrv) EditBatchTag(ctx context.Context, req req.BatchTagEditReq) error {
    db := s.dbm.GetDB(ctx.GetDbId())

    if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        batchTagRepo := repository.NewBatchTagRepo(tx)
        batchTag, err := batchTagRepo.GetBatchTagInfo(req.Uuid)
        if err != nil {
            return errors.WithMessage(err, "获取分批类型详情失败")
        }

        // 验证缩写字段（必填，长度 1-10 个字符）
        if req.Abbreviation == "" {
            return errors.New("名称缩写不能为空")
        }
        if len(req.Abbreviation) > 10 {
            return errors.New("名称缩写不能超过10个字符")
        }

        // 检查颜色是否已被其他分批类型使用（排除自己）
        if batchTagRepo.CheckColorExists(req.Color, req.Uuid) {
            return errors.New("该颜色已被其他分批类型使用")
        }

        // 更新多语言名称（已在 v2.9.0 实现，保持不变）
        batchTag.MultiLanguageName.InitByLocaleResponse(req.LocaleName)
        repository.NewMultiLanguageNameRepo(tx).UpdateMultiLanguageName(batchTag.MultiLanguageNameUuid, *batchTag.MultiLanguageName)

        // 更新分批类型
        batchTag.Color = req.Color
        batchTag.Name = req.LocaleName.ToJson()
        batchTag.Abbreviation = req.Abbreviation // 新增字段
        err = batchTagRepo.UpdateBatchTag(*batchTag)
        if err != nil {
            return errors.WithMessage(err, "更新分批类型失败")
        }

        return nil
    }); err != nil {
        return errors.WithMessage(err, "编辑分批类型失败")
    }

    return nil
}

// GetBatchTag 获取分批类型详情
func (s *productSrv) GetBatchTag(ctx context.Context, req req.BatchTagReq) (*product_resp.BatchTagDetail, error) {
    batchTagRepo := repository.NewBatchTagRepo(s.dbm.GetDB(ctx.GetDbId()))
    batchTag, err := batchTagRepo.GetBatchTagInfo(req.Uuid)
    if err != nil {
        return nil, errors.WithMessage(err, "获取分批类型详情失败")
    }

    return &product_resp.BatchTagDetail{
        Uuid:         batchTag.Uuid,
        LocaleName:   batchTag.MultiLanguageName.GetNames(), // 已在 v2.9.0 实现
        Abbreviation: batchTag.Abbreviation,                  // 新增字段
        Color:        batchTag.Color,
        Sort:         batchTag.Sort,
    }, nil
}
```

### Repository 层

#### Repository 接口（已存在，无需修改）

```go
// main/app/repository/i_batch_tag_repo.go
type IBatchTagRepo interface {
    GetBatchTags(opts ...DBOption) ([]*model.BatchTag, error)
    GetBatchTag(opts ...DBOption) (*model.BatchTag, error)
    GetBatchTagList() ([]*model.BatchTag, error)
    GetBatchTagInfo(uuid uint64) (*model.BatchTag, error)
    CreateBatchTag(batchTag model.BatchTag) error
    UpdateBatchTag(batchTag model.BatchTag) error
    DeleteBatchTag(uuid uint64) error
    // ... 其他方法
}
```

#### Repository 实现（已存在，无需修改）

Repository 实现已经支持所有需要的操作，只需要确保查询时包含 `Abbreviation` 字段即可。

### API 层

#### API Controller（需要修改）

```go
// main/app/api/v1/shop/shop_batch_product.go

// AddBatchTag 添加分批类型
func (h *BatchProductHandler) AddBatchTag(c *gin.Context) {
    ctx := helper.GetContext(c)
    addReq := req.BatchTagAddReq{}
    if err := c.ShouldBindJSON(&addReq); err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    // 验证缩写字段
    if addReq.Abbreviation == "" {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, errors.New("名称缩写不能为空"))
        return
    }
    if len(addReq.Abbreviation) > 10 {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, errors.New("名称缩写不能超过10个字符"))
        return
    }
    
    err := h.productSrv.AddBatchTag(ctx, addReq)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    helper.Success(c, nil, "添加成功")
}

// EditBatchTag 编辑分批类型
func (h *BatchProductHandler) EditBatchTag(c *gin.Context) {
    ctx := helper.GetContext(c)
    editReq := req.BatchTagEditReq{}
    if err := c.ShouldBindJSON(&editReq); err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    // 验证缩写字段
    if editReq.Abbreviation == "" {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, errors.New("名称缩写不能为空"))
        return
    }
    if len(editReq.Abbreviation) > 10 {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, errors.New("名称缩写不能超过10个字符"))
        return
    }
    
    err := h.productSrv.EditBatchTag(ctx, editReq)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    helper.Success(c, nil, "保存成功")
}

// GetBatchTag 获取分批类型详情（无需修改，Service 层已返回 Abbreviation）
func (h *BatchProductHandler) GetBatchTag(c *gin.Context) {
    // 现有实现无需修改
}
```

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:

- **Key 命名**: `ttpos:batch_tag:{uuid}`
- **过期时间**: 5 分钟
- **更新策略**: Cache-Aside Pattern

**示例**:

```go
// 缓存读取
key := fmt.Sprintf("ttpos:batch_tag:%d", uuid)
cached, err := cache.Get(key)
if err == nil {
    // 缓存命中
    return cached
}

// 缓存未命中，查询数据库
data, err := repo.GetBatchTagInfo(uuid)
if err != nil {
    return err
}

// 写入缓存
cache.Set(key, data, 5*time.Minute)
return data
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 缩写字段为空

- **处理方式**: 返回参数错误，提示"名称缩写不能为空"
- **用户影响**: 用户看到错误提示，无法提交
- **代码示例**:
  ```go
  if req.Abbreviation == "" {
      return errors.New("名称缩写不能为空")
  }
  ```

#### 场景 2: 缩写字段超过长度限制

- **处理方式**: 返回参数错误，提示"名称缩写不能超过10个字符"
- **用户影响**: 用户看到错误提示，无法提交


---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **Token 刷新**: 自动刷新机制

### 权限控制

- **RBAC**: 基于角色的访问控制
- **API 权限**: 每个 API 检查用户权限

### 数据安全

- **SQL 注入防护**: 使用参数化查询（GORM）
- **XSS 防护**: 前端输入校验

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- main/app/service: 70%+
- main/app/repository: 80%+

**测试内容**:

- Service 业务逻辑（AddBatchTag, EditBatchTag, GetBatchTag）
- Repository 数据访问
- DTO 数据转换
- 缩写字段验证（必填、长度限制）

**示例**:

```go
// main/app/service/product_test.go
func TestProductService_AddBatchTag(t *testing.T) {
    // 测试添加分批类型（包含缩写字段）
}

func TestProductService_AddBatchTag_AbbreviationRequired(t *testing.T) {
    // 测试缩写字段必填验证
}

func TestProductService_AddBatchTag_AbbreviationMaxLength(t *testing.T) {
    // 测试缩写字段长度限制（1-10个字符）
}
```

### API 测试

**测试内容**:

- API 接口调用
- 参数验证（缩写字段必填、长度限制）
- 响应格式（包含缩写字段）
- 错误处理

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 使用索引（uuid 已有唯一索引）
   - 优化 SQL 查询

2. **缓存优化**:
   - Redis 缓存热点数据
   - 缓存预热

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms
- 缓存命中率: > 80%

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件（增加 abbreviation 字段）
- [ ] 执行数据库迁移
- [ ] 更新 Go Model（增加 Abbreviation 字段）
- [ ] 编写数据迁移脚本（为现有数据设置默认缩写）

### Phase 2: 核心实现

- [ ] 更新 Request DTO（在 BatchTagAddReq 和 BatchTagEditReq 中增加 Abbreviation 字段）
- [ ] 更新 Response DTO（在 BatchTag 和 BatchTagDetail 中增加 Abbreviation 字段）
- [ ] 更新 Service 实现（AddBatchTag, EditBatchTag, GetBatchTag, GetBatchTagList 中处理缩写字段）
- [ ] 更新 API Controller（在 AddBatchTag 和 EditBatchTag 中增加参数验证）

### Phase 3: 测试

- [ ] 单元测试
- [ ] API 测试
- [ ] 集成测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: xiezhihuan  
**审核者**: 待定

