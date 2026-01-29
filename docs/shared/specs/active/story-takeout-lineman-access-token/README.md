# LINE MAN OAuth Access Token 缓存功能

> **状态**: 🚀 开发阶段 - 技术设计已完成

---

## 📁 文档结构

```
story-takeout-lineman-access-token/
├── README.md          # 本文件 - Spec 概览
├── requirements.md    # 需求文档 ✅ 已完成，已通过审核
├── design.md          # 技术设计 ✅ 已创建
└── tasks.md           # 任务分解 ✅ 已创建
```

---

## 📋 Spec 信息

| 项目 | 内容 |
|------|------|
| **Spec 名称** | story-takeout-lineman-access-token |
| **功能描述** | LINE MAN OAuth Access Token 自动获取与 Redis 缓存 |
| **来源 Proposal** | [v2.13.1-lineman-access-token](../../../../team/proposals/2026-01/v2.13.1-lineman-access-token.md) |
| **负责人** | rikugun |
| **目标版本** | v2.13.1 |
| **涉及技术栈** | Go (ttpos-bmp/) |
| **预估 SP** | 3 |

---

## 🎯 功能概述

实现 LINE MAN OAuth Access Token 的自动获取与 Redis 缓存机制，参考 Grab 平台的成熟实现：

**核心功能**：
1. ✅ OAuth Token 获取（调用 LINE MAN OAuth API）
2. ✅ Redis Token 缓存（TTL = expires_in - 60s）
3. ✅ 双重检查锁（避免并发重复请求）
4. ✅ Authorization Header 生成（Bearer Token）
5. ✅ 配置支持（endpoint 可配置）

**业务价值**：
- 🚀 减少 90% Token 请求次数
- ⚡ 缓存命中响应时间 < 10ms
- 🛡️ 避免频率限制风险
- 🔧 统一 Token 管理

---

## 📖 当前阶段

### ✅ 已完成

- [x] Proposal 创建和评审
- [x] 配置变更（endpoint 配置）
- [x] Requirements 文档创建
- [x] 产品审核通过
- [x] Design 文档创建
- [x] Tasks 文档创建

### 📝 当前任务

- [ ] **开始开发**
  - 按照 tasks.md 逐项实现
  - Phase 1: 配置和基础结构 ✅ 已完成
  - Phase 2: 核心实现（8 个任务）
  - Phase 3: 测试和文档（5 个任务）

### ⏭️ 下一步

1. 开始 Phase 2 - 核心实现
2. 按 tasks.md 中的任务顺序执行
3. 测试覆盖率 ≥ 80%
4. 完成后记录 Graphiti Episode

---

## 🔗 相关文档

### Spec 文档
- [需求文档](./requirements.md) - 详细需求和验收标准 ✅
- [设计文档](./design.md) - 技术设计和实现方案 ✅
- [任务分解](./tasks.md) - 详细执行任务清单 ✅

### 参考资料
- [Grab OAuth 实现](../../../../ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go)
- [配置变更文档](../../../../ttpos-bmp/app/ttpos-takeout/manifest/config/CHANGELOG-v2.13.1.md)
- [环境变量配置](../../../../ttpos-bmp/app/ttpos-takeout/manifest/config/lineman-env-example.md)

### 规范文档
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

---

## 📞 联系方式

- **负责人**: rikugun
- **技术栈**: Go + GoFrame 2.x + Redis
- **模块**: ttpos-bmp/app/ttpos-takeout
- **问题反馈**: 通过 DooTask 或团队沟通渠道

---

**创建日期**: 2026-01-07  
**最后更新**: 2026-01-07  
**版本**: v1.0.0

