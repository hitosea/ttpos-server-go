# 物品负库存控制 需求文档

> 本文档定义物品负库存控制功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/material-allow-negative-stock.md](../../../../team/proposals/2025-12/material-allow-negative-stock.md) |
| **创建日期**      | 2025-12-10                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

在物品管理功能中增加"允许负库存"的设置选项，允许商户针对不同物品配置是否允许负库存。当允许负库存时，即使当前库存不足，也可以进行出库操作；当不允许负库存时，保持原有的库存校验逻辑。

该功能主要面向商户管理员，用于灵活配置不同物品的库存管理策略，提升业务适应性。

## 🎯 产品对齐

该功能支持产品在库存管理方面的精细化控制需求，帮助商户在不同业务场景下灵活处理库存问题，减少因库存不足导致的业务中断。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在添加或编辑物品时设置是否允许负库存  
**以便于** 针对不同物品采用不同的库存管理策略，提高业务灵活性

---

## 功能需求

### Requirement 1: 数据库支持负库存字段

**用户故事**: 作为系统，我需要存储物品的负库存设置，以便于后续业务逻辑判断

#### 验收标准

1. **WHEN** 数据库迁移执行时 **THEN** 系统 **SHALL** 在 `ttpos_material` 表中添加 `allow_negative_stock` 字段
2. **IF** 字段不存在 **THEN** 系统 **SHALL** 创建字段类型为 `INT(1)`，默认值为 `0`
3. **WHEN** 查询物品信息时 **THEN** 系统 **SHALL** 返回 `allow_negative_stock` 字段值

#### 具体要求

- [ ] 1.1 创建数据库迁移文件，在 `ttpos_material` 表中添加 `allow_negative_stock` 字段
- [ ] 1.2 字段类型：`INT(1) NOT NULL DEFAULT 0 COMMENT '是否允许负库存：1-允许，0-不允许'`
- [ ] 1.3 更新 Go Main 模块的 `Material` 模型，添加 `AllowNegativeStock` 字段
- [ ] 1.4 更新 PHP Admin 模块的 `Material` 模型，添加 `allow_negative_stock` 字段
- [ ] 1.5 历史数据默认值为 0（不允许负库存），保持向后兼容

---

### Requirement 2: 添加物品接口支持负库存设置

**用户故事**: 作为商户管理员，我想在添加物品时设置是否允许负库存，以便于配置物品的库存管理策略

#### 验收标准

1. **WHEN** 商户管理员调用添加物品接口时 **THEN** 系统 **SHALL** 接受 `allow_negative_stock` 参数
2. **IF** `allow_negative_stock` 为 `1` **THEN** 系统 **SHALL** 保存为允许负库存
3. **IF** `allow_negative_stock` 为 `0` 或未传 **THEN** 系统 **SHALL** 保存为不允许负库存（默认）
4. **WHEN** 物品保存成功后 **THEN** 系统 **SHALL** 在响应中返回 `allow_negative_stock` 字段值

#### 具体要求

- [ ] 2.1 更新 `MaterialAddReq` 结构体，确保 `AllowNegativeStock` 字段存在（已存在，需验证）
- [ ] 2.2 更新 `AddMaterial` Service 方法，保存 `allow_negative_stock` 字段到数据库
- [ ] 2.3 更新物品响应结构，返回 `allow_negative_stock` 字段
- [ ] 2.4 当开启 ERP 同步时，将 `allow_negative_stock` 字段传递给 `erpSrv.AddMaterial`（已实现，需验证）

---

### Requirement 3: 编辑物品接口支持负库存设置

**用户故事**: 作为商户管理员，我想在编辑物品时修改是否允许负库存，以便于调整物品的库存管理策略

#### 验收标准

1. **WHEN** 商户管理员调用编辑物品接口时 **THEN** 系统 **SHALL** 接受 `allow_negative_stock` 参数
2. **IF** `allow_negative_stock` 参数存在 **THEN** 系统 **SHALL** 更新物品的负库存设置
3. **WHEN** 物品更新成功后 **THEN** 系统 **SHALL** 在响应中返回更新后的 `allow_negative_stock` 字段值
4. **WHEN** 开启 ERP 同步时 **THEN** 系统 **SHALL** 同步更新 ERP 系统中的负库存设置

#### 具体要求

- [ ] 3.1 更新 `MaterialEditReq` 结构体，添加 `AllowNegativeStock` 字段
- [ ] 3.2 更新 `UpdateMaterial` Service 方法，支持更新 `allow_negative_stock` 字段
- [ ] 3.3 更新 `MaterialEditErpReq` 结构体，添加 `AllowNegativeStock` 字段
- [ ] 3.4 更新 `UpdateMaterialByEprItem` 方法，同步负库存设置到 ERP
- [ ] 3.5 更新物品响应结构，返回 `allow_negative_stock` 字段

---

### Requirement 4: 同步物品接口支持负库存字段

**用户故事**: 作为系统，我需要在同步物品数据时传递负库存设置，以便于保持数据一致性

#### 验收标准

1. **WHEN** 从总部同步物品到子店时 **THEN** 系统 **SHALL** 同步 `allow_negative_stock` 字段
2. **WHEN** 从 ERP 同步物品到本地时 **THEN** 系统 **SHALL** 同步 `allow_negative_stock` 字段
3. **WHEN** 同步物品数据时 **THEN** 系统 **SHALL** 确保 `allow_negative_stock` 字段正确传递和保存

#### 具体要求

- [ ] 4.1 检查 `AddMaterialByEprItem` 方法，确保从 ERP 同步时读取 `allow_negative_stock` 字段
- [ ] 4.2 检查总部同步物品的逻辑，确保 `allow_negative_stock` 字段被正确传递
- [ ] 4.3 在同步过程中添加日志记录，便于排查问题

---

### Requirement 5: 前端 UI 支持负库存设置

**用户故事**: 作为商户管理员，我想在添加/编辑物品的界面上设置是否允许负库存，以便于直观地配置物品属性

#### 验收标准

1. **WHEN** 商户管理员打开添加物品页面时 **THEN** 系统 **SHALL** 显示"允许负库存"选项
2. **WHEN** 商户管理员打开编辑物品页面时 **THEN** 系统 **SHALL** 显示当前物品的"允许负库存"设置
3. **IF** 商户管理员勾选"允许负库存" **THEN** 系统 **SHALL** 在提交时传递 `allow_negative_stock: 1`
4. **IF** 商户管理员未勾选"允许负库存" **THEN** 系统 **SHALL** 在提交时传递 `allow_negative_stock: 0` 或不传该字段

#### 具体要求

- [ ] 5.1 在添加物品表单中添加"允许负库存"开关组件
- [ ] 5.2 在编辑物品表单中添加"允许负库存"开关组件，并绑定现有值
- [ ] 5.3 添加字段说明文案，解释负库存的含义和使用场景
- [ ] 5.4 确保表单提交时正确传递 `allow_negative_stock` 字段

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/material/add`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖同步流程
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [ ] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **数据库字段**: `ttpos_material` 表成功添加 `allow_negative_stock` 字段，默认值为 0
2. **添加物品**: 添加物品时可以设置 `allow_negative_stock`，保存后数据库字段值正确
3. **编辑物品**: 编辑物品时可以修改 `allow_negative_stock`，更新后数据库字段值正确
4. **ERP 同步**: 开启 ERP 同步时，`allow_negative_stock` 字段正确传递到 ERP 系统
5. **总部同步**: 从总部同步物品时，`allow_negative_stock` 字段正确同步到子店
6. **前端 UI**: 添加/编辑物品页面显示"允许负库存"选项，操作后数据正确保存

### 测试验收

1. **单元测试**: Service 层和 Repository 层测试覆盖率达标
2. **API 测试**: 添加物品和编辑物品接口测试通过
3. **集成测试**: ERP 同步和总部同步流程测试通过
4. **手动测试**: 前端 UI 功能测试通过，浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 历史数据默认不允许负库存（`allow_negative_stock = 0`），保持向后兼容
- 负库存设置仅影响出库时的库存校验逻辑，不影响其他业务逻辑
- 需要明确告知用户负库存的含义和使用风险

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/model/material.go` - Material 模型
- `main/app/service/material.go` - Material Service
- `main/app/service/rpc/erp/material.go` - ERP 同步服务
- `admin/app/common/model/product/Material.php` - PHP Material 模型

### 服务依赖

- **Admin → Main**: HTTP API 调用（添加/编辑物品接口）
- **Main → BMP**: gRPC 调用（ERP 同步，如需要）

### 业务依赖

- 物品管理功能（已存在）
- ERP 同步功能（已存在）
- 总部同步功能（已存在）

---

## 风险和缓解

### 风险 1: 负库存可能导致库存数据不准确

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 明确负库存的业务规则和使用场景
- 在 UI 上提供清晰的用户提示和警告
- 建议商户谨慎使用负库存功能

### 风险 2: 同步接口字段传递错误导致数据不一致

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 在同步接口中增加字段校验和日志记录
- 编写完整的集成测试覆盖同步流程
- 添加数据一致性检查机制

### 风险 3: 历史数据迁移问题

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 数据库迁移时设置默认值为 0（不允许负库存）
- 保持向后兼容，不影响现有功能
- 提供数据迁移回滚方案

---

## 时间表

- **Phase 1 - 数据库和模型**: 0.5 天
- **Phase 2 - API 接口开发**: 1 天
- **Phase 3 - 前端 UI 开发**: 0.5 天
- **Phase 4 - 测试和文档**: 1 天
- **总计**: 3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- ERP 库存管理模块相关文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-10  
**作者**: xiezhihuan  
**审核者**: {审核者}

