# 简化菜单更新方法返回值 需求文档

> 本文档定义简化 `UpdateMenuItem` 和 `UpdateMenuModifier` 返回值的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/simplify-menu-update-response.md](../../../../team/proposals/2025-12/simplify-menu-update-response.md) |
| **创建日期**      | 2025-12-18                                                                                                 |
| **负责人**        | AI Agent                                                                                                       |
| **目标 Sprint**   | -                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                  |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | -             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

移除 `grabDto.UpdateMenuResult` 结构体，将 `UpdateMenuItem` 和 `UpdateMenuModifier` 方法的返回值简化为 `error`，符合 Go 语言惯例并减少代码冗余。

## 🎯 产品对齐

此为技术重构任务，旨在：
- 简化 API 设计，减少维护成本
- 统一错误处理模式
- 符合 Go 语言惯例

## 📝 用户故事

**作为** 开发人员  
**我想** 简化菜单更新方法的返回值  
**以便于** 代码更简洁、错误处理更统一

---

## 功能需求

### Requirement 1: 移除 UpdateMenuResult 结构体

**用户故事**: 作为开发人员，我想移除冗余的返回结构体，以便于简化代码

#### 验收标准

1. **WHEN** 调用 `UpdateMenuItem` **THEN** 系统 **SHALL** 返回 `error` 类型
2. **WHEN** 调用 `UpdateMenuModifier` **THEN** 系统 **SHALL** 返回 `error` 类型
3. **WHEN** 操作成功 **THEN** 系统 **SHALL** 返回 `nil`
4. **WHEN** 操作失败 **THEN** 系统 **SHALL** 返回包含错误信息的 `error`

#### 具体要求

- [x] 1.1 删除 `grabDto.UpdateMenuResult` 结构体定义
- [x] 1.2 修改 `UpdateMenuItem` 方法签名为 `error` 返回值
- [x] 1.3 修改 `UpdateMenuModifier` 方法签名为 `error` 返回值
- [x] 1.4 更新 service interface 定义
- [x] 1.5 更新所有调用方代码

---

## 非功能需求

### 代码架构和模块化

- **遵循规范**: `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### 测试要求

- [ ] 确保现有功能不受影响
- [ ] 编译通过，无 lint 错误

---

## 验收标准

### 功能验收

1. **编译通过**: 代码编译无错误
2. **功能正常**: 菜单更新功能正常工作
3. **错误处理**: 错误信息正确返回

### 文档验收

1. **代码注释**: 方法注释更新

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1

---

## 依赖关系

### 涉及文件

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`
- `ttpos-bmp/app/ttpos-takeout/internal/service/grab_menu.go` (interface)

---

## 风险和缓解

### 风险 1: 调用方代码未同步更新

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 全局搜索确认所有调用点
- 编译验证

---

## 时间表

- **Phase 1 - 代码修改**: 0.5 天
- **总计**: 0.5 天（SP = 1）

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: AI Agent

