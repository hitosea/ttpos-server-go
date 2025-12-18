# 批量创建外卖商品 需求文档

> 本文档定义批量创建外卖商品功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-batch-create-takeout-products.md](../../../../team/proposals/2025-12/v2.12.0-batch-create-takeout-products.md) |
| **创建日期**      | 2025-12-18                                                                                                 |
| **负责人**        | 待定                                                                                                       |
| **目标 Sprint**   | 待定                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | 待定             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

商家端需要能够批量将TTPOS系统中的商品推送到外卖平台(Grab、LINE MAN等),并进行批量上架、下架、删除操作。该功能为商家提供高效的外卖商品管理能力,避免逐个商品手动操作。

当前系统已具备单商品推送到外卖平台的能力(`PushMenuToGrab`),但缺少批量操作功能。本需求旨在补充批量操作能力,支持多个外卖平台,并通过异步任务机制确保大批量操作的稳定性。

## 🎯 产品对齐

本功能支持TTPOS外卖管理系统的产品愿景,通过提供批量操作能力,帮助商家:
- 快速完成大量商品的外卖平台同步,提升运营效率90%以上
- 统一管理多个外卖平台的商品状态,降低操作成本
- 确保商品数据在TTPOS和外卖平台之间的一致性
- 增强系统对主流外卖平台的支持能力

## 📝 用户故事

**作为** 商户管理员  
**我想** 批量将TTPOS商品推送到外卖平台,并进行批量上架、下架、删除操作  
**以便于** 快速完成大量商品的平台同步,避免逐个操作,提升运营效率

---

## 功能需求

### Requirement 1: 批量创建外卖商品

**用户故事**: 作为商户管理员,我想批量将TTPOS商品推送到外卖平台(Grab/LINE MAN),以便于快速完成商品同步

#### 验收标准

1. **WHEN** 选择多个商品点击"批量创建Grab" **THEN** 系统 **SHALL** 使用 `/product/takeout/add` 接口的逻辑批量创建外卖商品映射关系
2. **WHEN** 批量操作提交后 **THEN** 系统 **SHALL** 异步处理并直接返回操作结果
3. **IF** 部分商品创建失败 **THEN** 系统 **SHALL** 在响应中标记失败商品并记录失败原因
4. **WHEN** 批量操作完成 **THEN** 系统 **SHALL** 在TTPOS中创建商品外卖映射关系
5. **WHEN** 商品映射成功创建 **THEN** 系统 **SHALL** 记录操作日志,包含操作人、时间、平台等信息

#### 具体要求

- [ ] 1.1 实现 POST `/shop/takeout/products/batch_create` 接口
- [ ] 1.2 支持请求参数:
  - `platform`: 外卖平台标识 (必填: grab/lineman)
  - `product_uuids`: 商品UUID列表 (必填,数组,最多100个)
- [ ] 1.3 接口返回格式:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "total": 100,
      "success": 95,
      "failed": 5,
      "failed_products": [
        {
          "product_uuid": "123456",
          "product_name": "商品A",
          "error": "平台限流"
        }
      ]
    }
  }
  ```
- [ ] 1.4 实现批量处理逻辑:
  - 使用 Goroutine 并发处理,提高效率
  - 实现限流控制,每秒不超过10个请求
  - 失败自动重试3次
  - 记录每个商品的处理结果
- [ ] 1.5 数据同步:
  - 创建或还原 `ttpos_product_package_takeout` 记录
  - 如已存在已删除的记录则还原
  - 更新商品的外卖平台映射状态
- [ ] 1.6 操作日志:
  - 记录到 `ttpos_operation_log` 表
  - 包含操作人、操作时间、平台、商品数量等信息

---

### Requirement 2: 批量上架外卖商品

**用户故事**: 作为商户管理员,我想批量上架外卖平台商品,以便于快速更新商品售卖状态

#### 验收标准

1. **WHEN** 选择多个商品点击"批量上架Grab" **THEN** 系统 **SHALL** 将选中商品批量在Grab平台上架
2. **WHEN** 批量上架完成后 **THEN** 系统 **SHALL** 返回操作结果
3. **IF** 商品未创建到平台 **THEN** 系统 **SHALL** 标记为失败并提示"商品未同步到平台"
4. **WHEN** 上架成功 **THEN** 系统 **SHALL** 更新TTPOS中的商品状态为"已上架"
5. **WHEN** 批量操作完成 **THEN** 系统 **SHALL** 显示成功和失败的商品数量

#### 具体要求

- [ ] 2.1 实现 POST `/shop/takeout/products/batch_online` 接口
- [ ] 2.2 支持请求参数:
  - `platform`: 外卖平台标识 (必填: grab/lineman)
  - `product_uuids`: 商品UUID列表 (必填,数组,最多100个)
- [ ] 2.3 前置条件检查:
  - 验证商品是否已同步到外卖平台
  - 检查商品是否已删除
- [ ] 2.4 批量上架逻辑:
  - 调用外卖平台上架API
  - 实现限流和重试机制
  - 更新商品上架状态
- [ ] 2.5 记录操作日志

---

### Requirement 3: 批量下架外卖商品

**用户故事**: 作为商户管理员,我想批量下架外卖平台商品,以便于快速停止商品售卖

#### 验收标准

1. **WHEN** 选择多个商品点击"批量下架LINE MAN" **THEN** 系统 **SHALL** 将选中商品批量在LINE MAN平台下架
2. **WHEN** 批量下架完成后 **THEN** 系统 **SHALL** 返回操作结果
3. **IF** 商品未创建到平台 **THEN** 系统 **SHALL** 标记为失败并提示"商品未同步到平台"
4. **WHEN** 下架成功 **THEN** 系统 **SHALL** 更新TTPOS中的商品状态为"已下架"

#### 具体要求

- [ ] 3.1 实现 POST `/shop/takeout/products/batch_offline` 接口
- [ ] 3.2 支持请求参数:
  - `platform`: 外卖平台标识 (必填: grab/lineman)
  - `product_uuids`: 商品UUID列表 (必填,数组,最多100个)
- [ ] 3.3 前置条件检查:
  - 验证商品是否已同步到外卖平台
  - 检查商品是否已删除
- [ ] 3.4 批量下架逻辑:
  - 调用外卖平台下架API
  - 实现限流和重试机制
  - 更新商品下架状态
- [ ] 3.5 记录操作日志

---

### Requirement 4: 批量删除外卖商品

**用户故事**: 作为商户管理员,我想批量删除外卖平台商品,以便于清理不需要的商品

#### 验收标准

1. **WHEN** 选择多个商品点击"批量删除Grab" **THEN** 系统 **SHALL** 将选中商品批量从Grab平台删除
2. **WHEN** 批量删除完成后 **THEN** 系统 **SHALL** 返回操作结果
3. **IF** 商品未创建到平台 **THEN** 系统 **SHALL** 标记为失败并提示"商品未同步到平台"
4. **WHEN** 删除成功 **THEN** 系统 **SHALL** 删除TTPOS中的 `ttpos_product_takeout` 记录
5. **WHEN** 删除操作执行前 **THEN** 系统 **SHALL** 显示确认对话框提示用户

#### 具体要求

- [ ] 4.1 实现 POST `/shop/takeout/products/batch_delete` 接口
- [ ] 4.2 支持请求参数:
  - `platform`: 外卖平台标识 (必填: grab/lineman)
  - `product_uuids`: 商品UUID列表 (必填,数组,最多100个)
- [ ] 4.3 前置条件检查:
  - 验证商品是否已同步到外卖平台
  - 检查商品是否已删除
- [ ] 4.4 批量删除逻辑:
  - 调用外卖平台删除API
  - 实现限流和重试机制
  - 软删除 `ttpos_product_takeout` 记录
- [ ] 4.5 记录操作日志

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **复用现有模块**: 
  - 复用 `main/app/modules/takeout/` 模块结构
  - 复用 `main/app/service/product_takeout.go` 服务
  - 扩展 `main/app/api/v1/shop/shop_takeout.go` 路由
  - 复用现有外卖平台API集成代码
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范
  - `.cursor/rules/api.mdc` - API 设计规范

### API 设计要求

- [x] URL 使用 snake_case 命名
  - `/shop/takeout/products/batch_create`
  - `/shop/takeout/products/batch_online`
  - `/shop/takeout/products/batch_offline`
  - `/shop/takeout/products/batch_delete`
- [x] data 字段必须是对象,不能是 null 或数组
- [x] 响应格式:`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

本需求使用现有表,不需要新增数据表:
- `ttpos_product_takeout`: 外卖商品映射表
- `ttpos_product`: 商品主表
- `ttpos_category`: 商品分类表
- `ttpos_product_i18n`: 商品多语言表

查询优化:
- [x] 使用现有索引: `idx_takeout_platform_company` (takeout_platform, company_uuid)
- [x] 使用 JOIN 优化多表查询
- [x] 避免 N+1 查询问题

### 性能要求

- [x] 本地响应时间 < 5秒 (100个商品)
- [x] 批量操作使用并发处理,提高效率
- [x] 实现限流控制,每秒不超过10个外卖平台API请求
- [x] 并发处理使用公司UUID隔离

### 浏览器兼容性(管理后台)

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口:
  - 批量创建接口测试
  - 批量上架接口测试
  - 批量下架接口测试
  - 批量删除接口测试
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言(中文、英文、泰语、日语、韩语等)
- [x] 错误提示支持多语言
- [x] 操作确认对话框支持多语言
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证(使用现有 middleware.Auth)
- [x] 数据隔离:只能操作当前登录商家的商品
- [x] SQL 注入防护(使用 GORM 参数化查询)
- [x] XSS 防护(前端输入校验)
- [x] 批量操作数量限制:最多100个商品
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 外卖平台API调用失败自动重试3次
- [x] 错误日志记录(使用 logger.Logger)
- [x] 参数校验和错误处理
- [x] 任务状态持久化,防止进程重启丢失

---

## 验收标准

### 功能验收

1. **批量创建**: 能够批量将商品推送到Grab/LINE MAN平台,直接返回操作结果
2. **批量上架**: 能够批量上架外卖平台商品,更新商品状态
3. **批量下架**: 能够批量下架外卖平台商品,更新商品状态
4. **批量删除**: 能够批量删除外卖平台商品,删除映射关系
5. **数据隔离**: 不同商家只能操作自己的商品数据
6. **限流控制**: 外卖平台API调用频率不超过限制
7. **失败重试**: 失败的商品自动重试,记录失败原因

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **性能测试**: 响应时间满足要求,异步任务正常工作
5. **并发测试**: 多用户同时操作不冲突

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Swagger注释完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头,实现以 `Srv` 结尾(如 `IProductTakeoutSrv`)
- Service 只能依赖其他 Service 接口
- 不使用 panic,返回 error
- Handler 层负责参数校验和响应封装
- Service 层负责业务逻辑和异步任务管理
- Repository 层负责数据访问

### 业务约束

- 支持多个外卖平台(grab、lineman等)
- 只能操作当前登录商家的商品数据
- 批量操作最多支持100个商品
- 商品必须先同步到平台才能进行上架、下架、删除操作

### 资源约束

- 开发时间: 3-5 天
- Story Point: SP3-SP5 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gin-gonic/gin` - Web框架
- `gorm.io/gorm` - ORM框架
- `ttpos-server-go/pkg/logger` - 日志系统
- `ttpos-server-go/app/service/product_takeout.go` - 外卖商品服务
- `ttpos-server-go/app/modules/takeout/` - 外卖模块

### 服务依赖

- **Main → Database**: 查询商品数据
- **Main → Grab API**: 推送商品到Grab平台
- **Main → LINE MAN API**: 推送商品到LINE MAN平台
- **Shop Frontend → Main**: HTTP API 调用

### 业务依赖

- 依赖商品管理功能
- 依赖外卖平台API集成
- 依赖 `ttpos_product` 表
- 依赖 `ttpos_product_takeout` 表
- 前置条件:商家已配置外卖平台凭证

---

## 风险和缓解

### 风险 1: 外卖平台API限流

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 实现请求限流机制,控制调用频率(每秒不超过10个请求)
- 使用队列机制平滑处理请求
- 失败自动重试,最多3次
- 提供任务状态查询,让用户了解执行结果

### 风险 2: 批量操作耗时

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用 Goroutine 并发处理,提高效率
- 限制批量操作数量最多100个
- 实现超时控制,超时返回部分结果

### 风险 3: 部分商品失败

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 记录每个商品的处理结果
- 失败商品自动重试3次
- 显示失败商品列表和失败原因
- 支持手动重试失败商品

### 风险 4: 数据一致性

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 操作成功后同步更新TTPOS的商品状态
- 记录操作日志便于追溯

### 风险 5: 平台API差异

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 为每个平台实现独立的适配器
- 统一接口规范,屏蔽平台差异
- 充分测试不同平台的API行为
- 提供平台配置选项

---

## 时间表

- **Phase 1 - 核心实现**: 2-3 天
  - 实现批量创建接口
  - 实现批量上架/下架/删除接口
  - 实现并发处理和限流控制
  - 实现失败重试机制
- **Phase 2 - 测试和文档**: 1-1.5 天
  - 单元测试
  - API测试
  - 集成测试
  - 完善API文档
- **总计**: 2.5-4.5 天(SP = SP2-SP3)

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 相关代码

- `main/app/modules/takeout/` - 外卖模块
- `main/app/service/product_takeout.go` - 外卖商品服务
- `main/app/api/v1/shop/shop_takeout.go` - 外卖路由
- `main/app/model/product_takeout.go` - 外卖商品模型

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板:`docs/agent/templates/graphiti-episode.md`
- 活动日志:`docs/team/activities/2025-12/2025-12-18.md`
- 提醒:需求评审或范围调整若形成经验,应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: weifashi  
**审核者**: 待定

