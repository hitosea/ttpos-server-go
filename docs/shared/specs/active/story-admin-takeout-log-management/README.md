# 云平台-日志管理(外卖相关)

> **状态**: 🔵 待开发  
> **版本**: v2.12.0  
> **负责人**: 待分配  
> **关联任务**: [DooTask #37765](dootask://task/37765)

---

## 📋 概述

为云平台(Admin)添加外卖相关的日志管理功能,让平台管理员和商户能够查看、筛选和追溯外卖订单同步日志(Grab/LINE MAN等)。

---

## 🎯 核心功能

1. **日志列表查询**: 分页查询外卖导入日志
2. **多维度筛选**: 支持按门店、平台、类型、状态筛选
3. **日志详情查看**: 查看完整的错误信息
4. **权限控制**: 商户只能查看自己的日志,平台管理员可查看所有日志

---

## 📚 文档结构

```
story-admin-takeout-log-management/
├── README.md                 # 本文件 - 规格概览
├── requirements.md           # 需求文档 - User Story + AC
├── design.md                 # 技术设计文档
└── tasks.md                  # 任务分解文档
```

---

## 🚀 快速导航

### 产品经理/BA

- 📄 [需求文档](./requirements.md) - 查看 User Story 和 AC
- 📋 [提案文档](../../../../team/proposals/2025-12/v2.12.0-cloud-platform-takeout-log-management.md)

### 开发人员

- 🏗️ [技术设计](./design.md) - 查看架构设计和实现方案
- ✅ [任务分解](./tasks.md) - 查看开发任务列表

### 测试人员

- ✅ [需求文档](./requirements.md) - 查看验收标准
- 🧪 [技术设计](./design.md#🧪-测试计划) - 查看测试计划

---

## 📊 进度跟踪

| 阶段 | 状态 | 负责人 | 完成时间 |
|------|------|--------|---------|
| 需求提案 | ✅ 已完成 | weifashi | 2025-12-17 |
| 需求文档 | 🔵 待开始 | 待分配 | - |
| 技术设计 | ✅ 已完成 | weifashi | 2025-12-17 |
| 任务分解 | 🔵 待开始 | 待分配 | - |
| 后端开发 | 🔵 待开始 | 待分配 | - |
| 前端开发 | 🔵 待开始 | 待分配 | - |
| 测试验收 | 🔵 待开始 | 待分配 | - |
| 上线部署 | 🔵 待开始 | 待分配 | - |

---

## 🔗 相关链接

### 需求相关

- [DooTask 任务](dootask://task/37765)
- [需求提案](../../../../team/proposals/2025-12/v2.12.0-cloud-platform-takeout-log-management.md)

### 技术相关

- [外卖菜单导入进度设计文档](../story-shop-takeout-import-progress/design.md) - 参考文档
- [Go Main 开发规范](../../../../../.cursor/rules/go-main.mdc)
- [API 设计规范](../../../../../.cursor/rules/api.mdc)
- [数据库开发规范](../../../../../.cursor/rules/database.mdc)

### 代码位置

**后端**:
- Model: `main/app/modules/takeout/domain/model/takeout_import_log.go`
- Repository: `main/app/modules/takeout/domain/repository/takeout_import_log_repository.go`
- Application Service: `main/app/modules/takeout/application/takeout_app_service.go`
- Admin API: `main/app/api/v1/admin/admin_takeout.go` (待新建)

**前端**:
- Admin 页面: `admin-frontend/src/views/takeout/logs.vue` (待新建)
- API 调用: `admin-frontend/src/api/admin/takeout.ts` (待新建)

---

## 💡 技术亮点

1. **代码复用**: 复用现有的 Shop 端日志基础设施,无需重复开发
2. **无需数据库迁移**: 表结构已存在,无需创建新表
3. **权限控制**: 基于角色的访问控制,确保数据安全
4. **性能优化**: 使用索引优化查询,支持大数据量场景
5. **扩展性**: 支持未来添加更多外卖平台

---

## 📝 Story Point 评估

**预估 SP**: 3-5

**评估依据**:
- **后端开发**: 2 SP (新增 Admin API Handler + 权限控制)
- **前端开发**: 2 SP (新增日志管理页面)
- **测试和联调**: 1 SP

**风险因素**:
- ⚠️ 权限系统集成复杂度(需确认现有权限模型)
- ⚠️ 大数据量查询性能(需优化索引)

---

## 📅 里程碑

| 里程碑 | 预计日期 | 实际日期 | 说明 |
|--------|---------|---------|------|
| 需求评审 | 2025-12-18 | - | 产品、开发、测试共同评审 |
| 技术评审 | 2025-12-19 | - | 技术团队评审设计方案 |
| 开发完成 | 2025-12-25 | - | 前后端开发完成 |
| 测试完成 | 2025-12-27 | - | 功能测试和集成测试完成 |
| 上线部署 | 2025-12-31 | - | 部署到生产环境 |

---

**创建日期**: 2025-12-17  
**最后更新**: 2025-12-17  
**维护者**: weifashi

