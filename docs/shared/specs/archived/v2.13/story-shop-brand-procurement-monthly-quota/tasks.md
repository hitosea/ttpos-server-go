# 品牌采购限额控制 任务分解

> 本文档定义品牌采购限额控制功能的详细执行任务清单（包含申请次数限制、物品数量限制、月度限购）。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 51  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

**预估工作量**: 5.0 天（SP = 5）

---

## Phase 0: 全局配置初始化

- [x] 0.1 添加全局配置项

  - File: 通过 SQL 或 Seeder 添加到 `ttpos_config` 表
  - Purpose: 初始化品牌采购申请次数和单次数量上限配置
  - Requirements: Requirement 6.1, 6.2
  - Leverage: 现有配置表 `ttpos_config`，参考其他配置项的初始化方式
  - SQL:
    ```sql
    INSERT INTO `ttpos_config` (`name`, `value`, `description`, `create_time`) VALUES
    ('purchase.brand.daily_limit', '2', '品牌采购每日申请次数上限', UNIX_TIMESTAMP()),
    ('purchase.brand.single_qty_limit', '100', '品牌采购单次物品数量上限', UNIX_TIMESTAMP())
    ON DUPLICATE KEY UPDATE `value`=VALUES(`value`), `description`=VALUES(`description`);
    ```
  - Success: 配置项添加成功，可通过 config 读取

---

## Phase 1: 数据库设计和迁移

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [ ] 1.1 创建限购配置主表迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_purchase_quota_config_table.php`
  - Purpose: 定义限购配置主表结构（支持多门店维度）
  - Requirements: Requirement 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_purchase_quota_config 表的迁移文件，遵循 requirements.md 中的数据库设计 | Context: 必须包含 apply_to_all_shops(默认1) 字段，添加 uk_uuid 唯一索引，添加 idx_material, idx_status 和 idx_delete_time 普通索引 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查表是否存在 | Success: 迁移文件创建成功，字段定义正确，索引完整

- [ ] 1.1a 创建限购配置门店关联表迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_purchase_quota_config_shop_table.php`
  - Purpose: 定义限购配置与门店的多对多关联关系
  - Requirements: Requirement 3.1, 3.4
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_purchase_quota_config_shop 关联表的迁移文件 | Context: 必须包含 config_uuid, shop_uuid, create_time, delete_time 字段，添加 uk_config_shop 联合唯一索引，添加 idx_config, idx_shop, idx_delete_time 普通索引 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，字段定义正确，索引完整

- [ ] 1.1b 创建门店配置表迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_shop_purchase_config_table.php`
  - Purpose: 创建门店级采购配置表
  - Requirements: Requirement 7.1-7.4
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_shop_purchase_config 表的迁移文件 | Context: 必须包含 shop_uuid(唯一索引), daily_limit(默认2), single_qty_limit(默认100), create_time, update_time, delete_time 字段 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建限购配置表
  - Requirements: Requirement 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表已创建，索引已创建

- [ ] 1.3 创建 Go Model（支持关联表）

  - File: `main/app/model/purchase_quota_config.go`
  - Purpose: 定义限购配置主表和关联表的 Go 数据模型
  - Requirements: Requirement 3.1-3.6
  - Leverage: 现有 Model: `main/app/model/`，迁移文件: Task 1.1, 1.1a
  - Prompt: Role: Go Developer | Task: 创建 PurchaseQuotaConfig 和 PurchaseQuotaConfigShop 两个结构体，映射到主表和关联表 | Context: 主表使用 gorm 标签，增加 ApplyToAllShops 字段，增加 Shops []PurchaseQuotaConfigShop 关联字段；实现 GetShopUuidList() 和 AppliesTo(shopUuid) 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，关联字段使用 foreignKey 和 references 标签 | Success: Model 创建成功，字段映射正确，关联关系正确，方法实现正确

- [ ] 1.3a 创建门店配置 Go Model

  - File: `main/app/model/shop_purchase_config.go`
  - Purpose: 定义门店采购配置 Go 数据模型
  - Requirements: Requirement 7.1-7.4
  - Leverage: 现有 Model: `main/app/model/`
  - Prompt: Role: Go Developer | Task: 创建 ShopPurchaseConfig 结构体，映射到 ttpos_shop_purchase_config 表 | Context: 包含 ShopUuid, DailyLimit, SingleQtyLimit 等字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 创建成功，字段映射正确

- [ ] 1.4 创建常量定义

  - File: `main/app/constant/purchase_quota.go`
  - Purpose: 定义限购相关常量（状态、周期类型、超限策略、配置来源）
  - Requirements: Requirement 1.3
  - Leverage: 现有常量文件: `main/app/constant/`
  - Prompt: Role: Go Developer | Task: 创建限购常量，包括配置状态（启用/禁用）、周期类型（按天(默认)/月度/季度/年度）、超限策略（严格拒绝）、配置来源（门店/总部） | Context: 使用 const 定义，按天为默认值0，遵循命名规范 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 常量定义完整，命名清晰

---

## Phase 2: Repository 层实现

- [ ] 2.1 创建 Repository 接口

  - File: `main/app/repository/i_purchase_quota_config_repo.go`
  - Purpose: 定义限购配置数据访问接口（包含门店关联操作）
  - Requirements: Requirement 1.1, 2.1, 3.1
  - Leverage: 现有 Repository 接口: `main/app/repository/i_*_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 IPurchaseQuotaConfigRepo 接口，定义 GetByMaterialUuidAndShop（支持门店过滤）, GetList, Create, Update, Delete, BatchCreateShops（批量创建门店关联）, DeleteShops（删除门店关联）方法，以及选项方法 WhereStatus, WhereMaterialUuid, WhereUuid | Context: 使用选项模式(DBOption)，接口方法返回 (*model.PurchaseQuotaConfig, error) 或 ([]*model.PurchaseQuotaConfig, int64, error) | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确，包含门店关联操作

- [ ] 2.2 实现 Repository（选项模式+事务处理）

  - File: `main/app/repository/purchase_quota_config_repo.go`
  - Purpose: 实现限购配置数据访问逻辑（包含门店关联事务处理）
  - Requirements: Requirement 1.1, 2.1, 3.1
  - Leverage: 现有 Repository 实现: `main/app/repository/*_repo.go`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 purchaseQuotaConfigRepoImpl，使用选项模式实现灵活查询 | Context: 只持有 db *gorm.DB，实现 GetByMaterialUuidAndShop（使用 EXISTS 子查询判断门店关联）, GetList, Create, Update, Delete（事务：同时软删除主表和关联表）, BatchCreateShops, DeleteShops, WhereStatus, WhereMaterialUuid, WhereUuid 方法 | Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0)，查询时过滤 status=1，Delete 和创建门店关联必须使用事务 | Success: Repository 实现完整，选项模式正确，软删除正确，状态过滤正确，事务处理正确

- [ ] 2.3 编写 Repository 单元测试

  - File: `main/app/repository/purchase_quota_config_repo_test.go`
  - Purpose: 确保 Repository 数据访问正确（包含门店关联测试）
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 PurchaseQuotaConfigRepo 编写单元测试，覆盖率 ≥ 80% | Context: 测试 GetByMaterialUuidAndShop（apply_to_all_shops=1/0 时门店过滤逻辑）, GetList（空/非空）, Create, Update, Delete（主表和关联表同时删除）, BatchCreateShops, DeleteShops, 选项方法，测试软删除，测试状态过滤 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过，门店关联逻辑测试充分

---

## Phase 3: Service 层集成

- [ ] 3.1 实现申请次数校验方法

  - File: `main/app/service/purchase_order/purchase_order.go`
  - Purpose: 检查门店每日申请次数是否超限
  - Requirements: Requirement 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 Service: `main/app/service/purchase_order/purchase_order.go`，店铺时区处理参考现有代码
  - Prompt: Role: Go Developer | Task: 在 purchaseOrderSrv 中新增 checkDailySubmitLimit 方法，实现每日申请次数限制校验 | Context: 读取全局配置 purchase.brand.daily_limit，获取店铺时区，使用时区计算当天起止时间戳，统计当天已提交（status!=0）的申请次数，超限返回错误 | Restrictions: 草稿(status=0)不计入，使用店铺时区，遵循 .cursor/rules/go-main.mdc | Success: 校验准确，时区处理正确，错误提示国际化

- [ ] 3.2 实现单次数量校验方法

  - File: `main/app/service/purchase_order/purchase_order.go`
  - Purpose: 检查单次申请物品总数量是否超限
  - Requirements: Requirement 2.1, 2.2, 2.3
  - Leverage: 现有 Service: `main/app/service/purchase_order/purchase_order.go`
  - Prompt: Role: Go Developer | Task: 在 purchaseOrderSrv 中新增 checkSingleQtyLimit 方法，实现单次数量限制校验 | Context: 读取全局配置 purchase.brand.single_qty_limit，遍历订单明细统计总数量，超限返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 校验准确，错误提示国际化

- [ ] 3.3 在 PurchaseOrderService 中新增月度限购校验方法

  - File: `main/app/service/purchase_order/purchase_order.go`
  - Purpose: 实现限购校验核心逻辑
  - Requirements: Requirement 2.1, 2.2, 2.3, 3.1, 3.2
  - Leverage: 现有 Service: `main/app/service/purchase_order/purchase_order.go`，Task 2.1-2.2 的 Repository
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 purchaseOrderSrv 中新增 checkPurchaseQuota 方法，实现限购校验逻辑 | Context: 遍历订单明细，查询限购配置，校验单位，统计已用额度，判断是否超限 | Restrictions: 仅对 PurchaseType=2（品牌采购）生效，使用 errors.WithMessage 包装错误，记录日志 | Success: 校验逻辑完整，错误提示明确，日志记录清晰

- [ ] 3.4 实现已用额度统计方法

  - File: `main/app/service/purchase_order/purchase_order.go`
  - Purpose: 实时查询本月已使用的采购额度
  - Requirements: Requirement 4.2, 4.3
  - Leverage: 现有 Service: `main/app/service/purchase_order/purchase_order.go`
  - Prompt: Role: Go Developer with SQL expertise | Task: 在 purchaseOrderSrv 中新增 getMonthlyUsedQuota 方法，实现实时统计逻辑 | Context: 使用 JOIN 查询 ttpos_purchase_order 和 ttpos_purchase_order_item，过滤条件：purchase_type=2, material_uuid, unit_uuid, status IN (1,2,4,5), 当前月份, 排除当前单据 | Restrictions: 使用 GORM，使用 COALESCE 处理空值，使用 FROM_UNIXTIME 格式化时间 | Success: 统计准确，查询性能良好，SQL 正确

- [ ] 3.5 在 SubmitPurchaseOrder 中集成三维度限购校验

  - File: `main/app/service/purchase_order/purchase_order.go`
  - Purpose: 在提交接口中按顺序调用三个校验方法
  - Requirements: Requirement 1.1-1.5, 2.1-2.3, 4.1-4.4
  - Leverage: Task 3.1, 3.2, 3.3 的校验方法
  - Prompt: Role: Go Developer | Task: 在 SubmitPurchaseOrder 方法中，订单状态校验后、提交前，依次调用 ① checkDailySubmitLimit ② checkSingleQtyLimit ③ checkPurchaseQuota | Context: 只对 PurchaseType=2 的订单进行校验，任一校验失败立即返回错误 | Restrictions: 不影响现有流程，遵循 .cursor/rules/go-main.mdc | Success: 集成成功，三维度校验生效，其他采购类型不受影响

- [ ] 3.6 编写 Service 单元测试

  - File: `main/app/service/purchase_order/purchase_order_test.go`
  - Purpose: 确保限购校验逻辑正确
  - Requirements: Requirement 1.1-1.5, 2.1-2.3, 4.1-4.4, 5.1, 5.2
  - Leverage: 现有测试: `main/app/service/purchase_order/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为三维度限购校验逻辑编写单元测试，覆盖率 ≥ 70% | Context: 测试申请次数限制（超限/未超限/草稿不计入），测试单次数量限制（超限/未超限），测试月度限购（无配置放行/单位匹配/单位不匹配/超限/未超限），测试已用额度统计 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过，边界情况已覆盖

---

## Phase 4: 国际化实现

- [ ] 4.1 添加中文错误提示

  - File: `main/i18n/zh.json`
  - Purpose: 添加中文错误文案
  - Requirements: 国际化要求
  - Leverage: 现有 i18n 文件: `main/i18n/zh.json`
  - Context: 添加 6 个 Key: purchase.daily_limit_exceeded, purchase.single_qty_exceeded, purchase.quota.exceeded, purchase.quota.unit_mismatch, purchase.quota.config_not_found, purchase.quota.used_query_failed
  - Success: 文案添加成功

- [ ] 4.2 添加英文错误提示

  - File: `main/i18n/en.json`
  - Purpose: 添加英文错误文案
  - Requirements: 国际化要求
  - Leverage: 现有 i18n 文件: `main/i18n/en.json`
  - Success: 文案添加成功

- [ ] 4.3 添加日语错误提示

  - File: `main/i18n/ja.json`
  - Purpose: 添加日语错误文案
  - Requirements: 国际化要求
  - Leverage: 现有 i18n 文件: `main/i18n/ja.json`
  - Success: 文案添加成功

- [ ] 4.4 添加韩语错误提示

  - File: `main/i18n/ko.json`
  - Purpose: 添加韩语错误文案
  - Requirements: 国际化要求
  - Leverage: 现有 i18n 文件: `main/i18n/ko.json`
  - Success: 文案添加成功

- [ ] 4.5 添加其他语言错误提示

  - File: `main/i18n/{th,de,sv,tr,my,zhtw}.json`
  - Purpose: 添加泰语、德语、瑞典语、土耳其语、缅甸语、繁体中文错误文案
  - Requirements: 国际化要求
  - Leverage: 现有 i18n 文件
  - Success: 所有语言文案添加成功

---

## Phase 6: API 开发

- [ ] 6.1 实现限购配置创建/更新 API（支持门店关联）

  - File: `main/app/api/v1/shop/shop_purchase_quota.go`
  - Purpose: 创建或更新物品限购配置（使用关联表存储门店关系）
  - Requirements: Requirement 9.1
  - Leverage: 现有 API: `main/app/api/v1/shop/`，Task 2.1-2.2 的 Repository
  - Prompt: Role: Go API Developer | Task: 实现 POST /api/v1/shop/purchase/quota/config 接口 | Context: 接收 material_uuid, unit_uuid, quota_limit, apply_to_all_shops, shop_uuids，校验参数（apply_to_all_shops=false时shop_uuids必填），使用事务：1)创建/更新主表 2)如果apply_to_all_shops=false则先DeleteShops再BatchCreateShops，如果=true则只DeleteShops | Restrictions: 遵循 .cursor/rules/api.mdc，使用统一响应格式，必须使用事务保证一致性 | Success: API 实现完成，参数校验完整，错误处理正确，事务处理正确

- [ ] 6.2 实现限购配置查询 API（包含门店列表）

  - File: `main/app/api/v1/shop/shop_purchase_quota.go`
  - Purpose: 查询指定物品的限购配置（返回关联的门店列表）
  - Requirements: Requirement 9.2
  - Leverage: Task 2.1-2.2 的 Repository
  - Prompt: Role: Go API Developer | Task: 实现 GET /api/v1/shop/purchase/quota/config/{material_uuid} 接口 | Context: 根据 material_uuid 查询配置，预加载 Shops 关联，返回配置详情（包含 shop_uuids 数组和 shop_count 统计） | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 实现完成，数据返回正确，包含门店列表

- [ ] 6.3 实现限购配置删除 API（同时删除关联）

  - File: `main/app/api/v1/shop/shop_purchase_quota.go`
  - Purpose: 删除限购配置（同时删除门店关联）
  - Requirements: Requirement 9.3
  - Leverage: Task 2.1-2.2 的 Repository
  - Prompt: Role: Go API Developer | Task: 实现 DELETE /api/v1/shop/purchase/quota/config/{uuid} 接口 | Context: 根据 uuid 调用 Repository.Delete（内部使用事务同时软删除主表和关联表） | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 实现完成，删除操作正确，事务处理正确

- [ ] 6.4 实现门店配置查询 API

  - File: `main/app/api/v1/shop/shop_config.go`
  - Purpose: 获取门店采购配置
  - Requirements: Requirement 10.1
  - Leverage: 现有门店配置相关代码
  - Prompt: Role: Go API Developer | Task: 实现 GET /api/v1/shop/config/{shop_uuid} 接口 | Context: 查询门店配置，如果不存在返回全局默认值 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 实现完成，配置读取正确

- [ ] 6.5 实现门店配置更新 API

  - File: `main/app/api/v1/shop/shop_config.go`
  - Purpose: 更新门店采购配置
  - Requirements: Requirement 10.2
  - Leverage: 现有门店配置相关代码
  - Prompt: Role: Go API Developer | Task: 实现 POST /api/v1/shop/config/{shop_uuid} 接口 | Context: 接收 purchase_daily_limit, purchase_single_qty_limit，创建或更新配置 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 实现完成，配置保存正确

- [ ] 6.6 API 单元测试

  - File: `main/app/api/v1/shop/shop_purchase_quota_test.go`
  - Purpose: 测试限购配置相关 API（包含门店关联测试）
  - Requirements: 所有 API 需求
  - Leverage: 现有 API 测试
  - Prompt: Role: QA Engineer | Task: 为限购配置 API 编写单元测试 | Context: 测试创建配置（apply_to_all_shops=true/false）、查询配置（返回门店列表）、更新配置（门店关联更新）、删除配置（同时删除关联）、参数校验（shop_uuids 为空时的校验）、错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过，门店关联逻辑测试充分

- [ ] 6.7 创建 API 控制器注册路由

  - File: `main/app/api/v1/shop/router.go` 或相关路由文件
  - Purpose: 注册限购配置相关 API 路由
  - Requirements: 所有 API 需求
  - Leverage: 现有路由注册代码
  - Prompt: Role: Go API Developer | Task: 在路由文件中注册限购配置相关路由 | Context: POST /shop/purchase/quota/config, GET /shop/purchase/quota/config/:material_uuid, DELETE /shop/purchase/quota/config/:uuid | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路由注册成功，接口可访问

---

## Phase 7: 前端开发

- [ ] 7.1 门店配置页面 - 入口

  - File: 前端代码（根据前端仓库路径）
  - Purpose: 在门店列表增加"门店配置"按钮
  - Requirements: Requirement 7.1
  - Leverage: 现有门店管理页面
  - Prompt: Role: Frontend Developer | Task: 在门店列表的每一行增加"门店配置"按钮 | Context: 点击后跳转到门店配置页面，传递 shop_uuid | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 按钮显示正确，点击跳转正确

- [ ] 7.2 门店配置页面 - UI 实现

  - File: 前端代码（根据前端仓库路径）
  - Purpose: 实现门店配置页面
  - Requirements: Requirement 7.2-7.4
  - Leverage: 现有表单页面
  - Prompt: Role: Frontend Developer | Task: 实现门店配置页面 | Context: 显示门店信息，包含"每日采购申请次数"输入框（数字类型），实现获取配置和保存配置功能 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Element Plus 组件 | Success: 页面显示正确，配置读取和保存正常

- [ ] 7.3 物品限购配置 - 入口

  - File: 前端代码（根据前端仓库路径）
  - Purpose: 在物品详情页增加"申请限额设置"入口
  - Requirements: Requirement 8.1
  - Leverage: 现有物品详情页面
  - Prompt: Role: Frontend Developer | Task: 在物品详情页增加"申请限额设置"按钮或链接 | Context: 点击后打开配置弹窗 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 入口显示正确，点击打开弹窗

- [ ] 7.4 物品限购配置弹窗 - UI 实现

  - File: 前端代码（根据前端仓库路径）
  - Purpose: 实现物品限购配置弹窗（支持门店选择）
  - Requirements: Requirement 8.2-8.7
  - Leverage: 现有弹窗组件
  - Prompt: Role: Frontend Developer | Task: 实现限购配置弹窗 | Context: 包含"应用到全部店铺"开关（默认开启），门店选择器（多选，支持搜索，当开关关闭时显示），限购数量输入框（数字键盘），确定和取消按钮 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Element Plus 组件，门店数据从关联表获取 | Success: 弹窗显示正确，交互流畅，门店选择正确

- [ ] 7.5 物品限购配置 - API 集成

  - File: 前端代码（根据前端仓库路径）
  - Purpose: 集成限购配置相关 API
  - Requirements: Requirement 9.1-9.3
  - Leverage: 现有 API 调用封装
  - Prompt: Role: Frontend Developer | Task: 集成限购配置 API | Context: 调用创建配置 API、查询配置 API、删除配置 API，处理成功和失败情况 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 调用正常，错误处理完善

- [ ] 7.6 前端表单校验

  - File: 前端代码（根据前端仓库路径）
  - Purpose: 实现前端表单校验
  - Requirements: Requirement 8.6
  - Leverage: Element Plus 表单校验
  - Prompt: Role: Frontend Developer | Task: 实现表单校验规则 | Context: 限购数量必填且 > 0，应用门店至少选择一个（当 apply_to_all_shops=false 时） | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 校验规则正确，提示信息友好

---

## Phase 8: 测试和文档

- [ ] 8.1 集成测试

  - File: `test/integration/purchase_quota_test.go`（可选）
  - Purpose: 测试端到端限购功能（包含门店维度）
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 1)创建限购配置（apply_to_all_shops=true/false） 2)提交品牌采购申请 3)验证限购校验（全部门店 vs 指定门店） 4)测试超限场景 5)测试驳回后释放 6)测试门店关联删除 | Restrictions: 测试真实用户场景 | Success: 集成测试通过，门店维度测试充分

- [ ] 8.2 性能测试

  - File: -
  - Purpose: 确保限购校验性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Context: 测试限购校验响应时间，测试统计查询性能
  - Success: 限购校验响应时间 < 100ms，统计查询 < 50ms

- [ ] 5.3 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG
  - Context: 在 v2.14.0 版本下添加"新增品牌采购月度限购功能"
  - Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Repository: ≥ 80%（包含门店关联逻辑测试）
  - Service（限购校验）: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - ✅ 限购配置主表已创建
  - ✅ 门店关联表已创建
  - ✅ 提交时限购校验生效
  - ✅ 单位约束生效
  - ✅ 无配置物品正常放行
  - ✅ 门店维度限购生效（apply_to_all_shops=1 全部门店，=0 指定门店）
  - ✅ 删除配置时同步删除门店关联

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] 数据库迁移文件已创建
- [ ] 所有国际化文案已添加

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-brand-procurement-monthly-quota/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-brand-procurement-monthly-quota/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-brand-procurement-monthly-quota/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-brand-procurement-monthly-quota/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-brand-procurement-monthly-quota/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计细节（特别注意关联表模式）
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能（注意事务处理）
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

### ⚠️ 关键注意事项

1. **数据库设计变更**：从 JSON 字段改为关联表模式
   - 主表：`ttpos_purchase_quota_config`（包含 apply_to_all_shops 字段）
   - 关联表：`ttpos_purchase_quota_config_shop`（存储门店关联）
   
2. **事务处理要求**：
   - 创建/更新配置时，需要同步处理关联表
   - 删除配置时，必须同时软删除主表和关联表
   - 使用 GORM 事务确保数据一致性

3. **查询逻辑**：
   - `apply_to_all_shops=1`：应用到全部门店，关联表无记录
   - `apply_to_all_shops=0`：应用到指定门店，关联表有记录
   - 使用 EXISTS 子查询判断门店是否在适用范围内

---

## 附录：标准 Prompt 模板

### Go 后端开发（Repository）

```
Role: Go Developer specializing in Repository Pattern with GORM expertise

Task: 实现 PurchaseQuotaConfigRepo，提供限购配置数据访问（支持关联表模式）

Context:
- Current file: main/app/repository/purchase_quota_config_repo.go
- Interface file: main/app/repository/i_purchase_quota_config_repo.go
- Leverage code: main/app/repository/*_repo.go (选项模式示例)
- Requirements: Requirement 1.1, 2.1, 3.1 in requirements.md
- Design: 使用关联表 ttpos_purchase_quota_config_shop 存储门店关联
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/database.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Repository 只持有 db *gorm.DB 实例，不能持有 DBManager
- 使用选项模式(DBOption) 实现灵活查询
- 软删除查询必须过滤 delete_time=0
- 查询限购配置时必须过滤 status=1（启用状态）
- GetByMaterialUuidAndShop 使用 EXISTS 子查询判断门店关联
- Delete 方法必须使用事务同时软删除主表和关联表
- 使用 errors.WithMessage 包装错误

Success Criteria:
- Repository 实现完整，选项模式正确
- 软删除过滤正确，状态过滤正确
- 门店关联查询逻辑正确（apply_to_all_shops=1 OR EXISTS）
- 事务处理正确（Delete, BatchCreateShops）
- 代码通过 go fmt 和 go vet
- 单元测试覆盖率 ≥ 80%
```

### Go 后端开发（Service）

```
Role: Go Developer with business logic expertise

Task: 在 PurchaseOrderService 中集成限购校验逻辑

Context:
- Current file: main/app/service/purchase_order/purchase_order.go
- Leverage code: Task 2.1-2.2 的 Repository 实现
- Requirements: Requirement 2.1, 2.2, 2.3, 3.1, 3.2 in requirements.md
- Project specs: 遵循 .cursor/rules/go-main.mdc

Restrictions:
- 只对 PurchaseType=2（品牌采购）进行限购校验
- 使用 errors.WithMessage 包装错误
- 记录详细的日志（使用 logger.Logger）
- 不使用 panic，返回 error
- 实时查询统计模式，不维护额外状态
- 统计时排除当前订单，只统计有效状态(1,2,4,5)

Success Criteria:
- 限购校验逻辑完整
- 错误提示明确（包含物品名、限额、已用、本次申请）
- 日志记录清晰
- 单元测试覆盖率 ≥ 70%
- 所有边界情况已测试
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: 为限购校验逻辑编写单元测试

Context:
- Target file: main/app/service/purchase_order/purchase_order.go (checkPurchaseQuota, getMonthlyUsedQuota)
- Test file: main/app/service/purchase_order/purchase_order_test.go
- Coverage target: ≥ 70%

Test Cases Required:
- 无限购配置时放行
- 单位匹配时继续校验
- 单位不匹配时拒绝
- 超限时拒绝
- 未超限时通过
- 已用额度统计准确性
- 驳回订单不计入统计
- 草稿订单不计入统计

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 必须包含边界情况测试
- 使用 mock 隔离外部依赖

Success Criteria:
- 测试覆盖率 ≥ 70%
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-07  
**维护者**: BenDaye

