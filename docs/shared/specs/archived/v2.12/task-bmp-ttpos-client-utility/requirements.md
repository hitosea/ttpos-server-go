> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# TTPOS HTTP Client 工具类抽取 需求文档

> 本文档定义 TTPOS HTTP Client 工具类抽取的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/ttpos-client-utility.md](../../../../team/proposals/2025-12/ttpos-client-utility.md) |
| **创建日期**      | 2025-12-18                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | -                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   |              |
| **审核日期** |              |
| **审核意见** |          |

---

## 📋 概述

抽取 `ttpos-takeout` 模块中 HTTP Client 创建逻辑为通用工具类，统一 TTPOS API 调用方式，便于添加通用中间件（如日志 dump）。

## 🎯 产品对齐

提升代码复用性和可维护性，与 `ttpos-erp` 模块的 `GetClient` 模式保持一致。

## 📝 用户故事

**作为** 开发人员  
**我想** 使用统一的 HTTP Client 工具类调用 TTPOS API  
**以便于** 减少重复代码，统一配置管理

---

## 功能需求

### Requirement 1: GetTtposClient 基础 Client

**用户故事**: 作为开发人员，我想获取预配置的 HTTP Client，以便于快速调用 TTPOS API

#### 验收标准

1. **WHEN** 调用 `GetTtposClient(ctx)` **THEN** 系统 **SHALL** 返回预配置的 `*gclient.Client`
2. **IF** 配置了 `app.ttposEndpoint` **THEN** Client **SHALL** 自动设置该值为 prefix
3. **WHEN** 配置 `app.ttpos-client.dump` 为 `true` **THEN** Client **SHALL** 自动执行 `resp.RawDump()`

#### 具体要求

- [x] 1.1 自动设置 `app.ttposEndpoint` 为 prefix
- [x] 1.2 自动设置超时时间为 10 秒
- [x] 1.3 自动设置 `ContentJson()`
- [x] 1.4 添加 dump 中间件，根据配置开关打印请求/响应

---

### Requirement 2: GetTtposClientWithAuth 带认证 Client

**用户故事**: 作为开发人员，我想获取带认证头的 HTTP Client，以便于调用需要认证的 TTPOS API

#### 验收标准

1. **WHEN** 调用 `GetTtposClientWithAuth(ctx, identifier)` **THEN** 系统 **SHALL** 返回带认证头的 Client
2. **IF** 认证生成失败 **THEN** 系统 **SHALL** 返回错误

#### 具体要求

- [x] 2.1 包含 GetTtposClient 的所有功能
- [x] 2.2 自动生成并设置 `X-TTPOS-SECRET` 认证头

---

## 非功能需求

### 代码架构和模块化

- **位置**: `ttpos-bmp/app/ttpos-takeout/utility/ttpos_client.go`
- **遵循规范**: `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### 性能要求

- [x] 无额外性能开销

### 测试要求

- [ ] 单元测试覆盖

---

## 验收标准

### 功能验收

1. **GetTtposClient**: 返回预配置的 Client，包含 prefix、超时、ContentJson
2. **GetTtposClientWithAuth**: 返回带认证头的 Client
3. **Dump 功能**: 配置开启时正确打印请求/响应

### 测试验收

1. **单元测试**: 覆盖两个工厂方法

### 文档验收

1. **代码注释**: 函数注释完整

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- `ttpos-bmp/app/ttpos-takeout/utility/ttpos_auth.go` - 认证生成

---

## 风险和缓解

无明显风险。

---

## 时间表

- **Phase 1 - 工具类开发**: 0.25 天
- **Phase 2 - 重构现有代码**: 0.25 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范

### 参考实现

- `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go` - GetClient 实现

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: rikugun  
**审核者**: -

