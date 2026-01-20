# story-purchase-limit-scheme-management

> 品牌采购限购方案管理（Phase 1: CRUD + 数据迁移）

---

## 📋 基本信息

| 项目              | 内容                     |
| ----------------- | ------------------------ |
| **Spec 名称**     | story-purchase-limit-scheme-management |
| **功能名称**      | 品牌采购限购方案管理         |
| **目标版本**      | v2.15.0                  |
| **创建日期**      | 2026-01-20               |
| **负责人**        | 待分配                   |
| **状态**          | 开发中                   |
| **Story Point**   | 3                        |
| **关联 Proposal** | [v2.15.0-purchase-limit-scheme-adjustment.md](../../../../team/proposals/2026-01/v2.15.0-purchase-limit-scheme-adjustment.md) |
| **关联任务**      | DooTask #38970           |
| **后续 Story**    | [story-purchase-limit-scheme-validation](../story-purchase-limit-scheme-validation/) |

---

## 📄 文档结构

- **[requirements.md](./requirements.md)** - 需求文档（User Story + AC 验收标准） ✅
- **[design.md](./design.md)** - 技术设计文档 ✅
- **[tasks.md](./tasks.md)** - 任务分解和进度追踪 ✅

---

## 🎯 功能概述

支持总部在"参数设置"中创建和管理多个限购方案，实现精细化的品牌采购申请限额管理。

### Phase 1 范围

1. **限购方案 CRUD**：创建、读取、更新、删除限购方案
2. **周期配置**：选择限购方案的生效周期（星期几）
3. **物品配置**：选择物品并为每个物品配置限额（数量+单位）
4. **门店配置**：选择限购方案适用的门店范围（全部/指定门店）
5. **数据迁移** ⚠️：迁移旧表数据到新方案表，然后删除旧表

### 涉及终端

- [x] Shop 商家管理端（总部）

### 涉及技术栈

- [x] Go (main/app/service/purchase_order/)
- [x] MySQL 8.0+ (新增 4 个表 + 数据迁移)
- [x] Redis 6.0+ (限购方案配置缓存)

---

## 🗄️ 数据库设计

### 新建 4 个表

1. `ttpos_purchase_limit_scheme` - 限购方案主表
2. `ttpos_purchase_limit_scheme_item` - 物品配置表
3. `ttpos_purchase_limit_scheme_shop` - 门店配置表
4. `ttpos_purchase_limit_scheme_weekday` - 星期配置表

### 数据迁移 ⚠️

**迁移任务**：
- 迁移 `ttpos_purchase_quota_config` → 新方案表
- 迁移 `ttpos_purchase_quota_config_shop` → 新方案门店表
- 删除旧表（2 个）
- 使用事务保证原子性
- 提供回滚脚本

---

## 🔌 API 设计

### 5 个接口

| API | Method | URL | 说明 |
| --- | --- | --- | --- |
| 列表 | GET | `/shop/purchase/limit_scheme/list` | 获取限购方案列表 |
| 详情 | GET | `/shop/purchase/limit_scheme/detail` | 获取限购方案详情 |
| 创建 | POST | `/shop/purchase/limit_scheme/create` | 创建限购方案 |
| 更新 | POST | `/shop/purchase/limit_scheme/update` | 更新限购方案 |
| 删除 | DELETE | `/shop/purchase/limit_scheme/delete` | 删除限购方案 |

---

## 📊 工作量评估

- **预计天数**: 3.5-4 天
- **Story Point**: 3

### Phase 分解

| Phase | 任务 | 时间 |
| --- | --- | --- |
| Phase 1 | 数据库设计和迁移 | 1 天 |
| Phase 2 | Repository + Service 实现 | 1.5 天 |
| Phase 3 | API 层集成 | 0.5 天 |
| Phase 4 | 测试与文档 | 0.5 天 |

---

## 🔗 快速链接

### 核心文档

- [需求文档](./requirements.md) - User Story + AC 验收标准
- [技术设计](./design.md) - 数据库设计 + API 设计
- [任务分解](./tasks.md) - 开发任务 + 测试任务 + 进度追踪

### 相关资源

- [提案文档](../../../../team/proposals/2026-01/v2.15.0-purchase-limit-scheme-adjustment.md)
- [原型链接](https://modao.cc/proto/NYlDfREZt0gr57g5xvn9XE/sharing?view_mode=device&screen=rbpV8FJlbQCm1ruKH)
- [DooTask #38970](http://t.hitosea.com/project/368/task/detail/38970)

### 参考代码

- `main/app/service/purchase_order/purchase_quota.go` - 现有限购功能
- `main/app/service/purchase_order/purchase_order.go` - 采购申请服务
- `main/app/api/v1/shop/shop_purchase.go` - 采购申请 API

---

## 📝 开发指引

### 开发顺序

1. **Phase 1**: 数据库设计和迁移
   - 创建迁移文件
   - 创建 Model 文件
   - 测试迁移脚本

2. **Phase 2**: Repository + Service
   - 实现 4 个 Repository
   - 实现 Service 业务逻辑
   - 创建 DTO 定义

3. **Phase 3**: API 层集成
   - 实现 5 个 Handler
   - 注册路由

4. **Phase 4**: 测试与文档
   - 单元测试
   - API 测试
   - 数据迁移验证
   - 文档更新

### 关键注意事项

⚠️ **数据迁移**是本 Story 的关键任务：
- 必须在测试环境充分验证
- 生产环境在维护窗口期执行
- 迁移前备份旧表数据
- 使用事务保证原子性
- 准备回滚方案

---

**版本**: v1.0.0  
**创建日期**: 2026-01-20  
**维护者**: weifashi  
**状态**: ✅ 设计完成，待开发
