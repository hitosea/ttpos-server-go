# 获取当前店内Grab商品列表 需求文档

> 本文档定义获取当前店内Grab商品列表的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-grab-products-list.md](../../../../team/proposals/2025-12/v2.12.0-grab-products-list.md) |
| **创建日期**      | 2025-12-18                                                                                                 |
| **负责人**        | 待定                                                                                                       |
| **目标 Sprint**   | 待定                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | weifashi             |
| **审核日期** | 2025-12-18             |
| **审核意见** | 需求简化为统计接口,支持多平台参数,通过审核         |

---

## 📋 概述

商家端需要能够查看当前店内已同步的外卖平台商品统计信息。该功能为商家提供直观的数据统计,帮助商家快速了解各外卖平台的商品同步数量。

当前系统已具备Grab等外卖平台的菜单导入、导出、推送等功能,但缺少对已导入商品的统计功能。本需求旨在补充这一统计能力,支持多个外卖平台(Grab、Lineman等)。

## 🎯 产品对齐

本功能支持TTPOS外卖管理系统的产品愿景,通过提供清晰的数据统计,帮助商家:
- 快速了解各外卖平台的商品同步数量
- 判断是否需要重新导入或更新商品数据
- 提升商家对外卖管理系统的可见性和信任度

## 📝 用户故事

**作为** 商户管理员  
**我想** 查看当前店内各外卖平台的商品统计数量  
**以便于** 快速了解外卖商品的同步状态

---

## 功能需求

### Requirement 1: 外卖商品统计接口

**用户故事**: 作为商户管理员,我想快速查看当前店内指定外卖平台的商品总数,以便于了解商品同步状态

#### 验收标准

1. **WHEN** 调用统计接口并指定平台 **THEN** 系统 **SHALL** 返回对应平台的商品总数
2. **WHEN** 商品被导入或删除后 **THEN** 系统 **SHALL** 在5分钟内更新统计数据
3. **WHEN** 查询不同店铺 **THEN** 系统 **SHALL** 返回对应店铺的商品统计
4. **IF** 店铺没有该平台商品 **THEN** 系统 **SHALL** 返回总数为0
5. **WHEN** 不传平台参数 **THEN** 系统 **SHALL** 返回所有外卖平台的商品总数

#### 具体要求

- [ ] 1.1 实现 GET `/shop/takeout/products/count` 接口
- [ ] 1.2 支持查询参数:
  - `platform`: 外卖平台标识 (可选: grab/lineman/等, 不传则统计所有平台)
- [ ] 1.3 接口返回格式: 
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "total": 100  // 商品总数
    }
  }
  ```
- [ ] 1.4 统计数据基于 `ttpos_product_takeout` 表,条件:
  - `takeout_platform = {platform}` (如果传了platform参数)
  - `company_uuid = 当前登录商家`
  - `delete_time = 0` (未删除)
- [ ] 1.5 实现Redis缓存:
  - 缓存key: `takeout:products:count:{company_uuid}:{platform}` 或 `takeout:products:count:{company_uuid}:all`
  - 有效期: 5分钟
- [ ] 1.6 支持通过 `force_refresh=1` 参数强制刷新缓存
- [ ] 1.7 支持的外卖平台:
  - `grab`: Grab平台
  - `lineman`: Lineman平台
  - 不传或传空: 所有平台商品总数

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
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范
  - `.cursor/rules/api.mdc` - API 设计规范

### API 设计要求

- [x] URL 使用 snake_case 命名
  - `/shop/takeout/products/count`
- [x] data 字段必须是对象,不能是 null 或数组
- [x] 响应格式:`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

本需求不需要新增数据表,使用现有表:
- `ttpos_product_takeout`: 外卖商品映射表
- `ttpos_product`: 商品主表
- `ttpos_category`: 商品分类表
- `ttpos_product_i18n`: 商品多语言表

查询优化:
- [x] 使用现有索引: `idx_takeout_platform_company` (takeout_platform, company_uuid)
- [x] 使用 JOIN 优化多表查询
- [x] 避免 N+1 查询问题

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 统计接口实现Redis缓存,缓存时间5分钟
- [x] 列表接口支持分页,避免一次查询大量数据
- [x] 使用数据库索引优化查询性能
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
  - 统计接口测试(传platform参数)
  - 统计接口测试(不传platform参数)
  - 统计接口测试(不同平台)
  - 缓存功能测试
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言(中文、英文、泰语、日语、韩语等)
- [x] 商品名称、分类名称使用多语言实现
- [x] 错误提示支持多语言
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证(使用现有 middleware.Auth)
- [x] 数据隔离:只能查询当前登录商家的商品
- [x] SQL 注入防护(使用 GORM 参数化查询)
- [x] XSS 防护(前端输入校验)
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 缓存失效时从数据库查询
- [x] 错误日志记录(使用 logger.Logger)
- [x] 参数校验和错误处理

---

## 验收标准

### 功能验收

1. **统计接口**: 能够正确返回指定平台或所有平台的商品总数,缓存生效
2. **平台参数**: 支持grab、lineman等平台,不传参数返回所有平台总数
3. **数据隔离**: 不同商家只能查询自己的商品数据
4. **缓存机制**: 缓存正常工作,强制刷新参数生效

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **性能测试**: 响应时间满足要求
5. **缓存测试**: 缓存功能正常工作

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
- Service 层负责业务逻辑
- Repository 层负责数据访问

### 业务约束

- 支持多个外卖平台(grab、lineman等)
- 只返回当前登录商家的商品数据
- 已删除的商品(delete_time != 0)不纳入统计

### 资源约束

- 开发时间: 0.5-1 天
- Story Point: SP1 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gin-gonic/gin` - Web框架
- `gorm.io/gorm` - ORM框架
- `ttpos-server-go/pkg/cache` - Redis缓存
- `ttpos-server-go/pkg/logger` - 日志系统
- `ttpos-server-go/app/service/product_takeout.go` - 外卖商品服务

### 服务依赖

- **Main → Database**: 查询商品数据
- **Main → Redis**: 缓存统计数据
- **Shop Frontend → Main**: HTTP API 调用

### 业务依赖

- 依赖外卖商品导入功能已完成
- 依赖 `ttpos_product_takeout` 表数据结构
- 前置条件:商家已导入Grab商品数据

---

## 风险和缓解

### 风险 1: 缓存不一致问题

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 设置合理的缓存过期时间(5分钟)
- 提供强制刷新缓存的参数
- 商品导入/删除时主动清除缓存

---

## 时间表

- **Phase 1 - 接口开发**: 0.5 天
  - 实现统计接口
  - 实现缓存机制
  - 支持多平台参数
- **Phase 2 - 测试和文档**: 0.5 天
  - 单元测试
  - API测试
  - 完善API文档
- **总计**: 0.5-1 天(SP = SP1)

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

