# 同步总部外卖商品功能 任务分解

> 本文档定义同步总部外卖商品到子店的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 5  
**已完成**: 4  
**进行中**: -  
**完成率**: 80%

---

## Phase 1: 核心实现

### Repository 层扩展

- [x] 1.1 添加 ProductPackageTakeout Repository 预加载方法

  - File: `main/app/repository/product_package_takeout.go`
  - Purpose: 添加预加载外卖规格价格和外卖属性价格的方法
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有预加载方法: `WithProductBomTakeouts`, `WithProductPackageAttributeTakeouts`（可能需要新增）
  - Action: 检查并确保 Repository 有以下预加载方法：
    - `WithProductBomTakeouts(opts ...DBOption) DBOption` - 预加载外卖规格价格
    - `WithProductPackageAttributeTakeouts(opts ...DBOption) DBOption` - 预加载外卖属性价格
  - Prompt: Role: Go Developer with GORM expertise | Task: 检查并添加 ProductPackageTakeoutRepo 的预加载方法 | Context: 需要预加载 ProductBomTakeouts 和 ProductPackageAttributeTakeouts 关联数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 GORM Preload | Success: 预加载方法添加完成，可以一次性加载所有关联数据

### Service 层实现

- [x] 1.2 实现 syncHeadquarterTakeoutProducts 辅助方法

  - File: `main/app/service/product.go`
  - Purpose: 实现同步总部外卖商品到子店的核心逻辑
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4
  - Leverage: 
    - 店内商品同步逻辑: `main/app/service/product.go:7754-8089`
    - ProductPackageTakeoutRepo: `main/app/repository/product_package_takeout.go`
    - ProductBomTakeoutRepo: `main/app/repository/product_bom_takeout.go`
    - ProductPackageAttributeTakeoutRepo: `main/app/repository/product_package_attribute_takeout.go`
  - Implementation Steps:
    1. 查询总部外卖商品（包含关联数据）
    2. 查询子店现有外卖商品（用于判断是否已同步）
    3. 构建子店已同步数据的 Map（用于快速查找 status 和 price）
    4. 遍历总部数据，准备新数据：
       - 外卖商品：首次同步 status=0，再次同步保留子店 status
       - 规格价格：首次同步使用总部 price，再次同步保留子店 price
       - 属性价格：始终使用总部 price
    5. 在事务中执行：
       - 批量删除子店现有外卖商品数据
       - 逐条插入新外卖商品数据（错误记录日志但不中断）
    6. 返回错误或 nil
  - Prompt: Role: Go Developer specializing in data synchronization | Task: 实现 syncHeadquarterTakeoutProducts 方法，同步总部外卖商品到子店 | Context: 参考店内商品同步逻辑（行7754-8089），使用相同的先删后建策略，保留子店的 status 和规格 price | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务，错误处理完善，使用 logger 记录日志 | Success: 方法实现完成，首次同步和再次同步逻辑正确，事务管理正确

- [x] 1.3 在 SyncProduct 方法中调用外卖商品同步逻辑

  - File: `main/app/service/product.go`
  - Purpose: 集成外卖商品同步到现有的商品同步流程
  - Requirements: Requirement 1, Requirement 4
  - Leverage: Task 1.2 的 syncHeadquarterTakeoutProducts 方法
  - Action: 在 `SyncProduct` 方法的店内商品同步逻辑内部添加：
    ```go
    // 在店内商品同步的 if 块内部
    if companySetting.IsSubShop() && syncHeadquarterData {
        // ... 店内商品同步逻辑 ...
        
        // 同步总店外卖商品到子店
        err = s.syncHeadquarterTakeoutProducts(ctx, db, headquarterDb, &companySetting)
        if err != nil {
            return errors.WithMessage(err, "同步总店外卖商品到子店失败")
        }
    }
    ```
  - Success: 外卖商品同步逻辑已集成到 SyncProduct 方法

- [x] 1.4 补充外卖商品图片文件同步功能

  - File: `main/app/service/product.go`
  - Purpose: 同步外卖商品的图片文件到子店
  - Requirements: Requirement 1 (外卖商品基本信息包含 ImageFileUuid)
  - Leverage: 
    - 店内商品图片同步逻辑: `main/app/service/product.go:8670-8750`
    - ProductPackageTakeout.ImageFileUuid 字段
  - Context: 
    - 外卖商品表有 `image_file_uuid` 字段关联 `ttpos_file` 表
    - 需要查询总部外卖商品的图片文件UUID
    - 将图片文件和文件分组同步到子店
  - Implementation Steps:
    1. 在 `SyncHeadquarterFile` 方法中，扩展查询逻辑：
       - 当前只查询 `ProductPackage.image_file_uuid`
       - 需要增加查询 `ProductPackageTakeout.image_file_uuid`
    2. 修改查询语句：
       ```go
       // 原有查询
       fileUuidQuery := headquarterDb.Model(&model.ProductPackage{}).Where("image_file_uuid > 0").Select("image_file_uuid")
       
       // 修改为联合查询
       productFileUuidQuery := headquarterDb.Model(&model.ProductPackage{}).Where("image_file_uuid > 0").Select("image_file_uuid")
       takeoutFileUuidQuery := headquarterDb.Model(&model.ProductPackageTakeout{}).Where("image_file_uuid > 0").Select("image_file_uuid")
       fileUuidQuery := headquarterDb.Raw("? UNION ?", productFileUuidQuery, takeoutFileUuidQuery)
       ```
    3. 其余逻辑保持不变（文件和文件分组的同步）
  - Prompt: Role: Go Developer with GORM expertise | Task: 在 SyncHeadquarterFile 方法中增加外卖商品图片同步 | Context: 外卖商品表 ProductPackageTakeout 有 image_file_uuid 字段，需要将这些图片文件也同步到子店 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 UNION 查询合并店内商品和外卖商品的图片UUID | Success: 外卖商品图片文件成功同步到子店

---

## Phase 2: 测试和验证

### 功能测试

- [ ] 2.1 手动测试验证

  - File: -
  - Purpose: 验证同步功能的正确性
  - Requirements: 所有功能需求
  - Test Cases:
    1. **首次同步测试**:
       - 准备总部外卖商品数据（包含规格价格、属性价格、图片文件、多语言名称和卖点描述）
       - 执行商品同步
       - 验证子店外卖商品 status = 0（下架）
       - 验证子店规格价格与总部一致
       - 验证子店属性价格与总部一致
       - 验证子店外卖商品图片文件已同步
       - **验证子店外卖商品多语言名称已同步**
       - **验证子店外卖商品卖点描述多语言已同步**
    2. **再次同步测试**:
       - 修改子店外卖商品的 status = 1（上架）
       - 修改子店某个规格的 price
       - 修改总部某个属性的 price
       - 修改总部外卖商品图片
       - 修改总部外卖商品的多语言名称和卖点描述
       - 执行商品同步
       - 验证子店外卖商品 status 保持为 1
       - 验证子店规格 price 保持不变
       - 验证子店属性 price 更新为总部最新值
       - 验证子店外卖商品图片更新为总部最新图片
       - **验证子店外卖商品多语言名称更新为总部最新值**
       - **验证子店外卖商品卖点描述多语言更新为总部最新值**
    3. **错误场景测试**:
       - 总部无外卖商品时执行同步
       - 子店无外卖商品时执行首次同步
       - 网络异常或数据库异常时的错误处理
  - Success: 所有测试用例通过，数据同步正确，错误处理完善

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 代码有完善的中文注释
- [ ] 错误处理使用 `errors.WithMessage`
- [ ] 日志记录使用 `logger.Logger.Error`

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
  - [x] Requirement 1: 同步总部外卖商品基本信息
  - [x] Requirement 2: 同步外卖规格价格
  - [x] Requirement 3: 同步外卖属性价格
  - [x] Requirement 4: 批量删除和批量插入
  - [x] Requirement 5: 错误处理和日志
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] CHANGELOG.md 已更新（如需要）
- [ ] 代码注释完整

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/structs.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-product-sync-headquarter-takeout/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-product-sync-headquarter-takeout/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-product-sync-headquarter-takeout/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-product-sync-headquarter-takeout/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-product-sync-headquarter-takeout/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的详细设计
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：实现参考

### 关键代码片段

#### 1. 查询总部外卖商品

```go
headTakeoutList, err := headTakeoutRepo.GetProductPackageTakeoutList(
    commonRepo.WhereByHeadquarterUuid(0),
    headTakeoutRepo.WithProductBomTakeouts(),
    headTakeoutRepo.WithProductPackageAttributeTakeouts(),
)
```

#### 2. 查询子店现有外卖商品

```go
subTakeoutList, err := subTakeoutRepo.GetProductPackageTakeoutList(
    commonRepo.WhereByHeadquarterUuid(companySetting.HeadquarterUuid),
    subTakeoutRepo.WithProductBomTakeouts(),
    subTakeoutRepo.WithProductPackageAttributeTakeouts(),
)
```

#### 3. 构建子店数据 Map

```go
subTakeoutMap := make(map[uint64]*model.ProductPackageTakeout)
subBomTakeoutMap := make(map[uint64]*model.ProductBomTakeout)
for _, takeout := range subTakeoutList {
    subTakeoutMap[takeout.Uuid] = takeout
    for _, bom := range takeout.ProductBomTakeouts {
        subBomTakeoutMap[bom.Uuid] = bom
    }
}
```

#### 4. 确定外卖商品状态

```go
status := uint(0) // 默认下架
if existsTakeout, ok := subTakeoutMap[headTakeout.Uuid]; ok {
    status = existsTakeout.Status // 保留子店状态
}
```

#### 5. 确定规格价格

```go
price := headBom.Price // 默认使用总部价格
if existsBom, ok := subBomTakeoutMap[headBom.Uuid]; ok {
    price = existsBom.Price // 保留子店价格
}
```

#### 6. 事务执行

```go
err = subDb.Transaction(func(tx *gorm.DB) error {
    // 1. 删除子店现有数据
    // 2. 插入新数据
    return nil
})
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/曾振华/2025-12/2025-12-18.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: 曾振华
