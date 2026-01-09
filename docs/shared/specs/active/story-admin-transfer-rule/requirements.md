# 新管理端-业务设置-调拨规则 需求文档

> 本文档定义调拨规则配置功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.13.0-admin-transfer-rule.md](../../../../team/proposals/2026-01/v2.13.0-admin-transfer-rule.md) |
| **创建日期**      | 2026-01-06                                                                                                 |
| **负责人**        | weifashi                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |
| **关联任务**      | DooTask #38422                                                                                              |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | 产品经理             |
| **审核日期** | 2026-01-06             |
| **审核意见** | 需求明确，验收标准完整，可以进入技术设计阶段         |

---

## 📋 概述

在新管理端的业务设置中增加"调拨规则"功能，允许总店管理员配置各门店的调拨发起类型（调入/调出），实现对门店调拨权限的统一管理和控制。通过规则限制，可以根据业务策略灵活管理库存流向，规范调拨流程。

## 🎯 产品对齐

该功能支持以下产品目标：
- **加强库存管控**：总部可以根据实际业务需求，灵活配置各门店的调拨权限
- **规范业务流程**：通过规则限制，避免不合理的调拨操作，提升库存管理水平
- **提升管理效率**：统一在总店进行配置，无需逐店沟通和管理
- **支持业务策略**：配合不同的库存管理策略（如中心仓模式、对等调拨模式等）

## 📝 用户故事

**作为** 总店管理员  
**我想** 配置各门店允许的调拨发起类型（调入/调出）  
**以便于** 根据业务策略统一管理门店的调拨权限，规范库存流向

---

## 功能需求

### Requirement 1: 调拨规则配置界面

**用户故事**: 作为总店管理员，我想在业务设置中配置各门店的调拨规则，以便于统一管理调拨权限

#### 验收标准

1. **WHEN** 总店管理员进入"业务设置 - 调拨规则"页面 **THEN** 系统 **SHALL** 显示门店列表，每个门店有"调入"和"调出"两个勾选选项
2. **WHEN** 管理员勾选某门店的"调入"选项 **THEN** 系统 **SHALL** 保存配置，该门店允许发起调入单
3. **WHEN** 管理员勾选某门店的"调出"选项 **THEN** 系统 **SHALL** 保存配置，该门店允许发起调出单
4. **WHEN** 管理员取消勾选某个选项 **THEN** 系统 **SHALL** 保存配置，该门店不允许发起对应类型的调拨单

#### 具体要求

- [ ] 1.1 新增"业务设置 - 调拨规则"菜单项
- [ ] 1.2 显示所有门店列表（包括总店和分店）
- [ ] 1.3 每个门店有两个勾选框："允许调入"和"允许调出"
- [ ] 1.4 支持批量配置（可选：全选、反选、批量应用）
- [ ] 1.5 配置保存后给出成功提示
- [ ] 1.6 支持搜索和筛选门店

---

### Requirement 2: 规则验证和边界处理

**用户故事**: 作为系统，我想确保每个门店至少保留一种调拨类型，以便于保证业务连续性

#### 验收标准

1. **WHEN** 某门店只剩最后一个调拨类型选项 **AND** 管理员尝试取消该选项 **THEN** 系统 **SHALL** 将该选项置灰，无法取消
2. **WHEN** 某门店只剩最后一个调拨类型选项 **AND** 管理员尝试取消该选项 **THEN** 系统 **SHALL** 显示提示"至少保留一个调拨类型"
3. **IF** 后端接收到不合法的配置（两个选项都为 false）**THEN** 系统 **SHALL** 返回错误，拒绝保存

#### 具体要求

- [ ] 2.1 前端实时检测勾选状态，最后一个选项自动置灰
- [ ] 2.2 前端显示提示信息，说明为什么选项被置灰
- [ ] 2.3 后端 API 验证参数，至少一个类型为 true
- [ ] 2.4 后端验证失败时返回明确的错误信息

---

### Requirement 3: 门店端权限控制

**用户故事**: 作为门店管理员，我想在发起调拨时只看到总店允许的调拨类型，以便于遵守总店的管理规则

#### 验收标准

1. **WHEN** 门店A的规则配置为"只允许调入" **AND** 门店A管理员进入调拨单发起页面 **THEN** 系统 **SHALL** 只显示"调入"选项
2. **WHEN** 门店B的规则配置为"只允许调出" **AND** 门店B管理员进入调拨单发起页面 **THEN** 系统 **SHALL** 只显示"调出"选项
3. **WHEN** 门店C的规则配置为"允许调入和调出" **AND** 门店C管理员进入调拨单发起页面 **THEN** 系统 **SHALL** 显示"调入"和"调出"两个选项
4. **IF** 门店未配置调拨规则 **THEN** 系统 **SHALL** 默认显示"调入"和"调出"两个选项（保持兼容性）

#### 具体要求

- [ ] 3.1 新增 API：根据门店 ID 查询调拨规则
- [ ] 3.2 调拨单发起页面调用 API 获取规则
- [ ] 3.3 根据规则过滤调拨类型选项
- [ ] 3.4 未配置规则时使用默认值（允许所有类型）
- [ ] 3.5 当门店不允许某种类型时，隐藏对应的入口或按钮

---

### Requirement 4: 规则即时生效

**用户故事**: 作为总店管理员，我想配置保存后立即生效，以便于快速响应业务需求

#### 验收标准

1. **WHEN** 总店管理员保存调拨规则配置 **THEN** 系统 **SHALL** 立即更新数据库
2. **WHEN** 门店管理员刷新或重新进入调拨单发起页面 **THEN** 系统 **SHALL** 显示最新的规则配置

#### 具体要求

- [ ] 4.1 配置保存后立即更新数据库
- [ ] 4.2 门店端每次进入调拨页面时重新获取规则
- [ ] 4.3 可选：使用 Redis 缓存规则，设置合理的过期时间（如 5 分钟）
- [ ] 4.4 可选：规则更新时清除相关缓存

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/transfer_rule`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

#### API 端点设计（初步）

**1. 获取调拨规则列表**
```
GET /api/v1/transfer_rule/list
Response: {
  code: 0,
  message: "success",
  data: {
    list: [
      {
        shop_id: 1,
        shop_name: "总店",
        allow_transfer_in: true,
        allow_transfer_out: true,
        update_time: 1704528000
      }
    ]
  }
}
```

**2. 保存调拨规则**
```
POST /api/v1/transfer_rule/save
Request: {
  shop_id: 1,
  allow_transfer_in: true,
  allow_transfer_out: false
}
Response: {
  code: 0,
  message: "保存成功",
  data: {}
}
```

**3. 获取门店的调拨规则（门店端）**
```
GET /api/v1/transfer_rule/get
Response: {
  code: 0,
  message: "success",
  data: {
    allow_transfer_in: true,
    allow_transfer_out: false
  }
}
```

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

#### 数据表设计（初步）

**表名**: `ttpos_transfer_rule`

| 字段名 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| id | bigint unsigned | 主键 | AUTO_INCREMENT |
| uuid | bigint unsigned | 唯一标识 | - |
| shop_id | bigint unsigned | 门店ID | - |
| merchant_id | bigint unsigned | 商户ID | - |
| allow_transfer_in | tinyint | 是否允许调入（1:是 0:否） | 1 |
| allow_transfer_out | tinyint | 是否允许调出（1:是 0:否） | 1 |
| create_time | int | 创建时间 | 0 |
| update_time | int | 更新时间 | 0 |
| delete_time | int | 删除时间（软删除） | 0 |

**索引设计**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_uuid` (`uuid`)
- UNIQUE KEY `uk_shop` (`shop_id`, `delete_time`) - 保证同一门店只有一条有效规则
- KEY `idx_merchant` (`merchant_id`)

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
- [x] 缓存策略（Redis，可选）
- [x] 支持并发配置（使用事务）

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 仅总店管理员可配置规则
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 配置失败时给出明确提示

---

## 验收标准

### 功能验收

1. **总店配置功能**: 总店管理员可以进入"业务设置 - 调拨规则"页面，配置各门店的调拨发起类型
2. **规则验证**: 每个门店至少保留一种调拨类型，最后一个选项自动置灰
3. **门店权限控制**: 门店管理员在发起调拨时，只能看到总店允许的调拨类型
4. **规则即时生效**: 总店配置保存后，门店端立即生效（刷新后可见）
5. **兼容性**: 未配置规则的门店默认显示所有调拨类型

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥70%, Repository ≥80%）
2. **API 测试**: 所有接口测试通过（规则列表、保存、查询）
3. **集成测试**: 端到端流程测试通过（配置 → 保存 → 门店端生效）
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
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

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 每个门店至少保留一种调拨类型（调入或调出）
- 只有总店管理员可以配置调拨规则
- 未配置规则的门店默认允许所有调拨类型（保持兼容性）
- 规则配置后立即生效

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5（待技术评审确认，必须 ≤ 5）

---

## 依赖关系

### 技术依赖

- `gin-gonic/gin` - Web 框架
- `gorm.io/gorm` - ORM 框架
- `Vue 3` - 前端框架
- `Element Plus` - UI 组件库

### 服务依赖

- **Frontend → Main**: HTTP API 调用（调拨规则配置、查询）
- **无 BMP 依赖**: 本功能不涉及微服务调用

### 业务依赖

- 依赖门店管理功能（获取门店列表）
- 依赖权限管理功能（验证总店管理员权限）
- 依赖调拨单功能（应用规则限制）

---

## 风险和缓解

### 风险 1: 数据迁移风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 规则表设计支持默认值，未配置门店默认允许所有类型
- 现有调拨单无需迁移，新规则只影响新发起的调拨单
- 提供数据修复脚本，必要时可批量初始化规则

### 风险 2: 用户体验风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在门店端给出明确的提示信息，说明调拨类型受总店规则限制
- 提供帮助文档，说明如何联系总店管理员调整规则
- 总店配置界面提供批量操作，方便快速调整

### 风险 3: 业务连续性风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 前端强制校验至少保留一个类型
- 后端双重校验，拒绝不合法的配置
- 未配置规则时使用默认值（允许所有类型）

---

## 时间表

- **Phase 1 - 数据库设计和迁移**: 0.5 天
- **Phase 2 - 后端 API 开发**: 1.5 天
- **Phase 3 - 前端页面开发**: 1.5 天
- **Phase 4 - 测试和优化**: 0.5 天
- **总计**: 4 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关文档

- DooTask 任务: #38422
- 提案文档: `docs/team/proposals/2026-01/v2.13.0-admin-transfer-rule.md`
- 调拨模块代码: `main/app/repository/transfer_order.go`
- 调拨单优化需求: `docs/team/proposals/2026-01/v2.13.0-admin-transfer-order-optimization.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2026-01/2026-01-06.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-06  
**作者**: weifashi  
**审核者**: {产品经理}

