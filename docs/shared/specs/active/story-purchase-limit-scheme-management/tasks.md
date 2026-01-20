# story-purchase-limit-scheme-management 任务清单

> 本文档定义限购方案管理功能的开发任务和进度追踪。

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| **总 SP** | 3 |
| **总任务数** | 14 |
| **已完成** | 0 |
| **完成率** | 0% |
| **预计时间** | 3.5-4 天 |

---

## Phase 1: 数据库设计和迁移（1 天）

### 1.1 设计数据库表结构

| 项目 | 内容 |
|------|------|
| **File** | `admin/database/migrations/{timestamp}_migrate_purchase_quota_to_limit_scheme.php` |
| **Purpose** | 创建 4 个新表 + 迁移数据 + 删除旧表 |
| **Requirements** | Requirement 6: 数据迁移 |
| **Leverage** | 参考现有迁移文件格式 |

**子任务**：
- [ ] 1.1.1 设计 `ttpos_purchase_limit_scheme` 表结构
- [ ] 1.1.2 设计 `ttpos_purchase_limit_scheme_item` 表结构
- [ ] 1.1.3 设计 `ttpos_purchase_limit_scheme_shop` 表结构
- [ ] 1.1.4 设计 `ttpos_purchase_limit_scheme_weekday` 表结构
- [ ] 1.1.5 设计索引和外键约束

**验收标准**：
- 表结构符合数据库规范（`id`, `uuid`, `create_time`, `update_time`, `delete_time`）
- 索引设计合理（company_uuid, status, scheme_id, material_code 等）
- 字段类型和长度符合需求

---

### 1.2 编写数据迁移脚本

| 项目 | 内容 |
|------|------|
| **File** | `admin/database/migrations/{timestamp}_migrate_purchase_quota_to_limit_scheme.php` |
| **Purpose** | 迁移旧表数据到新方案表，然后删除旧表 |
| **Requirements** | Requirement 6: 数据迁移 |

**子任务**：
- [ ] 1.2.1 实现 `up()` 方法：创建 4 个新表
- [ ] 1.2.2 实现数据迁移逻辑：
  - [ ] 从 `ttpos_purchase_quota_config` 迁移到 `ttpos_purchase_limit_scheme` + `ttpos_purchase_limit_scheme_item`
  - [ ] 从 `ttpos_purchase_quota_config_shop` 迁移到 `ttpos_purchase_limit_scheme_shop`
  - [ ] 默认周期：全周（周一到周日，插入 7 条记录）
- [ ] 1.2.3 实现旧表删除逻辑
- [ ] 1.2.4 实现 `down()` 方法：回滚脚本
- [ ] 1.2.5 使用事务保证原子性

**验收标准**：
- 迁移脚本在测试环境成功执行
- 旧表数据完整迁移到新表
- 数据一致性得到验证
- 旧表成功删除
- 回滚脚本可以恢复数据

**迁移逻辑示例**：
```php
// 为每个 ttpos_purchase_quota_config 记录创建一个限购方案
foreach ($old_configs as $config) {
    // 1. 创建限购方案主记录
    $scheme_id = DB::table('ttpos_purchase_limit_scheme')->insertGetId([
        'uuid' => generateUuid(),
        'company_uuid' => $config->company_uuid,
        'name' => $config->material_name . '-限购方案',
        'status' => 1,
        'apply_to_all_shops' => $config->apply_to_all_shops,
        'daily_limit' => 0,
        'create_time' => time(),
        'update_time' => time(),
    ]);
    
    // 2. 创建物品配置记录
    DB::table('ttpos_purchase_limit_scheme_item')->insert([
        'uuid' => generateUuid(),
        'scheme_id' => $scheme_id,
        'material_code' => $config->material_code,
        'unit_code' => $config->unit_code,
        'quota_limit' => $config->quota_limit,
        'create_time' => time(),
        'update_time' => time(),
    ]);
    
    // 3. 创建周期配置记录（默认全周）
    for ($weekday = 1; $weekday <= 7; $weekday++) {
        DB::table('ttpos_purchase_limit_scheme_weekday')->insert([
            'uuid' => generateUuid(),
            'scheme_id' => $scheme_id,
            'weekday' => $weekday,
            'create_time' => time(),
            'update_time' => time(),
        ]);
    }
    
    // 4. 迁移门店配置
    if ($config->apply_to_all_shops == 0) {
        $shops = DB::table('ttpos_purchase_quota_config_shop')
            ->where('config_id', $config->id)
            ->get();
        foreach ($shops as $shop) {
            DB::table('ttpos_purchase_limit_scheme_shop')->insert([
                'uuid' => generateUuid(),
                'scheme_id' => $scheme_id,
                'company_uuid' => $shop->company_uuid,
                'create_time' => time(),
                'update_time' => time(),
            ]);
        }
    }
}

// 5. 删除旧表
DB::statement('DROP TABLE IF EXISTS ttpos_purchase_quota_config_shop');
DB::statement('DROP TABLE IF EXISTS ttpos_purchase_quota_config');
```

---

### 1.3 创建 Go Model 文件

| 项目 | 内容 |
|------|------|
| **Files** | `main/app/model/purchase_limit_scheme*.go`（4 个文件） |
| **Purpose** | 数据模型定义 |
| **Requirements** | 所有 Requirements |

**子任务**：
- [ ] 1.3.1 创建 `purchase_limit_scheme.go`（主表）
- [ ] 1.3.2 创建 `purchase_limit_scheme_item.go`（物品配置表）
- [ ] 1.3.3 创建 `purchase_limit_scheme_shop.go`（门店配置表）
- [ ] 1.3.4 创建 `purchase_limit_scheme_weekday.go`（星期配置表）

**验收标准**：
- Model 结构体包含所有必需字段
- TableName() 方法正确
- GORM 标签正确

---

### 1.4 测试数据迁移脚本

| 项目 | 内容 |
|------|------|
| **Environment** | 测试环境 |
| **Purpose** | 验证数据迁移脚本正确性 |

**子任务**：
- [ ] 1.4.1 在测试环境准备旧表数据
- [ ] 1.4.2 执行迁移脚本
- [ ] 1.4.3 验证数据完整性
- [ ] 1.4.4 测试回滚脚本

---

## Phase 2: Repository + Service 核心实现（1.5 天）

### 2.1 实现 Repository 层

| 项目 | 内容 |
|------|------|
| **Files** | `main/app/repository/purchase_limit_scheme_*.go`（4 个文件） |
| **Purpose** | 数据访问层实现 |
| **Requirements** | 所有 Requirements |
| **Leverage** | 参考 `purchase_quota_config.go` 的 Repository 设计 |

**子任务**：
- [ ] 2.1.1 实现 `purchase_limit_scheme_repo.go`（主表 CRUD）
- [ ] 2.1.2 实现 `purchase_limit_scheme_item_repo.go`（物品配置 CRUD）
- [ ] 2.1.3 实现 `purchase_limit_scheme_shop_repo.go`（门店配置 CRUD）
- [ ] 2.1.4 实现 `purchase_limit_scheme_weekday_repo.go`（星期配置 CRUD）
- [ ] 2.1.5 实现选项方法（Where*, Paginate 等）

**验收标准**：
- Repository 只持有 db 实例（不持有 DBManager）
- 使用选项模式（DBOption）
- 支持软删除（delete_time）
- 错误处理完整

---

### 2.2 实现 Service 层

| 项目 | 内容 |
|------|------|
| **File** | `main/app/service/purchase_order/purchase_limit_scheme.go` |
| **Purpose** | 限购方案管理业务逻辑 |
| **Requirements** | Requirements 1-5 |
| **Leverage** | 参考 `purchase_order.go` 的 Service 结构 |

**子任务**：
- [ ] 2.2.1 定义 Service 接口 `IPurchaseLimitSchemeSrv`
- [ ] 2.2.2 实现 `Create()` 方法（创建方案 + 关联表）
- [ ] 2.2.3 实现 `Update()` 方法（更新方案 + 关联表）
- [ ] 2.2.4 实现 `GetByUuid()` 方法（查询方案详情 + 关联数据）
- [ ] 2.2.5 实现 `GetList()` 方法（查询方案列表 + 统计信息）
- [ ] 2.2.6 实现 `Delete()` 方法（软删除方案 + 关联表）
- [ ] 2.2.7 实现业务校验逻辑（名称重复检查、周期/物品必填等）

**验收标准**：
- Service 只依赖其他 Service 接口（不依赖 Repository）
- 使用事务保证多表操作的原子性
- 错误处理完整，返回 error 不使用 panic
- 日志记录完整

**核心逻辑**：

```go
func (s *purchaseLimitSchemeSrv) Create(ctx context.Context, req req.PurchaseLimitSchemeCreateReq) (uint64, error) {
    db := ctx.GetDB()
    
    // 1. 校验名称重复
    schemeRepo := repository.NewPurchaseLimitSchemeRepo(db)
    existing, _ := schemeRepo.GetList(
        schemeRepo.WhereCompanyUuid(ctx.GetCompanyUuid()),
        schemeRepo.WhereName(req.Name),
    )
    if len(existing) > 0 {
        return 0, errors.New("限购方案名称已存在")
    }
    
    // 2. 开启事务
    return db.Transaction(func(tx *gorm.DB) error {
        // 2.1 创建限购方案主记录
        scheme := &model.PurchaseLimitScheme{
            Uuid:            utils.GenerateUuid(),
            CompanyUuid:     ctx.GetCompanyUuid(),
            Name:            req.Name,
            Status:          req.Status,
            ApplyToAllShops: req.ApplyToAllShops,
            DailyLimit:      req.DailyLimit,
            CreateTime:      time.Now().Unix(),
            UpdateTime:      time.Now().Unix(),
        }
        if err := repository.NewPurchaseLimitSchemeRepo(tx).Create(scheme); err != nil {
            return err
        }
        
        // 2.2 批量插入星期配置
        weekdayRepo := repository.NewPurchaseLimitSchemeWeekdayRepo(tx)
        for _, weekday := range req.Weekdays {
            weekdayRecord := &model.PurchaseLimitSchemeWeekday{
                Uuid:       utils.GenerateUuid(),
                SchemeId:   scheme.Id,
                Weekday:    weekday,
                CreateTime: time.Now().Unix(),
                UpdateTime: time.Now().Unix(),
            }
            if err := weekdayRepo.Create(weekdayRecord); err != nil {
                return err
            }
        }
        
        // 2.3 批量插入物品配置
        itemRepo := repository.NewPurchaseLimitSchemeItemRepo(tx)
        for _, item := range req.Items {
            itemRecord := &model.PurchaseLimitSchemeItem{
                Uuid:         utils.GenerateUuid(),
                SchemeId:     scheme.Id,
                MaterialCode: item.MaterialCode,
                UnitCode:     item.UnitCode,
                QuotaLimit:   item.QuotaLimit,
                CreateTime:   time.Now().Unix(),
                UpdateTime:   time.Now().Unix(),
            }
            if err := itemRepo.Create(itemRecord); err != nil {
                return err
            }
        }
        
        // 2.4 插入门店配置（如果不是全部门店）
        if req.ApplyToAllShops == 0 && len(req.Shops) > 0 {
            shopRepo := repository.NewPurchaseLimitSchemeShopRepo(tx)
            for _, shopUuid := range req.Shops {
                shopRecord := &model.PurchaseLimitSchemeShop{
                    Uuid:        utils.GenerateUuid(),
                    SchemeId:    scheme.Id,
                    CompanyUuid: shopUuid,
                    CreateTime:  time.Now().Unix(),
                    UpdateTime:  time.Now().Unix(),
                }
                if err := shopRepo.Create(shopRecord); err != nil {
                    return err
                }
            }
        }
        
        return scheme.Uuid, nil
    })
}
```

---

## Phase 2: Repository + Service 核心实现（1.5 天）

### 2.1 实现 PurchaseLimitSchemeRepo

| 项目 | 内容 |
|------|------|
| **File** | `main/app/repository/purchase_limit_scheme_repo.go` |
| **Purpose** | 限购方案主表 CRUD |
| **Requirements** | Requirements 1-2 |
| **Leverage** | 参考 `purchase_quota_config.go` Repository |

**子任务**：
- [ ] 2.1.1 定义 Repository 接口
- [ ] 2.1.2 实现 `Create()` 方法
- [ ] 2.1.3 实现 `Update()` 方法
- [ ] 2.1.4 实现 `GetByUuid()` 方法
- [ ] 2.1.5 实现 `GetList()` 方法（支持分页和筛选）
- [ ] 2.1.6 实现 `Delete()` 方法（软删除）
- [ ] 2.1.7 实现选项方法（WhereCompanyUuid, WhereStatus, WhereName, Paginate）

---

### 2.2 实现 PurchaseLimitSchemeItemRepo

| 项目 | 内容 |
|------|------|
| **File** | `main/app/repository/purchase_limit_scheme_item_repo.go` |
| **Purpose** | 物品配置表 CRUD |
| **Requirements** | Requirement 4 |

**子任务**：
- [ ] 2.2.1 定义 Repository 接口
- [ ] 2.2.2 实现 CRUD 方法
- [ ] 2.2.3 实现 `GetBySchemeId()` 方法（查询方案的所有物品）
- [ ] 2.2.4 实现 `DeleteBySchemeId()` 方法（批量删除）

---

### 2.3 实现 PurchaseLimitSchemeShopRepo

| 项目 | 内容 |
|------|------|
| **File** | `main/app/repository/purchase_limit_scheme_shop_repo.go` |
| **Purpose** | 门店配置表 CRUD |
| **Requirements** | Requirement 5 |

**子任务**：
- [ ] 2.3.1 定义 Repository 接口
- [ ] 2.3.2 实现 CRUD 方法
- [ ] 2.3.3 实现 `GetBySchemeId()` 方法（查询方案的所有门店）
- [ ] 2.3.4 实现 `DeleteBySchemeId()` 方法（批量删除）

---

### 2.4 实现 PurchaseLimitSchemeWeekdayRepo

| 项目 | 内容 |
|------|------|
| **File** | `main/app/repository/purchase_limit_scheme_weekday_repo.go` |
| **Purpose** | 星期配置表 CRUD |
| **Requirements** | Requirement 3 |

**子任务**：
- [ ] 2.4.1 定义 Repository 接口
- [ ] 2.4.2 实现 CRUD 方法
- [ ] 2.4.3 实现 `GetBySchemeId()` 方法（查询方案的所有星期）
- [ ] 2.4.4 实现 `DeleteBySchemeId()` 方法（批量删除）

---

### 2.5 实现 PurchaseLimitSchemeService

| 项目 | 内容 |
|------|------|
| **File** | `main/app/service/purchase_order/purchase_limit_scheme.go` |
| **Purpose** | 限购方案管理业务逻辑 |
| **Requirements** | Requirements 1-5 |
| **Leverage** | 参考 `purchase_order.go` Service |

**子任务**：
- [ ] 2.5.1 定义 Service 接口 `IPurchaseLimitSchemeSrv`
- [ ] 2.5.2 实现 `Create()` 方法（事务处理，创建主表 + 3 个关联表）
- [ ] 2.5.3 实现 `Update()` 方法（事务处理，更新主表 + 删除旧关联 + 插入新关联）
- [ ] 2.5.4 实现 `GetByUuid()` 方法（查询主表 + 关联表，组装完整数据）
- [ ] 2.5.5 实现 `GetList()` 方法（查询主表 + 统计关联表数量）
- [ ] 2.5.6 实现 `Delete()` 方法（软删除主表 + 关联表）
- [ ] 2.5.7 实现业务校验（名称重复、周期必填、物品必填等）

**验收标准**：
- 使用事务保证多表操作原子性
- 错误处理完整
- 日志记录关键操作

---

### 2.6 创建 DTO 定义

| 项目 | 内容 |
|------|------|
| **Files** | `main/app/dto/req/purchase_limit_scheme_req.go` + `main/app/dto/resp/purchase_limit_scheme_resp.go` |
| **Purpose** | 请求和响应数据结构 |

**子任务**：
- [ ] 2.6.1 创建 Request DTO（Create, Update, Get, List, Delete）
- [ ] 2.6.2 创建 Response DTO（Detail, Summary, List）
- [ ] 2.6.3 添加参数验证标签（binding）

---

## Phase 3: API 层集成（0.5 天）

### 3.1 实现 API Handler

| 项目 | 内容 |
|------|------|
| **File** | `main/app/api/v1/shop/shop_purchase.go` |
| **Purpose** | 新增 5 个限购方案接口 |
| **Requirements** | Requirements 1-5 |
| **Leverage** | 参考现有 API Handler 模式 |

**子任务**：
- [ ] 3.1.1 在 `PurchaseHandler` 中注入 `IPurchaseLimitSchemeSrv`
- [ ] 3.1.2 实现 `GetLimitSchemeList()` Handler
- [ ] 3.1.3 实现 `GetLimitSchemeDetail()` Handler
- [ ] 3.1.4 实现 `CreateLimitScheme()` Handler
- [ ] 3.1.5 实现 `UpdateLimitScheme()` Handler
- [ ] 3.1.6 实现 `DeleteLimitScheme()` Handler

**验收标准**：
- 参数验证完整
- 响应格式符合规范（data 为对象）
- 错误处理统一

---

### 3.2 注册路由

| 项目 | 内容 |
|------|------|
| **File** | `main/app/api/v1/shop/shop_purchase.go` |
| **Purpose** | 注册限购方案 API 路由 |

**子任务**：
- [ ] 3.2.1 在 `RegisterRoutes()` 方法中添加 5 个路由
- [ ] 3.2.2 使用正确的路由格式：`/purchase/limit_scheme/{action}`
- [ ] 3.2.3 应用认证中间件

**路由清单**：
```go
privateApi.GET("/purchase/limit_scheme/list", wrapper.GetLimitSchemeList)
privateApi.GET("/purchase/limit_scheme/detail", wrapper.GetLimitSchemeDetail)
privateApi.POST("/purchase/limit_scheme/create", wrapper.CreateLimitScheme)
privateApi.POST("/purchase/limit_scheme/update", wrapper.UpdateLimitScheme)
privateApi.DELETE("/purchase/limit_scheme/delete", wrapper.DeleteLimitScheme)
```

---

## Phase 4: 测试与文档（0.5 天）

### 4.1 编写单元测试

| 项目 | 内容 |
|------|------|
| **Files** | `main/app/service/purchase_order/purchase_limit_scheme_test.go` + Repository 测试文件 |
| **Purpose** | 单元测试覆盖 |
| **Requirements** | 覆盖率 Service ≥ 70%, Repository ≥ 80% |

**子任务**：
- [ ] 4.1.1 Service 层测试（Create, Update, GetByUuid, GetList, Delete）
- [ ] 4.1.2 Repository 层测试（CRUD + 选项方法）
- [ ] 4.1.3 业务逻辑测试（名称重复、周期必填等）

**测试命令**：
```bash
cd main
go test -v ./app/service/purchase_order/... -run TestPurchaseLimitScheme
go test -coverprofile=coverage.out ./app/service/purchase_order/
go tool cover -html=coverage.out
```

---

### 4.2 API 测试

| 项目 | 内容 |
|------|------|
| **Tool** | Postman / curl |
| **Purpose** | API 接口测试 |

**子任务**：
- [ ] 4.2.1 测试创建限购方案 API
- [ ] 4.2.2 测试更新限购方案 API
- [ ] 4.2.3 测试查询限购方案列表 API
- [ ] 4.2.4 测试查询限购方案详情 API
- [ ] 4.2.5 测试删除限购方案 API
- [ ] 4.2.6 测试参数验证（缺少必填项、格式错误等）
- [ ] 4.2.7 测试错误处理（记录不存在、权限不足等）

---

### 4.3 数据迁移验证

| 项目 | 内容 |
|------|------|
| **Environment** | 测试环境 → 生产环境 |
| **Purpose** | 验证数据迁移的正确性和完整性 |

**子任务**：
- [ ] 4.3.1 在测试环境执行迁移脚本
- [ ] 4.3.2 验证数据完整性（记录数量、字段值）
- [ ] 4.3.3 验证业务连续性（新方案功能正常）
- [ ] 4.3.4 准备生产环境迁移计划
- [ ] 4.3.5 备份生产环境旧表数据
- [ ] 4.3.6 在维护窗口期执行生产迁移
- [ ] 4.3.7 迁移后验证

---

### 4.4 文档更新

| 项目 | 内容 |
|------|------|
| **Files** | API 文档、数据库文档 |
| **Purpose** | 完善技术文档 |

**子任务**：
- [ ] 4.4.1 补充 API 接口文档（请求示例、响应示例）
- [ ] 4.4.2 补充数据库表结构文档
- [ ] 4.4.3 补充数据迁移流程文档
- [ ] 4.4.4 更新 design.md（补充实际实现细节）

---

## 提交清单

### 代码质量

- [ ] `cd main && go mod tidy` 执行
- [ ] `cd main && go fmt ./...` 执行
- [ ] `cd main && go vet ./...` 通过
- [ ] 测试通过: `cd main && go test ./app/service/purchase_order/...`
- [ ] 测试覆盖率达标

### 功能完整性

- [ ] 所有验收标准通过（Requirements 1-6）
- [ ] API 响应格式正确（data 为对象）
- [ ] 多语言字段使用 LocaleResponse（如需要）
- [ ] 错误提示清晰明确

### 数据库迁移

- [ ] 迁移文件已创建
- [ ] 迁移脚本在测试环境验证通过
- [ ] 旧表数据备份完成
- [ ] 生产环境迁移计划制定
- [ ] 回滚脚本已准备

### 安全检查

- [ ] 所有 API 需要身份验证
- [ ] 权限控制正确（总部用户可管理）
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 敏感操作有日志记录

---

## 文件清单

### 新增文件

#### Model 层（4 个文件）
- `main/app/model/purchase_limit_scheme.go`
- `main/app/model/purchase_limit_scheme_item.go`
- `main/app/model/purchase_limit_scheme_shop.go`
- `main/app/model/purchase_limit_scheme_weekday.go`

#### Repository 层（4 个文件）
- `main/app/repository/purchase_limit_scheme_repo.go`
- `main/app/repository/purchase_limit_scheme_item_repo.go`
- `main/app/repository/purchase_limit_scheme_shop_repo.go`
- `main/app/repository/purchase_limit_scheme_weekday_repo.go`

#### Service 层（1 个文件）
- `main/app/service/purchase_order/purchase_limit_scheme.go`

#### DTO 层（2 个文件）
- `main/app/dto/req/purchase_limit_scheme_req.go`
- `main/app/dto/resp/purchase_limit_scheme_resp.go`

#### 数据迁移（1 个文件）
- `admin/database/migrations/{timestamp}_migrate_purchase_quota_to_limit_scheme.php`

### 修改文件

#### API 层（1 个文件）
- `main/app/api/v1/shop/shop_purchase.go`
  - 新增 5 个 Handler 方法
  - 新增 5 个路由注册
  - 注入 `IPurchaseLimitSchemeSrv` 依赖

---

## 风险识别

### 风险 1: 数据迁移失败导致数据丢失

**影响**: 高  
**概率**: 中  
**缓解措施**:
- 迁移前备份旧表数据
- 使用事务保证原子性
- 在测试环境充分验证
- 提供回滚脚本
- 生产环境在维护窗口期执行

### 风险 2: 多表事务性能问题

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 批量插入优化
- 使用数据库连接池
- 监控事务执行时间

---

## 下一步

- [ ] 执行任务：按 Phase 顺序开发
- [ ] 代码审查：提交 PR 前进行 Code Review
- [ ] 测试验证：确保所有测试通过
- [ ] 部署上线：执行数据迁移和功能发布

---

**版本**: v1.0.0  
**创建日期**: 2026-01-20  
**作者**: weifashi  
**Story Point**: 3  
**状态**: ✅ 设计完成，待开发
